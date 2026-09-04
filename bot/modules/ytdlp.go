package modules

import (
	"fmt"
	"os"
	"path/filepath"

	"go-mirror-bot/bot"
	"go-mirror-bot/bot/helper/ext_utils"
	"go-mirror-bot/bot/helper/mirror_utils/download_utils"
	"go-mirror-bot/bot/helper/mirror_utils/upload_utils"
	"go-mirror-bot/bot/helper/telegram_helper"
	"go-mirror-bot/bot/helper/themes"

	tele "gopkg.in/telebot.v3"
)

func InitYtDlp(b *tele.Bot) {
	dispatch := func(c tele.Context, isLeech, isZip bool) error {
		raw := c.Message().Text
		if c.Message().ReplyTo != nil && c.Message().ReplyTo.Text != "" && len(c.Args()) == 0 {
			raw = "/cmd " + c.Message().ReplyTo.Text
		}

		args := ext_utils.ArgParser(raw)
		if args.Link == "" {
			return c.Send("⚠️ Masukkan URL video. Contoh: <code>/ytdl https://youtu.be/...</code>", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}

		sender := c.Sender().Username
		if sender == "" {
			sender = c.Sender().FirstName
		}

		modeStr := "YT-DLP Mirror"
		if isLeech {
			modeStr = "YT-DLP Leech"
		}

		statusMsg, _ := c.Bot().Send(c.Recipient(), fmt.Sprintf("<b><i>Task Started</i></b>\n┠ <b>Mode:</b> %s\n┖ <b>By:</b> @%s\n\n🚀 <i>Mengunduh video via yt-dlp...</i>", modeStr, sender), &tele.SendOptions{ParseMode: tele.ModeHTML})

		go func() {
			filePath, err := download_utils.YtDlpDownload(args.Link, bot.ConfigDict.DownloadDir, args.CustomName, nil)
			if err != nil {
				c.Bot().Edit(statusMsg, fmt.Sprintf("❌ <b>YT-DLP Gagal:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
				return
			}

			fileName := filepath.Base(filePath)
			fi, _ := os.Stat(filePath)
			var size int64
			if fi != nil {
				size = fi.Size()
			}

			if isZip {
				c.Bot().Edit(statusMsg, "🗜 <b>Mengompres ke ZIP...</b>", &tele.SendOptions{ParseMode: tele.ModeHTML})
				if zipPath, err := ext_utils.CompressToZip(filePath); err == nil {
					filePath = zipPath
					fileName = filepath.Base(filePath)
				}
			}

			if isLeech {
				c.Bot().Edit(statusMsg, fmt.Sprintf("📤 <b>Unduhan Selesai!</b>\n📁 <code>%s</code>\n🚀 <i>Mengirim ke Telegram...</i>", fileName), &tele.SendOptions{ParseMode: tele.ModeHTML})
				if err := upload_utils.TelegramUpload(b, c.Recipient(), filePath, "@"+sender); err != nil {
					c.Bot().Send(c.Recipient(), fmt.Sprintf("❌ <b>Leech Gagal:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
				} else {
					c.Bot().Send(c.Recipient(), themes.FormatCompleteMsg(fileName, size, "Leech", "", "@"+sender), &tele.SendOptions{ParseMode: tele.ModeHTML})
				}
			} else {
				c.Bot().Edit(statusMsg, fmt.Sprintf("📤 <b>Unduhan Selesai!</b>\n📁 <code>%s</code>\n🚀 <i>Mengunggah ke Cloud...</i>", fileName), &tele.SendOptions{ParseMode: tele.ModeHTML})
				dest := bot.ConfigDict.RclonePath
				if args.CustomRemote != "" {
					dest = args.CustomRemote
				}
				if err := upload_utils.RcloneTransfer(filePath, dest, nil); err != nil {
					c.Bot().Send(c.Recipient(), fmt.Sprintf("❌ <b>Upload Gagal:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
				} else {
					if dest == "" {
						dest = "Disimpan di server"
					}
					c.Bot().Send(c.Recipient(), themes.FormatCompleteMsg(fileName, size, "Mirror", dest, "@"+sender), &tele.SendOptions{ParseMode: tele.ModeHTML})
				}
			}
			ext_utils.CleanPath(filePath)
		}()
		return nil
	}

	for _, cmd := range telegram_helper.BotCommands.YtdlCommand {
		b.Handle(cmd, telegram_helper.AuthGuard(func(c tele.Context) error { return dispatch(c, false, false) }))
	}
	for _, cmd := range telegram_helper.BotCommands.YtdlZipCommand {
		b.Handle(cmd, telegram_helper.AuthGuard(func(c tele.Context) error { return dispatch(c, false, true) }))
	}
	for _, cmd := range telegram_helper.BotCommands.YtdlLeechCommand {
		b.Handle(cmd, telegram_helper.AuthGuard(func(c tele.Context) error { return dispatch(c, true, false) }))
	}
	for _, cmd := range telegram_helper.BotCommands.YtdlZipLeech {
		b.Handle(cmd, telegram_helper.AuthGuard(func(c tele.Context) error { return dispatch(c, true, true) }))
	}
}
