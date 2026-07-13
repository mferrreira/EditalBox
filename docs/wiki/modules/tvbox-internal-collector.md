# Module: `tvbox/internal/collector`

Coletor dirigido de páginas do IFNMG. Faz o fetch HTTP, extrai título/conteúdo/datas/documentos e normaliza tudo para um `domain.Notice` canônico. É deliberadamente tolerante a mudanças de HTML — não usa seletores fixos.

## Responsibilities

- Buscar uma URL (timeout 20s) e limitar o body a 2 MiB.
- Extrair título (meta og:title/twitter:title → h1 → h2 → `<title>`, rejeitando títulos genéricos do portal).
- Normalizar texto (remove tags, NBSP, colapsa espaços) e truncar `BodyText` em 12k runas.
- Detectar eventos de datas (inscrição, resultado, matrícula, recurso, etc.) em dois formatos PT-BR.
- Classificar `SourceType` (`news`, `document`, `staff_selection`, `student_selection`, `index`) e `Status` derivado.
- Extrair links seguíveis conforme `followRules` e documentos oficiais `documento.ifnmg.edu.br`.

## Key Files

- [`collector.go`](../../tvbox/internal/collector/collector.go) — `Collector`, `Fetch`, `PageResult`, funções de extração/normalização/derivação de datas e status.
- [`collector_test.go`](../../tvbox/internal/collector/collector_test.go) — testes do coletor.

## Public API

- `func New(userAgent string, followRules []string) *Collector`
- `func (c *Collector) Fetch(ctx context.Context, rawURL string) (PageResult, error)`

`PageResult` carrega `Notice`, `Documents`, `Events` e `Links`.

## Internal Structure

`Fetch` é um pipeline puro: request → `extractTitle` → `normalizeText`/`stripTags` → `extractEvents` → `canonicalize` → `deriveImportantDates` → monta `Notice` → filtra `Documents`/`Links`. Funções utilitárias (`canonicalize`, `classifySource`, `deriveStatus`, `parseDate`, `monthFromPT`, `isGenericTitle`) são independentes e testáveis.

## Dependencies

- **Used by:** `internal/app`.
- **Uses:** `internal/domain`.

## Notable Patterns / Gotchas

- `classifySource` usa substring da URL (ex.: `/professor-substituto`), não conteúdo — rápido e estável.
- `deriveStatus` é a lógica de negócio mais sensível: combina `registrationEnd`/`finalEventAt` com palavras-chave (`encerrad`, `resultado`, `matricula`).
- Datas em PT-BR reconhecidas em formato `dd/mm/aaaa` e `dd de mês de aaaa`, incluindo intervalos ("inscrições de X a Y").
- `extractLinks` resolve URLs relativas contra a base e canôniza antes de retornar.
