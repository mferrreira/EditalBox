# Module: `agent/src/agent/store`

Persistência do índice local do agente. Mantém editais indexados e seus chunks com keywords para busca por sobreposição.

## Responsibilities

- Migrar schema (`indexed_notices`, `indexed_chunks` com FK cascade).
- `upsert_notices` — insere/atualiza editais e recalcula os chunks (apaga os antigos do notice e reindexa).
- `search` — ranqueia chunks por overlap de tokens com a consulta.
- `indexed_count` — total de editais indexados.

## Key Files

- [`store.py`](../../agent/src/agent/store.py) — `Store`, `IndexedNotice`, `_chunk_notice`, `search`, `tokenize`, `normalize_token`.

## Public API

- `def __init__(self, db_path: str)`
- `def upsert_notices(self, notices) -> tuple[int, int]` — (indexed, chunks).
- `def search(self, query: str, limit: int) -> list[sqlite3.Row]`
- `def indexed_count(self) -> int`

## Internal Structure

`_chunk_notice` divide o texto (título+excerpt+body) em janelas de 900 chars com overlap de 150, extraindo keywords (tokens >=4 chars). `search` faz JOIN com `indexed_chunks`, pontua por `overlap*10 + matches no chunk`, deduplica por `notice_id` e limita. `tokenize`/`normalize_token` remove acentos e plural simples (PT-BR).

## Dependencies

- **Used by:** `service`, `main`.
- **Uses:** `sqlite3` (stdlib).

## Notable Patterns / Gotchas

- Chunking por janela deslizante com overlap de 150 chars reduz perda de contexto na borda.
- `ON DELETE CASCADE` + delete explícito de chunks no reupsert mantém o índice coerente com o edital.
- `normalize_token` remove acento e corta "s" final (>4 chars) — heurística leve de stemming PT-BR.
