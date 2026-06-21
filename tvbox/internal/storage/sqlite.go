package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/marcio/editalbox/tvbox/internal/domain"
)

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	store := &Store{db: db}
	if err := store.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) DB() *sql.DB {
	return s.db
}

func (s *Store) migrate() error {
	schema := `
PRAGMA journal_mode=WAL;
CREATE TABLE IF NOT EXISTS notices (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  source_url TEXT NOT NULL,
  canonical_url TEXT NOT NULL UNIQUE,
  title TEXT NOT NULL,
  excerpt TEXT NOT NULL DEFAULT '',
  body_text TEXT NOT NULL DEFAULT '',
  source_type TEXT NOT NULL DEFAULT 'unknown',
  status TEXT NOT NULL DEFAULT 'unknown',
  registration_end TEXT,
  final_event_at TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS notice_documents (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  notice_id INTEGER NOT NULL,
  url TEXT NOT NULL UNIQUE,
  kind TEXT NOT NULL DEFAULT 'unknown',
  title TEXT NOT NULL DEFAULT '',
  updated_at TEXT NOT NULL,
  FOREIGN KEY(notice_id) REFERENCES notices(id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS notice_events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  notice_id INTEGER NOT NULL,
  event_type TEXT NOT NULL,
  label TEXT NOT NULL,
  event_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE(notice_id, event_type, label, event_at),
  FOREIGN KEY(notice_id) REFERENCES notices(id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS sync_runs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  started_at TEXT NOT NULL,
  finished_at TEXT,
  status TEXT NOT NULL,
  seeds_visited INTEGER NOT NULL DEFAULT 0,
  pages_fetched INTEGER NOT NULL DEFAULT 0,
  notices_upserted INTEGER NOT NULL DEFAULT 0,
  error_message TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS telegram_sessions (
  chat_id INTEGER PRIMARY KEY,
  summary TEXT NOT NULL DEFAULT '',
  last_activity TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS telegram_messages (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  chat_id INTEGER NOT NULL,
  role TEXT NOT NULL,
  content TEXT NOT NULL,
  created_at TEXT NOT NULL
);`
	_, err := s.db.Exec(schema)
	return err
}

func (s *Store) UpsertNotice(ctx context.Context, notice domain.Notice) (int64, bool, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	var id int64
	err := s.db.QueryRowContext(ctx, `SELECT id FROM notices WHERE canonical_url = ?`, notice.CanonicalURL).Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		return 0, false, err
	}

	regEnd := nullableTime(notice.RegistrationEnd)
	finalEvent := nullableTime(notice.FinalEventAt)
	if err == sql.ErrNoRows {
		result, execErr := s.db.ExecContext(ctx, `
INSERT INTO notices (source_url, canonical_url, title, excerpt, body_text, source_type, status, registration_end, final_event_at, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			notice.SourceURL, notice.CanonicalURL, notice.Title, notice.Excerpt, notice.BodyText, notice.SourceType, notice.Status, regEnd, finalEvent, now, now)
		if execErr != nil {
			return 0, false, execErr
		}
		insertID, _ := result.LastInsertId()
		return insertID, true, nil
	}

	_, err = s.db.ExecContext(ctx, `
UPDATE notices
SET source_url = ?, title = ?, excerpt = ?, body_text = ?, source_type = ?, status = ?, registration_end = ?, final_event_at = ?, updated_at = ?
WHERE id = ?`,
		notice.SourceURL, notice.Title, notice.Excerpt, notice.BodyText, notice.SourceType, notice.Status, regEnd, finalEvent, now, id)
	return id, false, err
}

func (s *Store) UpsertDocument(ctx context.Context, doc domain.NoticeDocument) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO notice_documents (notice_id, url, kind, title, updated_at)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(url) DO UPDATE SET
  notice_id = excluded.notice_id,
  kind = excluded.kind,
  title = excluded.title,
  updated_at = excluded.updated_at`,
		doc.NoticeID, doc.URL, doc.Kind, doc.Title, time.Now().UTC().Format(time.RFC3339))
	return err
}

