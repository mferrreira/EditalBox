# Module: `tvbox/cmd/editalbox`

Ponto de entrada do serviço principal (nó da TV Box). Não contém lógica de domínio — apenas monta o `App` a partir da config e trata o ciclo de vida do processo.

## Responsibilities

- Carregar configuração via `config.Load()`.
- Construir a aplicação com `app.New(cfg)`.
- Ficar em execução até receber `SIGINT`/`SIGTERM` (via `signal.NotifyContext`), garantindo `application.Close()` no encerramento.

## Key Files

- [`main.go`](../../tvbox/cmd/editalbox/main.go) — `main()`: load → New → Run até sinal.

## Public API

- `func main()` — entrypoint; sem retorno.

## Internal Structure

Sequencial e mínima: `config.Load()` → `app.New()` → `app.Run(ctx)`. O `defer application.Close()` garante fechamento do SQLite.

## Dependencies

- **Used by:** nenhum (é o binário).
- **Uses:** `internal/app`, `internal/config`.

## Notable Patterns / Gotchas

- O contexto de cancelamento vem de sinais do SO; `Run` faz shutdown gracioso do HTTP em até 5s.
