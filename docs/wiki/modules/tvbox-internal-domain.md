# Module: `tvbox/internal/domain`

Modelos de domínio compartilhados entre os pacotes do `tvbox`. São structs simples de dados (DTOs) — sem comportamento.

## Responsibilities

- Definir as entidades centrais do sistema: `Notice`, `NoticeDocument`, `NoticeEvent`, `SyncRun`, `SessionSummary`, `CandidateAnswer`.

## Key Files

- [`models.go`](../../tvbox/internal/domain/models.go) — todas as structs de domínio.

## Public API

- `type Notice struct` — campos: `ID`, `SourceURL`, `CanonicalURL`, `Title`, `Excerpt`, `BodyText`, `SourceType`, `Status`, `RegistrationEnd *time.Time`, `FinalEventAt *time.Time`, `UpdatedAt`, `CreatedAt`.
- `type NoticeDocument struct` — `ID`, `NoticeID`, `URL`, `Kind`, `Title`.
- `type NoticeEvent struct` — `ID`, `NoticeID`, `EventType`, `Label`, `EventAt`.
- `type SyncRun struct` — contadores de uma sincronização.
- `type CandidateAnswer struct` — representação JSON enviada ao agente (com `Keywords []string`).

## Internal Structure

Structs planas; ponteiros de tempo (`*time.Time`) sinalizam "sem data". `CandidateAnswer` traz `json:` tags porque é serializada para o agente auxiliar.

## Dependencies

- **Used by:** `collector`, `storage`, `app`, `agent`.
- **Uses:** apenas `time`.

## Notable Patterns / Gotchas

- `CandidateAnswer` duplica campos de `Notice` com nomes JSON-friendly — é o contrato de API com o agente, não um ORM.