func (s *Store) UpsertEvent(ctx context.Context, event domain.NoticeEvent) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO notice_events (notice_id, event_type, label, event_at, updated_at)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(notice_id, event_type, label, event_at) DO UPDATE SET
  updated_at = excluded.updated_at`,
		event.NoticeID, event.EventType, event.Label, event.EventAt.UTC().Format(time.RFC3339), time.Now().UTC().Format(time.RFC3339))
	return err
}

func (s *Store) StartSyncRun(ctx context.Context) (int64, error) {
	result, err := s.db.ExecContext(ctx, `INSERT INTO sync_runs (started_at, status) VALUES (?, 'running')`, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *Store) FinishSyncRun(ctx context.Context, id int64, status string, seedsVisited, pagesFetched, noticesUpserted int, errMsg string) error {
	_, err := s.db.ExecContext(ctx, `
UPDATE sync_runs
SET finished_at = ?, status = ?, seeds_visited = ?, pages_fetched = ?, notices_upserted = ?, error_message = ?
WHERE id = ?`,
		time.Now().UTC().Format(time.RFC3339), status, seedsVisited, pagesFetched, noticesUpserted, errMsg, id)
	return err
}

func (s *Store) GetStatus(ctx context.Context) (int, string, error) {
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM notices`).Scan(&total); err != nil {
		return 0, "", err
	}
	var lastSync sql.NullString
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(finished_at), '') FROM sync_runs WHERE status = 'success'`).Scan(&lastSync); err != nil {
		return 0, "", err
	}
	return total, lastSync.String, nil
}

func (s *Store) SearchNotices(ctx context.Context, query string, limit int) ([]domain.Notice, error) {
	candidates, err := s.RecentNotices(ctx, 500)
	if err != nil {
		return nil, err
	}
	scored := scoreNotices(candidates, query)
	if len(scored) == 0 {
		return []domain.Notice{}, nil
	}
	if len(scored) > limit {
		scored = scored[:limit]
	}
	return scored, nil
}

func (s *Store) RecentNotices(ctx context.Context, limit int) ([]domain.Notice, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, source_url, canonical_url, title, excerpt, body_text, source_type, status, registration_end, final_event_at, created_at, updated_at
FROM notices
ORDER BY updated_at DESC
LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanNotices(rows)
}

func (s *Store) SaveMessage(ctx context.Context, chatID int64, role, content string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, `INSERT INTO telegram_messages (chat_id, role, content, created_at) VALUES (?, ?, ?, ?)`, chatID, role, content, now)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
INSERT INTO telegram_sessions (chat_id, summary, last_activity)
VALUES (?, '', ?)
ON CONFLICT(chat_id) DO UPDATE SET last_activity = excluded.last_activity`, chatID, now)
	return err
}

func (s *Store) SessionSummary(ctx context.Context, chatID int64, retention time.Duration) (string, error) {
	cutoff := time.Now().UTC().Add(-retention).Format(time.RFC3339)
	rows, err := s.db.QueryContext(ctx, `
SELECT role, content
FROM telegram_messages
WHERE chat_id = ? AND created_at >= ?
ORDER BY created_at ASC
LIMIT 12`, chatID, cutoff)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var parts []string
	for rows.Next() {
		var role, content string
		if err := rows.Scan(&role, &content); err != nil {
			return "", err
		}
		parts = append(parts, fmt.Sprintf("%s: %s", role, content))
	}
	return strings.Join(parts, "\n"), rows.Err()
}

