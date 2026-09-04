package bot

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
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
	// Guard autorisasi
	authGuard := func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			if !config.Auth.IsAuthorized(c.Sender().ID, c.Chat().ID) {
				return c.Send("<i>You Are not authorized user! Deploy your own WZML-X Mirror-Leech bot</i>\n\n"+
					fmt.Sprintf("ℹ️ <b>Chat ID:</b> <code>%d</code> | <b>User ID:</b> <code>%d</code>", c.Chat().ID, c.Sender().ID),
					&tele.SendOptions{ParseMode: tele.ModeHTML})
			}
			return next(c)
		}
	}

	// ── 1. Start, Help, Ping, Stats ───────────────────────────────────────────

	b.Handle("/start", func(c tele.Context) error {
		msg := "<i>This bot can mirror all your links|files|torrents to Google Drive or any rclone cloud or to telegram.</i>\n\n" +
			"<b>Type /help to get a list of available commands</b>"
		return c.Send(msg, &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	handleHelp := func(c tele.Context) error {
		helpText := "㊂ <b><i>WZML-X Go Help Guide Menu!</i></b>\n\n" +
			"<b>Aria2 & DDL Mirror / Leech:</b>\n" +
			"• <code>/mirror</code>, <code>/m</code> - Direct Link, GDrive, MediaFire, Magnet, Torrent ke Cloud\n" +
			"• <code>/leech</code>, <code>/l</code> - Download & kirim langsung ke Telegram\n" +
			"• <code>/zipmirror</code>, <code>/zm</code> | <code>/zipleech</code>, <code>/zl</code> - Kompres Zip\n" +
			"• <code>/unzipmirror</code>, <code>/uzm</code> | <code>/unzipleech</code>, <code>/uzl</code> - Ekstrak Arsip\n\n" +
			"<b>YT-DLP (YouTube & Video Sites):</b>\n" +
			"• <code>/ytdl</code>, <code>/y</code> - Unduh video/audio ke Cloud\n" +
			"• <code>/ytdlleech</code>, <code>/yl</code> - Unduh video/audio ke Telegram\n" +
			"• <code>/ytdlzip</code>, <code>/yz</code> | <code>/ytdlzipleech</code>, <code>/yzl</code>\n\n" +
			"<b>Cloud & GDrive Tools:</b>\n" +
			"• <code>/clone &lt;src&gt; &lt;dst&gt;</code>, <code>/c</code> - Salin antar remote Cloud\n" +
			"• <code>/count &lt;remote:path&gt;</code> - Hitung total file & ukuran remote\n\n" +
			"<b>Status & Admin:</b>\n" +
			"• <code>/status</code>, <code>/s</code> | <code>/cancel &lt;gid&gt;</code> | <code>/cancelall</code>\n" +
			"• <code>/stats</code>, <code>/st</code> | <code>/ping</code>, <code>/p</code> | <code>/shell &lt;cmd&gt;</code>"
		return c.Send(helpText, &tele.SendOptions{ParseMode: tele.ModeHTML})
	}
	b.Handle("/help", handleHelp)
	b.Handle("/h", handleHelp)

	handlePing := func(c tele.Context) error {
		start := time.Now()
		latency := time.Since(start).Milliseconds()
		return c.Send(fmt.Sprintf("<b>Pong</b>\n<code>%d ms..</code>", latency), &tele.SendOptions{ParseMode: tele.ModeHTML})
	}
	b.Handle("/ping", handlePing)
	b.Handle("/p", handlePing)

	handleStats := func(c tele.Context) error {
		st := core.GetSystemStats(task.TaskMgr.BotStartTime)
		statsMsg := fmt.Sprintf(
			"⌬ <b><i>BOT STATISTICS :</i></b>\n"+
				"┖ <b>Bot Uptime :</b> %s\n\n"+
				"┎ <b><i>RAM ( MEMORY ) :</i></b>\n"+
				"┃ %s %.1f%%\n"+
				"┖ <b>U :</b> %s | <b>F :</b> %s | <b>T :</b> %s\n\n"+
				"┎ <b><i>DISK :</i></b>\n"+
				"┃ %s %.1f%%\n"+
				"┖ <b>U :</b> %s | <b>F :</b> %s | <b>T :</b> %s\n\n"+
				"⌬ <b><i>OS SYSTEM :</i></b>\n"+
				"┠ <b>OS Arch :</b> %s\n"+
				"┖ <b>Go Runtime :</b> %s",
			st.Uptime,
			core.GenerateWZMLBar(st.RAMPercent), st.RAMPercent,
			st.RAMUsed, st.RAMFree, st.RAMTotal,
			core.GenerateWZMLBar(st.DiskPercent), st.DiskPercent,
			st.DiskUsed, st.DiskFree, st.DiskTotal,
			st.OSArch,
			st.GoVersion,
		)
		return c.Send(statsMsg, &tele.SendOptions{ParseMode: tele.ModeHTML})
	}
	b.Handle("/stats", handleStats)
	b.Handle("/st", handleStats)

	// ── 2. Admin & Shell Commands ─────────────────────────────────────────────

	b.Handle("/auth", func(c tele.Context) error {
		if !config.Auth.IsOwnerOrSudo(c.Sender().ID) {
			return c.Send("⚠️ Hanya Owner atau Sudo yang dapat menggunakan perintah ini.")
		}
		var targetID int64
		args := c.Args()
		if len(args) > 0 {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return c.Send("⚠️ Format ID tidak valid.", &tele.SendOptions{ParseMode: tele.ModeHTML})
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

	b.Handle("/unauth", func(c tele.Context) error {
		if !config.Auth.IsOwnerOrSudo(c.Sender().ID) {
			return c.Send("⚠️ Hanya Owner atau Sudo yang dapat menggunakan perintah ini.")
		}
		var targetID int64
		args := c.Args()
		if len(args) > 0 {
			id, _ := strconv.ParseInt(args[0], 10, 64)
			targetID = id
		} else if c.Message().ReplyTo != nil && c.Message().ReplyTo.Sender != nil {
			targetID = c.Message().ReplyTo.Sender.ID
		} else {
			targetID = c.Chat().ID
		}

		config.Auth.UnauthorizeChat(targetID)
		return c.Send(fmt.Sprintf("🛑 <b>Otorisasi ID <code>%d</code> berhasil dicabut!</b>", targetID), &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	b.Handle("/authlist", func(c tele.Context) error {
		if !config.Auth.IsOwnerOrSudo(c.Sender().ID) {
			return c.Send("⚠️ Hanya Owner atau Sudo yang dapat menggunakan perintah ini.")
		}
		list := config.Auth.GetAllAuthorized()
		var sb strings.Builder
		sb.WriteString("<b>📋 Daftar ID Terotorisasi:</b>\n")
		for _, id := range list {
			sb.WriteString(fmt.Sprintf("• <code>%d</code>\n", id))
		}
		return c.Send(sb.String(), &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	b.Handle("/shell", func(c tele.Context) error {
		if !config.Auth.IsOwnerOrSudo(c.Sender().ID) {
			return c.Send("⚠️ Hanya Owner yang diizinkan mengeksekusi shell.")
		}
		cmdStr := strings.TrimSpace(strings.TrimPrefix(c.Text(), "/shell"))
		if cmdStr == "" {
			return c.Send("⚠️ Masukkan perintah shell.", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}
		out, err := exec.Command("bash", "-c", cmdStr).CombinedOutput()
		resText := string(out)
		if err != nil {
			resText += fmt.Sprintf("\n[Error: %v]", err)
		}
		if len(resText) > 4000 {
			resText = resText[:4000] + "\n...(terpotong)"
		}
		return c.Send(fmt.Sprintf("<b>Terminal Output:</b>\n<pre>%s</pre>", resText), &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	// ── 3. Status & Cancel ────────────────────────────────────────────────────

	handleStatus := authGuard(func(c tele.Context) error {
		inlineMarkup := &tele.ReplyMarkup{}
		btnRefresh := inlineMarkup.Data("🔄 Refresh", "refresh_status")
		inlineMarkup.Inline(inlineMarkup.Row(btnRefresh))
		return c.Send(task.TaskMgr.FormatStatusView(), inlineMarkup, &tele.SendOptions{ParseMode: tele.ModeHTML})
	})
	b.Handle("/status", handleStatus)
	b.Handle("/s", handleStatus)
	b.Handle("/statusall", handleStatus)

	b.Handle(tele.OnCallback, func(c tele.Context) error {
		data := c.Callback().Data
		if data == "refresh_status" {
			inlineMarkup := &tele.ReplyMarkup{}
			btnRefresh := inlineMarkup.Data("🔄 Refresh", "refresh_status")
			inlineMarkup.Inline(inlineMarkup.Row(btnRefresh))
			c.Edit(task.TaskMgr.FormatStatusView(), inlineMarkup, &tele.SendOptions{ParseMode: tele.ModeHTML})
			return c.Respond(&tele.CallbackResponse{Text: "Status diperbarui!"})
		}

		if strings.HasPrefix(data, "cancel_") {
			gid := strings.TrimPrefix(data, "cancel_")
			t := task.TaskMgr.Get(gid)
			if t != nil {
				sender := c.Sender().Username
				if sender == "" {
					sender = c.Sender().FirstName
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

	handleCancel := authGuard(func(c tele.Context) error {
		args := c.Args()
		if len(args) == 0 {
			return c.Send("⚠️ Masukkan GID tugas. Contoh:\n<code>/cancel 2089b05ecca3d829</code>", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}
		gid := args[0]
		core.Aria.Remove(gid)
		task.TaskMgr.Remove(gid)
		return c.Send(fmt.Sprintf("🛑 Berhasil membatalkan unduhan GID: <code>%s</code>", gid), &tele.SendOptions{ParseMode: tele.ModeHTML})
	})
	b.Handle("/cancel", handleCancel)
	b.Handle("/stop", handleCancel)

	b.Handle("/cancelall", authGuard(func(c tele.Context) error {
		if !config.Auth.IsOwnerOrSudo(c.Sender().ID) {
			return c.Send("⚠️ Hanya Owner atau Sudo yang dapat membatalkan semua proses.")
		}
		count := task.TaskMgr.CancelAll()
		return c.Send(fmt.Sprintf("🛑 <b>Berhasil membatalkan seluruh proses aktif (%d tugas).</b>", count), &tele.SendOptions{ParseMode: tele.ModeHTML})
	}))

	// ── 4. Cloud Clone & Count (Google Drive / Rclone) ─────────────────────────

	handleClone := authGuard(func(c tele.Context) error {
		args := c.Args()
		if len(args) < 2 {
			return c.Send("⚠️ Format salah! Gunakan:\n<code>/clone &lt;source_remote:path&gt; &lt;dest_remote:path&gt;</code>\n\nContoh: <code>/clone gdrive:SourceFolder gdrive:DestFolder</code>", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}
		src, dst := args[0], args[1]
		statusMsg, _ := c.Bot().Send(c.Recipient(), fmt.Sprintf("♻️ <b>Memulai Cloud Clone...</b>\n\n➲ <b>Src:</b> <code>%s</code>\n➲ <b>Dst:</b> <code>%s</code>", src, dst), &tele.SendOptions{ParseMode: tele.ModeHTML})

		go func() {
			err := uploader.CloneCloud(src, dst, nil)
			if err != nil {
				c.Bot().Edit(statusMsg, fmt.Sprintf("❌ <b>Clone Gagal:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
			} else {
				c.Bot().Edit(statusMsg, fmt.Sprintf("✅ <b>Clone Berhasil!</b>\n\n➲ <b>Src:</b> <code>%s</code>\n➲ <b>Dst:</b> <code>%s</code>", src, dst), &tele.SendOptions{ParseMode: tele.ModeHTML})
			}
		}()
		return nil
	})
	b.Handle("/clone", handleClone)
	b.Handle("/c", handleClone)

	b.Handle("/count", authGuard(func(c tele.Context) error {
		args := c.Args()
		if len(args) == 0 {
			return c.Send("⚠️ Format salah! Gunakan:\n<code>/count &lt;remote:path&gt;</code>", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}
		remote := args[0]
		statusMsg, _ := c.Bot().Send(c.Recipient(), fmt.Sprintf("🔍 <i>Menghitung isi remote <code>%s</code>...</i>", remote), &tele.SendOptions{ParseMode: tele.ModeHTML})
		out, err := uploader.CountRemote(remote)
		if err != nil {
			_, err = c.Bot().Edit(statusMsg, fmt.Sprintf("❌ <b>Gagal menghitung:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
			return err
		}
		_, err = c.Bot().Edit(statusMsg, fmt.Sprintf("📊 <b>Hasil Count Remote:</b>\n<code>%s</code>\n\n<pre>%s</pre>", remote, out), &tele.SendOptions{ParseMode: tele.ModeHTML})
		return err
	}))

	// ── 5. Aria2 + DDL Resolver (Direct Link, GDrive, MediaFire, Torrent) ─────

	dispatchDownload := func(c tele.Context, isLeech bool, forceZip bool, forceExtract bool) error {
		rawText := c.Message().Text
		if c.Message().ReplyTo != nil && c.Message().ReplyTo.Text != "" && len(c.Args()) == 0 {
			rawText = "/cmd " + c.Message().ReplyTo.Text
		}

		parsed := core.ParseMirrorArgs(rawText)
		if forceZip {
			parsed.IsZip = true
		}
		if forceExtract {
			parsed.IsExtract = true
		}

		if parsed.Link == "" {
			return c.Send("⚠️ Format tautan kosong! Gunakan contoh:\n"+
				"• <code>/mirror https://drive.google.com/...</code>\n"+
				"• <code>/mirror https://mediafire.com/...</code>\n"+
				"• <code>/mirror magnet:?xt=...</code>", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}

		// DDL Resolver: Ubah Google Drive, Mediafire, Pixeldrain, dll. ke Direct Link
		resolvedURL, resolvedName, ddlErr := core.ResolveDDL(parsed.Link)
		if ddlErr == nil && resolvedURL != parsed.Link {
			log.Printf("[INFO] DDL Resolver berhasil mengonversi: %s -> %s", parsed.Link, resolvedURL)
			parsed.Link = resolvedURL
			if parsed.CustomName == "" && resolvedName != "" {
				parsed.CustomName = resolvedName
			}
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

		modeStr := "Mirror"
		if isLeech {
			modeStr = "Leech"
		}

		t := &task.Task{
			GID:       gid,
			Name:      initialName,
			Status:    "Downloading",
			Mode:      modeStr,
			User:      "@" + sender,
			UserID:    c.Sender().ID,
			StartTime: time.Now(),
		}
		task.TaskMgr.Add(t)

		inlineMarkup := &tele.ReplyMarkup{}
		btnCancel := inlineMarkup.Data("🛑 Batalkan", "cancel_"+gid)
		inlineMarkup.Inline(inlineMarkup.Row(btnCancel))

		headerText := "<b><i>Task Started</i></b>\n┠ <b>Mode:</b> " + modeStr + "\n┖ <b>By:</b> @" + sender + "\n\n" +
			fmt.Sprintf("➲ <b>GID:</b> <code>%s</code>\n➲ <b>Name:</b> <code>%s</code>", gid, initialName)

		initialMsg, _ := c.Bot().Send(c.Recipient(), headerText, inlineMarkup, &tele.SendOptions{ParseMode: tele.ModeHTML})
		go processMirrorLifecycle(b, c.Recipient(), initialMsg, gid, rcloneDest, isLeech, parsed.IsZip, parsed.IsExtract, inlineMarkup)
		return nil
	}

	b.Handle("/mirror", authGuard(func(c tele.Context) error { return dispatchDownload(c, false, false, false) }))
	b.Handle("/m", authGuard(func(c tele.Context) error { return dispatchDownload(c, false, false, false) }))
	b.Handle("/zipmirror", authGuard(func(c tele.Context) error { return dispatchDownload(c, false, true, false) }))
	b.Handle("/zm", authGuard(func(c tele.Context) error { return dispatchDownload(c, false, true, false) }))
	b.Handle("/unzipmirror", authGuard(func(c tele.Context) error { return dispatchDownload(c, false, false, true) }))
	b.Handle("/uzm", authGuard(func(c tele.Context) error { return dispatchDownload(c, false, false, true) }))

	b.Handle("/leech", authGuard(func(c tele.Context) error { return dispatchDownload(c, true, false, false) }))
	b.Handle("/l", authGuard(func(c tele.Context) error { return dispatchDownload(c, true, false, false) }))
	b.Handle("/zipleech", authGuard(func(c tele.Context) error { return dispatchDownload(c, true, true, false) }))
	b.Handle("/zl", authGuard(func(c tele.Context) error { return dispatchDownload(c, true, true, false) }))
	b.Handle("/unzipleech", authGuard(func(c tele.Context) error { return dispatchDownload(c, true, false, true) }))
	b.Handle("/uzl", authGuard(func(c tele.Context) error { return dispatchDownload(c, true, false, true) }))

	// ── 6. YT-DLP Dispatcher (YouTube, TikTok, Instagram, Twitter, dll.) ──────

	dispatchYtDlp := func(c tele.Context, isLeech bool, isZip bool) error {
		rawText := c.Message().Text
		if c.Message().ReplyTo != nil && c.Message().ReplyTo.Text != "" && len(c.Args()) == 0 {
			rawText = "/cmd " + c.Message().ReplyTo.Text
		}

		parsed := core.ParseMirrorArgs(rawText)
		if parsed.Link == "" {
			return c.Send("⚠️ Format link salah! Contoh:\n<code>/ytdl https://www.youtube.com/watch?v=...</code>\n<code>/ytdlleech https://vt.tiktok.com/...</code>", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}

		sender := c.Sender().Username
		if sender == "" {
			sender = c.Sender().FirstName
		}

		modeStr := "YT-DLP Mirror"
		if isLeech {
			modeStr = "YT-DLP Leech"
		}

		statusMsg, _ := c.Bot().Send(c.Recipient(), fmt.Sprintf("<b><i>Task Started</i></b>\n┠ <b>Mode:</b> %s\n┖ <b>By:</b> @%s\n\n🚀 <i>Mengambil informasi video via yt-dlp...</i>", modeStr, sender), &tele.SendOptions{ParseMode: tele.ModeHTML})

		go func() {
			filePath, err := core.DownloadYtDlp(parsed.Link, config.AppConfig.DownloadDir, parsed.CustomName, func(p core.YtDlpProgress) {
				// Throttle/update status secara berkala jika perlu
			})

			if err != nil {
				c.Bot().Edit(statusMsg, fmt.Sprintf("❌ <b>YT-DLP Gagal:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
				return
			}

			fileName := filepath.Base(filePath)
			fi, _ := os.Stat(filePath)
			var totalSize int64
			if fi != nil {
				totalSize = fi.Size()
			}

			// Jika flag zip aktif
			if isZip {
				zipPath, zipErr := core.CompressToZip(filePath)
				if zipErr == nil {
					filePath = zipPath
					fileName = filepath.Base(filePath)
				}
			}

			if isLeech {
				c.Bot().Edit(statusMsg, fmt.Sprintf("📤 <b>Unduhan Selesai!</b>\n📁 <code>%s</code>\n🚀 <i>Mengirim ke Telegram...</i>", fileName), &tele.SendOptions{ParseMode: tele.ModeHTML})
				leechErr := uploader.LeechToTelegram(b, c.Recipient(), filePath, "@"+sender)
				if leechErr != nil {
					c.Bot().Send(c.Recipient(), fmt.Sprintf("❌ <b>Leech Gagal:</b> %v", leechErr), &tele.SendOptions{ParseMode: tele.ModeHTML})
				} else {
					c.Bot().Send(c.Recipient(), core.FormatWZMLComplete(fileName, totalSize, "0s", "Leech", "", "@"+sender), &tele.SendOptions{ParseMode: tele.ModeHTML})
				}
			} else {
				c.Bot().Edit(statusMsg, fmt.Sprintf("📤 <b>Unduhan Selesai!</b>\n📁 <code>%s</code>\n🚀 <i>Mengunggah ke Cloud Storage...</i>", fileName), &tele.SendOptions{ParseMode: tele.ModeHTML})
				rcloneDest := config.AppConfig.RclonePath
				if parsed.CustomRemote != "" {
					rcloneDest = parsed.CustomRemote
				}
				uploadErr := uploader.UploadFile(filePath, rcloneDest, nil)
				if uploadErr != nil {
					c.Bot().Send(c.Recipient(), fmt.Sprintf("❌ <b>Upload Gagal:</b> %v", uploadErr), &tele.SendOptions{ParseMode: tele.ModeHTML})
				} else {
					destText := rcloneDest
					if destText == "" {
						destText = "Disimpan di server (RCLONE_PATH belum diset)"
					}
					c.Bot().Send(c.Recipient(), core.FormatWZMLComplete(fileName, totalSize, "0s", "Mirror", destText, "@"+sender), &tele.SendOptions{ParseMode: tele.ModeHTML})
				}
			}

			uploader.CleanLocal(filePath)
		}()

		return nil
	}

	b.Handle("/ytdl", authGuard(func(c tele.Context) error { return dispatchYtDlp(c, false, false) }))
	b.Handle("/y", authGuard(func(c tele.Context) error { return dispatchYtDlp(c, false, false) }))
	b.Handle("/ytdlzip", authGuard(func(c tele.Context) error { return dispatchYtDlp(c, false, true) }))
	b.Handle("/yz", authGuard(func(c tele.Context) error { return dispatchYtDlp(c, false, true) }))

	b.Handle("/ytdlleech", authGuard(func(c tele.Context) error { return dispatchYtDlp(c, true, false) }))
	b.Handle("/yl", authGuard(func(c tele.Context) error { return dispatchYtDlp(c, true, false) }))
	b.Handle("/ytdlzipleech", authGuard(func(c tele.Context) error { return dispatchYtDlp(c, true, true) }))
	b.Handle("/yzl", authGuard(func(c tele.Context) error { return dispatchYtDlp(c, true, true) }))

	// ── 7. Handle File .torrent ───────────────────────────────────────────────

	b.Handle(tele.OnDocument, authGuard(func(c tele.Context) error {
		doc := c.Message().Document
		if doc == nil || !strings.HasSuffix(strings.ToLower(doc.FileName), ".torrent") {
			return nil
		}

		caption := c.Message().Caption
		isLeech := strings.Contains(strings.ToLower(caption), "leech") || strings.HasPrefix(strings.ToLower(caption), "/l")
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

		modeStr := "Mirror"
		if isLeech {
			modeStr = "Leech"
		}

		t := &task.Task{
			GID:       gid,
			Name:      initialName,
			Status:    "Downloading",
			Mode:      modeStr,
			User:      "@" + sender,
			UserID:    c.Sender().ID,
			StartTime: time.Now(),
		}
		task.TaskMgr.Add(t)

		inlineMarkup := &tele.ReplyMarkup{}
		btnCancel := inlineMarkup.Data("🛑 Batalkan", "cancel_"+gid)
		inlineMarkup.Inline(inlineMarkup.Row(btnCancel))

		headerText := fmt.Sprintf("<b><i>Torrent Task Started</i></b>\n┠ <b>Mode:</b> %s\n┖ <b>By:</b> @%s\n\n➲ <b>File:</b> <code>%s</code>\n➲ <b>GID:</b> <code>%s</code>", modeStr, sender, initialName, gid)
		initialMsg, _ := c.Bot().Send(c.Recipient(), headerText, inlineMarkup, &tele.SendOptions{ParseMode: tele.ModeHTML})

		go processMirrorLifecycle(b, c.Recipient(), initialMsg, gid, config.AppConfig.RclonePath, isLeech, parsed.IsZip, parsed.IsExtract, inlineMarkup)
		return nil
	}))
}

// processMirrorLifecycle memantau download live -> archive (zip/extract) -> upload -> cleanup
func processMirrorLifecycle(b *tele.Bot, recipient tele.Recipient, statusMsg *tele.Message, gid string, rcloneDest string, isLeech bool, isZip bool, isExtract bool, markup *tele.ReplyMarkup) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	var filePath string
	var fileName string
	var totalSize int64
	var lastStatusText string
	activeGID := gid
	taskStartTime := time.Now()

	for range ticker.C {
		st, err := core.Aria.TellStatus(activeGID)
		if err != nil {
			log.Printf("[WARN] Gagal membaca status GID %s: %v", activeGID, err)
			continue
		}

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

		// Update live message
		if st.Status == "active" && total > 0 && statusMsg != nil {
			newText := core.FormatWZMLTaskStatus(
				t.Name,
				t.Progress,
				completed,
				total,
				speed,
				t.ETA,
				"Downloading",
				t.Mode,
				t.User,
				t.UserID,
				activeGID,
				taskStartTime,
			)

			if newText != lastStatusText {
				lastStatusText = newText
				b.Edit(statusMsg, newText, markup, &tele.SendOptions{ParseMode: tele.ModeHTML})
			}
		}

		// Selesai Download
		if st.Status == "complete" {
			log.Printf("[INFO] Download GID %s selesai. Path: %s", activeGID, filePath)
			t.Progress = 100.0

			if isExtract && core.IsArchive(filePath) {
				if statusMsg != nil {
					b.Edit(statusMsg, "📦 <b>Mengekstrak arsip...</b> Mohon tunggu.", &tele.SendOptions{ParseMode: tele.ModeHTML})
				}
				extractedPath, extErr := core.ExtractArchive(filePath)
				if extErr == nil {
					filePath = extractedPath
					fileName = filepath.Base(filePath)
				}
			}

			if isZip {
				if statusMsg != nil {
					b.Edit(statusMsg, "🗜 <b>Mengompres ke ZIP...</b> Mohon tunggu.", &tele.SendOptions{ParseMode: tele.ModeHTML})
				}
				zipPath, zipErr := core.CompressToZip(filePath)
				if zipErr == nil {
					filePath = zipPath
					fileName = filepath.Base(filePath)
				}
			}

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
					b.Send(recipient, core.FormatWZMLComplete(fileName, totalSize, "0s", "Leech", "", t.User), &tele.SendOptions{ParseMode: tele.ModeHTML})
				}
			} else {
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
					b.Send(recipient, core.FormatWZMLComplete(fileName, totalSize, "0s", "Mirror", destText, t.User), &tele.SendOptions{ParseMode: tele.ModeHTML})
				}
			}

			uploader.CleanLocal(filePath)
			task.TaskMgr.Remove(activeGID)
			task.TaskMgr.Remove(gid)
			return
		}

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
