#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TVBOX_DIR="$ROOT_DIR/tvbox"
ENV_FILE="$TVBOX_DIR/.env"
ENV_EXAMPLE="$TVBOX_DIR/.env.example"
SERVICE_SRC="$ROOT_DIR/deploy/systemd/editalbox-tvbox.service"

INSTALL_ROOT="/opt/editalbox/tvbox"
INSTALL_BIN="$INSTALL_ROOT/editalbox"
INSTALL_ENV="$INSTALL_ROOT/.env"
SERVICE_DST="/etc/systemd/system/editalbox-tvbox.service"

info() {
  printf '\033[1;34m[install-tvbox]\033[0m %s\n' "$*"
}

warn() {
  printf '\033[1;33m[install-tvbox]\033[0m %s\n' "$*"
}

fail() {
  printf '\033[1;31m[install-tvbox]\033[0m %s\n' "$*" >&2
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

prompt_agent_url() {
  local current
  current="$(get_env_value "EDITALBOX_AGENT_BASE_URL" || true)"
  if [[ -z "${current// }" || "$current" == "http://127.0.0.1:8090" || "$current" == "http://localhost:8090" ]]; then
    warn "Na TV Box, o agent deve apontar para o MacBook na rede local."
    read -r -p "URL do agent no MacBook (ex.: http://192.168.0.10:8090): " current
    current="$(normalize_agent_url "$current")"
    validate_agent_url "$current"
    set_env_value "EDITALBOX_AGENT_BASE_URL" "$current"
  fi
}

check_agent_health() {
  local url="$1"
  if curl -fsS "$url/health" >/dev/null 2>&1; then
    info "Agent respondeu em $url"
  else
    warn "Nao foi possivel validar o agent em $url/health agora. A instalacao vai continuar."
  fi
}

build_binary() {
  local output_path
  output_path="$(mktemp)"
  rm -f "$output_path"
  info "Compilando binario Go para a TV Box"
  (
    cd "$TVBOX_DIR"
    CGO_ENABLED=0 go build -o "$output_path" ./cmd/editalbox
  )
  printf '%s' "$output_path"
}

install_files() {
  local binary_path="$1"
  info "Instalando arquivos em $INSTALL_ROOT"
  sudo install -d -m 755 "$INSTALL_ROOT"
  sudo install -d -m 755 "$INSTALL_ROOT/data"
  sudo install -m 755 "$binary_path" "$INSTALL_BIN"
  sudo install -m 600 "$ENV_FILE" "$INSTALL_ENV"
  sudo install -m 644 "$SERVICE_SRC" "$SERVICE_DST"
}

enable_service() {
  info "Registrando servico no systemd"
  sudo systemctl daemon-reload
  sudo systemctl enable --now editalbox-tvbox
}

show_status() {
  info "Status do servico"
  sudo systemctl --no-pager --full status editalbox-tvbox || true
}

print_summary() {
  cat <<EOF

Instalacao concluida.

Arquivos de producao:
- Binario: $INSTALL_BIN
- Ambiente: $INSTALL_ENV
- Unit file: $SERVICE_DST

Comandos uteis:
- sudo systemctl restart editalbox-tvbox
- sudo systemctl status editalbox-tvbox
- sudo journalctl -u editalbox-tvbox -n 100 --no-pager

Guia operacional:
- $ROOT_DIR/docs/startup.md

EOF
}

main() {
  require_cmd go
  require_cmd curl
  require_cmd sudo
  require_cmd systemctl

  copy_env_if_missing
  prompt_if_empty "EDITALBOX_TELEGRAM_TOKEN" "Token do bot Telegram"
  prompt_if_empty "EDITALBOX_TELEGRAM_ALLOWED_CHAT_IDS" "Chat ID permitido no Telegram"
  prompt_agent_url

  local agent_url
  agent_url="$(normalize_agent_url "$(get_env_value "EDITALBOX_AGENT_BASE_URL")")"
  validate_agent_url "$agent_url"
  set_env_value "EDITALBOX_AGENT_BASE_URL" "$agent_url"
  check_agent_health "$agent_url"

  local binary_path
  binary_path="$(build_binary)"
  trap 'rm -f "$binary_path"' EXIT

  install_files "$binary_path"
  enable_service
  show_status
  print_summary
}

main "$@"
