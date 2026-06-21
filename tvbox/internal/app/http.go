package app

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

func (a *App) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", a.handleHealth)
	mux.HandleFunc("/sync", a.handleSync)
	return mux
}

func (a *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	total, lastSync, err := a.store.GetStatus(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	health, _ := a.agent.Health(r.Context())
	writeJSON(w, http.StatusOK, map[string]any{
		"status":         "ok",
		"total_notices":  total,
		"last_sync":      lastSync,
		"uptime_seconds": int(time.Since(a.startedAt).Seconds()),
		"agent":          health,
	})
}

func (a *App) handleSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	if err := a.syncNow(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "queued"})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
