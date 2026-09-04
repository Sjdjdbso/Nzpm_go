#!/bin/bash

# Ambil trackers list dengan timeout 5s (supaya tidak hang jika jaringan lambat)
tracker_list=$(curl -Ns --connect-timeout 5 https://ngosang.github.io/trackerslist/trackers_all_http.txt 2>/dev/null | awk '$0' | tr '\n' ',' || echo "")

echo "[INFO] Menjalankan Aria2c Daemon..."
aria2c --allow-overwrite=true \
       --auto-file-renaming=true \
       --bt-enable-lpd=true \
       --bt-detach-seed-only=true \
       --bt-remove-unselected-file=true \
       --bt-tracker="[$tracker_list]" \
       --bt-max-peers=0 \
       --enable-rpc=true \
       --rpc-listen-all=false \
       --rpc-listen-port=6800 \
       --rpc-max-request-size=1024M \
       --max-connection-per-server=16 \
       --max-concurrent-downloads=10 \
       --split=16 \
       --seed-ratio=0 \
       --check-integrity=true \
       --continue=true \
       --daemon=true \
       --disk-cache=40M \
       --force-save=true \
       --min-split-size=10M \
       --follow-torrent=mem \
       --check-certificate=false \
       --optimize-concurrent-downloads=true \
       --http-accept-gzip=true \
       --max-file-not-found=0 \
       --max-tries=20 \
       --reuse-uri=true \
       --content-disposition-default-utf8=true \
       --quiet=true \
       --summary-interval=0
