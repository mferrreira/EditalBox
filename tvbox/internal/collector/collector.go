package collector

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/marcio/editalbox/tvbox/internal/domain"
)

type Collector struct {
	client      *http.Client
	userAgent   string
	followRules []string
}

type PageResult struct {
	Notice    domain.Notice
	Documents []domain.NoticeDocument
	Events    []domain.NoticeEvent
	Links     []string
}

func New(userAgent string, followRules []string) *Collector {
	return &Collector{
		client:      &http.Client{Timeout: 20 * time.Second},
		userAgent:   userAgent,
		followRules: followRules,
	}
}

func (c *Collector) Fetch(ctx context.Context, rawURL string) (PageResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return PageResult{}, err
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9,en;q=0.8")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")

	resp, err := c.client.Do(req)
	if err != nil {
		return PageResult{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return PageResult{}, err
	}

	html := string(body)
	title := extractTitle(html)
	text := normalizeText(stripTags(html))
	events := extractEvents(text)
	canonical := canonicalize(rawURL)
	registrationEnd, finalEventAt := deriveImportantDates(events)
	notice := domain.Notice{
		SourceURL:       rawURL,
		CanonicalURL:    canonical,
		Title:           firstNonEmpty(title, fallbackTitleFromURL(rawURL)),
		Excerpt:         buildExcerpt(text),
		BodyText:        truncate(text, 12000),
		SourceType:      classifySource(rawURL),
		Status:          deriveStatus(text, registrationEnd, finalEventAt, time.Now()),
		RegistrationEnd: registrationEnd,
		FinalEventAt:    finalEventAt,
	}

	var docs []domain.NoticeDocument
	for _, docURL := range extractLinks(rawURL, html) {
		if strings.Contains(docURL, "documento.ifnmg.edu.br/action.php") {
			docs = append(docs, domain.NoticeDocument{
				URL:   docURL,
				Kind:  classifyDocument(docURL),
				Title: "documento oficial",
			})
		}
	}

	return PageResult{
		Notice:    notice,
		Documents: docs,
		Events:    events,
		Links:     c.filterLinks(extractLinks(rawURL, html)),
	}, nil
}

func (c *Collector) filterLinks(links []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, link := range links {
		if _, ok := seen[link]; ok {
			continue
		}
		for _, rule := range c.followRules {
			if strings.Contains(link, rule) {
				seen[link] = struct{}{}
				out = append(out, link)
				break
			}
		}
	}
	return out
}

func extractTitle(html string) string {
	for _, re := range []*regexp.Regexp{
		regexp.MustCompile(`(?is)<meta[^>]+property=["']og:title["'][^>]+content=["'](.*?)["']`),
		regexp.MustCompile(`(?is)<meta[^>]+name=["']twitter:title["'][^>]+content=["'](.*?)["']`),
		regexp.MustCompile(`(?is)<h1[^>]*>(.*?)</h1>`),
		regexp.MustCompile(`(?is)<h2[^>]*>(.*?)</h2>`),
	} {
		match := re.FindStringSubmatch(html)
		if len(match) >= 2 {
			title := normalizeText(stripTags(match[1]))
			if title != "" && !isGenericTitle(title) {
				return title
			}
		}
	}
	match := regexp.MustCompile(`(?is)<title>(.*?)</title>`).FindStringSubmatch(html)
	if len(match) < 2 {
		return ""
	}
	title := normalizeText(stripTags(match[1]))
	if isGenericTitle(title) {
		return ""
	}
	return title
}

func extractLinks(baseURL, html string) []string {
	re := regexp.MustCompile(`(?i)href=["']([^"'#]+)["']`)
	matches := re.FindAllStringSubmatch(html, -1)
	base, _ := url.Parse(baseURL)
	var out []string
	for _, match := range matches {
		raw := strings.TrimSpace(match[1])
		parsed, err := url.Parse(raw)
		if err != nil {
			continue
		}
		resolved := base.ResolveReference(parsed)
		out = append(out, canonicalize(resolved.String()))
	}
	return out
}

func stripTags(input string) string {
	re := regexp.MustCompile(`(?s)<[^>]*>`)
	return re.ReplaceAllString(input, " ")
}

