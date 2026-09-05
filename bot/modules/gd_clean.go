package modules

import (
	"fmt"

	"go-mirror-bot/bot"
	"go-mirror-bot/bot/helper/mirror_utils/upload_utils"
	"go-mirror-bot/bot/helper/telegram_helper"

	tele "gopkg.in/telebot.v3"
)

func InitGDClean(b *tele.Bot) {
	handleClean := telegram_helper.SudoGuard(func(c tele.Context) error {
		args := c.Args()
		trash := true
		targetID := bot.ConfigDict.GdriveID

		for _, a := range args {
			if a == "-f" || a == "--force" {
				trash = false
			} else if a != "" {
				targetID = a
			}
		}

		if targetID == "" {
			return c.Send("⚠️ GDRIVE_ID belum ditentukan di konfigurasi bot!", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}

		action := "memindahkan ke sampah"
		if !trash {
			action = "menghapus PERMANEN"
		}

		statusMsg, _ := c.Bot().Send(c.Recipient(), fmt.Sprintf("🧹 <i>Sedang %s isi drive <code>%s</code>...</i>", action, targetID), &tele.SendOptions{ParseMode: tele.ModeHTML})

		go func() {
			gdHelper, err := upload_utils.NewGoogleDriveHelper()
			if err != nil {
				c.Bot().Edit(statusMsg, fmt.Sprintf("❌ <b>Google Drive Auth Gagal:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
				return
			}

			res, err := gdHelper.DriveClean(targetID, trash)
			if err != nil {
				c.Bot().Edit(statusMsg, fmt.Sprintf("❌ <b>Gagal Membersihkan Drive:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
				return
			}

			c.Bot().Edit(statusMsg, res, &tele.SendOptions{ParseMode: tele.ModeHTML})
		}()

		return nil
	})

	for _, cmd := range telegram_helper.BotCommands.GDCleanCommand {
		b.Handle(cmd, handleClean)
	}
}
