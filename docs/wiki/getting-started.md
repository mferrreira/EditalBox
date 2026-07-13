# Getting Started

## Prerequisites

- **tvbox (Go):** toolchain Go instalado; `go.mod` em `tvbox/`. Sem cgo (driver SQLite `modernc.org/sqlite`).
- **agent (Python):** Python 3.10+; dependências são apenas stdlib (sem `requirements.txt` — `urllib`, `sqlite3`, `http.server`).
- **Opcional:** Ollama rodando localmente (default `http://127.0.0.1:11434`) com o modelo `qwen2.5:7b-instruct` para respostas em LN. Sem ele, o agente usa fallback determinístico.

## Installation

```bash
# clone
git clone git@github.com:mferrreira/EditalBox.git
cd EditalBox

# tvbox
cd tvbox && go build ./... && cd ..

# agent (sem dependências externas; só sobe o módulo)
# não há pip install necessário
```

## First Run (desenvolvimento local)

Em dois terminais:

```bash
# terminal 1 — agente auxiliar
chmod +x start-mac.sh      # ou start-linux.sh / start-windows.ps1
./start-mac.sh

# terminal 2 — tvbox (coleta + telegram)
chmod +x start-tvbox.sh
./start-tvbox.sh
```

O agente escuta em `http://127.0.0.1:8090`; o tvbox em `:8080`. A primeira sincronização roda no startup do tvbox.

## Common Workflows

### Forçar sincronização
```bash
curl -X POST http://127.0.0.1:8080/sync
```

### Checar saúde
```bash
curl http://127.0.0.1:8080/health
curl http://127.0.0.1:8090/health
```

### Comandos no Telegram
- `/status` — total de editais, última sincronização e estado do Ollama.
- `/sync` — dispara sincronização imediata.
- `/recent` — lista os 5 editais mais recentes.
- Qualquer outro texto — consulta em linguagem natural.

## Configuration

Variáveis de ambiente (prefixo `EDITALBOX_` no tvbox, `EDITALBOX_AGENT_` no agente). Principais:

- `EDITALBOX_HTTP_ADDR` (`:8080`), `EDITALBOX_DB_PATH` (`./data/editalbox.db`), `EDITALBOX_SYNC_INTERVAL` (`24h`).
- `EDITALBOX_TELEGRAM_TOKEN`, `EDITALBOX_TELEGRAM_ALLOWED_CHAT_IDS` (CSV; vazio = todos).
- `EDITALBOX_AGENT_BASE_URL` (`http://127.0.0.1:8090`), `EDITALBOX_AGENT_TIMEOUT` (`20s`).
- `EDITALBOX_AGENT_OLLAMA_URL` (`http://127.0.0.1:11434`), `EDITALBOX_AGENT_OLLAMA_MODEL` (`qwen2.5:7b-instruct`).

## Where to Go Next

- Arquitetura: [architecture.md](architecture.md)
- Mapa de módulos: [README.md#module-map](README.md#module-map)
- Documentação existente do projeto: `docs/architecture.md`, `docs/startup.md`
