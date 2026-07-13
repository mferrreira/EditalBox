# Arquitetura Técnica

## Componentes

### TV Box (`tvbox/`, Go)

Responsabilidades:

- coleta dirigida de páginas do IFNMG;
- extração genérica de links candidatos;
- persistência local em SQLite;
- deduplicação por URL canônica e documento oficial;
- classificação inicial de relevância;
- bot do Telegram com polling;
- manutenção de sessão temporária por 24h;
- busca textual local;
- envio incremental dos itens novos ou alterados para o serviço auxiliar;
- fallback quando o serviço auxiliar estiver indisponível.

### Nó auxiliar (`agent/`, Python)

Responsabilidades:

- recepção incremental dos editais e documentos alterados;
- chunking e indexação local;
- ranqueamento lexical inicial;
- geração de resposta com modelo local via Ollama;
- consolidação de evidências e justificativas;
- healthcheck simples para a TV Box.

## Fontes de coleta priorizadas

Seeds principais:

- `https://www.ifnmg.edu.br/januaria`
- `https://www.ifnmg.edu.br/editais-ifnmg`
- `https://www.ifnmg.edu.br/mais-noticias-januaria`
- `https://www.ifnmg.edu.br/assistenciaestudantil-januaria/editais-assistenciaestudantil-januaria`
- `https://www.ifnmg.edu.br/extensao-januaria/editais`
- `https://www.ifnmg.edu.br/pesquisa-januaria/pesquisa/editais-pesquisa-januaria`
- `https://www.ifnmg.edu.br/processoseletivo`
- `https://www.ifnmg.edu.br/professor-substituto`

Padrões de seguimento:

- `/mais-noticias-januaria/`
- `/processoseletivo`
- `/professor-substituto/55-portal/januaria/`
- `/assistenciaestudantil-januaria/`
- `/extensao-januaria/`
- `/pesquisa-januaria/`
- `documento.ifnmg.edu.br/action.php?...fDocumentId=`

## Modelo de dados principal

### TV Box

- `notices`
- `notice_documents`
- `notice_events`
- `sync_runs`
- `telegram_sessions`
- `telegram_messages`

### Serviço auxiliar

- `indexed_notices`
- `indexed_chunks`
- `ingest_runs`

## Regra de validade

- Um edital sai das respostas padrão quando a data de inscrição expira.
- O edital continua armazenado enquanto houver eventos posteriores relevantes.
- O edital só é marcado como `finalizado` quando a última data terminal conhecida expira.
- A exclusão física não é o comportamento padrão.

## Estados do edital

- `open`
- `registration_closed`
- `in_progress`
- `finalized`
- `unavailable`
- `unknown`

## Sessões do Telegram

- contexto temporário por sessão;
- expiração após 24h sem interação;
- sem perfil persistente de usuário;
- contexto enviado ao serviço auxiliar junto da pergunta atual.

## Contrato entre TV Box e serviço auxiliar

### `GET /health`

Retorna:

- status do serviço;
- disponibilidade do modelo local;
- quantidade de itens indexados.

### `POST /v1/ingest`

Entrada:

- editais novos ou alterados;
- texto bruto essencial;
- documentos associados;
- datas e status derivados.

Saída:

- quantidade de editais indexados;
- quantidade de chunks gerados.

### `POST /v1/answer`

Entrada:

- pergunta atual;
- resumo da sessão;
- lista de candidatos selecionados pela TV Box;
- limite de resultados.

Saída:

- resposta textual;
- resultados estruturados;
- justificativa curta por item;
- indicador se houve uso de Ollama.
