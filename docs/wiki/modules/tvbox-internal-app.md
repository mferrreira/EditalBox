# Module: `tvbox/internal/app`

Coração do serviço principal. Orquestra coleta, persistência, servidor HTTP, polling do Telegram e roteamento de consultas para o agente auxiliar.

## Responsibilities

- Inicializar `Store`, `Collector`, `Agent Client` e `Telegram Bot` a partir da config.
- Rodar em paralelo: servidor HTTP, scheduler de sincronização e polling do Telegram.
- Executar a sincronização em BFS limitado (até 60 páginas por ciclo) a partir das seeds.
- Empurrar editais alterados para o agente (`/v1/ingest`).
- Responder comandos (`/status`, `/sync`, `/recent`) e perguntas em LN via agente.

## Key Files

- [`app.go`](../../tvbox/internal/app/app.go) — `App`, `Run`, `runScheduler`, `syncNow`, `runTelegram`, `handleTelegramMessage`, `answerQuery`, `buildCandidatePool`.
- [`http.go`](../../tvbox/internal/app/http.go) — `routes()`, `handleHealth`, `handleSync` (POST `/sync`).

## Public API

- `func New(cfg config.Config) (*App, error)`
- `func (a *App) Run(ctx context.Context) error`
- `func (a *App) syncNow(ctx context.Context) error` — dispara uma coleta completa.
- `func (a *App) answerQuery(ctx, chatID, question) string` — devolve texto (resposta do agente ou fallback local).

## Internal Structure

`Run` lança 3 goroutines (HTTP, scheduler, telegram se `bot.Enabled()`) e aguarda `ctx.Done()` ou erro. O `syncNow` faz BFS com fila + conjunto `visited`, caps em `pagesFetched < 60`, grava um `SyncRun` e chama `pushChangedNotices` (janela de 25h antes do último sync, ou 1 ano na primeira vez). `handleTelegramMessage` persiste a mensagem, decide comando vs. consulta e devolve ao usuário; consultas caem em `answerQuery`, que monta um pool de candidatos (busca textual + recentes, deduplicados) e chama o agente.

## Dependencies

- **Used by:** `cmd/editalbox`.
- **Uses:** `internal/collector`, `internal/storage`, `internal/agent`, `internal/telegram`, `internal/config`, `internal/domain`.

## Notable Patterns / Gotchas

- `pagesFetched < 60` é um limite de custo/segurança por ciclo de sync.
- `pushChangedNotices` usa `lastSync.Add(-25h)` de propósito para recuperar editais do ciclo anterior que possam ter sido perdidos.
- Fallback: se o agente falha ou não devolve `Structured`, `answerQuery` devolve a listagem textual dos candidatos (`formatNotices`).
- `isAllowed` bloqueia chats fora de `AllowedChatIDs` (vazio = todos).
