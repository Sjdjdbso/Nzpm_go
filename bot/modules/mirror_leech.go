package modules

import (
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"go-mirror-bot/bot"
	"go-mirror-bot/bot/helper/ext_utils"
	"go-mirror-bot/bot/helper/listeners"
	"go-mirror-bot/bot/helper/mirror_utils/download_utils"
	"go-mirror-bot/bot/helper/telegram_helper"

	tele "gopkg.in/telebot.v3"
)

func InitMirrorLeech(b *tele.Bot) {
	dispatch := func(c tele.Context, isLeech, forceZip, forceExtract bool) error {
		rawText := c.Message().Text
		if c.Message().ReplyTo != nil && c.Message().ReplyTo.Text != "" && len(c.Args()) == 0 {
			rawText = "/cmd " + c.Message().ReplyTo.Text
		}

		args := ext_utils.ArgParser(rawText)
		if forceZip {
			args.IsZip = true
		}
		if forceExtract {
			args.IsExtract = true
		}

		if args.Link == "" {
			return c.Send("⚠️ Format tautan kosong! Masukkan link direct, GDrive, MediaFire, atau Magnet.", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}

		// DDL Generator
		resolvedURL, resolvedName, ddlErr := download_utils.DirectLinkGenerator(args.Link)
		if ddlErr == nil && resolvedURL != args.Link {
			args.Link = resolvedURL
			if args.CustomName == "" && resolvedName != "" {
				args.CustomName = resolvedName
			}
		}

		sender := c.Sender().Username
		if sender == "" {
			sender = c.Sender().FirstName
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

		rcloneDest := bot.ConfigDict.RclonePath
		if args.CustomRemote != "" {
			rcloneDest = args.CustomRemote
		}

		modeStr := "Mirror"
		if isLeech {
			modeStr = "Leech"
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

	// Handle File .torrent
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
