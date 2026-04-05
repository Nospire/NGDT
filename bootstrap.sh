#!/usr/bin/env bash

set -euo pipefail

INSTALL_DIR="${HOME}/.scripts/ngdt"
REPO="Nospire/NGDT"
BASE_URL="https://github.com/${REPO}/releases/latest/download"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

ok()   { printf "${GREEN}[OK]${NC} %s\n" "$*"; }
info() { printf "${YELLOW}[..] %s${NC}\n" "$*"; }
err()  { printf "${RED}[ERR]${NC} %s\n" "$*" >&2; }

echo ""
echo "========================================"
echo "  Geekcom Deck Tools — NGDT installer"
echo "========================================"
echo ""

# Check SteamOS
if ! grep -q 'steamos' /etc/os-release 2>/dev/null; then
    err "This script is intended for Steam Deck (SteamOS) only."
    exit 1
fi

# Check deck user password
PASSWD_STATUS="$(passwd -S deck 2>/dev/null | awk '{print $2}')"
if [[ "$PASSWD_STATUS" == "L" || "$PASSWD_STATUS" == "NP" ]]; then
    echo "No password set for user deck."
    echo "A password is required to perform system updates."
    echo "Please create a password for user deck:"
    echo ""
    passwd deck
    echo ""
fi

# Ensure TTY is available
if [[ ! -t 0 ]]; then
    # No TTY (curl | bash) — re-execute script from a file
    SCRIPT_PATH="$(mktemp /tmp/ngdt-install-XXXXX.sh)"
    curl -fsSL https://gdt.geekcom.org/tui -o "$SCRIPT_PATH"
    chmod +x "$SCRIPT_PATH"
    exec bash "$SCRIPT_PATH"
fi

# Read sudo password
printf "Enter sudo password (input will be hidden): " > /dev/tty
stty -echo </dev/tty
IFS= read -r GDT_SUDO_PASS </dev/tty || true
stty echo </dev/tty
printf "\n" > /dev/tty

if [[ -z "$GDT_SUDO_PASS" ]]; then
    printf "[ERR] Empty password.\n" >&2
    exit 1
fi

if ! printf '%s\n' "$GDT_SUDO_PASS" | sudo -S -k -p '' true >/dev/null 2>&1; then
    printf "[ERR] Wrong sudo password.\n" >&2
    exit 1
fi

printf "[OK] sudo activated\n"
export GDT_SUDO_PASS

# Install files
info "Creating ${INSTALL_DIR}..."
mkdir -p "$INSTALL_DIR"

info "Downloading ngdt..."
curl -fsSL --progress-bar -o "$INSTALL_DIR/ngdt" "${BASE_URL}/ngdt"
chmod +x "$INSTALL_DIR/ngdt"
ok "ngdt ready"

info "Downloading gdt-tui..."
curl -fsSL --progress-bar -o "$INSTALL_DIR/gdt-tui" "${BASE_URL}/gdt-tui"
chmod +x "$INSTALL_DIR/gdt-tui"
ok "gdt-tui ready"

info "Downloading sing-box..."
curl -fsSL --progress-bar -o "$INSTALL_DIR/sing-box" "${BASE_URL}/sing-box"
chmod +x "$INSTALL_DIR/sing-box"
ok "sing-box ready"

echo ""
ok "Installation complete. Starting update..."
echo ""

# Launch TUI — reads GDT_SUDO_PASS from env
cd "$INSTALL_DIR" && ./gdt-tui
