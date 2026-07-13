# Sequence Diagrams

## Workflow: Sincronização periódica (tvbox)

![Sync sequence](images/seq_sync.png)

```mermaid
sequenceDiagram
    participant Sched as Scheduler
    participant App
    participant Coll as Collector
    participant Store as SQLite (tvbox)
    participant Ag as Agent Client
    participant Agent as Agent HTTP
    Sched->>App: syncNow (a cada 24h ou POST /sync)
    loop BFS até 60 páginas
        App->>Coll: Fetch(url)
        Coll->>Coll: extrai título/datas/docs/status
        Coll-->>App: PageResult
        App->>Store: UpsertNotice / UpsertDocument / UpsertEvent
    end
    App->>Store: NoticesUpdatedSince(janela)
    App->>Ag: Ingest(editais alterados)
    Ag->>Agent: POST /v1/ingest
    Agent-->>Ag: {indexed, chunks}
    App->>Store: FinishSyncRun(success)
```

## Workflow: Consulta em linguagem natural (via Telegram)

![Query sequence](images/seq_query.png)

```mermaid
sequenceDiagram
    participant User
    participant TG as Telegram Bot
    participant App
    participant Store as SQLite (tvbox)
    participant Ag as Agent Client
    participant Agent as Agent Service
    participant Oll as Ollama
    User->>TG: mensagem (pergunta)
    TG->>App: handleTelegramMessage
    App->>Store: SaveMessage(user)
    App->>App: buildCandidatePool (SearchNotices + RecentNotices)
    App->>Ag: Answer(question, session_summary, candidates)
    Ag->>Agent: POST /v1/answer
    Agent->>Agent: _rank (sinônimos + intent_bonus)
    alt Ollama pronto
        Agent->>Oll: generate(prompt grounding-only)
        Oll-->>Agent: texto LN
    else fallback
        Agent-->>Ag: texto determinístico
    end
    Agent-->>Ag: {text, structured, used_ollama}
    Ag-->>App: AnswerResult
    App->>Store: SaveMessage(assistant)
    App->>TG: SendMessage(resposta)
    TG-->>User: resposta
```

## Walkthrough (consulta)

1. **Entrada** — [`tvbox/internal/telegram/bot.go`](../../tvbox/internal/telegram/bot.go) faz `getUpdates`; [`app.go:runTelegram`](../../tvbox/internal/app/app.go) consome.
2. **Roteamento** — [`app.go:handleTelegramMessage`](../../tvbox/internal/app/app.go) decide comando vs. consulta; consultas caem em `answerQuery`.
3. **Candidatos** — [`app.go:buildCandidatePool`](../../tvbox/internal/app/app.go) mistura busca textual (`SearchNotices`) e recentes (`RecentNotices`).
4. **Resposta** — [`agent/src/agent/service.py:AgentService.answer`](../../agent/src/agent/service.py) ranqueia e (se houver Ollama) gera texto com `_prompt` restrito aos candidatos.

## Notes

- Todo fallback de `answerQuery` devolve `formatNotices` (listagem textual) se o agente não retornar `Structured`.
- O push de ingestão é assíncrono em relação à consulta: a sincronização roda no scheduler e empurra delta; a consulta só lê o que já está indexado.
