# Class Diagram

## Core Types (tvbox, Go)

![Classes tvbox](images/classes_tvbox.png)

```mermaid
classDiagram
    class App {
        +Config cfg
        +Store store
        +Collector collector
        +Client agent
        +Bot bot
        +syncNow(ctx) error
        +answerQuery(ctx, chatID, question) string
    }
    class Collector {
        +http.Client client
        +string userAgent
        +[]string followRules
        +Fetch(ctx, rawURL) PageResult
    }
    class Store {
        +*sql.DB db
        +UpsertNotice(ctx, Notice) int64,bool,error
        +SearchNotices(ctx, query, limit) []Notice
    }
    class Bot {
        +string token
        +time.Duration pollTimeout
        +GetUpdates(ctx, offset) []Update
        +SendMessage(ctx, chatID, text) error
    }
    class Client {
        +string baseURL
        +Health(ctx) Health
        +Ingest(ctx, []IngestNotice) error
        +Answer(ctx, AnswerRequest) AnswerResult
    }
    class Notice {
        +int64 ID
        +string CanonicalURL
        +string Title
        +string Status
        +*time.Time RegistrationEnd
        +*time.Time FinalEventAt
    }
    class Config {
        +string HTTPAddr
        +string DBPath
        +time.Duration SyncInterval
        +string AgentBaseURL
    }
    App --> Collector : uses
    App --> Store : uses
    App --> Bot : uses
    App --> Client : uses
    App --> Config : holds
    Collector --> Notice : builds
    Store --> Notice : persists
```

## Core Types (agent, Python)

![Classes agent](images/classes_agent.png)

```mermaid
classDiagram
    class AgentService {
        +Store store
        +OllamaClient ollama
        +ingest(payload) tuple
        +answer(question, session_summary, candidates, limit) dict
    }
    class OllamaClient {
        +string base_url
        +string model
        +health() bool
        +generate(prompt) str
    }
    class Store {
        +upsert_notices(notices) tuple
        +search(query, limit) list
        +indexed_count() int
    }
    class Config {
        +host
        +port
        +ollama_url
        +ollama_model
    }
    AgentService --> Store : uses
    AgentService --> OllamaClient : uses
    OllamaClient --> Config : reads
    AgentService ..> IndexedNotice : builds
```

## Notes

- O tvbox e o agent são processos separados; a única relação real é HTTP (`Client` → rotas do agente). Os diagramas acima mostram os grafos internos de cada lado.
- `Notice`/`IndexedNotice` são DTOs quase idênticos — o tvbox os serializa para o agente via `CandidateAnswer`/`IngestNotice`.
- Não há herança relevante; o estilo é composição + structs planas (Go) e dataclasses (Python).
