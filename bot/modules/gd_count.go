package modules

import (
	"fmt"
	"strings"

	"go-mirror-bot/bot/helper/ext_utils"
	"go-mirror-bot/bot/helper/mirror_utils/upload_utils"
	"go-mirror-bot/bot/helper/telegram_helper"

	tele "gopkg.in/telebot.v3"
)

func InitGDCount(b *tele.Bot) {
	b.Handle(telegram_helper.BotCommands.CountCommand, telegram_helper.AuthGuard(func(c tele.Context) error {
		args := c.Args()
		var target string

		if len(args) > 0 {
			target = args[0]
		} else if c.Message().ReplyTo != nil && c.Message().ReplyTo.Text != "" {
			parts := strings.Fields(c.Message().ReplyTo.Text)
			if len(parts) > 0 {
				target = parts[0]
			}
		}

		if target == "" {
			return c.Send("⚠️ Format salah! Gunakan:\n<code>/count &lt;gdrive_link / remote:path&gt;</code>\natau balas pesan berisi link Google Drive dengan <code>/count</code>", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}

		sender := c.Sender().Username
		if sender == "" {
			sender = c.Sender().FirstName
		}

		// 1. Jika link Google Drive
		if ext_utils.IsGDriveLink(target) {
			statusMsg, _ := c.Bot().Send(c.Recipient(), fmt.Sprintf("🔍 <i>Menghitung isi Google Drive...</i>\n<code>%s</code>", target), &tele.SendOptions{ParseMode: tele.ModeHTML})

			go func() {
				gdHelper, err := upload_utils.NewGoogleDriveHelper()
				if err != nil {
					c.Bot().Edit(statusMsg, fmt.Sprintf("❌ <b>Google Drive Auth Gagal:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
					return
				}

				name, mimeType, totalBytes, totalFiles, totalFolders, err := gdHelper.Count(target)
				if err != nil {
					c.Bot().Edit(statusMsg, fmt.Sprintf("❌ <b>Gagal Count GDrive:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
					return
				}

				result := fmt.Sprintf(
					"⌬ <b><i>Google Drive Info:</i></b>\n\n"+
						"➲ <b>Name:</b> <code>%s</code>\n"+
						"┠ <b>Size:</b> <code>%s</code>\n"+
						"┠ <b>Type:</b> <code>%s</code>\n",
					name, ext_utils.FormatBytes(totalBytes), mimeType,
				)

				if mimeType == "Folder" {
					result += fmt.Sprintf("┠ <b>SubFolders:</b> <code>%d</code>\n┠ <b>Files:</b> <code>%d</code>\n", totalFolders, totalFiles)
				}
				result += fmt.Sprintf("┖ <b>By:</b> @%s", sender)

				c.Bot().Edit(statusMsg, result, &tele.SendOptions{ParseMode: tele.ModeHTML})
			}()
			return nil
		}

		// 2. Jika Rclone path
		statusMsg, _ := c.Bot().Send(c.Recipient(), fmt.Sprintf("🔍 <i>Menghitung isi remote <code>%s</code>...</i>", target), &tele.SendOptions{ParseMode: tele.ModeHTML})
		out, err := upload_utils.RcloneCount(target)
		if err != nil {
			_, err = c.Bot().Edit(statusMsg, fmt.Sprintf("❌ <b>Gagal:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
			return err
		}
		_, err = c.Bot().Edit(statusMsg, fmt.Sprintf("📊 <b>Hasil Count Remote:</b>\n<code>%s</code>\n\n<pre>%s</pre>", target, out), &tele.SendOptions{ParseMode: tele.ModeHTML})
		return err
	}))
}
