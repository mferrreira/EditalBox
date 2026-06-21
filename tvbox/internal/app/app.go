package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/marcio/editalbox/tvbox/internal/agent"
	"github.com/marcio/editalbox/tvbox/internal/collector"
	"github.com/marcio/editalbox/tvbox/internal/config"
	"github.com/marcio/editalbox/tvbox/internal/domain"
	"github.com/marcio/editalbox/tvbox/internal/storage"
	"github.com/marcio/editalbox/tvbox/internal/telegram"
)

type App struct {
	cfg       config.Config
	store     *storage.Store
	collector *collector.Collector
	agent     *agent.Client
	bot       *telegram.Bot
	httpSrv   *http.Server
	mu        sync.Mutex
	lastSync  time.Time
}

func New(cfg config.Config) (*App, error) {
	store, err := storage.Open(cfg.DBPath)
	if err != nil {
		return nil, err
	}
	app := &App{
		cfg:       cfg,
		store:     store,
		collector: collector.New(cfg.CollectorUserAgent, cfg.CollectorFollowRules),
		agent:     agent.New(cfg.AgentBaseURL, cfg.AgentTimeout),
		bot:       telegram.New(cfg.TelegramToken),
	}
	app.httpSrv = &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: app.routes(),
	}
	return app, nil
}

func (a *App) Close() error {
	return a.store.Close()
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 3)
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		log.Printf("http listening on %s", a.cfg.HTTPAddr)
		if err := a.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	go func() {
		if err := a.runScheduler(runCtx); err != nil {
			errCh <- err
		}
	}()

	if a.bot.Enabled() {
		go func() {
			if err := a.runTelegram(runCtx); err != nil {
				errCh <- err
			}
		}()
	}

	select {
	case <-ctx.Done():
	case err := <-errCh:
		cancel()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		_ = a.httpSrv.Shutdown(shutdownCtx)
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = a.httpSrv.Shutdown(shutdownCtx)
	return nil
}

func (a *App) runScheduler(ctx context.Context) error {
	if err := a.syncNow(ctx); err != nil {
		log.Printf("initial sync failed: %v", err)
	}
	ticker := time.NewTicker(a.cfg.SyncInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := a.syncNow(ctx); err != nil {
				log.Printf("scheduled sync failed: %v", err)
			}
		}
	}
}

func (a *App) syncNow(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	runID, err := a.store.StartSyncRun(ctx)
	if err != nil {
		return err
	}

	visited := map[string]struct{}{}
	queue := append([]string{}, a.cfg.CollectorSeedURLs...)
	seedsVisited := 0
	pagesFetched := 0
	noticesUpserted := 0

	for len(queue) > 0 && pagesFetched < 60 {
		rawURL := queue[0]
		queue = queue[1:]
		if _, ok := visited[rawURL]; ok {
			continue
		}
		visited[rawURL] = struct{}{}
		seedsVisited++

		page, fetchErr := a.collector.Fetch(ctx, rawURL)
		if fetchErr != nil {
			log.Printf("fetch %s: %v", rawURL, fetchErr)
			continue
		}
		pagesFetched++

		noticeID, _, upsertErr := a.store.UpsertNotice(ctx, page.Notice)
		if upsertErr != nil {
			log.Printf("upsert notice %s: %v", rawURL, upsertErr)
			continue
		}
		noticesUpserted++

		for _, doc := range page.Documents {
			doc.NoticeID = noticeID
			_ = a.store.UpsertDocument(ctx, doc)
		}
		for _, event := range page.Events {
			event.NoticeID = noticeID
			_ = a.store.UpsertEvent(ctx, event)
		}
		for _, link := range page.Links {
			if _, ok := visited[link]; !ok {
				queue = append(queue, link)
			}
		}
	}

	if err := a.pushChangedNotices(ctx); err != nil {
		log.Printf("push changed notices: %v", err)
	}
	a.lastSync = time.Now().UTC()
	return a.store.FinishSyncRun(ctx, runID, "success", seedsVisited, pagesFetched, noticesUpserted, "")
}

