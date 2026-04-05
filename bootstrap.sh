#!/usr/bin/env bash
set -euo pipefail

INSTALL_DIR="${HOME}/.scripts/ngdt"
REPO="Nospire/NGDT"
BASE_URL="https://github.com/${REPO}/releases/latest/download"

# ===== Output helpers (ASCII only — TTY has no unicode/color) =====
ok()   { printf "[OK] %s\n" "$*"; }
info() { printf "[..] %s\n" "$*"; }
err()  { printf "[ERR] %s\n" "$*" >&2; }

# ===== Restore terminal on exit =====
trap 'stty echo 2>/dev/null || true' EXIT INT TERM

echo ""
echo "========================================"
echo "  Geekcom Deck Tools - NGDT installer"
echo "========================================"
echo ""

# ===== Check SteamOS =====
if ! grep -q 'steamos' /etc/os-release 2>/dev/null; then
    err "This script is for Steam Deck (SteamOS) only."
    exit 1
fi

# ===== Desktop mode warning =====
if [[ -n "${DISPLAY:-}" || -n "${WAYLAND_DISPLAY:-}" ]]; then
    echo "WARNING: You appear to be running in desktop mode."
    echo ""
    echo "NGDT is designed for TTY (press Ctrl+Alt+F4 to switch)."
    echo ""
    echo "For desktop mode, use GDT instead:"
    echo "  https://github.com/Nospire/GDT/releases/latest"
    echo ""
    printf "Continue anyway? [y/N]: "
    IFS= read -r REPLY </dev/tty || true
    if [[ "${REPLY:-n}" != "y" && "${REPLY:-n}" != "Y" ]]; then
        echo "Aborted."
        exit 0
    fi
    echo ""
fi

# ===== Check/create deck password =====
PASSWD_STATUS="$(passwd -S deck 2>/dev/null | awk '{print $2}')"
if [[ "$PASSWD_STATUS" == "L" || "$PASSWD_STATUS" == "NP" ]]; then
    echo "No password set for user deck."
    echo "A password is required for system updates."
    echo "Please create a password:"
    echo ""
    passwd deck
    echo ""
fi

# ===== Read sudo password =====
printf "Enter sudo password (hidden): " >/dev/tty
stty -echo </dev/tty
IFS= read -r GDT_SUDO_PASS </dev/tty || true
stty echo </dev/tty
printf "\n" >/dev/tty

if [[ -z "$GDT_SUDO_PASS" ]]; then
    err "Empty password."
    exit 1
fi

if ! printf '%s\n' "$GDT_SUDO_PASS" | sudo -S -k -p '' true >/dev/null 2>&1; then
    err "Wrong sudo password."
    exit 1
fi

ok "sudo activated"
export GDT_SUDO_PASS

# ===== Download binaries =====
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

# ===== Launch TUI =====
cd "$INSTALL_DIR" && ./gdt-tui