func normalizeText(input string) string {
	input = strings.ReplaceAll(input, "\u00a0", " ")
	return strings.Join(strings.Fields(input), " ")
}

func buildExcerpt(text string) string {
	for _, marker := range []string{"inscri", "edital", "processo seletivo", "bolsa", "auxilio", "resultado"} {
		if idx := strings.Index(strings.ToLower(text), marker); idx >= 0 {
			return truncate(text[idx:], 320)
		}
	}
	return truncate(text, 320)
}

func truncate(input string, size int) string {
	runes := []rune(input)
	if len(runes) <= size {
		return input
	}
	return string(runes[:size])
}

func canonicalize(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.Replace(raw, "http://", "https://", 1)
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	parsed.Fragment = ""
	if strings.HasSuffix(parsed.Path, "/") && len(parsed.Path) > 1 {
		parsed.Path = strings.TrimSuffix(parsed.Path, "/")
	}
	return parsed.String()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func fallbackTitleFromURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	path := strings.Trim(parsed.Path, "/")
	if path == "" {
		return parsed.Host
	}
	parts := strings.Split(path, "/")
	return strings.ReplaceAll(parts[len(parts)-1], "-", " ")
}

func classifySource(rawURL string) string {
	switch {
	case strings.Contains(rawURL, "/mais-noticias-januaria/"):
		return "news"
	case strings.Contains(rawURL, "documento.ifnmg.edu.br"):
		return "document"
	case strings.Contains(rawURL, "/professor-substituto"):
		return "staff_selection"
	case strings.Contains(rawURL, "/processoseletivo"):
		return "student_selection"
	default:
		return "index"
	}
}

func classifyDocument(rawURL string) string {
	if strings.Contains(rawURL, "fDocumentId=") {
		return "ifnmg_document"
	}
	return "external"
}

func deriveStatus(text string, registrationEnd, finalEventAt *time.Time, now time.Time) string {
	low := strings.ToLower(text)
	if finalEventAt != nil && now.After(finalEventAt.Add(24*time.Hour)) {
		return "finalized"
	}
	if registrationEnd != nil {
		if now.After(registrationEnd.Add(24 * time.Hour)) {
			if finalEventAt != nil && now.Before(finalEventAt.Add(24*time.Hour)) {
				return "registration_closed"
			}
			if strings.Contains(low, "resultado") || strings.Contains(low, "cronograma") || strings.Contains(low, "matricula") {
				return "in_progress"
			}
			return "registration_closed"
		}
		return "open"
	}
	switch {
	case strings.Contains(low, "encerrad"), strings.Contains(low, "finalizad"):
		return "finalized"
	case strings.Contains(low, "resultado final"), strings.Contains(low, "resultado preliminar"), strings.Contains(low, "cronograma"), strings.Contains(low, "matricula"):
		return "in_progress"
	case strings.Contains(low, "inscri") && (strings.Contains(low, "até") || strings.Contains(low, "aberta")):
		return "open"
	default:
		return "unknown"
	}
}

