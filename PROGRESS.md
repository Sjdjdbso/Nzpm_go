# 📋 Catatan Progres & Arsitektur 1:1: WZML-X Go (wzv3 Port)

Arsitektur bot ini telah di-rebuild secara **1:1 persis mengikuti struktur direktori dan modul WZML-X (`wzv3`)**, dengan mengganti seluruh codebase dari Python ke **Golang**.

---

## 🗂 Pemetaan Struktur Folder 1:1 (Python ➔ Golang)

| Modul Asli (`wzv3` Python) | Modul Porting (`wzv3` Go) | Fungsi |
| :--- | :--- | :--- |
| `bot/config.py` | `bot/config.go` | Konfigurasi, Envs loader, dan Auth guard |
| `bot/startup.py` | `bot/startup.go` | Inisialisasi Aria2c daemon & Web Health Check port 8080 |
| `bot/version.py` | `bot/version.go` | Informasi versi bot dan framework |
| `bot/__main__.py` | `main.go` | Entrypoint utama pendaftaran semua modul |
| `bot/helper/ext_utils/bot_utils.py` | `bot/helper/ext_utils/bot_utils.go` | Format bytes, speed, ETA, dan ArgParser |
| `bot/helper/ext_utils/fs_utils.py` | `bot/helper/ext_utils/fs_utils.go` | Kompresi 7z, ekstrak arsip, dan split 49MB |
| `bot/helper/ext_utils/task_manager.py`| `bot/helper/ext_utils/task_manager.go` | In-memory thread-safe Task Queue |
| `bot/helper/listeners/tasks_listener.py`| `bot/helper/listeners/tasks_listener.go`| Lifecycle unduhan (download -> post-process -> upload -> clean) |
| `bot/helper/mirror_utils/download_utils/aria2_download.py` | `bot/helper/mirror_utils/download_utils/aria2_download.go` | Klien JSON-RPC Aria2c multi-thread & torrent |
| `bot/helper/mirror_utils/download_utils/direct_link_generator.py` | `bot/helper/mirror_utils/download_utils/direct_link_generator.go` | DDL resolver (GDrive, Mediafire, Pixeldrain, Dropbox, Solidfiles) |
| `bot/helper/mirror_utils/download_utils/yt_dlp_download.py` | `bot/helper/mirror_utils/download_utils/yt_dlp_download.go` | Downloader YT-DLP ratusan situs video/audio |
| `bot/helper/mirror_utils/upload_utils/rclone_transfer.py` | `bot/helper/mirror_utils/upload_utils/rclone_transfer.go` | Uploader Rclone ke Google Drive & Cloud storage |
| `bot/helper/mirror_utils/upload_utils/pyrogramEngine.py` | `bot/helper/mirror_utils/upload_utils/tg_uploader.go` | Telegram Leech engine & auto-splitter |
| `bot/helper/mirror_utils/status_utils/` | `bot/helper/mirror_utils/status_utils/status_manager.go` | Pengelola tampilan pesan status aktif |
| `bot/helper/telegram_helper/bot_commands.py` | `bot/helper/telegram_helper/bot_commands.go` | Definisi seluruh perintah & alias WZML |
| `bot/helper/telegram_helper/filters.py` | `bot/helper/telegram_helper/filters.go` | Middleware proteksi autorisasi chat |
| `bot/helper/themes/wzml_minimal.py` | `bot/helper/themes/wzml_minimal.go` | Tampilan tema WZML (bar `■■■□□`, pohon status, stats server) |
| `bot/modules/mirror_leech.py` | `bot/modules/mirror_leech.go` | Handlers: `/mirror`, `/leech`, `/zm`, `/zl`, `/uzm`, `/uzl`, torrent |
| `bot/modules/ytdlp.py` | `bot/modules/ytdlp.go` | Handlers: `/ytdl`, `/y`, `/ytdlleech`, `/yl`, `/yz`, `/yzl` |
| `bot/modules/clone.py` & `gd_count.py`| `bot/modules/clone.go` | Handlers: `/clone`, `/c`, `/count` |
| `bot/modules/status.py` | `bot/modules/status.go` | Handlers: `/status`, `/s`, `/statusall` + tombol refresh |
| `bot/modules/cancel_mirror.py` | `bot/modules/cancel_mirror.go` | Handlers: `/cancel`, `/stop`, `/cancelall` + tombol cancel |
| `bot/modules/authorize.py` | `bot/modules/authorize.go` | Handlers: `/authorize`, `/a`, `/unauthorize`, `/ua`, `/authlist` |
| `bot/modules/shell.py` | `bot/modules/shell.go` | Handlers: `/shell` terminal executor |
| `bot/modules/status.py` (stats part) | `bot/modules/stats.go` | Handlers: `/stats`, `/st`, `/ping`, `/p`, `/start`, `/help` |

---

## 🚀 Status Kinerja
- **Bahasa:** Golang 1.22
- **RAM Usage:** ~12.5 MB - 14 MB (Idle)
- **Ukuran Docker Image:** ~140 MB (Koyeb Ready)
