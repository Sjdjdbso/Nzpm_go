package modules

import (
	"fmt"
	"strings"

	"go-mirror-bot/bot/helper/ext_utils"
	"go-mirror-bot/bot/helper/mirror_utils/upload_utils"
	"go-mirror-bot/bot/helper/telegram_helper"

	tele "gopkg.in/telebot.v3"
)

func InitGDList(b *tele.Bot) {
	handleList := telegram_helper.AuthGuard(func(c tele.Context) error {
		args := c.Args()
		if len(args) == 0 {
			return c.Send("⚠️ Masukkan kata kunci pencarian!\nContoh: <code>/list Avengers</code> atau <code>/search Linux</code>", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}

		key := strings.Join(args, " ")
		statusMsg, _ := c.Bot().Send(c.Recipient(), fmt.Sprintf("🔍 <i>Mencari <code>%s</code> di Google Drive...</i>", key), &tele.SendOptions{ParseMode: tele.ModeHTML})

		go func() {
			gdHelper, err := upload_utils.NewGoogleDriveHelper()
			if err != nil {
				c.Bot().Edit(statusMsg, fmt.Sprintf("❌ <b>Google Drive Auth Gagal:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
				return
			}

			files, err := gdHelper.DriveList(key, true, "both")
			if err != nil {
				c.Bot().Edit(statusMsg, fmt.Sprintf("❌ <b>Gagal Mencari di GDrive:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
				return
			}

			if len(files) == 0 {
				c.Bot().Edit(statusMsg, fmt.Sprintf("❌ Tidak ditemukan file atau folder dengan nama <code>%s</code> di Google Drive.", key), &tele.SendOptions{ParseMode: tele.ModeHTML})
				return
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("🔍 <b>Hasil Pencarian Google Drive:</b> <code>%s</code> (Total: %d)\n\n", key, len(files)))

			maxItems := 15
			if len(files) < maxItems {
				maxItems = len(files)
			}

			for i := 0; i < maxItems; i++ {
				f := files[i]
				icon := "📄"
				link := fmt.Sprintf("https://drive.google.com/file/d/%s/view", f.Id)
				if f.MimeType == upload_utils.GDriveFolderMime {
					icon = "📁"
					link = fmt.Sprintf("https://drive.google.com/drive/folders/%s", f.Id)
				}
				sb.WriteString(fmt.Sprintf("%d. %s <a href=\"%s\">%s</a> (<code>%s</code>)\n",
					i+1, icon, link, f.Name, ext_utils.FormatBytes(f.Size)))
			}

			if len(files) > maxItems {
				sb.WriteString(fmt.Sprintf("\n<i>...dan %d item lainnya.</i>", len(files)-maxItems))
			}

			c.Bot().Edit(statusMsg, sb.String(), &tele.SendOptions{ParseMode: tele.ModeHTML})
		}()

		return nil
	})

	for _, cmd := range telegram_helper.BotCommands.ListCommand {
		b.Handle(cmd, handleList)
	}
}
