package collector

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestExtractTitlePrefersHeadingOverGenericTitle(t *testing.T) {
	html := `<html><head><title>Portal IFNMG - Mais Noticias Januaria</title></head><body><h1>Edital de Bolsa Treinamento 2026</h1></body></html>`
	if got := extractTitle(html); got != "Edital de Bolsa Treinamento 2026" {
		t.Fatalf("unexpected title: %q", got)
	}
}

func TestExtractEventsAndStatusFromDateRange(t *testing.T) {
	text := "As inscricoes estarao abertas de 22 de junho a 05 de julho de 2026. Resultado final em 20/07/2026."
	events := extractEvents(text)
	registrationEnd, finalEventAt := deriveImportantDates(events)
	if registrationEnd == nil {
		t.Fatalf("expected registration end")
	}
	if got := registrationEnd.Format("2006-01-02"); got != "2026-07-05" {
		t.Fatalf("unexpected registration end: %s", got)
	}
	if finalEventAt == nil {
		t.Fatalf("expected final event")
	}
	if got := finalEventAt.Format("2006-01-02"); got != "2026-07-20" {
		t.Fatalf("unexpected final event: %s", got)
	}
	status := deriveStatus(text, registrationEnd, finalEventAt, time.Date(2026, 6, 25, 0, 0, 0, 0, time.UTC))
	if status != "open" {
		t.Fatalf("unexpected status: %s", status)
	}
}

func TestFetchBuildsMeaningfulNotice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html><head><title>Portal IFNMG - Editais</title></head><body><h1>Bolsa de Pesquisa 2026</h1><p>Inscricoes ate 12/08/2026.</p><a href="/doc">Documento</a></body></html>`))
	}))
	defer server.Close()

	c := New("test-agent", []string{"/doc"})
	result, err := c.Fetch(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if result.Notice.Title != "Bolsa de Pesquisa 2026" {
		t.Fatalf("unexpected title: %q", result.Notice.Title)
	}
	if !strings.Contains(strings.ToLower(result.Notice.Excerpt), "inscricoes") {
		t.Fatalf("unexpected excerpt: %q", result.Notice.Excerpt)
	}
}
