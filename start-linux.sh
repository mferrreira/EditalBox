#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
AGENT_DIR="$ROOT_DIR/agent"
ENV_FILE="$AGENT_DIR/.env"
ENV_EXAMPLE="$AGENT_DIR/.env.example"
LOG_DIR="$AGENT_DIR/logs"

mkdir -p "$LOG_DIR"

info() {
  printf '\033[1;34m[linux]\033[0m %s\n' "$*"
}

warn() {
  printf '\033[1;33m[linux]\033[0m %s\n' "$*"
}

fail() {
  printf '\033[1;31m[linux]\033[0m %s\n' "$*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "Comando obrigatorio ausente: $1"
}

copy_env_if_missing() {
  if [[ ! -f "$ENV_FILE" ]]; then
    cp "$ENV_EXAMPLE" "$ENV_FILE"
    info "Arquivo $ENV_FILE criado a partir de $ENV_EXAMPLE"
  fi
}

load_env() {
  set -a
  # shellcheck disable=SC1090
  . "$ENV_FILE"
  set +a
}

get_env_value() {
  local key="$1"
  awk -F= -v k="$key" '$1 == k { sub(/^[^=]*=/, ""); print; exit }' "$ENV_FILE"
}

set_env_value() {
  local key="$1"
  local value="$2"
  local tmp
  tmp="$(mktemp)"
  awk -v k="$key" -v v="$value" '
    BEGIN { done=0 }
    index($0, k "=") == 1 {
      print k "=" v
      done=1
      next
    }
    { print }
    END {
      if (!done) print k "=" v
    }
  ' "$ENV_FILE" > "$tmp"
  mv "$tmp" "$ENV_FILE"
}

prompt_if_empty() {
  local key="$1"
  local label="$2"
  local current
  current="$(get_env_value "$key" || true)"
  if [[ -z "${current// }" ]]; then
    read -r -p "$label: " current
    [[ -n "${current// }" ]] || fail "Valor obrigatorio nao informado para $key"
    set_env_value "$key" "$current"
  fi
}

ensure_ollama_installed() {
  if command -v ollama >/dev/null 2>&1; then
    info "Ollama encontrado em $(command -v ollama)"
    return
  fi

  warn "Ollama nao encontrado. Tentando instalar com o instalador oficial."
  require_cmd curl
  curl -fsSL https://ollama.com/install.sh | sh
  command -v ollama >/dev/null 2>&1 || fail "Falha ao instalar o Ollama."
}

ensure_ollama_running() {
  local api_url="$1"
  local base_url="${api_url%/api*}"
  if curl -fsS "$base_url/api/tags" >/dev/null 2>&1; then
    info "Ollama API ja esta respondendo em $base_url"
    return
  fi

  warn "Ollama nao esta respondendo em $base_url. Iniciando 'ollama serve'."
  nohup ollama serve > "$LOG_DIR/ollama.log" 2>&1 &

  local retries=20
  while (( retries > 0 )); do
    if curl -fsS "$base_url/api/tags" >/dev/null 2>&1; then
      info "Ollama inicializado com sucesso."
      return
    fi
    sleep 1
    retries=$((retries - 1))
  done

  fail "Ollama nao respondeu apos a inicializacao. Veja $LOG_DIR/ollama.log"
}

ensure_model_installed() {
  local model="$1"
  if ollama show "$model" >/dev/null 2>&1; then
    info "Modelo $model ja instalado."
    return
  fi

  warn "Modelo $model nao encontrado. Executando 'ollama pull $model'."
  ollama pull "$model"
}

print_tutorial() {
  cat <<EOF

EditalBox Linux

O que este script faz:
- cria agent/.env se faltar
- carrega as variaveis de ambiente
- instala o Ollama se necessario
- inicia o Ollama se ele nao estiver rodando
- baixa o modelo configurado se faltar
- inicia o agent Python local

Links uteis:
- Tutorial completo: $ROOT_DIR/docs/startup.md
- Instalacao do Ollama: https://ollama.com/download/linux
- Biblioteca de modelos: https://ollama.com/library

Logs:
- Ollama: $LOG_DIR/ollama.log

EOF
}

main() {
  require_cmd python3
  require_cmd curl

  copy_env_if_missing
  prompt_if_empty "EDITALBOX_AGENT_OLLAMA_MODEL" "Modelo Ollama desejado"
  load_env

  ensure_ollama_installed
  ensure_ollama_running "${EDITALBOX_AGENT_OLLAMA_URL}"
  ensure_model_installed "${EDITALBOX_AGENT_OLLAMA_MODEL}"

  print_tutorial
  info "Iniciando agent em http://${EDITALBOX_AGENT_HOST}:${EDITALBOX_AGENT_PORT}"
  cd "$AGENT_DIR"
  PYTHONPATH="./src" python3 -m agent.main
}

main "$@"
