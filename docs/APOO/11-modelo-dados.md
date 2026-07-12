# 11. Modelo de Dados

## Tabelas/recursos principais
TV Box:
- notices
- notice_documents
- notice_events
- sync_runs
- telegram_sessions
- telegram_messages

Agent:
- indexed_notices
- indexed_chunks

## Decisões de persistência
- SQLite local na TV Box e no agent.
- Separação de índices leves e estado operacional.
- Keys referenciáveis por source_url e id local.

## Estratégia de herança/extension
- Dados separados por componente; sincronização incremental.
