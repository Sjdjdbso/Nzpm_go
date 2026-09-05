package modules

import (
	"fmt"
	"strings"

	"go-mirror-bot/bot"
	"go-mirror-bot/bot/helper/ext_utils"
	"go-mirror-bot/bot/helper/mirror_utils/upload_utils"
	"go-mirror-bot/bot/helper/telegram_helper"

	tele "gopkg.in/telebot.v3"
)

func InitClone(b *tele.Bot) {
	handleClone := telegram_helper.AuthGuard(func(c tele.Context) error {
		args := c.Args()
		var src, dst string

		if len(args) == 0 && c.Message().ReplyTo != nil && c.Message().ReplyTo.Text != "" {
			parts := strings.Fields(c.Message().ReplyTo.Text)
			if len(parts) > 0 {
				src = parts[0]
			}
		} else if len(args) == 1 {
			src = args[0]
		} else if len(args) >= 2 {
			src = args[0]
			dst = args[1]
		}

		if src == "" {
			return c.Send("⚠️ Format salah! Gunakan:\n• <code>/clone &lt;gdrive_link&gt; [dest_folder_id]</code>\n• <code>/clone &lt;src_remote:path&gt; &lt;dst_remote:path&gt;</code>\natau balas link dengan <code>/clone</code>", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}

		sender := c.Sender().Username
		if sender == "" {
			sender = c.Sender().FirstName
		}

		// 1. Google Drive Native Clone
		if ext_utils.IsGDriveLink(src) {
			if dst == "" {
				dst = bot.ConfigDict.GdriveID
			}

			statusMsg, _ := c.Bot().Send(c.Recipient(), fmt.Sprintf("♻️ <b>Memulai Google Drive Clone...</b>\n\n➲ <b>Src:</b> <code>%s</code>\n➲ <b>Dst:</b> <code>%s</code>\n🚀 <i>Memproses...</i>", src, dst), &tele.SendOptions{ParseMode: tele.ModeHTML})

			go func() {
				gdHelper, err := upload_utils.NewGoogleDriveHelper()
				if err != nil {
					c.Bot().Edit(statusMsg, fmt.Sprintf("❌ <b>Google Drive Auth Gagal:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
					return
				}

				resLink, totalBytes, mimeType, totalFiles, totalFolders, err := gdHelper.Clone(src, dst, func(p, t int64, f, fo int) {
					// Progress update
				})

				if err != nil {
					c.Bot().Edit(statusMsg, fmt.Sprintf("❌ <b>GDrive Clone Gagal:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
					return
				}

				completeText := fmt.Sprintf(
					"<b><i>Google Drive Clone Selesai!</i></b>\n\n"+
						"➲ <b>Size:</b> <code>%s</code>\n"+
						"┠ <b>Type:</b> <code>%s</code>\n",
					ext_utils.FormatBytes(totalBytes), mimeType,
				)

				if mimeType == "Folder" {
					completeText += fmt.Sprintf("┠ <b>SubFolders:</b> <code>%d</code>\n┠ <b>Files:</b> <code>%d</code>\n", totalFolders, totalFiles)
				}
				completeText += fmt.Sprintf("┠ <b>Link:</b> <a href=\"%s\">Google Drive</a>\n┖ <b>By:</b> @%s", resLink, sender)

				markup := &tele.ReplyMarkup{}
				btnURL := markup.URL("🔗 Buka Hasil Clone", resLink)
				markup.Inline(markup.Row(btnURL))

				c.Bot().Edit(statusMsg, completeText, markup, &tele.SendOptions{ParseMode: tele.ModeHTML})
			}()
			return nil
		}

		// 2. Rclone Cloud Clone
		if dst == "" {
			return c.Send("⚠️ Untuk rclone remote, tujuan (destination) wajib diisi!\nContoh: <code>/clone gdrive:FolderA gdrive:FolderB</code>", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}

		statusMsg, _ := c.Bot().Send(c.Recipient(), fmt.Sprintf("♻️ <b>Memulai Cloud Clone...</b>\n\n➲ <b>Src:</b> <code>%s</code>\n➲ <b>Dst:</b> <code>%s</code>", src, dst), &tele.SendOptions{ParseMode: tele.ModeHTML})

		go func() {
			err := upload_utils.RcloneClone(src, dst)
			if err != nil {
				c.Bot().Edit(statusMsg, fmt.Sprintf("❌ <b>Clone Gagal:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
			} else {
				c.Bot().Edit(statusMsg, fmt.Sprintf("✅ <b>Clone Berhasil!</b>\n\n➲ <b>Src:</b> <code>%s</code>\n➲ <b>Dst:</b> <code>%s</code>", src, dst), &tele.SendOptions{ParseMode: tele.ModeHTML})
			}
		}()
		return nil
	})

	for _, cmd := range telegram_helper.BotCommands.CloneCommand {
		b.Handle(cmd, handleClone)
	}
}
