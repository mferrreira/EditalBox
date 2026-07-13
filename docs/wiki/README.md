# EditalBox

Sistema embarcado para coleta, manutenção e consulta inteligente de editais publicados no portal do IFNMG Campus Januária. Roda como um nó principal em uma TV Box com Armbian (baixo consumo de CPU/memória) e delega o processamento pesado de linguagem natural a um serviço auxiliar na rede local. Evita dependências de APIs pagas: a "inteligência" vem de um modelo local via Ollama.

## Key Concepts

- **Coleta dirigida e tolerante a HTML** — o coletor não depende de um seletor CSS fixo; extrai links e conteúdo de forma genérica e normaliza para uma URL canônica.
- **Divisão de carga** — a TV Box (Go) faz coleta, persistência e busca textual; o agente (Python) faz indexação por chunks e geração de resposta em linguagem natural com Ollama.
- **Status derivado** — cada edital recebe um status (`open`, `registration_closed`, `in_progress`, `finalized`, `unknown`) calculado a partir de datas extraídas do texto e palavras-chave.
- **Consulta híbrida** — uma pergunta em linguagem natural é ranqueada por sobreposição de tokens (expandidos por sinônimos) e, se o Ollama estiver disponível, reformulada em linguagem natural com base apenas nos candidatos locais.

## Entry Points

- [`tvbox/cmd/editalbox/main.go`](tvbox/cmd/editalbox/main.go) — ponto de entrada do serviço principal (TV Box). Carrega config, cria o `App` e fica ouvindo sinais de término.
- [`agent/src/agent/main.py`](agent/src/agent/main.py) — ponto de entrada do serviço auxiliar (agente). Sobe um `ThreadingHTTPServer` com rotas `/health`, `/v1/ingest`, `/v1/answer`.

## High-Level Architecture

O sistema é um monorepo com dois componentes que se falam por HTTP na LAN. O componente `tvbox/` (Go) é o nó de coleta e armazenamento: varre páginas do IFNMG, extrai editais, persiste em SQLite local e responde consultas simples (comandos Telegram + busca textual). Periodicamente ele empurra os editais alterados para o componente `agent/` (Python), que indexa por chunks e serve busca contextual + geração de resposta em LN via Ollama. Detalhe em [architecture.md](architecture.md).

## Module Map

| Module | Purpose |
|---|---|
| [`tvbox/cmd/editalbox`](modules/tvbox-cmd-editalbox.md) | Bootstrap do serviço principal: carrega config e sobe o `App`. |
| [`tvbox/internal/app`](modules/tvbox-internal-app.md) | Orquestração: scheduler de sync, servidor HTTP, polling do Telegram e roteamento de consultas. |
| [`tvbox/internal/collector`](modules/tvbox-internal-collector.md) | Coleta dirigida de páginas, extração de links/conteúdo e derivação de status/datas. |
| [`tvbox/internal/storage`](modules/tvbox-internal-storage.md) | Persistência SQLite: upsert de editais/documentos/eventos, busca textual ranqueada e sessões de Telegram. |
| [`tvbox/internal/domain`](modules/tvbox-internal-domain.md) | Modelos de domínio (Notice, NoticeDocument, NoticeEvent, SyncRun, CandidateAnswer). |
| [`tvbox/internal/telegram`](modules/tvbox-internal-telegram.md) | Cliente do Bot API do Telegram com polling e retrying com backoff exponencial. |
| [`tvbox/internal/config`](modules/tvbox-internal-config.md) | Carga de configuração a partir de variáveis de ambiente com defaults sensíveis. |
| [`tvbox/internal/agent`](modules/tvbox-internal-agent.md) | Cliente HTTP do tvbox para o serviço auxiliar (ingest, answer, health). |
| [`agent/src/agent/main`](modules/agent-src-agent-main.md) | Bootstrap do agente: servidor HTTP das rotas `/v1/ingest` e `/v1/answer`. |
| [`agent/src/agent/service`](modules/agent-src-agent-service.md) | Núcleo do agente: ingestão, ranqueamento por sinônimos e geração de resposta. |
| [`agent/src/agent/ollama`](modules/agent-src-agent-ollama.md) | Cliente do Ollama (health check + generate). |
| [`agent/src/agent/store`](modules/agent-src-agent-store.md) | Persistência do índice local em SQLite (notices + chunks com keywords). |

## Getting Started

Veja [getting-started.md](getting-started.md).
