# 05. Casos de Uso Expandidos

## CU-001 | Coletar/importar edital

### Atores
- Serviço de coleta (TV Box)
- Administrador

### Interessados
- TV Box, usuário final, operação.

### Pré-condições
- Seeds configuradas; site acessível.

### Pós-condições
- Notice/doc/events atualizados.

### Fluxo principal de eventos
1. `[IN]` Serviço acessa seeds.
2. `[OUT]` Sistema normaliza links/conteúdos.
3. `[IN]` Sistema aplica deduplicação.
4. `[OUT]` Sistema persiste notices/docs/events.

### Fluxos alternativos e variantes
- 2a. HTML alterado: parser genérico compensa e registra warning.

### Regras de negócio
- Manter dados principais locais.
- Registrar source_url e datas relevantes.

### Requisitos correlacionados
- RF-01, RF-02.

### Questões em aberto
- Confirmar tolerância minima a mudanças HTML.

---

## CU-005 | Consultar por NLP/local agent

### Atores
- Usuário Telegram
- Serviço auxiliar

### Interessados
- Usuário final, curadoria.

### Pré-condições
- Índice atualizado; Ollama/agent disponíveis.

### Pós-condições
- Resposta gerada com justificativa e source.

### Fluxo principal de eventos
1. `[IN]` Usuário envia pergunta.
2. `[OUT]` Sistema recebe contexto do índice local.
3. `[IN]` Agent consulta RAG/local LLM.
4. `[OUT]` Sistema retorna resposta e referência.

### Fluxos alternativos e variantes
- 3a. índice insuficiente: sistema retorna busca textual simples.

### Regras de negócio
- Resposta deve ser grounded em fontes locais.
- Sem API paga obrigatória.

### Requisitos correlacionados
- RF-05, RF-06, RF-07.

### Questões em aberto
- Confirmar limiar de confiança para grounding.
