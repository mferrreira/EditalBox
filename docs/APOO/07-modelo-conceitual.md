# 07. Modelo Conceitual

## Entidades principais
- Edital/Notícia
- Documento do edital
- Evento do edital
- Fonte/Coleta
- Sessão Telegram
- Mensagem Telegram
- Índice local
- Configuração Operacional
- Sync/Job
- Estado da Base

## Relacionamentos
- Edital `1:N` Documento/Evento.
- Fonte alimenta coleta.
- Sync/Job atualiza Edital + Índice local.
- Sessão agrupa Mensagens.
- Índice local referencia chunks/runs.

## Regras conceituais relevantes
- Base principal local.
- Deduplicação por URL/conteúdo.
- Consulta textual/NLP sem API paga obrigatória.
