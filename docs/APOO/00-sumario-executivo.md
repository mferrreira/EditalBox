# 00. Sumário Executivo

## Visão geral
Sistema embarcado + agente local para coletar, manter e consultar editais do IFNMG Campus Januária com bot Telegram, SQLite e processamento NLP local.

## Objetivos principais
- Atualizar base local de editais sem depender de APIs pagas.
- Permitir consulta textual e por linguagem natural.
- Manter baixo consumo na TV Box e abstrair NLP pesado para agente local.

## Atores principais
- Usuário Telegram
- Administrador/Operador de TV Box
- Serviço de coleta (TV Box)
- Serviço auxiliar (agent)

## Escopo
Inclui coleta/parser genérico, deduplicação incremental, SQLite local, bot Telegram, busca textual, indexação local e consulta NLP via Ollama local. Não inclui painel web administrativo extenso.

## Marcos
- TV Box em Go: scraping, extração, jobs de sync.
- Agent em Python: ingest, indexação, RAG local, NLP.
- Telegram como interface conversacional principal.

## Decisões relevantes
- Dados principais locais na TV Box.
- NLP pesado isolado em agente na LAN.
- Sem API paga obrigatória.
