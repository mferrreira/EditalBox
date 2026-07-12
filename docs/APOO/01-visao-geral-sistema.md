# 01. Visão Geral e Escopo do Sistema

## Problema
Editais ficam dispersos no portal e sem base local estruturada, dificultando consulta confiável e duplicidade/integridade da informação.

## Solução proposta
Sistema com dois componentes: TV Box (coleta/armazenamento/consulta leve) e agent local (interpretação e NLP), com interface Telegram.

## Atores envolvidos
- Usuário Telegram
- Administrador/Operador de TV Box
- Serviço de coleta (TV Box)
- Serviço auxiliar (agent)

## Contorno do sistema
Inclui scraping/parser genérico, extração de links, deduplicação incremental, SQLite, bot Telegram, busca textual, indexação e RAG local.

## Restrições devem ser respeitadas
- Baixo consumo na TV Box.
- Sem API paga obrigatória.
- Tolerar mudanças leves no HTML do portal.
- Dados preferencialmente locais.
