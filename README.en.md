# NGDT — Geekcom Deck Tools (No GUI)

A utility for updating SteamOS on Steam Deck in Russia.

## Why is this needed?

Roskomnadzor has blocked Valve's servers, causing `steamos-update`
to stop working — the Deck cannot receive system updates.

NGDT solves this problem: it automatically brings up a temporary tunnel,
runs the update through it, and then the tunnel closes on its own.

## Installation

Connect a USB keyboard to your Steam Deck, switch to TTY
(Ctrl+Alt+F4), log in as user `deck` and run:

```bash
curl -fsSL https://fix.geekcom.org/ngdt | bash
```

The script will do everything itself: download the required files, ask for
the sudo password and start the update. After a successful update, reboot
the Deck:

```bash
sudo reboot
```

## Requirements

- Steam Deck with SteamOS 3.x
- USB keyboard (Bluetooth does not work in TTY)
- TTY mode: Ctrl+Alt+F4
- Password for user deck (if not set — the script will offer to create one)

## How it works

```
Install script
    ↓
Requests a temporary tunnel from the Geekcom server
    ↓
Brings up the tunnel
    ↓
Runs steamos-update through the tunnel
    ↓
Tunnel closes automatically
```

## Community

- 📢 [News](https://t.me/geekcomdeck_news)
- 🎮 [Games](https://t.me/geekcom_deck_games)
- 💬 [Chat](https://t.me/Geekcom_hub)
- ☕ [Support on Boosty](https://boosty.to/steamdecks)
