# 04. Catálogo de Casos de Uso

## CU-001 | Coletar/importar edital
- Atores: Serviço de coleta; Administrador.
- Objetivo: Trazer editais novos/alterados.
- Pré-condições: seeds configuradas e site acessível.
- Pós-condições: notice/doc/events atualizados.
- Partes interessadas: sistema, operação.

## CU-002 | Deduplicar/atualizar base <<CRUD>>
- Atores: Serviço de coleta.
- Objetivo: Evitar duplicidade e manter delta.
- Pré-condições: notices anteriores existem.
- Pós-condições: base coerente.
- Partes interessadas: sistema.

## CU-003 | Marcar status
- Atores: Serviço de coleta.
- Objetivo: Refletir disponibilidade/finalização.
- Pré-condições: notice identificada.
- Pós-condições: status atualizado.
- Partes interessadas: usuários.

## CU-004 | Consultar textualmente
- Atores: Usuário Telegram; Sistema.
- Objetivo: Buscar por título/link/palavras-chave.
- Pré-condições: query disponível.
- Pós-condições: resultados ranqueados.
- Partes interessadas: usuários.

## CU-005 | Consultar por NLP/local agent
- Atores: Usuário Telegram; Serviço auxiliar.
- Objetivo: Responder em linguagem natural.
- Pré-condições: índice atualizado.
- Pós-condições: resposta gerada com justificativa/fonte.
- Partes interessadas: usuários.

## CU-006 | Manter indexação do agent <<CRUD>>
- Atores: Serviço auxiliar; Serviço de coleta.
- Objetivo: Indexar chunks e atualizar recall.
- Pré-condições: notice/chunk prontos.
- Pós-condições: índice atualizado.
- Partes interessadas: sistema.

## CU-007 | Telegram bot <<rep>>
- Atores: Usuário Telegram; Sistema.
- Objetivo: Interface conversacional.
- Pré-condições: token/config válidos.
- Pós-condições: interações recebidas/respondidas.
- Partes interessadas: usuários.

## CU-008 | Health/base status <<rep>>
- Atores: Administrador; Usuário Telegram.
- Objetivo: Informar estado da base.
- Pré-condições: sync/jobs monitorados.
- Pós-condições: contadores reportados.
- Partes interessadas: administração.

## CU-009 | Configuração operacional <<CRUD>>
- Atores: Administrador.
- Objetivo: Ajustar seeds/modelos/tokens.
- Pré-condições: acesso administrativo.
- Pós-condições: configuração aplicada.
- Partes interessadas: administração.

## CU-010 | Logging/auditoria <<rep>>
- Atores: Serviços; Administrador.
- Objetivo: Registrar sync/consulta/mensagens.
- Pré-condições: eventos relevantes ocorrendo.
- Pós-condições: logs estruturados emitidos.
- Partes interessadas: administração.
