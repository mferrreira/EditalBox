#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TVBOX_DIR="$ROOT_DIR/tvbox"
ENV_FILE="$TVBOX_DIR/.env"
ENV_EXAMPLE="$TVBOX_DIR/.env.example"

info() {
  printf '\033[1;34m[tvbox]\033[0m %s\n' "$*"
}

warn() {
  printf '\033[1;33m[tvbox]\033[0m %s\n' "$*"
}

fail() {
  printf '\033[1;31m[tvbox]\033[0m %s\n' "$*" >&2
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

prompt_agent_url_if_localhost() {
  local current
  current="$(get_env_value "EDITALBOX_AGENT_BASE_URL" || true)"
  if [[ -z "${current// }" || "$current" == "http://127.0.0.1:8090" || "$current" == "http://localhost:8090" ]]; then
    warn "Para producao na TV Box, use o IP do MacBook na LAN. Para desenvolvimento local neste Mac, localhost tambem pode ser usado."
    read -r -p "URL do agent (ex.: http://192.168.0.10:8090 ou localhost:8090): " current
    current="$(normalize_agent_url "$current")"
    validate_agent_url "$current"
    set_env_value "EDITALBOX_AGENT_BASE_URL" "$current"
  fi
}

normalize_agent_url() {
  local value="$1"
  value="${value// /}"
  if [[ -z "$value" ]]; then
    printf '%s' "$value"
    return
  fi
  if [[ ! "$value" =~ ^https?:// ]]; then
    value="http://$value"
  fi
  printf '%s' "$value"
}

validate_agent_url() {
  local value="$1"
  [[ "$value" =~ ^https?://[^/]+:[0-9]+$ ]] || fail "URL do agent invalida. Use o formato http://HOST:PORTA"
}

check_agent_health() {
  local url="$1"
  if command -v curl >/dev/null 2>&1; then
    if curl -fsS "$url/health" >/dev/null 2>&1; then
      info "Agent respondeu em $url"
    else
      warn "Nao consegui validar o agent em $url/health agora. O servico Go ainda sera iniciado."
    fi
  fi
}

print_tutorial() {
  cat <<EOF

EditalBox TV Box

O que este script faz:
- cria tvbox/.env se faltar
- pede o token do Telegram se estiver ausente
- pede a URL do agent no MacBook se a configuracao estiver apontando para localhost
- valida os binarios locais
- inicia o servico principal em Go

Links uteis:
- Tutorial completo: $ROOT_DIR/docs/startup.md
- Criar bot no Telegram: https://t.me/BotFather
- Documentacao Telegram Bots: https://core.telegram.org/bots

EOF
}

main() {
  require_cmd go

  copy_env_if_missing
  prompt_if_empty "EDITALBOX_TELEGRAM_TOKEN" "Token do bot Telegram"
  prompt_if_empty "EDITALBOX_TELEGRAM_ALLOWED_CHAT_IDS" "Chat ID permitido no Telegram"
  prompt_agent_url_if_localhost
  load_env
  validate_agent_url "${EDITALBOX_AGENT_BASE_URL}"

  check_agent_health "${EDITALBOX_AGENT_BASE_URL}"
  print_tutorial

  info "Iniciando servico Go em ${EDITALBOX_HTTP_ADDR}"
  cd "$TVBOX_DIR"
  go run ./cmd/editalbox
}

main "$@"
