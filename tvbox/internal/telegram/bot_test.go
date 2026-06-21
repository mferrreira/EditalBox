package telegram

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestBotRetriesTemporaryFailures(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := atomic.AddInt32(&attempts, 1)
		if current < 3 {
			http.Error(w, "temporary", http.StatusBadGateway)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok": true,
			"result": []map[string]any{
				{
					"update_id": 1,
					"message": map[string]any{
						"message_id": 10,
						"chat":       map[string]any{"id": 20},
						"text":       "teste",
					},
				},
			},
		})
	}))
	defer server.Close()

	bot := New("token", 2*time.Second, 3, 10*time.Millisecond)
	bot.baseURL = server.URL
	updates, err := bot.GetUpdates(context.Background(), 0)
	if err != nil {
		t.Fatalf("get updates: %v", err)
	}
	if len(updates) != 1 {
		t.Fatalf("unexpected updates count: %d", len(updates))
	}
	if atomic.LoadInt32(&attempts) != 3 {
		t.Fatalf("unexpected attempts: %d", attempts)
	}
}
