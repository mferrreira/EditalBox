package collector

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"regexp"
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
	canonical := canonicalize(rawURL)
	notice := domain.Notice{
		SourceURL:    rawURL,
		CanonicalURL: canonical,
		Title:        firstNonEmpty(title, fallbackTitleFromURL(rawURL)),
		Excerpt:      truncate(text, 320),
		BodyText:     truncate(text, 12000),
		SourceType:   classifySource(rawURL),
		Status:       deriveStatus(text, time.Now()),
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
		Events:    extractEvents(text),
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
		regexp.MustCompile(`(?is)<h1[^>]*>(.*?)</h1>`),
		regexp.MustCompile(`(?is)<h2[^>]*>(.*?)</h2>`),
	} {
		match := re.FindStringSubmatch(html)
		if len(match) >= 2 {
			title := normalizeText(stripTags(match[1]))
			if title != "" {
				return title
			}
		}
	}
	match := regexp.MustCompile(`(?is)<title>(.*?)</title>`).FindStringSubmatch(html)
	if len(match) < 2 {
		return ""
	}
	return normalizeText(stripTags(match[1]))
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

func deriveStatus(text string, now time.Time) string {
	low := strings.ToLower(text)
	switch {
	case strings.Contains(low, "resultado final"), strings.Contains(low, "resultado preliminar"), strings.Contains(low, "cronograma"):
		return "in_progress"
	case strings.Contains(low, "inscri") && strings.Contains(low, "até"):
		return "open"
	default:
		return "unknown"
	}
}

func extractEvents(text string) []domain.NoticeEvent {
	re := regexp.MustCompile(`(?i)(inscri[çc][aã]o|resultado final|resultado preliminar|matr[ií]cula|recurso)[^0-9]{0,30}(\d{1,2}/\d{1,2}/\d{4})`)
	matches := re.FindAllStringSubmatch(text, -1)
	var out []domain.NoticeEvent
	for _, match := range matches {
		when, err := time.Parse("02/01/2006", match[2])
		if err != nil {
			continue
		}
		label := normalizeText(match[1])
		out = append(out, domain.NoticeEvent{
			EventType: normalizeEventType(label),
			Label:     label,
			EventAt:   when,
		})
	}
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
