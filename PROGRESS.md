# 📋 Catatan Progres & Roadmap: Go-Mirror-Bot

Dokumentasi pelacakan perkembangan implementasi fitur bot mirror berbasis Golang, mengadopsi fungsionalitas dari **WZML-X (wzv3)** secara bertahap untuk memastikan stabilitas dan performa optimal di Koyeb (Docker).

---

## 📊 Status Ringkas Saat Ini
- **Status Bot:** 🟢 Berjalan Aktif & Responsif
- **Konsumsi RAM:** ~13 MB - 18 MB (Idle)
- **Engine Unduhan:** Aria2c JSON-RPC, YT-DLP, DDL Resolver, Rclone Cloud
- **Health Check Web:** HTTP Port 8080 (Aktif untuk Koyeb)
- **Deployment Target:** Koyeb / Docker Ready

---

## 🗺 Roadmap Fitur (Adaptasi dari wzv3)

### ✅ Fase 1: Fondasi Inti (Selesai)
- [x] Struktur modular project Golang.
- [x] Daemon manager Aria2c (`aria.sh` & `core/daemon.go`).
- [x] Klien JSON-RPC Aria2c (`AddURI`, `TellStatus`, `Remove`, `GetGlobalStat`).
- [x] Thread-safe in-memory task manager dengan visual progress bar.
- [x] Command dasar: `/start`, `/ping`, `/status`, `/cancel <gid>`.
- [x] Mini HTTP web server di port 8080 untuk Koyeb Health Check.
- [x] Dockerfile multi-stage build super ringan (~140MB).

---

### ✅ Fase 2: Keamanan & Autorisasi Chat (Selesai)
- [x] Validasi `OWNER_ID` dan daftar `SUDO_USERS`.
- [x] Dukungan `AUTHORIZED_CHATS` dari environment.
- [x] Perintah dinamis admin: `/auth <id>`, `/unauth <id>`, `/authlist`.
- [x] Middleware `authGuard` untuk memproteksi perintah unduhan & status dari pengguna liar.

---

### ✅ Fase 3: Fleksibilitas Download & Argument Parser (Selesai)
- [x] Dukungan **Magnet Link** (`magnet:?xt=...`) otomatis via Aria2.
- [x] Dukungan pengiriman file **`.torrent`** langsung ke Telegram bot.
- [x] Argument parser mirip wzv3:
  - Custom filename: `/mirror <url> -n nama_baru.zip` atau `/mirror <url> | nama_baru.zip`.
  - Custom Rclone destination: `/mirror <url> -rc remote:path`.
- [x] Deteksi tautan via reply pesan.

---

### ✅ Fase 4: UX & Dynamic Live Status Message (Selesai)
- [x] **Live Message Editing**: Pesan progress terupdate secara otomatis dan halus tiap 3 detik selama pengunduhan.
- [x] **Tombol Inline Keyboard**: Tombol `[ 🛑 Batalkan ]` interaktif langsung menempel pada pesan live status.
- [x] Proteksi pembatalan: hanya Owner, Sudo, atau pengunggah tugas yang bisa membatalkan proses.

---

### ✅ Fase 5: Kompresi, Dekompresi, & Telegram Leech Engine (Selesai)
- [x] **Perintah `/leech <url>`**: Mengirim hasil unduhan langsung ke Telegram chat pengguna sebagai Document / Video / Audio.
- [x] **Auto-Split Telegram**: Memecah file menjadi part-part 49MB jika ukuran file melebihi batas 50MB Telegram Bot API.
- [x] **Flag Kompresi `-z`**: Mengompres file/folder unduhan menjadi file `.zip` sebelum di-mirror atau di-leech.
- [x] **Flag Dekompresi `-e`**: Mengekstrak arsip (`.zip`, `.rar`, `.7z`, `.tar.gz`) secara otomatis setelah unduhan selesai.
- [x] Penambahan tool `p7zip` ke Dockerfile runtime.

---

### ✅ Fase 6: Engine DDL, YT-DLP, & Cloud Clone (Selesai)
- [x] **DDL Resolver Otomatis**: Mem-bypass dan mengubah link hosting menjadi direct download link untuk:
  - Google Drive (`drive.google.com`)
  - Mediafire (`mediafire.com`)
  - Pixeldrain (`pixeldrain.com`)
  - Dropbox (`dropbox.com`)
  - Solidfiles (`solidfiles.com`)
- [x] **YT-DLP Engine**:
  - Command `/ytdl`, `/y`, `/ytdlzip`, `/yz` (Download video/audio ke Cloud).
  - Command `/ytdlleech`, `/yl`, `/ytdlzipleech`, `/yzl` (Download video/audio ke Telegram).
  - Dukungan YouTube, TikTok, Instagram, Twitter, Facebook, SoundCloud, dll.
  - Integrasi `ffmpeg` untuk penggabungan stream video + audio resolusi tinggi.
- [x] **Cloud & GDrive Tools**:
  - Command `/clone <src> <dst>`, `/c` (Salin antar remote Cloud / Google Drive langsung tanpa makan disk lokal).
  - Command `/count <remote:path>` (Hitung total file dan ukuran folder remote).

---

## 📝 Log Perubahan (Changelog)

### [2026-09-04] - Versi 1.4.0 (DDL Resolver, YT-DLP & GDrive/Cloud Clone)
- Menambahkan DDL resolver untuk Google Drive, Mediafire, Pixeldrain, Dropbox, dan Solidfiles.
- Menambahkan engine YT-DLP dengan dukungan FFmpeg untuk ratusan situs video.
- Menambahkan perintah Cloud Clone (`/clone`) dan penghitung ukuran remote (`/count`).
- Menambahkan dependensi `ffmpeg` dan `yt-dlp` ke Docker image.

### [2026-09-04] - Versi 1.3.0 (Telegram Leech Engine, Auto-Split, & 7z Archiving)
- Menambahkan modul `/leech` untuk upload langsung ke Telegram chat.
- Menambahkan auto-splitter file 49MB untuk mengatasi limit Bot API 50MB.
- Menambahkan flag `-z` (zip) dan `-e` (extract) otomatis.

### [2026-09-04] - Versi 1.2.0 (Security, Magnets, Torrents & Live Progress)
- Menambahkan sistem autorisasi dinamis (`/auth`, `/unauth`, `/authlist`).
- Menambahkan parser argumen (`-n`, `|`, `-rc`).
- Menambahkan integrasi unduhan torrent file & magnet link.

### [2026-09-04] - Versi 1.0.0 (Initial Working Release)
- Membangun seluruh arsitektur bot Go menggantikan dependensi Python.
