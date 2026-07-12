# 12. Código e Testes

## Código relevante
- `tvbox/internal/domain/models.go`: entidades de domínio.
- `tvbox/internal/collector/collector.go`: coleta e seeds.
- `tvbox/internal/app/app.go`: orquestração da consulta inteligente.
- `tvbox/internal/telegram/bot.go`: bot Telegram.
- `agent/src/agent/main.py`: serviço auxiliar.
- `agent/src/agent/ollama.py`: integração com modelo local.
- `agent/src/agent/service.py`: ingestão/indexação/ranqueamento.

## Testes existentes
- `agent/tests/test_service.py`

## Cobertura de qualidade atual
- Testes pontuais em agent; considerar expandir para tvbox/sync/telegram.

## Decisões técnicas relevantes
- Separação forte TV Box/agent.
- SQLite como store principal local.
- Ollama local sem API paga.
