#!/usr/bin/env bash
set -euo pipefail

INSTALL_DIR="${HOME}/NGDT"
REPO="Nospire/NGDT"
BASE_URL="https://github.com/${REPO}/releases/latest/download"

echo "=== Geekcom Deck Tools — установка ==="

# Проверка пароля
if ! passwd -S deck 2>/dev/null | grep -q ' P '; then
    echo "Пароль не задан. Задайте пароль для пользователя deck:"
    passwd deck
fi

# Проверка sudo
echo "Введите пароль sudo:"
if ! sudo -v; then
    echo "Ошибка: не удалось активировать sudo"
    exit 1
fi

# Создаём папку
mkdir -p "$INSTALL_DIR"

echo "Скачиваем файлы..."
curl -fsSL -o "$INSTALL_DIR/ngdt" "${BASE_URL}/ngdt"
curl -fsSL -o "$INSTALL_DIR/gdt-tui" "${BASE_URL}/gdt-tui"
curl -fsSL -o "$INSTALL_DIR/sing-box" "${BASE_URL}/sing-box"

chmod +x "$INSTALL_DIR/ngdt" "$INSTALL_DIR/gdt-tui" "$INSTALL_DIR/sing-box"

echo "Запускаем обновление SteamOS..."
cd "$INSTALL_DIR" && ./gdt-tui
