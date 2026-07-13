# Module: `tvbox/internal/telegram`

Cliente mínimo do Bot API do Telegram. Faz polling de updates (`getUpdates`) e envio de mensagens (`sendMessage`) com retrying exponencial.

## Responsibilities

- Polling longo de updates a partir de um `offset`.
- Envio de mensagens de texto.
- Retrying com backoff exponencial (`base * 2^attempt`) apenas em erros temporários (5xx, 429).
- Desativar silenciosamente quando não há token configurado (`Enabled()`).

## Key Files

- [`bot.go`](../../tvbox/internal/telegram/bot.go) — `Bot`, `GetUpdates`, `SendMessage`, `call`, `backoffForAttempt`, `isRetryable`.
- [`bot_test.go`](../../tvbox/internal/telegram/bot_test.go) — testes.

## Public API

- `func New(token string, pollTimeout time.Duration, retryLimit int, retryBackoff time.Duration) *Bot`
- `func (b *Bot) Enabled() bool`
- `func (b *Bot) GetUpdates(ctx, offset int64) ([]Update, error)`
- `func (b *Bot) SendMessage(ctx, chatID int64, text string) error`

## Internal Structure

`call` encapsula POST JSON com loop de retry. `backoffForAttempt` dobra o tempo a cada tentativa; `isRetryable` retorna falso para `ctx.Err()` (para não retryar cancelled). `Update`/`Message` espelham o JSON da API.

## Dependencies

- **Used by:** `internal/app`.
- **Uses:** apenas stdlib (`net/http`, `encoding/json`).

## Notable Patterns / Gotchas

- `Enabled()`=false (sem token) faz `GetUpdates`/`SendMessage` retornarem nil sem erro — o app roda sem Telegram.
- O backoff só dispara em 5xx/429; 4xx é erro definitivo (retorna imediatamente).
- `truncateBody` limita o corpo de erro logado a 220 chars.
