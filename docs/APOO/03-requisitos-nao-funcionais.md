# 03. Requisitos Não Funcionais

## Requisitos não funcionais e suplementares
- **Usabilidade:** bot Telegram com respostas curtas e justificativas quando possível.
- **Confiabilidade:** base local mantém integridade com sync incremental.
- **Desempenho:** TV Box deve operar com baixo consumo; queries simples rápidas.
- **Segurança:** tokens/chaves armazenados fora de código; telegram sem vazamento de sessão.
- **Disponibilidade local:** stack tolerante a quedas momentâneas do portal.
- **Manutenibilidade:** componentes separados (TV Box vs agent).
- **Observabilidade:** logs de sync/consulta para operação.
