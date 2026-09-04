package bot

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"go-mirror-bot/config"
	"go-mirror-bot/core"
	"go-mirror-bot/task"
	"go-mirror-bot/uploader"

	tele "gopkg.in/telebot.v3"
)

func RegisterHandlers(b *tele.Bot) {
	// /start
	b.Handle("/start", func(c tele.Context) error {
		msg := "👋 <b>Halo! Bot Mirror Go Siap Digunakan.</b>\n\n" +
			"<b>Perintah Tersedia:</b>\n" +
			"• <code>/mirror &lt;url&gt;</code> - Download & Mirror ke Cloud\n" +
			"• <code>/status</code> - Pantau progres unduhan/unggahan aktif\n" +
			"• <code>/cancel &lt;gid&gt;</code> - Batalkan unduhan\n" +
			"• <code>/ping</code> - Cek responsivitas bot"
		return c.Send(msg, &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	// /ping
	b.Handle("/ping", func(c tele.Context) error {
		start := time.Now()
		uptime := time.Since(StartTime).Round(time.Second).String()
		latency := time.Since(start).Milliseconds()
		return c.Send(fmt.Sprintf("🏓 <b>Pong!</b> (%d ms)\n⏱ <b>Uptime:</b> %s", latency, uptime), &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	// /status
	b.Handle("/status", func(c tele.Context) error {
		return c.Send(task.TaskMgr.FormatStatusView(), &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	// /cancel <gid>
	b.Handle("/cancel", func(c tele.Context) error {
		args := c.Args()
		if len(args) == 0 {
			return c.Send("⚠️ Masukkan GID tugas yang ingin dibatalkan. Contoh:\n<code>/cancel 2089b05ecca3d829</code>", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}
		gid := args[0]
		err := core.Aria.Remove(gid)
		if err != nil {
			return c.Send(fmt.Sprintf("❌ Gagal membatalkan GID %s: %v", gid, err))
		}
		task.TaskMgr.Remove(gid)
		return c.Send(fmt.Sprintf("🛑 Berhasil membatalkan unduhan GID: <code>%s</code>", gid), &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	// /mirror <url>
	b.Handle("/mirror", func(c tele.Context) error {
		args := c.Args()
		if len(args) == 0 {
			return c.Send("⚠️ Format salah! Gunakan:\n<code>/mirror https://domain.com/file.zip</code>", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}

		targetURL := strings.TrimSpace(args[0])
		sender := c.Sender().Username
		if sender == "" {
			sender = c.Sender().FirstName
		}

		// Tambahkan ke Aria2
		gid, err := core.Aria.AddURI(targetURL, config.AppConfig.DownloadDir, "")
		if err != nil {
			return c.Send(fmt.Sprintf("❌ Gagal memulai download: %v", err))
		}

		// Catat tugas baru
		t := &task.Task{
			GID:       gid,
			Name:      "Mengambil info...",
			Status:    "Downloading",
			User:      "@" + sender,
			StartTime: time.Now(),
		}
		task.TaskMgr.Add(t)

		initialMsg, err := c.Bot().Send(c.Recipient(), fmt.Sprintf("📥 <b>Unduhan Dimulai!</b>\n🆔 <b>GID:</b> <code>%s</code>\n🔗 <b>URL:</b> <code>%s</code>", gid, targetURL), &tele.SendOptions{ParseMode: tele.ModeHTML})
		if err != nil {
			log.Printf("[ERROR] Gagal mengirim pesan awal: %v", err)
		}

		// Goroutine untuk memantau siklus tugas hingga selesai
		go processMirrorLifecycle(b, c.Recipient(), initialMsg, gid)

		return nil
	})
}

// processMirrorLifecycle memantau download -> upload -> cleanup
func processMirrorLifecycle(b *tele.Bot, recipient tele.Recipient, statusMsg *tele.Message, gid string) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	var filePath string
	var fileName string
	var totalSize int64

	for range ticker.C {
		st, err := core.Aria.TellStatus(gid)
		if err != nil {
			log.Printf("[WARN] Gagal membaca status GID %s: %v", gid, err)
			continue
		}

		t := task.TaskMgr.Get(gid)
		if t == nil {
			return
		}

		// Ambil nama file dan path jika tersedia
		if len(st.Files) > 0 && st.Files[0].Path != "" {
			filePath = st.Files[0].Path
			fileName = filepath.Base(filePath)
			t.Name = fileName
			t.FilePath = filePath
		}

		total := core.StringToInt64(st.TotalLength)
		completed := core.StringToInt64(st.CompletedLength)
		speed := core.StringToInt64(st.DownloadSpeed)
		totalSize = total

		t.TotalSize = total
		t.CompletedSize = completed
		t.Speed = speed
		t.ETA = core.CalculateETA(total, completed, speed)
		if total > 0 {
			t.Progress = float64(completed) / float64(total) * 100.0
		}

		// Cek apakah selesai
		if st.Status == "complete" {
			log.Printf("[INFO] Download GID %s selesai. Path: %s", gid, filePath)
			t.Status = "Uploading"
			t.Progress = 100.0

			b.Send(recipient, fmt.Sprintf("📤 <b>Unduhan Selesai!</b>\n📁 File: <code>%s</code>\n🚀 Memulai upload ke Cloud via Rclone...", fileName), &tele.SendOptions{ParseMode: tele.ModeHTML})

			// Jalankan upload Rclone
			uploadErr := uploader.UploadFile(filePath, config.AppConfig.RclonePath, nil)
			if uploadErr != nil {
				t.Status = "Error"
				t.ErrorMessage = uploadErr.Error()
				b.Send(recipient, fmt.Sprintf("❌ <b>Upload Gagal:</b> %v", uploadErr), &tele.SendOptions{ParseMode: tele.ModeHTML})
			} else {
				t.Status = "Completed"
				destText := config.AppConfig.RclonePath
				if destText == "" {
					destText = "Disimpan di server (RCLONE_PATH belum diset)"
				}
				b.Send(recipient, fmt.Sprintf(
					"✅ <b>Mirror Berhasil!</b>\n\n"+
						"📁 <b>File:</b> <code>%s</code>\n"+
						"📦 <b>Ukuran:</b> <code>%s</code>\n"+
						"☁️ <b>Tujuan:</b> <code>%s</code>\n"+
						"👤 <b>Pengguna:</b> %s",
					fileName,
					core.FormatBytes(totalSize),
					destText,
					t.User,
				), &tele.SendOptions{ParseMode: tele.ModeHTML})
			}

			// Bersihkan file lokal
			uploader.CleanLocal(filePath)
			task.TaskMgr.Remove(gid)
			return
		}

		// Cek jika error atau dibatalkan
		if st.Status == "error" {
			t.Status = "Error"
			t.ErrorMessage = st.ErrorMessage
			b.Send(recipient, fmt.Sprintf("❌ <b>Unduhan Gagal (GID %s):</b> %s", gid, st.ErrorMessage), &tele.SendOptions{ParseMode: tele.ModeHTML})
			task.TaskMgr.Remove(gid)
			return
		}

		if st.Status == "removed" {
			task.TaskMgr.Remove(gid)
			return
		}
	}
}