func (a *App) pushChangedNotices(ctx context.Context) error {
	since := a.lastSync.Add(-25 * time.Hour)
	if a.lastSync.IsZero() {
		since = time.Now().UTC().Add(-365 * 24 * time.Hour)
	}
	notices, err := a.store.NoticesUpdatedSince(ctx, since, 200)
	if err != nil {
		return err
	}
	if len(notices) == 0 {
		return nil
	}

	var payload []agent.IngestNotice
	for _, notice := range notices {
		payload = append(payload, agent.IngestNotice{
			ID:        notice.ID,
			Title:     notice.Title,
			URL:       notice.CanonicalURL,
			Status:    notice.Status,
			Excerpt:   notice.Excerpt,
			BodyText:  notice.BodyText,
			UpdatedAt: notice.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}
	return a.agent.Ingest(ctx, payload)
}

func (a *App) runTelegram(ctx context.Context) error {
	var offset int64
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		updates, err := a.bot.GetUpdates(ctx, offset)
		if err != nil {
			log.Printf("telegram get updates: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		for _, update := range updates {
			offset = update.UpdateID + 1
			if !a.isAllowed(update.Message.Chat.ID) {
				continue
			}
			if err := a.handleTelegramMessage(ctx, update.Message); err != nil {
				log.Printf("telegram handle message: %v", err)
			}
		}
	}
}

func (a *App) isAllowed(chatID int64) bool {
	if len(a.cfg.AllowedChatIDs) == 0 {
		return true
	}
	_, ok := a.cfg.AllowedChatIDs[chatID]
	return ok
}

func (a *App) handleTelegramMessage(ctx context.Context, message telegram.Message) error {
	text := strings.TrimSpace(message.Text)
	if text == "" {
		return nil
	}
	if err := a.store.SaveMessage(ctx, message.Chat.ID, "user", text); err != nil {
		return err
	}

	reply := ""
	switch {
	case strings.EqualFold(text, "/status"):
		total, lastSync, err := a.store.GetStatus(ctx)
		if err != nil {
			reply = "Nao foi possivel consultar o estado do sistema."
			break
		}
		health, _ := a.agent.Health(ctx)
		reply = fmt.Sprintf("Base local: %d editais\nUltima sincronizacao: %s\nAgent: %s\nOllama: %t", total, fallback(lastSync, "nenhuma"), fallback(health.Status, "indisponivel"), health.OllamaReady)
	case strings.EqualFold(text, "/sync"):
		if err := a.syncNow(ctx); err != nil {
			reply = "Sincronizacao falhou."
		} else {
			reply = "Sincronizacao concluida."
		}
	case strings.HasPrefix(strings.ToLower(text), "/recent"):
		notices, err := a.store.RecentNotices(ctx, 5)
		if err != nil {
			reply = "Nao foi possivel listar os editais recentes."
			break
		}
		reply = formatNotices(notices)
	default:
		reply = a.answerQuery(ctx, message.Chat.ID, text)
	}

	if err := a.store.SaveMessage(ctx, message.Chat.ID, "assistant", reply); err != nil {
		return err
	}
	return a.bot.SendMessage(ctx, message.Chat.ID, reply)
}

func (a *App) answerQuery(ctx context.Context, chatID int64, question string) string {
	candidatesPool, err := a.buildCandidatePool(ctx, question)
	if err != nil || len(candidatesPool) == 0 {
		return "Nao encontrei editais compativeis na base local."
	}

	sessionSummary, _ := a.store.SessionSummary(ctx, chatID, a.cfg.SessionRetention)
	var candidates []domain.CandidateAnswer
	for _, notice := range candidatesPool {
		candidates = append(candidates, domain.CandidateAnswer{
			ID:       notice.ID,
			Title:    notice.Title,
			URL:      notice.CanonicalURL,
			Status:   notice.Status,
			Excerpt:  notice.Excerpt,
			BodyText: notice.BodyText,
		})
	}

	result, err := a.agent.Answer(ctx, agent.AnswerRequest{
		Question:       question,
		SessionSummary: sessionSummary,
		Limit:          5,
		Candidates:     candidates,
	})
	if err != nil || len(result.Structured) == 0 {
		return formatNotices(candidatesPool[:min(5, len(candidatesPool))])
	}

	var lines []string
	lines = append(lines, result.Text)
	for _, item := range result.Structured {
		lines = append(lines, fmt.Sprintf("- %s\n%s\nMotivo: %s", item.Title, item.URL, item.Justification))
	}
	return strings.Join(lines, "\n\n")
}

func (a *App) buildCandidatePool(ctx context.Context, question string) ([]domain.Notice, error) {
	primary, err := a.store.SearchNotices(ctx, question, 12)
	if err != nil {
		return nil, err
	}
	recent, err := a.store.RecentNotices(ctx, 20)
	if err != nil {
		return nil, err
	}
	merged := make([]domain.Notice, 0, len(primary)+len(recent))
	seen := map[int64]struct{}{}
	for _, notice := range append(primary, recent...) {
		if _, ok := seen[notice.ID]; ok {
			continue
		}
		seen[notice.ID] = struct{}{}
		merged = append(merged, notice)
	}
	return merged, nil
}

func formatNotices(notices []domain.Notice) string {
	if len(notices) == 0 {
		return "Nenhum edital encontrado."
	}
	var parts []string
	for _, notice := range notices {
		parts = append(parts, fmt.Sprintf("- %s\n%s\nStatus: %s", notice.Title, notice.CanonicalURL, notice.Status))
	}
	return strings.Join(parts, "\n\n")
}

func fallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
