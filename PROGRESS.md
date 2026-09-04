# 📋 Catatan Progres & Roadmap: Go-Mirror-Bot

Dokumentasi pelacakan perkembangan implementasi fitur bot mirror berbasis Golang, mengadopsi fungsionalitas dari **WZML-X (wzv3)** secara bertahap untuk memastikan stabilitas dan performa optimal di Koyeb (Docker).

---

## 📊 Status Ringkas Saat Ini
- **Status Bot:** 🟢 Berjalan Aktif & Responsif
- **Konsumsi RAM:** ~13 MB - 16 MB (Idle)
- **Engine Download:** Aria2c JSON-RPC (Aktif)
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
- [x] Perintah `/mirror <url>` untuk direct download.
- [x] Auto-cleanup file lokal setelah proses selesai.
- [x] Mini HTTP web server di port 8080 untuk Koyeb Health Check.
- [x] Dockerfile multi-stage build super ringan (~42MB compressed).

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

### ⏳ Fase 5: Kompresi, Dekompresi, & Leech Engine (Berikutnya)
- [ ] Flag kompresi `-z` (Zip otomatis sebelum upload).
- [ ] Flag dekompresi `-e` (Extract `.zip`, `.rar`, `.7z` otomatis setelah download).
- [ ] Perintah `/leech <url>` untuk mengirim hasil unduhan langsung ke obrolan Telegram.

---

## 📝 Log Perubahan (Changelog)

### [2026-09-04] - Versi 1.2.0 (Security, Magnets, Torrents & Live Progress)
- Menambahkan sistem autorisasi dinamis (`/auth`, `/unauth`, `/authlist`).
- Menambahkan parser argumen (`-n`, `|`, `-rc`).
- Menambahkan integrasi unduhan torrent file & magnet link.
- Menambahkan auto-updating message dengan tombol inline cancel interaktif.

### [2026-09-04] - Versi 1.0.0 (Initial Working Release)
- Membangun seluruh arsitektur bot Go menggantikan dependensi Python.
- Menguji langsung di server dengan bot Telegram `@Tesmirorbot`.
- Menghasilkan waktu build Docker cepat dan konsumsi memori sangat rendah (~13 MB).
- Memperbaiki bug permission denied Aria2 dengan auto-resolusi absolute path direktori download.
