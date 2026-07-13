# Module: `agent/src/agent/main`

Bootstrap do serviço auxiliar (nó Python). Sobe um `ThreadingHTTPServer` que expõe as rotas consumidas pelo tvbox.

## Responsibilities

- Carregar `Config`, abrir o `Store` do agente e montar o `AgentService`.
- Servir `GET /health`, `POST /v1/ingest`, `POST /v1/answer`.
- Fechar o `Store` no encerramento.

## Key Files

- [`main.py`](../../agent/src/agent/main.py) — `main()`, `Handler` (HTTP), rotas.

## Public API

- `def main() -> None` — entrypoint.
- `GET /health` → `{status, ollama_ready, indexed_count}`.
- `POST /v1/ingest` (body `{notices:[...]}`) → `{indexed, chunks}`.
- `POST /v1/answer` (body `{question, session_summary, candidates, limit}`) → `{text, structured, used_ollama}`.

## Internal Structure

`Handler` (interno) implementa `do_GET`/`do_POST`. O `serve_forever()` roda até Ctrl-C; `finally` chama `store.close()`. `log_message` é sobrescrito para silenciar logs de acesso.

## Dependencies

- **Used by:** scripts `start-*.sh`.
- **Uses:** `.config`, `.ollama`, `.service`, `.store`.

## Notable Patterns / Gotchas

- `ThreadingHTTPServer` atende múltiplas requisições concorrentes (o tvbox pode ingestar enquanto responde consultas).
- `_json` escreve o corpo e os headers manualmente; o `Content-Length` é obrigatório.
