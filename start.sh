#!/bin/bash

set -e

# Pastikan permission eksekusi untuk skrip pendukung
chmod +x aria.sh 2>/dev/null || true

# Jika binary go-mirror-bot belum terkompilasi tapi ada compiler Go, build otomatis
if [ ! -f "go-mirror-bot" ] && command -v go &> /dev/null; then
    echo "[INFO] Mengompilasi binary go-mirror-bot..."
    go build -ldflags="-s -w" -o go-mirror-bot .
fi

# Jalankan Aria2c Daemon jika belum aktif
if ! pgrep -x "aria2c" > /dev/null; then
    echo "[INFO] Memulai Aria2c Daemon..."
    ./aria.sh
fi

# Jalankan Go-Mirror-Bot
echo "[INFO] Menjalankan Bot Telegram..."
exec ./go-mirror-bot
