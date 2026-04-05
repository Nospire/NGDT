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

# Проверка SteamOS
if ! grep -q 'steamos' /etc/os-release 2>/dev/null; then
    err "Этот скрипт предназначен только для Steam Deck (SteamOS)."
    exit 1
fi

# Проверка пароля deck
PASSWD_STATUS="$(passwd -S deck 2>/dev/null | awk '{print $2}')"
if [[ "$PASSWD_STATUS" == "L" || "$PASSWD_STATUS" == "NP" ]]; then
    echo "Пароль пользователя deck не задан."
    echo "Без пароля невозможно выполнить обновление системы."
    echo "Придумайте и введите пароль (он понадобится далее):"
    echo ""
    passwd deck
    echo ""
fi

# Читаем пароль напрямую с TTY
if [[ ! -t 0 ]]; then
    # Нет TTY (curl | bash) — перезапускаем скрипт из файла
    SCRIPT_PATH="$(mktemp /tmp/ngdt-install-XXXXX.sh)"
    curl -fsSL https://gdt.geekcom.org/tui -o "$SCRIPT_PATH"
    chmod +x "$SCRIPT_PATH"
    exec bash "$SCRIPT_PATH"
fi

printf "Введите пароль sudo для пользователя deck: " > /dev/tty
IFS= read -rs GDT_SUDO_PASS < /dev/tty
printf "\n" > /dev/tty

if ! printf '%s\n' "$GDT_SUDO_PASS" | sudo -S -k -p '' true >/dev/null 2>&1; then
    err "Неверный пароль sudo."
    exit 1
fi

ok "Пароль проверен"
export GDT_SUDO_PASS

# Установка файлов
info "Создаём ${INSTALL_DIR}..."
mkdir -p "$INSTALL_DIR"

info "Скачиваем ngdt..."
curl -fsSL --progress-bar -o "$INSTALL_DIR/ngdt" "${BASE_URL}/ngdt"
chmod +x "$INSTALL_DIR/ngdt"
ok "ngdt готов"

info "Скачиваем gdt-tui..."
curl -fsSL --progress-bar -o "$INSTALL_DIR/gdt-tui" "${BASE_URL}/gdt-tui"
chmod +x "$INSTALL_DIR/gdt-tui"
ok "gdt-tui готов"

info "Скачиваем sing-box..."
curl -fsSL --progress-bar -o "$INSTALL_DIR/sing-box" "${BASE_URL}/sing-box"
chmod +x "$INSTALL_DIR/sing-box"
ok "sing-box готов"

echo ""
ok "Установка завершена. Запускаем..."
echo ""

# Запуск TUI — он уже знает про sudo через env
cd "$INSTALL_DIR" && ./gdt-tui
