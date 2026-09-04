package bot

import (
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go-mirror-bot/config"
	"go-mirror-bot/core"
	"go-mirror-bot/task"
	"go-mirror-bot/uploader"

	tele "gopkg.in/telebot.v3"
)

func RegisterHandlers(b *tele.Bot) {
	// Middleware untuk memeriksa autorisasi pengguna
	authGuard := func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			if !config.Auth.IsAuthorized(c.Sender().ID, c.Chat().ID) {
				return c.Send("⛔ <b>Akses Ditolak!</b> Anda atau grup ini belum diotorisasi untuk menggunakan bot.\n"+
					fmt.Sprintf("ℹ️ Chat ID: <code>%d</code> | User ID: <code>%d</code>", c.Chat().ID, c.Sender().ID),
					&tele.SendOptions{ParseMode: tele.ModeHTML})
			}
			return next(c)
		}
	}

	// ── 1. Perintah Publik / Umum ─────────────────────────────────────────────

	// /start
	b.Handle("/start", func(c tele.Context) error {
		msg := "👋 <b>Halo! Bot Mirror & Leech Go Siap Digunakan.</b>\n\n" +
			"<b>Perintah Unduhan Cloud (Mirror):</b>\n" +
			"• <code>/mirror &lt;link&gt;</code> - Download & upload ke Cloud via Rclone\n" +
			"• <code>/mirror &lt;link&gt; -n nama.ext</code> - Rename file unduhan\n" +
			"• <code>/mirror &lt;link&gt; -z</code> - Zip sebelum diupload\n" +
			"• <code>/mirror &lt;link&gt; -e</code> - Ekstrak arsip otomatis\n" +
			"• <code>/mirror &lt;link&gt; -rc remote:path</code> - Upload ke remote khusus\n\n" +
			"<b>Perintah Unduhan Telegram (Leech):</b>\n" +
			"• <code>/leech &lt;link&gt;</code> - Kirim file langsung ke obrolan Telegram\n" +
			"• <code>/leech &lt;link&gt; -n nama.ext</code> - Leech dengan nama baru\n" +
			"• <code>/leech &lt;link&gt; -z</code> atau <code>-e</code> - Leech dengan zip/ekstrak\n\n" +
			"<b>Fitur Lain:</b>\n" +
			"• Kirim file <code>.torrent</code> untuk unduh langsung via Torrent\n" +
			"• <code>/status</code> - Pantau progres unduhan aktif\n" +
			"• <code>/cancel &lt;gid&gt;</code> - Batalkan unduhan\n" +
			"• <code>/ping</code> - Cek responsivitas bot\n\n" +
			"<b>Perintah Admin:</b> <code>/auth</code>, <code>/unauth</code>, <code>/authlist</code>"
		return c.Send(msg, &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	// /ping
	b.Handle("/ping", func(c tele.Context) error {
		start := time.Now()
		uptime := time.Since(StartTime).Round(time.Second).String()
		latency := time.Since(start).Milliseconds()
		return c.Send(fmt.Sprintf("🏓 <b>Pong!</b> (%d ms)\n⏱ <b>Uptime:</b> %s", latency, uptime), &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	// ── 2. Perintah Administratif ─────────────────────────────────────────────

	// /auth [id]
	b.Handle("/auth", func(c tele.Context) error {
		if !config.Auth.IsOwnerOrSudo(c.Sender().ID) {
			return c.Send("⚠️ Hanya Owner atau Sudo yang dapat menggunakan perintah ini.")
		}

		var targetID int64
		args := c.Args()
		if len(args) > 0 {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return c.Send("⚠️ Format ID tidak valid. Contoh: <code>/auth -10012345678</code>", &tele.SendOptions{ParseMode: tele.ModeHTML})
			}
			targetID = id
		} else if c.Message().ReplyTo != nil && c.Message().ReplyTo.Sender != nil {
			targetID = c.Message().ReplyTo.Sender.ID
		} else {
			targetID = c.Chat().ID
		}

		config.Auth.AuthorizeChat(targetID)
		return c.Send(fmt.Sprintf("✅ <b>ID <code>%d</code> berhasil diotorisasi!</b>", targetID), &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	// /unauth [id]
	b.Handle("/unauth", func(c tele.Context) error {
		if !config.Auth.IsOwnerOrSudo(c.Sender().ID) {
			return c.Send("⚠️ Hanya Owner atau Sudo yang dapat menggunakan perintah ini.")
		}

		var targetID int64
		args := c.Args()
		if len(args) > 0 {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return c.Send("⚠️ Format ID tidak valid. Contoh: <code>/unauth -10012345678</code>", &tele.SendOptions{ParseMode: tele.ModeHTML})
			}
			targetID = id
		} else if c.Message().ReplyTo != nil && c.Message().ReplyTo.Sender != nil {
			targetID = c.Message().ReplyTo.Sender.ID
		} else {
			targetID = c.Chat().ID
		}

		config.Auth.UnauthorizeChat(targetID)
		return c.Send(fmt.Sprintf("🛑 <b>Otorisasi untuk ID <code>%d</code> berhasil dicabut!</b>", targetID), &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	// /authlist
	b.Handle("/authlist", func(c tele.Context) error {
		if !config.Auth.IsOwnerOrSudo(c.Sender().ID) {
			return c.Send("⚠️ Hanya Owner atau Sudo yang dapat menggunakan perintah ini.")
		}
		list := config.Auth.GetAllAuthorized()
		if len(list) == 0 {
			return c.Send("ℹ️ Belum ada chat atau user yang diotorisasi secara dinamis.")
		}
		var sb strings.Builder
		sb.WriteString("<b>📋 Daftar ID Terotorisasi:</b>\n")
		for _, id := range list {
			sb.WriteString(fmt.Sprintf("• <code>%d</code>\n", id))
		}
		return c.Send(sb.String(), &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	// ── 3. Perintah Download & Mirror / Leech ──────────────────────────────────

	// /status
	b.Handle("/status", authGuard(func(c tele.Context) error {
		return c.Send(task.TaskMgr.FormatStatusView(), &tele.SendOptions{ParseMode: tele.ModeHTML})
	}))

	// /cancel <gid>
	b.Handle("/cancel", authGuard(func(c tele.Context) error {
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
	}))

	// Callback tombol inline [ 🛑 Batalkan ]
	b.Handle(tele.OnCallback, func(c tele.Context) error {
		data := c.Callback().Data
		if strings.HasPrefix(data, "cancel_") {
			gid := strings.TrimPrefix(data, "cancel_")
			t := task.TaskMgr.Get(gid)
			if t != nil {
				sender := c.Sender().Username
				if sender == "" {
					sender = c.Sender().FirstName
				}
				isOwner := config.Auth.IsOwnerOrSudo(c.Sender().ID)
				isRequester := strings.TrimPrefix(t.User, "@") == sender

				if !isOwner && !isRequester {
					return c.Respond(&tele.CallbackResponse{
						Text:      "⚠️ Anda tidak memiliki izin untuk membatalkan proses orang lain!",
						ShowAlert: true,
					})
				}

				core.Aria.Remove(gid)
				task.TaskMgr.Remove(gid)
				c.Respond(&tele.CallbackResponse{Text: "Unduhan dibatalkan!"})
				return c.Edit(fmt.Sprintf("🛑 <b>Unduhan Dibatalkan!</b>\n🆔 GID: <code>%s</code>\n👤 Dibatalkan oleh: @%s", gid, sender), &tele.SendOptions{ParseMode: tele.ModeHTML})
			}
			return c.Respond(&tele.CallbackResponse{Text: "Tugas sudah selesai atau tidak ditemukan."})
		}
		return nil
	})

	// Helper eksekusi task unduhan (dipakai oleh /mirror dan /leech)
	handleDownloadCommand := func(c tele.Context, isLeech bool) error {
		rawText := c.Message().Text
		if c.Message().ReplyTo != nil && c.Message().ReplyTo.Text != "" && len(c.Args()) == 0 {
			prefix := "/mirror "
			if isLeech {
				prefix = "/leech "
			}
			rawText = prefix + c.Message().ReplyTo.Text
		}

		parsed := core.ParseMirrorArgs(rawText)
		if parsed.Link == "" {
			cmdName := "/mirror"
			if isLeech {
				cmdName = "/leech"
			}
			return c.Send(fmt.Sprintf("⚠️ Format salah! Gunakan salah satu:\n"+
				"• <code>%s https://domain.com/file.zip</code>\n"+
				"• <code>%s https://domain.com/file.zip -n Baru.zip</code>\n"+
				"• <code>%s https://domain.com/file.zip -z</code> (Kompres Zip)\n"+
				"• <code>%s https://domain.com/archive.zip -e</code> (Ekstrak)\n"+
				"• <code>%s magnet:?xt=...</code>", cmdName, cmdName, cmdName, cmdName, cmdName), &tele.SendOptions{ParseMode: tele.ModeHTML})
		}

		sender := c.Sender().Username
		if sender == "" {
			sender = c.Sender().FirstName
		}

		gid, err := core.Aria.AddURI(parsed.Link, config.AppConfig.DownloadDir, parsed.CustomName)
		if err != nil {
			return c.Send(fmt.Sprintf("❌ Gagal memulai download: %v", err))
		}

		initialName := parsed.CustomName
		if initialName == "" {
			if parsed.IsMagnet {
				initialName = "🧲 Mengambil metadata torrent..."
			} else {
				initialName = "Mengambil info..."
			}
		}

		rcloneDest := config.AppConfig.RclonePath
		if parsed.CustomRemote != "" {
			rcloneDest = parsed.CustomRemote
		}

		t := &task.Task{
			GID:       gid,
			Name:      initialName,
			Status:    "Downloading",
			User:      "@" + sender,
			StartTime: time.Now(),
		}
		task.TaskMgr.Add(t)

		inlineMarkup := &tele.ReplyMarkup{}
		btnCancel := inlineMarkup.Data("🛑 Batalkan", "cancel_"+gid)
		inlineMarkup.Inline(inlineMarkup.Row(btnCancel))

		taskType := "📥 <b>Mirror Unduhan Dimulai!</b>"
		if isLeech {
			taskType = "📥 <b>Leech Unduhan Dimulai!</b>"
		}
		if parsed.IsMagnet {
			taskType = "🧲 <b>Magnet Torrent Ditambahkan!</b>"
		}

		initialMsg, err := c.Bot().Send(c.Recipient(),
			fmt.Sprintf("%s\n🆔 <b>GID:</b> <code>%s</code>\n📁 <b>Nama:</b> <code>%s</code>\n👤 <b>Pengguna:</b> @%s",
				taskType, gid, initialName, sender),
			inlineMarkup,
			&tele.SendOptions{ParseMode: tele.ModeHTML},
		)
		if err != nil {
			log.Printf("[ERROR] Gagal mengirim pesan awal: %v", err)
		}

		// Jalankan siklus pemantauan
		go processMirrorLifecycle(b, c.Recipient(), initialMsg, gid, rcloneDest, isLeech, parsed.IsZip, parsed.IsExtract, inlineMarkup)
		return nil
	}

	// Handler /mirror
	b.Handle("/mirror", authGuard(func(c tele.Context) error {
		return handleDownloadCommand(c, false)
	}))

	// Handler /leech
	b.Handle("/leech", authGuard(func(c tele.Context) error {
		return handleDownloadCommand(c, true)
	}))

	// Handle file .torrent upload
	b.Handle(tele.OnDocument, authGuard(func(c tele.Context) error {
		doc := c.Message().Document
		if doc == nil || !strings.HasSuffix(strings.ToLower(doc.FileName), ".torrent") {
			return nil
		}

		caption := c.Message().Caption
		isLeech := strings.HasPrefix(strings.ToLower(caption), "/leech")
		parsed := core.ParseMirrorArgs(caption)

		sender := c.Sender().Username
		if sender == "" {
			sender = c.Sender().FirstName
		}

		fileReader, err := c.Bot().File(&doc.File)
		if err != nil {
			return c.Send(fmt.Sprintf("❌ Gagal membaca file torrent: %v", err))
		}
		defer fileReader.Close()

		torrentBytes, err := io.ReadAll(fileReader)
		if err != nil {
			return c.Send(fmt.Sprintf("❌ Gagal memuat data torrent: %v", err))
		}

		gid, err := core.Aria.AddTorrent(torrentBytes, config.AppConfig.DownloadDir, parsed.CustomName)
		if err != nil {
			return c.Send(fmt.Sprintf("❌ Gagal menambahkan torrent ke Aria2: %v", err))
		}

		initialName := doc.FileName
		if parsed.CustomName != "" {
			initialName = parsed.CustomName
		}

		t := &task.Task{
			GID:       gid,
			Name:      initialName,
			Status:    "Downloading",
			User:      "@" + sender,
			StartTime: time.Now(),
		}
		task.TaskMgr.Add(t)

		inlineMarkup := &tele.ReplyMarkup{}
		btnCancel := inlineMarkup.Data("🛑 Batalkan", "cancel_"+gid)
		inlineMarkup.Inline(inlineMarkup.Row(btnCancel))

		headerText := "📂 <b>Torrent Mirror Dimulai!</b>"
		if isLeech {
			headerText = "📂 <b>Torrent Leech Dimulai!</b>"
		}

		initialMsg, _ := c.Bot().Send(c.Recipient(),
			fmt.Sprintf("%s\n🆔 <b>GID:</b> <code>%s</code>\n📁 <b>File:</b> <code>%s</code>\n👤 <b>Pengguna:</b> @%s",
				headerText, gid, initialName, sender),
			inlineMarkup,
			&tele.SendOptions{ParseMode: tele.ModeHTML},
		)

		go processMirrorLifecycle(b, c.Recipient(), initialMsg, gid, config.AppConfig.RclonePath, isLeech, parsed.IsZip, parsed.IsExtract, inlineMarkup)
		return nil
	}))
}

// processMirrorLifecycle memantau download secara live -> post-process (zip/extract) -> upload (rclone/telegram) -> cleanup
func processMirrorLifecycle(b *tele.Bot, recipient tele.Recipient, statusMsg *tele.Message, gid string, rcloneDest string, isLeech bool, isZip bool, isExtract bool, markup *tele.ReplyMarkup) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	var filePath string
	var fileName string
	var totalSize int64
	var lastStatusText string
	activeGID := gid

	for range ticker.C {
		st, err := core.Aria.TellStatus(activeGID)
		if err != nil {
			log.Printf("[WARN] Gagal membaca status GID %s: %v", activeGID, err)
			continue
		}

		// Handle transisi metadata magnet
		if len(st.FollowedBy) > 0 {
			activeGID = st.FollowedBy[0]
			log.Printf("[INFO] Transisi metadata magnet ke download utama GID: %s", activeGID)
			continue
		}

		t := task.TaskMgr.Get(activeGID)
		if t == nil {
			t = task.TaskMgr.Get(gid)
			if t == nil {
				return
			}
		}

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

		// Update pesan live
		if st.Status == "active" && total > 0 && statusMsg != nil {
			bar := task.GenerateProgressBar(t.Progress)
			newText := fmt.Sprintf(
				"📥 <b>Mengunduh:</b> <code>%s</code>\n"+
					"├ <b>Progres:</b> %s %.1f%%\n"+
					"├ <b>Ukuran:</b> %s / %s\n"+
					"├ <b>Speed:</b> %s | <b>ETA:</b> %s\n"+
					"└ <b>Pengguna:</b> %s",
				t.Name, bar, t.Progress,
				core.FormatBytes(completed), core.FormatBytes(total),
				core.FormatSpeed(speed), t.ETA,
				t.User,
			)

			if newText != lastStatusText {
				lastStatusText = newText
				b.Edit(statusMsg, newText, markup, &tele.SendOptions{ParseMode: tele.ModeHTML})
			}
		}

		// Selesai Download -> Mulai Post-Processing & Upload
		if st.Status == "complete" {
			log.Printf("[INFO] Download GID %s selesai. Path: %s", activeGID, filePath)
			t.Progress = 100.0

			// 1. Ekstraksi otomatis jika flag -e aktif
			if isExtract && core.IsArchive(filePath) {
				if statusMsg != nil {
					b.Edit(statusMsg, "📦 <b>Mengekstrak arsip...</b> Mohon tunggu.", &tele.SendOptions{ParseMode: tele.ModeHTML})
				}
				extractedPath, extErr := core.ExtractArchive(filePath)
				if extErr == nil {
					filePath = extractedPath
					fileName = filepath.Base(filePath)
				} else {
					log.Printf("[WARN] Gagal mengekstrak: %v", extErr)
				}
			}

			// 2. Kompresi zip jika flag -z aktif
			if isZip {
				if statusMsg != nil {
					b.Edit(statusMsg, "🗜 <b>Mengompres ke ZIP...</b> Mohon tunggu.", &tele.SendOptions{ParseMode: tele.ModeHTML})
				}
				zipPath, zipErr := core.CompressToZip(filePath)
				if zipErr == nil {
					filePath = zipPath
					fileName = filepath.Base(filePath)
				} else {
					log.Printf("[WARN] Gagal mengompres: %v", zipErr)
				}
			}

			// 3. Eksekusi Leech (Telegram) atau Mirror (Cloud Rclone)
			if isLeech {
				t.Status = "Leeching ke Telegram"
				if statusMsg != nil {
					b.Edit(statusMsg, fmt.Sprintf("📤 <b>Unduhan Selesai!</b>\n📁 <b>File:</b> <code>%s</code>\n🚀 <i>Sedang mengirim ke Telegram...</i>", fileName), &tele.SendOptions{ParseMode: tele.ModeHTML})
				}

				leechErr := uploader.LeechToTelegram(b, recipient, filePath, t.User)
				if leechErr != nil {
					t.Status = "Error"
					t.ErrorMessage = leechErr.Error()
					b.Send(recipient, fmt.Sprintf("❌ <b>Leech Gagal:</b> %v", leechErr), &tele.SendOptions{ParseMode: tele.ModeHTML})
				} else {
					t.Status = "Completed"
					b.Send(recipient, fmt.Sprintf("✅ <b>Leech Berhasil!</b>\n📁 <code>%s</code> telah dikirim ke obrolan.", fileName), &tele.SendOptions{ParseMode: tele.ModeHTML})
				}
			} else {
				// Mirror ke Cloud via Rclone
				t.Status = "Uploading ke Cloud"
				if statusMsg != nil {
					b.Edit(statusMsg, fmt.Sprintf("📤 <b>Unduhan Selesai!</b>\n📁 <b>File:</b> <code>%s</code>\n🚀 <i>Sedang mengunggah ke Cloud Storage...</i>", fileName), &tele.SendOptions{ParseMode: tele.ModeHTML})
				}

				uploadErr := uploader.UploadFile(filePath, rcloneDest, nil)
				if uploadErr != nil {
					t.Status = "Error"
					t.ErrorMessage = uploadErr.Error()
					b.Send(recipient, fmt.Sprintf("❌ <b>Upload Gagal:</b> %v", uploadErr), &tele.SendOptions{ParseMode: tele.ModeHTML})
				} else {
					t.Status = "Completed"
					destText := rcloneDest
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
			}

			// Bersihkan file lokal & hapus dari task manager
			uploader.CleanLocal(filePath)
			task.TaskMgr.Remove(activeGID)
			task.TaskMgr.Remove(gid)
			return
		}

		// Error handling
		if st.Status == "error" {
			t.Status = "Error"
			t.ErrorMessage = st.ErrorMessage
			if statusMsg != nil {
				b.Edit(statusMsg, fmt.Sprintf("❌ <b>Unduhan Gagal (GID %s):</b> %s", activeGID, st.ErrorMessage), &tele.SendOptions{ParseMode: tele.ModeHTML})
			}
			task.TaskMgr.Remove(activeGID)
			task.TaskMgr.Remove(gid)
			return
		}

		if st.Status == "removed" {
			task.TaskMgr.Remove(activeGID)
			task.TaskMgr.Remove(gid)
			return
		}
	}
}
