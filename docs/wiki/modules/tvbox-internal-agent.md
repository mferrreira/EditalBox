# Module: `tvbox/internal/agent`

Cliente HTTP do tvbox para o serviço auxiliar (agente Python). É a única ponte de rede do nó principal para o nó de processamento.

## Responsibilities

- Verificar saúde do agente (`/health`).
- Empurrar editais novos/alterados (`POST /v1/ingest`).
- Solicitar resposta em LN para uma consulta (`POST /v1/answer`) com candidatos e resumo de sessão.

## Key Files

- [`client.go`](../../tvbox/internal/agent/client.go) — `Client`, `Health`, `Ingest`, `Answer`, `post`.

## Public API

- `func New(baseURL string, timeout time.Duration) *Client`
- `func (c *Client) Health(ctx) (Health, error)`
- `func (c *Client) Ingest(ctx, notices []IngestNotice) error`
- `func (c *Client) Answer(ctx, AnswerRequest) (AnswerResult, error)`

Tipos de payload: `Health{Status, OllamaReady, IndexedCount}`, `AnswerRequest{Question, SessionSummary, Limit, Candidates}`, `AnswerResult{Text, Structured[], UsedOllama}`, `AnswerItem{Title, URL, Status, Justification}`, `IngestNotice{...}`.

## Internal Structure

`post` é o único método de transporte (POST JSON, decode opcional). `New` normaliza a base URL (remove barra final). `Ingest` envelopa a lista em `{"notices": [...]}`.

## Dependencies

- **Used by:** `internal/app`.
- **Uses:** `internal/domain` (tipos de payload).

## Notable Patterns / Gotchas

- O contrato de API (`/v1/ingest`, `/v1/answer`, `/health`) é espelhado em `agent/src/agent/main.py`.
- `UsedOllama` no `AnswerResult` indica se a resposta veio do modelo local ou foi só o texto de fallback.
