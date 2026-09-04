# Go Mirror Bot (Lightweight & Koyeb-Ready)

Bot mirror Telegram berkinerja tinggi yang ditulis menggunakan **Golang**, terinspirasi dari arsitektur **WZML-X (wzv3)**, namun dioptimalkan khusus agar super ringan dan hemat sumber daya (konsumsi RAM hanya ~20MB - 35MB).

Sangat cocok untuk di-deploy di **Koyeb (Free Tier 512MB RAM)** tanpa risiko Out Of Memory (OOM).

---

## 🚀 Keunggulan Dibanding Bot Python
- **Ukuran Image Sangat Kecil:** Hanya ~160 MB (dibanding Python wzv3 yang > 1.5 GB).
- **RAM Super Irit:** ~25 MB idle, aman dari crash OOM di VPS/PaaS kentang.
- **Koyeb Health Check Ready:** Terintegrasi dengan web server mini internal pada port `8080` sehingga status Koyeb selalu **Healthy**.
- **Aria2 Multi-Connection:** Download kecepatan maksimal dengan multi-connection engine.
- **Rclone Cloud Sync:** Terintegrasi langsung dengan CLI Rclone untuk upload ke Google Drive, Mega, OneDrive, dll.

---

## 🛠 Perintah Bot
| Perintah | Deskripsi |
| :--- | :--- |
| `/start` | Memulai bot dan menampilkan daftar perintah |
| `/mirror <url>` | Download file via Aria2 dan upload ke Cloud via Rclone |
| `/status` | Melihat progres download aktif (Progress bar, Speed, ETA) |
| `/cancel <gid>` | Membatalkan unduhan berdasarkan GID |
| `/ping` | Mengecek responsivitas bot dan waktu aktif (uptime) |

---

## ⚙️ Variabel Konfigurasi (Environment Variables)

| Variabel | Wajib? | Deskripsi |
| :--- | :---: | :--- |
| `BOT_TOKEN` | **Ya** | Token bot dari [@BotFather](https://t.me/BotFather) |
| `OWNER_ID` | Tidak | Telegram User ID pemilik bot |
| `DOWNLOAD_DIR`| Tidak | Direktori unduhan sementara (default: `downloads`) |
| `RCLONE_PATH` | Tidak | Remote Rclone target (misal: `gdrive:Mirror` atau `mega:Uploads`) |
| `PORT` | Tidak | Port HTTP Health Check untuk Koyeb (default: `8080`) |

> **Catatan untuk Rclone:** Jika menggunakan remote Google Drive / Mega, Anda dapat meletakkan file `rclone.conf` di root direktori project.

---

## 🚢 Panduan Deploy ke Koyeb via GitHub

1. **Inisialisasi Git dan Push ke Repository GitHub Anda:**
   ```bash
   cd ~/go-mirror-bot
   git init
   git add .
   git commit -m "Initial commit Go mirror bot"
   git branch -M main
   git remote add origin https://github.com/<username>/<repo-kamu>.git
   git push -u origin main
   ```

2. **Deploy di Dashboard Koyeb:**
   * Buka [Koyeb Dashboard](https://app.koyeb.com/).
   * Klik **Create Service** -> Pilih **GitHub**.
   * Pilih repositori yang baru saja Anda buat.
   * Pada bagian **Builder**, pilih **Dockerfile**.
   * Di bagian **Environment variables**, tambahkan:
     * `BOT_TOKEN` : `Token bot Telegram Anda`
     * `RCLONE_PATH` : `remote:folder` (opsional)
     * `PORT` : `8080`
   * Di bagian **Health checks**, pastikan protocol **HTTP** dan path `/` atau `/healthz` pada port `8080`.
   * Klik **Deploy**!

Proses build di Koyeb biasanya hanya memakan waktu 1–2 menit!
