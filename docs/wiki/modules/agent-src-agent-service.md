# Module: `agent/src/agent/service`

Núcleo do nó auxiliar. Recebe editais do tvbox, indexa e responde consultas em linguagem natural com ranqueamento por sinônimos e (opcionalmente) geração via Ollama.

## Responsibilities

- `ingest` — converter payload em `IndexedNotice` e upsertar no `Store` (que já faz o chunking).
- `answer` — ranquear candidatos, montar justificativas e, se o Ollama estiver pronto, gerar texto em LN grounding-only.
- Expansão de consulta por sinônimos PT-BR e bônus por intenção (bolsa, professor, curso, resultado).

## Key Files

- [`service.py`](../../agent/src/agent/service.py) — `AgentService`, `AnswerItem`, funções de tokenização/ranqueamento.

## Public API

- `def ingest(self, payload: list[dict]) -> tuple[int, int]` — (notices indexados, chunks).
- `def answer(self, question, session_summary, candidates, limit) -> dict` — `{text, structured, used_ollama}`.

## Internal Structure

`answer` chama `_rank`, que pontua cada candidato por sobreposição de tokens expandidos (`expanded_query_tokens`, com 8 grupos de sinônimos) + `intent_bonus`, com bônus de status (`open`+3, `in_progress`+1). Se não houver overlap, cai no `store.search` (chunks indexados) e, em último caso, nos próprios candidatos. `_justify` explica o match; `_prompt` monta o prompt do Ollama com apenas os itens selecionados; `looks_safe` descarta respostas suspeitas.

## Dependencies

- **Used by:** `main`.
- **Uses:** `.ollama`, `.store`.

## Notable Patterns / Gotchas

- `expanded_query_tokens` injeta sinônimos inteiros quando um termo da consulta bate com o grupo (ex.: "bolsa" → também "auxilio", "moradia", "monitoria").
- `looks_safe` bloqueia respostas com caracteres CJK (0x4E00–0x9FFF) — defesa contra saída inesperada do modelo.
- `used_ollama=false` significa que o `text` é o fallback determinístico (sem geração).
