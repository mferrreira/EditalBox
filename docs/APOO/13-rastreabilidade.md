# 13. Rastreabilidade

## Requisitos funcionais -> Casos de uso
- RF-01 -> CU-001
- RF-02 -> CU-001
- RF-03 -> CU-002
- RF-04 -> CU-003
- RF-05 -> CU-007
- RF-06 -> CU-004
- RF-07 -> CU-005
- RF-08 -> CU-006
- RF-09 -> CU-008
- RF-10 -> CU-009
- RF-11 -> CU-010

## Casos de uso -> Modelo conceitual
- CU-001/002/003 -> Edital, Documento, Evento, Sync/Job
- CU-004/005 -> Índice local, Edital, Agent
- CU-006 -> Índice local, Agent
- CU-007 -> Sessão/Mensagem Telegram
- CU-008 -> Estado da Base, Sync/Job
- CU-009 -> Configuração Operacional
- CU-010 -> Log/Telemetria

## Casos de uso -> Contratos OCL
- CU-001/002/003 -> invariantes notice/url/duplicidade/status
- CU-004/005/006 -> invariantes índice/chunk/grounding
- CU-007 -> invariante sessão/bot
- CU-008 -> invariante health/counters
- CU-010 -> invariante logging

## Casos de uso -> DCP
- CU-001/002/003 -> coleta/parser/tvbox
- CU-004/005 -> bot/busca/agent
- CU-006/007/008/009 -> tvbox/agent/config
- CU-010 -> ambos componentes

## Casos de uso -> Modelo de dados
- CU-001/002/003 -> notices, notice_documents, notice_events, sync_runs
- CU-004/005 -> indexed_notices, indexed_chunks
- CU-006 -> agent/indexação
- CU-007 -> telegram_sessions, telegram_messages
- CU-008 -> estado base
- CU-009 -> settings
- CU-010 -> logs
