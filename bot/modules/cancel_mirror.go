package modules

import (
	"fmt"
	"strings"

	"go-mirror-bot/bot"
	"go-mirror-bot/bot/helper/ext_utils"
	"go-mirror-bot/bot/helper/mirror_utils/download_utils"
	"go-mirror-bot/bot/helper/telegram_helper"

	tele "gopkg.in/telebot.v3"
)

func InitCancel(b *tele.Bot) {
	handleCancel := telegram_helper.AuthGuard(func(c tele.Context) error {
		args := c.Args()
		if len(args) == 0 {
			return c.Send("⚠️ Masukkan GID tugas. Contoh: <code>/cancel 2089b05ecca3d829</code>", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}
		gid := args[0]
		download_utils.Aria.Remove(gid)
		ext_utils.TaskMgr.Remove(gid)
		return c.Send(fmt.Sprintf("🛑 Berhasil membatalkan unduhan GID: <code>%s</code>", gid), &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	for _, cmd := range telegram_helper.BotCommands.CancelCommand {
		b.Handle(cmd, handleCancel)
	}

	b.Handle(telegram_helper.BotCommands.CancelAllCommand, telegram_helper.AuthGuard(func(c tele.Context) error {
		if !bot.ConfigDict.IsOwnerOrSudo(c.Sender().ID) {
			return c.Send("⚠️ Hanya Owner atau Sudo yang dapat membatalkan semua proses.")
		}
		tasks := ext_utils.TaskMgr.All()
		for _, t := range tasks {
			download_utils.Aria.Remove(t.GID)
			ext_utils.TaskMgr.Remove(t.GID)
		}
		return c.Send(fmt.Sprintf("🛑 <b>Berhasil membatalkan %d proses aktif.</b>", len(tasks)), &tele.SendOptions{ParseMode: tele.ModeHTML})
	}))

	b.Handle(tele.OnCallback, func(c tele.Context) error {
		data := c.Callback().Data
		if strings.HasPrefix(data, "cancel_") {
			gid := strings.TrimPrefix(data, "cancel_")
			t := ext_utils.TaskMgr.Get(gid)
			if t != nil {
				sender := c.Sender().Username
				if sender == "" {
					sender = c.Sender().FirstName
				}
				download_utils.Aria.Remove(gid)
				ext_utils.TaskMgr.Remove(gid)
				c.Respond(&tele.CallbackResponse{Text: "Unduhan dibatalkan!"})
				return c.Edit(fmt.Sprintf("🛑 <b>Unduhan Dibatalkan!</b>\n🆔 GID: <code>%s</code>\n👤 Dibatalkan oleh: @%s", gid, sender), &tele.SendOptions{ParseMode: tele.ModeHTML})
			}
			return c.Respond(&tele.CallbackResponse{Text: "Tugas sudah selesai atau tidak ditemukan."})
		}
		return nil
	})
}
