#!/bin/bash

# Jalankan Aria2c Daemon
chmod +x aria.sh
./aria.sh

# Jalankan Binary Bot Go
echo "[INFO] Menjalankan Bot Telegram..."
exec ./go-mirror-bot
