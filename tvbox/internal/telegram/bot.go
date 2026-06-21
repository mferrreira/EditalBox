package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Bot struct {
	token        string
	baseURL      string
	http         *http.Client
	pollTimeout  time.Duration
	retryLimit   int
	retryBackoff time.Duration
}

type Update struct {
	UpdateID int64   `json:"update_id"`
	Message  Message `json:"message"`
}

type Message struct {
	MessageID int64 `json:"message_id"`
	Chat      struct {
		ID int64 `json:"id"`
	} `json:"chat"`
	Text string `json:"text"`
}

func New(token string, pollTimeout time.Duration, retryLimit int, retryBackoff time.Duration) *Bot {
	return &Bot{
		token:        token,
		baseURL:      "https://api.telegram.org",
		http:         &http.Client{Timeout: pollTimeout + 10*time.Second},
		pollTimeout:  pollTimeout,
		retryLimit:   retryLimit,
		retryBackoff: retryBackoff,
	}
}

func (b *Bot) Enabled() bool {
	return strings.TrimSpace(b.token) != ""
}

func (b *Bot) GetUpdates(ctx context.Context, offset int64) ([]Update, error) {
	if !b.Enabled() {
		return nil, nil
	}
	body := map[string]any{
		"offset":  offset,
		"timeout": int(b.pollTimeout.Seconds()),
	}
	var response struct {
		OK     bool     `json:"ok"`
		Result []Update `json:"result"`
	}
	if err := b.call(ctx, "getUpdates", body, &response); err != nil {
		return nil, err
	}
	return response.Result, nil
}

func (b *Bot) SendMessage(ctx context.Context, chatID int64, text string) error {
	if !b.Enabled() {
		return nil
	}
	payload := map[string]any{
		"chat_id": chatID,
		"text":    text,
	}
	return b.call(ctx, "sendMessage", payload, nil)
}

func (b *Bot) call(ctx context.Context, method string, payload any, out any) error {
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		return err
	}
	var lastErr error
	for attempt := 0; attempt <= b.retryLimit; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/bot%s/%s", b.baseURL, b.token, method), bytes.NewReader(buf.Bytes()))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := b.http.Do(req)
		if err != nil {
			lastErr = err
		} else {
			body, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()
			if readErr != nil {
				lastErr = readErr
			} else if resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests {
				lastErr = fmt.Errorf("telegram temporary error: status=%d body=%s", resp.StatusCode, truncateBody(string(body)))
			} else if resp.StatusCode >= 400 {
				return fmt.Errorf("telegram request failed: status=%d body=%s", resp.StatusCode, truncateBody(string(body)))
			} else if out != nil {
				if err := json.Unmarshal(body, out); err != nil {
					return err
				}
				return nil
			} else {
				return nil
			}
		}

		if attempt == b.retryLimit || !isRetryable(ctx, lastErr) {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoffForAttempt(b.retryBackoff, attempt)):
		}
	}
	return lastErr
}

func backoffForAttempt(base time.Duration, attempt int) time.Duration {
	if base <= 0 {
		base = time.Second
	}
	return base * time.Duration(1<<attempt)
}

func isRetryable(ctx context.Context, err error) bool {
	if err == nil {
		return false
	}
	return !errors.Is(err, ctx.Err())
}

func truncateBody(body string) string {
	body = strings.Join(strings.Fields(body), " ")
	if len(body) > 220 {
		return body[:220]
	}
	return body
}