func (s *Store) NoticesUpdatedSince(ctx context.Context, since time.Time, limit int) ([]domain.Notice, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, source_url, canonical_url, title, excerpt, body_text, source_type, status, registration_end, final_event_at, created_at, updated_at
FROM notices
WHERE updated_at >= ?
ORDER BY updated_at ASC
LIMIT ?`, since.UTC().Format(time.RFC3339), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanNotices(rows)
}

func like(query string) string {
	return "%" + strings.ToLower(strings.TrimSpace(query)) + "%"
}

func scoreNotices(notices []domain.Notice, query string) []domain.Notice {
	tokens := normalizedTokens(query)
	if len(tokens) == 0 {
		return notices
	}

	type scoredNotice struct {
		score  int
		notice domain.Notice
	}
	var scored []scoredNotice
	for _, notice := range notices {
		score := noticeScore(notice, tokens)
		if score <= 0 {
			continue
		}
		scored = append(scored, scoredNotice{score: score, notice: notice})
	}

	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].score == scored[j].score {
			return scored[i].notice.UpdatedAt.After(scored[j].notice.UpdatedAt)
		}
		return scored[i].score > scored[j].score
	})

	out := make([]domain.Notice, 0, len(scored))
	for _, item := range scored {
		out = append(out, item.notice)
	}
	return out
}

func noticeScore(notice domain.Notice, tokens []string) int {
	text := strings.ToLower(strings.Join([]string{notice.Title, notice.Excerpt, notice.BodyText, notice.Status}, " "))
	score := 0
	for _, token := range tokens {
		switch {
		case strings.Contains(strings.ToLower(notice.Title), token):
			score += 8
		case strings.Contains(strings.ToLower(notice.Excerpt), token):
			score += 4
		case strings.Contains(text, token):
			score += 2
		}
	}
	if notice.Status == "open" {
		score += 2
	}
	return score
}

func normalizedTokens(query string) []string {
	raw := strings.Fields(strings.ToLower(query))
	stopwords := map[string]struct{}{
		"a": {}, "as": {}, "o": {}, "os": {}, "de": {}, "da": {}, "das": {}, "do": {}, "dos": {},
		"e": {}, "ou": {}, "um": {}, "uma": {}, "para": {}, "por": {}, "com": {}, "sem": {},
		"quais": {}, "qual": {}, "me": {}, "mostra": {}, "mostrar": {}, "tem": {}, "sobre": {},
		"que": {}, "eu": {}, "preciso": {}, "quero": {}, "abertos": {}, "aberto": {},
	}
	var out []string
	seen := map[string]struct{}{}
	for _, token := range raw {
		token = strings.Trim(token, ".,;:!?()[]{}\"'")
		if len(token) < 3 {
			continue
		}
		if _, ok := stopwords[token]; ok {
			continue
		}
		if _, ok := seen[token]; ok {
			continue
		}
		seen[token] = struct{}{}
		out = append(out, token)
	}
	return out
}

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format(time.RFC3339)
}

func scanNotices(rows *sql.Rows) ([]domain.Notice, error) {
	var out []domain.Notice
	for rows.Next() {
		var notice domain.Notice
		var regEnd, finalAt sql.NullString
		var createdAt, updatedAt string
		if err := rows.Scan(
			&notice.ID,
			&notice.SourceURL,
			&notice.CanonicalURL,
			&notice.Title,
			&notice.Excerpt,
			&notice.BodyText,
			&notice.SourceType,
			&notice.Status,
			&regEnd,
			&finalAt,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, err
		}
		if parsed, err := time.Parse(time.RFC3339, createdAt); err == nil {
			notice.CreatedAt = parsed
		}
		if parsed, err := time.Parse(time.RFC3339, updatedAt); err == nil {
			notice.UpdatedAt = parsed
		}
		if regEnd.Valid {
			if parsed, err := time.Parse(time.RFC3339, regEnd.String); err == nil {
				notice.RegistrationEnd = &parsed
			}
		}
		if finalAt.Valid {
			if parsed, err := time.Parse(time.RFC3339, finalAt.String); err == nil {
				notice.FinalEventAt = &parsed
			}
		}
		out = append(out, notice)
	}
	return out, rows.Err()
}
