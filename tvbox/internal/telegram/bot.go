package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Bot struct {
	token string
	http  *http.Client
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

func New(token string) *Bot {
	return &Bot{
		token: token,
		http:  &http.Client{Timeout: 20 * time.Second},
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
		"timeout": 25,
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
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("https://api.telegram.org/bot%s/%s", b.token, method), buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := b.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}
