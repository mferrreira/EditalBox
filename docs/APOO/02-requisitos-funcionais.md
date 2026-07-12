# 02. Requisitos Funcionais

## RF-01 | Coleta/parse genérico de editais
- Descrição: extrair editais das seeds configuradas sem dependência de seletor único.
- Atores: Serviço de coleta; Administrador.
- Prioridade: Alta.
- Critérios de aceitação: URLs/títulos extraídos e normalizados.

## RF-02 | Armazenar notices/docs/events localmente
- Descrição: persistir na TV Box.
- Atores: Serviço de coleta; Administrador.
- Prioridade: Alta.
- Critérios de aceitação: base local atualizada e auditável.

## RF-03 | Deduplicação e atualização incremental
- Descrição: evitar duplicatas e sincronizar delta.
- Atores: Serviço de coleta; Administrador.
- Prioridade: Alta.
- Critérios de aceitação: re-coleta não duplica notices conhecidas.

## RF-04 | Marcar status/indisponibilidade
- Descrição: marcar editais indisponíveis/encerrados/finalizados.
- Atores: Serviço de coleta.
- Prioridade: Média.
- Critérios de aceitação: status refletido nas respostas.

## RF-05 | Bot do Telegram
- Descrição: interface conversacional para consultas.
- Atores: Usuário Telegram; Sistema.
- Prioridade: Alta.
- Critérios de aceitação: bot responde comandos e perguntas.

## RF-06 | Busca textual local
- Descrição: consultar por comandos, palavras-chave, título/link.
- Atores: Usuário Telegram; Sistema.
- Prioridade: Alta.
- Critérios de aceitação: ranking simples funcional.

## RF-07 | Consulta NLP local via agent
- Descrição: responder em linguagem natural com base em fontes locais.
- Atores: Usuário Telegram; Serviço auxiliar.
- Prioridade: Alta.
- Critérios de aceitação: resposta citável, grounded.

## RF-08 | Ingestão/indexação do agent <<CRUD>>
- Descrição: indexar chunks e atualizar índice local.
- Atores: Serviço auxiliar; Serviço de coleta.
- Prioridade: Alta.
- Critérios de aceitação: consultas retornam contexto relevante.

## RF-09 | Health/base status
- Descrição: informar estado da base local.
- Atores: Administrador; Usuário Telegram.
- Prioridade: Média.
- Critérios de aceitação: contadores/status disponíveis.

## RF-10 | Configuração operacional
- Descrição: parametrizar seeds, modelos/tokens e limites.
- Atores: Administrador.
- Prioridade: Média.
- Critérios de aceitação: deploy sem hardcodes.

## RF-11 | Observabilidade/Logs <<rep>>
- Descrição: registrar jobs/sync/mensagens.
- Atores: Administrador; Serviços.
- Prioridade: Média.
- Critérios de aceitação: fluxo auditorável.
