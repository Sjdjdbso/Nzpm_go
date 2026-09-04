package modules

import (
	"fmt"

	"go-mirror-bot/bot/helper/mirror_utils/upload_utils"
	"go-mirror-bot/bot/helper/telegram_helper"

	tele "gopkg.in/telebot.v3"
)

func InitClone(b *tele.Bot) {
	handleClone := telegram_helper.AuthGuard(func(c tele.Context) error {
		args := c.Args()
		if len(args) < 2 {
			return c.Send("⚠️ Format salah! Gunakan:\n<code>/clone &lt;source:path&gt; &lt;dest:path&gt;</code>\n\nContoh: <code>/clone gdrive:FolderA gdrive:FolderB</code>", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}
		src, dst := args[0], args[1]
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

	b.Handle(telegram_helper.BotCommands.CountCommand, telegram_helper.AuthGuard(func(c tele.Context) error {
		args := c.Args()
		if len(args) == 0 {
			return c.Send("⚠️ Format salah! Gunakan:\n<code>/count &lt;remote:path&gt;</code>", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}
		remote := args[0]
		statusMsg, _ := c.Bot().Send(c.Recipient(), fmt.Sprintf("🔍 <i>Menghitung isi remote <code>%s</code>...</i>", remote), &tele.SendOptions{ParseMode: tele.ModeHTML})
		out, err := upload_utils.RcloneCount(remote)
		if err != nil {
			_, err = c.Bot().Edit(statusMsg, fmt.Sprintf("❌ <b>Gagal:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
			return err
		}
		_, err = c.Bot().Edit(statusMsg, fmt.Sprintf("📊 <b>Hasil Count Remote:</b>\n<code>%s</code>\n\n<pre>%s</pre>", remote, out), &tele.SendOptions{ParseMode: tele.ModeHTML})
		return err
	}))
}
