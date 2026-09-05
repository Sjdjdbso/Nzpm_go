package modules

import (
	"fmt"
	"strings"

	"go-mirror-bot/bot/helper/mirror_utils/upload_utils"
	"go-mirror-bot/bot/helper/telegram_helper"

	tele "gopkg.in/telebot.v3"
)

func InitGDDelete(b *tele.Bot) {
	handleDelete := telegram_helper.AuthGuard(func(c tele.Context) error {
		args := c.Args()
		var link string

		if len(args) > 0 {
			link = args[0]
		} else if c.Message().ReplyTo != nil && c.Message().ReplyTo.Text != "" {
			parts := strings.Fields(c.Message().ReplyTo.Text)
			if len(parts) > 0 {
				link = parts[0]
			}
		}

		if link == "" {
			return c.Send("⚠️ Format salah! Gunakan:\n<code>/del &lt;gdrive_link / file_id&gt;</code>\natau balas pesan berisi link Google Drive dengan <code>/del</code>", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}

		statusMsg, _ := c.Bot().Send(c.Recipient(), "🗑 <i>Menghapus dari Google Drive...</i>", &tele.SendOptions{ParseMode: tele.ModeHTML})

		go func() {
			gdHelper, err := upload_utils.NewGoogleDriveHelper()
			if err != nil {
				c.Bot().Edit(statusMsg, fmt.Sprintf("❌ <b>Google Drive Auth Gagal:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
				return
			}

			res, err := gdHelper.DeleteFile(link)
			if err != nil {
				c.Bot().Edit(statusMsg, fmt.Sprintf("❌ <b>Gagal Menghapus:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
				return
			}

			c.Bot().Edit(statusMsg, res, &tele.SendOptions{ParseMode: tele.ModeHTML})
		}()

		return nil
	})

	for _, cmd := range telegram_helper.BotCommands.DeleteCommand {
		b.Handle(cmd, handleDelete)
	}
}
