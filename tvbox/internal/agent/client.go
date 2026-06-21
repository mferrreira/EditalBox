package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/marcio/editalbox/tvbox/internal/domain"
)

type Client struct {
	baseURL string
	http    *http.Client
}

type Health struct {
	Status       string `json:"status"`
	OllamaReady  bool   `json:"ollama_ready"`
	IndexedCount int    `json:"indexed_count"`
}

type AnswerRequest struct {
	Question       string                   `json:"question"`
	SessionSummary string                   `json:"session_summary"`
	Limit          int                      `json:"limit"`
	Candidates     []domain.CandidateAnswer `json:"candidates"`
}

type AnswerResult struct {
	Text       string       `json:"text"`
	Structured []AnswerItem `json:"structured"`
	UsedOllama bool         `json:"used_ollama"`
}

type AnswerItem struct {
	Title         string `json:"title"`
	URL           string `json:"url"`
	Status        string `json:"status"`
	Justification string `json:"justification"`
}

type IngestNotice struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Status    string `json:"status"`
	Excerpt   string `json:"excerpt"`
	BodyText  string `json:"body_text"`
	UpdatedAt string `json:"updated_at"`
}

func New(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: timeout},
	}
}

func (c *Client) Health(ctx context.Context) (Health, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return Health{}, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return Health{}, err
	}
	defer resp.Body.Close()
	var health Health
	err = json.NewDecoder(resp.Body).Decode(&health)
	return health, err
}

func (c *Client) Ingest(ctx context.Context, notices []IngestNotice) error {
	payload := map[string]any{"notices": notices}
	return c.post(ctx, "/v1/ingest", payload, nil)
}

func (c *Client) Answer(ctx context.Context, request AnswerRequest) (AnswerResult, error) {
	var out AnswerResult
	err := c.post(ctx, "/v1/answer", request, &out)
	return out, err
}

func (c *Client) post(ctx context.Context, path string, body any, out any) error {
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
