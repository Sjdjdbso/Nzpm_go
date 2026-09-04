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

### ⏳ Fase 2: Keamanan & Autorisasi Chat (Berikutnya)
- [ ] Validasi `OWNER_ID` dan daftar `SUDO_USERS`.
- [ ] Dukungan `AUTHORIZED_CHATS` (membatasi penggunaan bot hanya di grup atau user tertentu).
- [ ] Perintah admin `/authorize`, `/unauthorize`, `/users`.
- [ ] Filter pesan agar bot tidak merespon pengguna ilegal.

---

### ⏳ Fase 3: Fleksibilitas Download & Argument Parser
- [ ] Dukungan **Magnet Link** dan **Torrent File** (`.torrent` upload ke bot) via Aria2.
- [ ] Argument parser mirip wzv3:
  - Custom filename: `/mirror <url> -n nama_baru.zip` atau `/mirror <url> | nama_baru.zip`.
  - Custom Rclone destination: `/mirror <url> -rc remote:path`.
- [ ] Kompresi & Dekompresi Otomatis:
  - Flag `-z` (Zip folder/file sebelum upload).
  - Flag `-e` (Extract archive `.zip`, `.rar`, `.7z`, `.tar` setelah download).

---

### ⏳ Fase 4: UX & Dynamic Status Message
- [ ] Auto-refreshing status message (mengedit pesan progress setiap interval 3-5 detik secara halus tanpa spam API Telegram).
- [ ] Tombol inline keyboard `[ 🛑 Batalkan ]` langsung di pesan status tiap tugas.
- [ ] Format pesan selesai yang lebih rapi (Kecepatan rata-rata, waktu proses, tautan cloud jika tersedia).

---

### ⏳ Fase 5: Telegram Leech Engine (`/leech`)
- [ ] Perintah `/leech <url>` untuk mengirim hasil unduhan langsung ke Telegram chat (bukan ke cloud).
- [ ] Support custom thumbnail & caption.
- [ ] Split file otomatis jika melebihi batas upload Telegram Bot API.

---

### ⏳ Fase 6: Rclone Cloud Multi-Config & Sync
- [ ] Integrasi penuh file `rclone.conf`.
- [ ] Command `/clone <remote1:path> <remote2:path>`.
- [ ] Command `/list <query>` untuk mencari file di Google Drive / Cloud.

---

## 📝 Log Perubahan (Changelog)

### [2026-09-04] - Versi 1.0.0 (Initial Working Release)
- Membangun seluruh arsitektur bot Go menggantikan dependensi Python.
- Menguji langsung di server dengan bot Telegram `@Tesmirorbot`.
- Menghasilkan waktu build Docker cepat dan konsumsi memori sangat rendah (~13 MB).
- Memperbaiki bug permission denied Aria2 dengan auto-resolusi absolute path direktori download.
