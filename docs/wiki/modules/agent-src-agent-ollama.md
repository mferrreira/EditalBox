# Module: `agent/src/agent/ollama`

Cliente mínimo do Ollama. Faz health check e geração de texto contra a API local.

## Responsibilities

- `health()` — verifica se o Ollama responde em `/api/tags` (200).
- `generate(prompt)` — chama `/api/generate` (stream desligado) e devolve o texto ou `None` em falha.

## Key Files

- [`ollama.py`](../../agent/src/agent/ollama.py) — `OllamaClient`.

## Public API

- `def __init__(self, config: Config)`
- `def health(self) -> bool`
- `def generate(self, prompt: str) -> str | None`

## Internal Structure

Usa `urllib.request` (sem dependências externas). `health` tem timeout de 3s; `generate` tem timeout de 20s. Modelo vindo de `config.ollama_model`.

## Dependencies

- **Used by:** `service`, `main` (health no `/health`).
- **Uses:** `.config`.

## Notable Patterns / Gotchas

- Sem `requests`/httpx — só stdlib, para manter o serviço enxuto.
- `generate` retorna `None` em `URLError` (o service então cai no fallback textual).
- `health` swallow qualquer exceção → `False` (degradação graciosa).
