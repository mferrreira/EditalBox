# Module: `tvbox/internal/storage`

Camada de persistência do nó principal. Usa SQLite puro (driver `modernc.org/sqlite`, sem cgo) com WAL. Concentra migração de schema, upserts idempotentes e a busca textual ranqueada.

## Responsibilities

- Criar/abrir o banco e aplicar o schema (6 tabelas).
- Upsert de `Notice` (por `canonical_url`), `NoticeDocument` (por `url`) e `NoticeEvent` (por tupla única).
- Registrar `SyncRun` (início/fim, contadores).
- Busca textual ranqueada (`SearchNotices`) e listagem recente.
- Persistir mensagens e resumos de sessão do Telegram.

## Key Files

- [`sqlite.go`](../../tvbox/internal/storage/sqlite.go) — `Store`, `migrate`, `Upsert*`, `SearchNotices`, `RecentNotices`, `SaveMessage`, `SessionSummary`, `NoticesUpdatedSince`.

## Public API

- `func Open(path string) (*Store, error)`
- `func (s *Store) UpsertNotice(ctx, domain.Notice) (id int64, inserted bool, err error)`
- `func (s *Store) SearchNotices(ctx, query string, limit int) ([]domain.Notice, error)`
- `func (s *Store) GetStatus(ctx) (total int, lastSync string, err error)`
- `func (s *Store) NoticesUpdatedSince(ctx, since time.Time, limit int) ([]domain.Notice, error)`

## Internal Structure

`migrate()` cria `notices` (canonical_url UNIQUE), `notice_documents`, `notice_events` (com ON DELETE CASCADE), `sync_runs`, `telegram_sessions`, `telegram_messages`. A busca (`scoreNotices`/`noticeScore`) pontua candidatos por tokens (título +8, excerpt +4, corpo +2) com stopwords PT e bônus +2 para `status=open`. `SessionSummary` recupera até 12 mensagens dentro de `retention`.

## Dependencies

- **Used by:** `internal/app`.
- **Uses:** `internal/domain`.

## Notable Patterns / Gotchas

- `canonical_url` é a chave de deduplicação — re-syncs atualizam em vez de duplicar.
- `notices` trunca datas em RFC3339 (strings), não `time.Time` — cuidado ao comparar.
- `SearchNotices` materializa até 500 recentes e ranqueia em memória (adequado ao volume de uma TV Box, não para milhões de linhas).
