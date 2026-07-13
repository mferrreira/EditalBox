# Module: `tvbox/internal/config`

Carga de configuração a partir de variáveis de ambiente, com defaults sensíveis para rodar em dev/local sem nenhuma configuração.

## Responsibilities

- Expor a struct `Config` com todos os parâmetros do tvbox.
- Resolver cada campo de `os.Getenv` com fallback.
- Embutir as seeds de coleta e regras de follow (URLs do IFNMG Januária).
- Definir o `User-Agent` de coleta (Chrome/Linux armv7, típico de TV Box).

## Key Files

- [`config.go`](../../tvbox/internal/config/config.go) — `Config`, `Load`, `env`, `envDuration`, `envInt`, `parseChatIDs`.

## Public API

- `func Load() Config` — lê o ambiente e retorna `Config` pronto.

`Config` traz: `Environment`, `HTTPAddr`, `DBPath`, `SyncInterval`, `TelegramToken`, `AllowedChatIDs`, `AgentBaseURL`, `AgentTimeout`, `SessionRetention`, `TelegramPollTimeout/RetryLimit/RetryBackoff`, `CollectorUserAgent`, `CollectorSeedURLs`, `CollectorFollowRules`.

## Internal Structure

`Load()` popula tudo via helpers `env*`; `parseChatIDs` aceita uma lista CSV de IDs numéricos. As seeds e follow rules são hardcoded em `Load()` (fonte do IFNMG Januária).

## Dependencies

- **Used by:** `cmd/editalbox`, `internal/app`.
- **Uses:** `os`, `time`.

## Notable Patterns / Gotchas

- Defaults: `HTTPAddr=:8080`, `DBPath=./data/editalbox.db`, `SyncInterval=24h`, `AgentBaseURL=http://127.0.0.1:8090`, `SessionRetention=24h`.
- `AllowedChatIDs` vazio = aceita qualquer chat.
- Seeds/follow rules são fixas no código, não em env — mudar origens exige recompilar.
