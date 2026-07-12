# 09. Dados e Modelo de Classes de Projeto

## Classes do tvbox (Go)
- Coletor/parser/seeds.
- Normalizador de links/conteúdos.
- Persistência SQLite: notices, notice_documents, notice_events, sync_runs, telegram_sessions, telegram_messages.
- Bot Telegram polling.
- Busca textual local.
- Health/base status.
- Configuração operacional.

## Classes do agent (Python)
- Ingestão incremental.
- Indexação/local chunks.
- Recuperação/ranqueamento contextual.
- Integração Ollama local.
- Serviço NLP/de resposta.
- Logging/observabilidade.

## Mapeamento
- TV Box: coleta + banco local + Telegram.
- Agent: ingestão + índice + NLP em LAN.
- Separação por componente reduz carga na TV Box.
