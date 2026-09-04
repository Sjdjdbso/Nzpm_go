package modules

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go-mirror-bot/bot"
	"go-mirror-bot/bot/helper/ext_utils"
	"go-mirror-bot/bot/helper/listeners"
	"go-mirror-bot/bot/helper/mirror_utils/download_utils"
	"go-mirror-bot/bot/helper/telegram_helper"
	"go-mirror-bot/bot/helper/themes"

	tele "gopkg.in/telebot.v3"
)

func InitMirrorLeech(b *tele.Bot) {
	dispatch := func(c tele.Context, isLeech, forceZip, forceExtract bool) error {
		rawText := c.Message().Text
		reply := c.Message().ReplyTo

		if reply != nil && reply.Text != "" && len(c.Args()) == 0 {
			rawText = "/cmd " + reply.Text
		}

		args := ext_utils.ArgParser(rawText)
		if forceZip {
			args.IsZip = true
		}
		if forceExtract {
			args.IsExtract = true
		}

		sender := c.Sender().Username
		if sender == "" {
			sender = c.Sender().FirstName
		}

		modeStr := "Mirror"
		if isLeech {
			modeStr = "Leech"
		}

		rcloneDest := bot.ConfigDict.RclonePath
		if args.CustomRemote != "" {
			rcloneDest = args.CustomRemote
		}

		// 1. Cek jika mereply File Media Telegram (Document, Video, Audio, Photo)
		if args.Link == "" && reply != nil {
			var file *tele.File
			var fileName string
			var fileSize int64

			if reply.Document != nil {
				file = &reply.Document.File
				fileName = reply.Document.FileName
				fileSize = reply.Document.FileSize
			} else if reply.Video != nil {
				file = &reply.Video.File
				fileName = reply.Video.FileName
				if fileName == "" {
					fileName = "video.mp4"
				}
				fileSize = reply.Video.FileSize
			} else if reply.Audio != nil {
				file = &reply.Audio.File
				fileName = reply.Audio.FileName
				if fileName == "" {
					fileName = "audio.mp3"
				}
				fileSize = reply.Audio.FileSize
			} else if reply.Photo != nil {
				file = &reply.Photo.File
				fileName = "photo.jpg"
				fileSize = reply.Photo.FileSize
			}

			if file != nil {
				// Cek jika file .torrent direply
				if strings.HasSuffix(strings.ToLower(fileName), ".torrent") {
					reader, err := c.Bot().File(file)
					if err == nil {
						defer reader.Close()
						torrentBytes, err := io.ReadAll(reader)
						if err == nil {
							gid, err := download_utils.Aria.AddTorrent(torrentBytes, bot.ConfigDict.DownloadDir, args.CustomName)
							if err == nil {
								tName := fileName
								if args.CustomName != "" {
									tName = args.CustomName
								}
								t := &ext_utils.Task{
									GID:       gid,
									Name:      tName,
									Status:    "Downloading",
									Mode:      modeStr,
									Engine:    "Aria2c",
									User:      "@" + sender,
									UserID:    c.Sender().ID,
									StartTime: time.Now(),
								}
								ext_utils.TaskMgr.Add(t)

								markup := &tele.ReplyMarkup{}
								btnCancel := markup.Data("🛑 Batalkan", "cancel_"+gid)
								markup.Inline(markup.Row(btnCancel))

								header := fmt.Sprintf("<b><i>Torrent Task Started</i></b>\n┠ <b>Mode:</b> %s\n┖ <b>By:</b> @%s\n\n➲ <b>File:</b> <code>%s</code>\n➲ <b>GID:</b> <code>%s</code>",
									modeStr, sender, tName, gid)
								statusMsg, _ := c.Bot().Send(c.Recipient(), header, markup, &tele.SendOptions{ParseMode: tele.ModeHTML})

								listener := &listeners.MirrorLeechListener{
									Bot:        b,
									Recipient:  c.Recipient(),
									StatusMsg:  statusMsg,
									GID:        gid,
									RcloneDest: rcloneDest,
									IsLeech:    isLeech,
									IsZip:      args.IsZip,
									IsExtract:  args.IsExtract,
									Markup:     markup,
								}
								go listener.Start()
								return nil
							}
						}
					}
				}

				// Unduh File Media Telegram Langsung
				if args.CustomName != "" {
					fileName = args.CustomName
				}
				gid := fmt.Sprintf("tg_%d", time.Now().UnixNano()%1000000)
				t := &ext_utils.Task{
					GID:       gid,
					Name:      fileName,
					Status:    "Downloading",
					Mode:      modeStr,
					Engine:    "Telegram",
					User:      "@" + sender,
					UserID:    c.Sender().ID,
					StartTime: time.Now(),
					TotalSize: fileSize,
				}
				ext_utils.TaskMgr.Add(t)

				markup := &tele.ReplyMarkup{}
				btnCancel := markup.Data("🛑 Batalkan", "cancel_"+gid)
				markup.Inline(markup.Row(btnCancel))

				header := fmt.Sprintf("<b><i>Telegram Task Started</i></b>\n┠ <b>Mode:</b> %s\n┖ <b>By:</b> @%s\n\n➲ <b>File:</b> <code>%s</code>\n➲ <b>GID:</b> <code>%s</code>",
					modeStr, sender, fileName, gid)
				statusMsg, _ := c.Bot().Send(c.Recipient(), header, markup, &tele.SendOptions{ParseMode: tele.ModeHTML})

				listener := &listeners.MirrorLeechListener{
					Bot:        b,
					Recipient:  c.Recipient(),
					StatusMsg:  statusMsg,
					GID:        gid,
					RcloneDest: rcloneDest,
					IsLeech:    isLeech,
					IsZip:      args.IsZip,
					IsExtract:  args.IsExtract,
					Markup:     markup,
				}

				go func() {
					reader, err := c.Bot().File(file)
					if err != nil {
						c.Bot().Edit(statusMsg, fmt.Sprintf("❌ Gagal mengambil file TG: %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
						ext_utils.TaskMgr.Remove(gid)
						return
					}
					defer reader.Close()

					_ = os.MkdirAll("downloads", 0755)
					destPath := filepath.Join("downloads", fmt.Sprintf("tg_%d_%s", time.Now().Unix(), fileName))
					out, err := os.Create(destPath)
					if err != nil {
						c.Bot().Edit(statusMsg, fmt.Sprintf("❌ Gagal membuat file lokal: %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
						ext_utils.TaskMgr.Remove(gid)
						return
					}
					defer out.Close()

					buf := make([]byte, 64*1024)
					var downloaded int64
					startTime := time.Now()
					lastUpdate := time.Now()

					for {
						n, readErr := reader.Read(buf)
						if n > 0 {
							_, _ = out.Write(buf[:n])
							downloaded += int64(n)
							t.CompletedSize = downloaded

							if fileSize > 0 {
								t.Progress = float64(downloaded) / float64(fileSize) * 100.0
								elapsed := time.Since(startTime).Seconds()
								if elapsed > 0 {
									t.Speed = int64(float64(downloaded) / elapsed)
									t.ETA = ext_utils.CalculateETA(fileSize, downloaded, t.Speed)
								}
							}

							if time.Since(lastUpdate) >= 3*time.Second {
								lastUpdate = time.Now()
								c.Bot().Edit(statusMsg, themes.FormatStatusMsg(t), markup, &tele.SendOptions{ParseMode: tele.ModeHTML})
							}
						}
						if readErr != nil {
							if readErr == io.EOF {
								break
							}
							c.Bot().Edit(statusMsg, fmt.Sprintf("❌ Error unduh file TG: %v", readErr), &tele.SendOptions{ParseMode: tele.ModeHTML})
							ext_utils.TaskMgr.Remove(gid)
							return
						}
					}

					t.CompletedSize = downloaded
					t.TotalSize = downloaded
					t.Progress = 100.0
					listener.ProcessCompletedDownload(destPath, downloaded, t)
				}()

				return nil
			}
		}

		if args.Link == "" {
			return c.Send("⚠️ Format tautan kosong! Masukkan link direct, GDrive, Mega, MediaFire, atau Magnet.", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}

		// 2. Cek jika tautan Mega.nz
		if strings.Contains(args.Link, "mega.nz") || strings.Contains(args.Link, "mega.co.nz") {
			gid := fmt.Sprintf("mega_%d", time.Now().UnixNano()%1000000)
			initialName := args.CustomName
			if initialName == "" {
				initialName = "Unduhan Mega.nz..."
			}

			t := &ext_utils.Task{
				GID:       gid,
				Name:      initialName,
				Status:    "Downloading",
				Mode:      modeStr,
				Engine:    "Megatools",
				User:      "@" + sender,
				UserID:    c.Sender().ID,
				StartTime: time.Now(),
			}
			ext_utils.TaskMgr.Add(t)

			markup := &tele.ReplyMarkup{}
			btnCancel := markup.Data("🛑 Batalkan", "cancel_"+gid)
			markup.Inline(markup.Row(btnCancel))

			header := fmt.Sprintf("<b><i>Mega Task Started</i></b>\n┠ <b>Mode:</b> %s\n┖ <b>By:</b> @%s\n\n➲ <b>Name:</b> <code>%s</code>\n➲ <b>GID:</b> <code>%s</code>",
				modeStr, sender, initialName, gid)
			statusMsg, _ := c.Bot().Send(c.Recipient(), header, markup, &tele.SendOptions{ParseMode: tele.ModeHTML})

			listener := &listeners.MirrorLeechListener{
				Bot:        b,
				Recipient:  c.Recipient(),
				StatusMsg:  statusMsg,
				GID:        gid,
				RcloneDest: rcloneDest,
				IsLeech:    isLeech,
				IsZip:      args.IsZip,
				IsExtract:  args.IsExtract,
				Markup:     markup,
			}

			go func() {
				outDir := filepath.Join("downloads", gid)
				downloadedPath, err := download_utils.MegaDownload(args.Link, outDir, nil)
				if err != nil {
					c.Bot().Edit(statusMsg, fmt.Sprintf("❌ <b>Mega Download Gagal:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
					ext_utils.TaskMgr.Remove(gid)
					return
				}

				var totalSize int64
				if fi, err := os.Stat(downloadedPath); err == nil {
					totalSize = fi.Size()
				}
				t.Name = filepath.Base(downloadedPath)
				listener.ProcessCompletedDownload(downloadedPath, totalSize, t)
			}()

			return nil
		}

		// 3. Direct Link Generator (GDrive, Mediafire, Pixeldrain, Solidfiles, Dropbox, dll)
		resolvedURL, resolvedName, ddlErr := download_utils.DirectLinkGenerator(args.Link)
		if ddlErr == nil && resolvedURL != args.Link {
			args.Link = resolvedURL
			if args.CustomName == "" && resolvedName != "" {
				args.CustomName = resolvedName
			}
		}

		gid, err := download_utils.Aria.AddURI(args.Link, bot.ConfigDict.DownloadDir, args.CustomName)
		if err != nil {
			return c.Send(fmt.Sprintf("❌ Gagal menambahkan ke Aria2: %v", err))
		}

		initialName := args.CustomName
		if initialName == "" {
			if args.IsMagnet {
				initialName = "🧲 Mengambil metadata torrent..."
			} else {
				initialName = "Mengambil info..."
			}
		}

		t := &ext_utils.Task{
			GID:       gid,
			Name:      initialName,
			Status:    "Downloading",
			Mode:      modeStr,
			Engine:    "Aria2c",
			User:      "@" + sender,
			UserID:    c.Sender().ID,
			StartTime: time.Now(),
		}
		ext_utils.TaskMgr.Add(t)

		inlineMarkup := &tele.ReplyMarkup{}
		btnCancel := inlineMarkup.Data("🛑 Batalkan", "cancel_"+gid)
		inlineMarkup.Inline(inlineMarkup.Row(btnCancel))

		header := fmt.Sprintf("<b><i>Task Started</i></b>\n┠ <b>Mode:</b> %s\n┖ <b>By:</b> @%s\n\n➲ <b>GID:</b> <code>%s</code>\n➲ <b>Name:</b> <code>%s</code>",
			modeStr, sender, gid, initialName)

		statusMsg, err := c.Bot().Send(c.Recipient(), header, inlineMarkup, &tele.SendOptions{ParseMode: tele.ModeHTML})
		if err != nil {
			log.Printf("[ERROR] Gagal mengirim initial msg: %v", err)
		}

		listener := &listeners.MirrorLeechListener{
			Bot:        b,
			Recipient:  c.Recipient(),
			StatusMsg:  statusMsg,
			GID:        gid,
			RcloneDest: rcloneDest,
			IsLeech:    isLeech,
			IsZip:      args.IsZip,
			IsExtract:  args.IsExtract,
			Markup:     inlineMarkup,
		}
		go listener.Start()
		return nil
	}

	// Mirror Commands
	for _, cmd := range telegram_helper.BotCommands.MirrorCommand {
		b.Handle(cmd, telegram_helper.AuthGuard(func(c tele.Context) error { return dispatch(c, false, false, false) }))
	}
	for _, cmd := range telegram_helper.BotCommands.ZipMirrorCommand {
		b.Handle(cmd, telegram_helper.AuthGuard(func(c tele.Context) error { return dispatch(c, false, true, false) }))
	}
	for _, cmd := range telegram_helper.BotCommands.UnzipMirrorCommand {
		b.Handle(cmd, telegram_helper.AuthGuard(func(c tele.Context) error { return dispatch(c, false, false, true) }))
	}

	// Leech Commands
	for _, cmd := range telegram_helper.BotCommands.LeechCommand {
		b.Handle(cmd, telegram_helper.AuthGuard(func(c tele.Context) error { return dispatch(c, true, false, false) }))
	}
	for _, cmd := range telegram_helper.BotCommands.ZipLeechCommand {
		b.Handle(cmd, telegram_helper.AuthGuard(func(c tele.Context) error { return dispatch(c, true, true, false) }))
	}
	for _, cmd := range telegram_helper.BotCommands.UnzipLeechCommand {
		b.Handle(cmd, telegram_helper.AuthGuard(func(c tele.Context) error { return dispatch(c, true, false, true) }))
	}

	// Handle File .torrent langsung
	b.Handle(tele.OnDocument, telegram_helper.AuthGuard(func(c tele.Context) error {
		doc := c.Message().Document
		if doc == nil || !strings.HasSuffix(strings.ToLower(doc.FileName), ".torrent") {
			return nil
		}

		caption := c.Message().Caption
		isLeech := strings.Contains(strings.ToLower(caption), "leech") || strings.HasPrefix(strings.ToLower(caption), "/l")
		args := ext_utils.ArgParser(caption)

		sender := c.Sender().Username
		if sender == "" {
			sender = c.Sender().FirstName
		}

		fileReader, err := c.Bot().File(&doc.File)
		if err != nil {
			return c.Send(fmt.Sprintf("❌ Gagal membaca file: %v", err))
		}
		defer fileReader.Close()

		torrentBytes, err := io.ReadAll(fileReader)
		if err != nil {
			return c.Send(fmt.Sprintf("❌ Gagal membaca konten torrent: %v", err))
		}

		gid, err := download_utils.Aria.AddTorrent(torrentBytes, bot.ConfigDict.DownloadDir, args.CustomName)
		if err != nil {
			return c.Send(fmt.Sprintf("❌ Gagal menambahkan torrent ke Aria2: %v", err))
		}

		name := doc.FileName
		if args.CustomName != "" {
			name = args.CustomName
		}
		modeStr := "Mirror"
		if isLeech {
			modeStr = "Leech"
		}

		t := &ext_utils.Task{
			GID:       gid,
			Name:      name,
			Status:    "Downloading",
			Mode:      modeStr,
			Engine:    "Aria2c",
			User:      "@" + sender,
			UserID:    c.Sender().ID,
			StartTime: time.Now(),
		}
		ext_utils.TaskMgr.Add(t)

		inlineMarkup := &tele.ReplyMarkup{}
		btnCancel := inlineMarkup.Data("🛑 Batalkan", "cancel_"+gid)
		inlineMarkup.Inline(inlineMarkup.Row(btnCancel))

		header := fmt.Sprintf("<b><i>Torrent Task Started</i></b>\n┠ <b>Mode:</b> %s\n┖ <b>By:</b> @%s\n\n➲ <b>File:</b> <code>%s</code>\n➲ <b>GID:</b> <code>%s</code>",
			modeStr, sender, name, gid)
		statusMsg, _ := c.Bot().Send(c.Recipient(), header, inlineMarkup, &tele.SendOptions{ParseMode: tele.ModeHTML})

		listener := &listeners.MirrorLeechListener{
			Bot:        b,
			Recipient:  c.Recipient(),
			StatusMsg:  statusMsg,
			GID:        gid,
			RcloneDest: bot.ConfigDict.RclonePath,
			IsLeech:    isLeech,
			IsZip:      args.IsZip,
			IsExtract:  args.IsExtract,
			Markup:     inlineMarkup,
		}
		go listener.Start()
		return nil
	}))
}