func extractEvents(text string) []domain.NoticeEvent {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(inscri[\pL]*|resultado final|resultado preliminar|matr[ií]cula[\pL]*|recurso[\pL]*|homologa[çc][aã]o|entrevista|chamada final)[^0-9]{0,40}(\d{1,2}/\d{1,2}/\d{4})`),
		regexp.MustCompile(`(?i)(inscri[\pL]*|resultado final|resultado preliminar|matr[ií]cula[\pL]*|recurso[\pL]*|homologa[çc][aã]o|entrevista|chamada final)[^0-9]{0,40}(\d{1,2}\s+de\s+[[:alpha:]çãéíóú]+\s+de\s+\d{4})`),
	}
	var out []domain.NoticeEvent
	seen := map[string]struct{}{}
	for _, re := range patterns {
		matches := re.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			when, err := parseDate(match[2])
			if err != nil {
				continue
			}
			label := normalizeText(match[1])
			key := label + "|" + when.Format(time.RFC3339)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, domain.NoticeEvent{
				EventType: normalizeEventType(label),
				Label:     label,
				EventAt:   when,
			})
		}
	}
	out = append(out, extractRangeEvents(text)...)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].EventAt.Before(out[j].EventAt)
	})
	return out
}

func normalizeEventType(label string) string {
	label = strings.ToLower(label)
	switch {
	case strings.Contains(label, "inscri"):
		return "registration"
	case strings.Contains(label, "resultado final"):
		return "final_result"
	case strings.Contains(label, "resultado preliminar"):
		return "preliminary_result"
	case strings.Contains(label, "matr"):
		return "enrollment"
	case strings.Contains(label, "recurso"):
		return "appeal"
	default:
		return "other"
	}
}

func deriveImportantDates(events []domain.NoticeEvent) (*time.Time, *time.Time) {
	var registrationEnd *time.Time
	var finalEventAt *time.Time
	for _, event := range events {
		event := event
		switch event.EventType {
		case "registration", "enrollment":
			if registrationEnd == nil || event.EventAt.After(*registrationEnd) {
				registrationEnd = &event.EventAt
			}
		case "final_result":
			if finalEventAt == nil || event.EventAt.After(*finalEventAt) {
				finalEventAt = &event.EventAt
			}
		default:
			if finalEventAt == nil || event.EventAt.After(*finalEventAt) {
				finalEventAt = &event.EventAt
			}
		}
	}
	return registrationEnd, finalEventAt
}

func extractRangeEvents(text string) []domain.NoticeEvent {
	re := regexp.MustCompile(`(?i)(inscri[\pL]*|matr[ií]cula[\pL]*)[^0-9]{0,50}(\d{1,2})\s+de\s+([[:alpha:]çãéíóú]+)\s+a\s+(\d{1,2})\s+de\s+([[:alpha:]çãéíóú]+)\s+de\s+(\d{4})`)
	matches := re.FindAllStringSubmatch(text, -1)
	var out []domain.NoticeEvent
	for _, match := range matches {
		dayEnd, _ := strconv.Atoi(match[4])
		year, _ := strconv.Atoi(match[6])
		monthEnd := monthFromPT(match[5])
		if monthEnd == time.Month(0) {
			continue
		}
		eventAt := time.Date(year, monthEnd, dayEnd, 0, 0, 0, 0, time.UTC)
		label := normalizeText(match[1])
		out = append(out, domain.NoticeEvent{
			EventType: normalizeEventType(label),
			Label:     label,
			EventAt:   eventAt,
		})
	}
	return out
}

func parseDate(value string) (time.Time, error) {
	value = strings.ToLower(normalizeText(value))
	if parsed, err := time.Parse("02/01/2006", value); err == nil {
		return parsed, nil
	}
	re := regexp.MustCompile(`(\d{1,2})\s+de\s+([[:alpha:]çãéíóú]+)\s+de\s+(\d{4})`)
	match := re.FindStringSubmatch(value)
	if len(match) != 4 {
		return time.Time{}, fmt.Errorf("unsupported date format: %s", value)
	}
	day, err := strconv.Atoi(match[1])
	if err != nil {
		return time.Time{}, err
	}
	year, err := strconv.Atoi(match[3])
	if err != nil {
		return time.Time{}, err
	}
	month := monthFromPT(match[2])
	if month == time.Month(0) {
		return time.Time{}, fmt.Errorf("unsupported month: %s", match[2])
	}
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC), nil
}

func monthFromPT(value string) time.Month {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "janeiro":
		return time.January
	case "fevereiro":
		return time.February
	case "marco", "março":
		return time.March
	case "abril":
		return time.April
	case "maio":
		return time.May
	case "junho":
		return time.June
	case "julho":
		return time.July
	case "agosto":
		return time.August
	case "setembro":
		return time.September
	case "outubro":
		return time.October
	case "novembro":
		return time.November
	case "dezembro":
		return time.December
	default:
		return time.Month(0)
	}
}

func isGenericTitle(title string) bool {
	title = strings.ToLower(normalizeText(title))
	generic := []string{
		"portal ifnmg - editais",
		"portal ifnmg - mais noticias januaria",
		"portal ifnmg - mais noticias",
		"portal ifnmg - januaria",
		"editais",
	}
	for _, item := range generic {
		if title == item {
			return true
		}
	}
	return false
}
