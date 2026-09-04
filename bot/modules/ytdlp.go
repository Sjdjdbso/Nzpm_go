package modules

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go-mirror-bot/bot"
	"go-mirror-bot/bot/helper/ext_utils"
	"go-mirror-bot/bot/helper/mirror_utils/download_utils"
	"go-mirror-bot/bot/helper/mirror_utils/upload_utils"
	"go-mirror-bot/bot/helper/mirror_utils/upload_utils/ddlserver"
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

		gid := fmt.Sprintf("yt_%d", time.Now().UnixNano()%1000000)
		t := &ext_utils.Task{
			GID:       gid,
			Name:      "Mengambil info YT-DLP...",
			Status:    "Downloading",
			Mode:      modeStr,
			Engine:    "yt-dlp",
			User:      "@" + sender,
			UserID:    c.Sender().ID,
			StartTime: time.Now(),
		}
		ext_utils.TaskMgr.Add(t)

		markup := &tele.ReplyMarkup{}
		btnCancel := markup.Data("🛑 Batalkan", "cancel_"+gid)
		markup.Inline(markup.Row(btnCancel))

		header := fmt.Sprintf("<b><i>Task Started</i></b>\n┠ <b>Mode:</b> %s\n┖ <b>By:</b> @%s\n\n➲ <b>GID:</b> <code>%s</code>\n🚀 <i>Mengunduh video via yt-dlp...</i>", modeStr, sender, gid)
		statusMsg, _ := c.Bot().Send(c.Recipient(), header, markup, &tele.SendOptions{ParseMode: tele.ModeHTML})

		go func() {
			filePath, err := download_utils.YtDlpDownload(args.Link, bot.ConfigDict.DownloadDir, args.CustomName, nil)
			if err != nil {
				c.Bot().Edit(statusMsg, fmt.Sprintf("❌ <b>YT-DLP Gagal:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
				ext_utils.TaskMgr.Remove(gid)
				return
			}

			fileName := filepath.Base(filePath)
			t.Name = fileName
			fi, _ := os.Stat(filePath)
			var size int64
			if fi != nil {
				size = fi.Size()
			}
			t.TotalSize = size
			t.CompletedSize = size
			t.Progress = 100.0

			if isZip {
				c.Bot().Edit(statusMsg, "🗜 <b>Mengompres ke ZIP...</b>", &tele.SendOptions{ParseMode: tele.ModeHTML})
				if zipPath, err := ext_utils.CompressToZip(filePath); err == nil {
					filePath = zipPath
					fileName = filepath.Base(filePath)
				}
			}

			if isLeech {
				t.Status = "Leeching"
				c.Bot().Edit(statusMsg, fmt.Sprintf("📤 <b>Unduhan Selesai!</b>\n📁 <code>%s</code>\n🚀 <i>Mengirim ke Telegram...</i>", fileName), &tele.SendOptions{ParseMode: tele.ModeHTML})
				if err := upload_utils.TelegramUpload(b, c.Recipient(), filePath, "@"+sender, c.Sender().ID); err != nil {
					c.Bot().Send(c.Recipient(), fmt.Sprintf("❌ <b>Leech Gagal:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
				} else {
					c.Bot().Send(c.Recipient(), themes.FormatCompleteMsg(fileName, size, "Leech", "", "@"+sender), &tele.SendOptions{ParseMode: tele.ModeHTML})
				}
			} else {
				t.Status = "Uploading"
				c.Bot().Edit(statusMsg, fmt.Sprintf("📤 <b>Unduhan Selesai!</b>\n📁 <code>%s</code>\n🚀 <i>Mengunggah...</i>", fileName), &tele.SendOptions{ParseMode: tele.ModeHTML})
				dest := bot.ConfigDict.RclonePath
				if args.CustomRemote != "" {
					dest = args.CustomRemote
				}
				destLower := strings.ToLower(dest)
				isDDL := destLower == "ddl" || destLower == "pixeldrain" || (dest == "" && (bot.ConfigDict.DefaultUpload == "ddl" || bot.ConfigDict.DefaultUpload == "pixeldrain"))

				if isDDL {
					apiKey := bot.ConfigDict.PixeldrainAPI
					if uCfg := ext_utils.UserStore.Get(c.Sender().ID); uCfg != nil && uCfg.PixeldrainAPI != "" {
						apiKey = uCfg.PixeldrainAPI
					}
					pd := ddlserver.NewPixeldrain(apiKey)
					pdLink, err := pd.Upload(filePath)
					if err != nil {
						c.Bot().Send(c.Recipient(), fmt.Sprintf("❌ <b>Pixeldrain Upload Gagal:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
					} else {
						completeText := fmt.Sprintf(
							"<b><i>YT-DLP Mirror DDL Selesai!</i></b>\n\n"+
								"➲ <b>File:</b> <code>%s</code>\n"+
								"┠ <b>Size:</b> <code>%s</code>\n"+
								"┠ <b>Server:</b> <code>Pixeldrain DDL</code>\n"+
								"┠ <b>Link:</b> <a href=\"%s\">%s</a>\n"+
								"┖ <b>By:</b> @%s",
							fileName, ext_utils.FormatBytes(size), pdLink, pdLink, sender,
						)
						markup := &tele.ReplyMarkup{}
						btnURL := markup.URL("🔗 Unduh Pixeldrain", pdLink)
						markup.Inline(markup.Row(btnURL))
						c.Bot().Send(c.Recipient(), completeText, markup, &tele.SendOptions{ParseMode: tele.ModeHTML})
					}
				} else {
					if err := upload_utils.RcloneTransfer(filePath, dest, nil); err != nil {
						c.Bot().Send(c.Recipient(), fmt.Sprintf("❌ <b>Upload Gagal:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
					} else {
						if dest == "" {
							dest = "Disimpan di server"
						}
						c.Bot().Send(c.Recipient(), themes.FormatCompleteMsg(fileName, size, "Mirror", dest, "@"+sender), &tele.SendOptions{ParseMode: tele.ModeHTML})
					}
				}
			}
			ext_utils.CleanPath(filePath)
			ext_utils.TaskMgr.Remove(gid)
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
