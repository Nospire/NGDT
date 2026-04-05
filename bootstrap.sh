#!/usr/bin/env bash
set -euo pipefail

INSTALL_DIR="${HOME}/.scripts/ngdt"
REPO="Nospire/NGDT"
BASE_URL="https://github.com/${REPO}/releases/latest/download"

# ===== Language detection =====
if [[ -n "${DISPLAY:-}" || -n "${WAYLAND_DISPLAY:-}" ]]; then
    LANG_MODE="ru"
else
    LANG_MODE="en"
fi

msg() {
    local ru="$1" en="$2"
    if [[ "$LANG_MODE" == "ru" ]]; then
        printf "%s\n" "$ru"
    else
        printf "%s\n" "$en"
    fi
}

# ===== Output helpers =====
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
    msg "ВНИМАНИЕ: вы в режиме рабочего стола." \
        "WARNING: You appear to be running in desktop mode."
    echo ""
    msg "NGDT предназначен для TTY (Ctrl+Alt+F4)." \
        "NGDT is designed for TTY (press Ctrl+Alt+F4 to switch)."
    echo ""
    msg "Для рабочего стола используйте GDT:" \
        "For desktop mode, use GDT instead:"
    echo "  https://github.com/Nospire/GDT/releases/latest"
    echo ""
    printf "[..] "
    msg "Продолжить всё равно? [y/N]: " \
        "Continue anyway? [y/N]: "
    IFS= read -r REPLY </dev/tty || true
    if [[ "${REPLY:-n}" != "y" && "${REPLY:-n}" != "Y" ]]; then
        msg "Отмена." "Aborted."
        exit 0
    fi
    echo ""
fi

# ===== Check/create deck password =====
PASSWD_STATUS="$(passwd -S deck 2>/dev/null | awk '{print $2}')"
if [[ "$PASSWD_STATUS" == "L" || "$PASSWD_STATUS" == "NP" ]]; then
    msg "Нет пароля для пользователя deck." \
        "No password set for user deck."
    msg "Пароль нужен для системных операций." \
        "A password is required for system updates."
    msg "Придумайте пароль:" \
        "Please create a password:"
    echo ""
    passwd deck
    echo ""
fi

# ===== Read sudo password =====
printf "[..] "
msg "Введите пароль sudo (скрыто): " \
    "Enter sudo password (hidden): "
stty -echo </dev/tty
IFS= read -r GDT_SUDO_PASS </dev/tty || true
stty echo </dev/tty
printf "\n" >/dev/tty

if [[ -z "$GDT_SUDO_PASS" ]]; then
    err "$(msg "Пустой пароль." "Empty password.")"
    exit 1
fi

if ! printf '%s\n' "$GDT_SUDO_PASS" | sudo -S -k -p '' true >/dev/null 2>&1; then
    err "$(msg "Неверный пароль sudo." "Wrong sudo password.")"
    exit 1
fi

ok "$(msg "sudo активирован" "sudo activated")"
export GDT_SUDO_PASS

# ===== Download binaries =====
info "$(msg "Создаём ${INSTALL_DIR}..." "Creating ${INSTALL_DIR}...")"
mkdir -p "$INSTALL_DIR"

info "$(msg "Скачиваем ngdt..." "Downloading ngdt...")"
curl -fsSL --progress-bar -o "$INSTALL_DIR/ngdt" "${BASE_URL}/ngdt"
chmod +x "$INSTALL_DIR/ngdt"
ok "$(msg "ngdt готов" "ngdt ready")"

info "$(msg "Скачиваем gdt-tui..." "Downloading gdt-tui...")"
curl -fsSL --progress-bar -o "$INSTALL_DIR/gdt-tui" "${BASE_URL}/gdt-tui"
chmod +x "$INSTALL_DIR/gdt-tui"
ok "$(msg "gdt-tui готов" "gdt-tui ready")"

info "$(msg "Скачиваем sing-box..." "Downloading sing-box...")"
curl -fsSL --progress-bar -o "$INSTALL_DIR/sing-box" "${BASE_URL}/sing-box"
chmod +x "$INSTALL_DIR/sing-box"
ok "$(msg "sing-box готов" "sing-box ready")"

echo ""
ok "$(msg "Установка завершена. Запускаем..." "Installation complete. Starting update...")"
echo ""

# ===== Launch TUI =====
cd "$INSTALL_DIR" && ./gdt-tui
