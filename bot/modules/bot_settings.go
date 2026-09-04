package modules

import (
	"fmt"

	"go-mirror-bot/bot"
	"go-mirror-bot/bot/helper/telegram_helper"

	tele "gopkg.in/telebot.v3"
)

func InitBotSettings(b *tele.Bot) {
	handleBotSettings := telegram_helper.SudoGuard(func(c tele.Context) error {
		rclonePath := bot.ConfigDict.RclonePath
		if rclonePath == "" {
			rclonePath = "❌ Belum disetel"
		} else {
			rclonePath = fmt.Sprintf("<code>%s</code>", rclonePath)
		}

		cmdSuffix := bot.ConfigDict.CmdSuffix
		if cmdSuffix == "" {
			cmdSuffix = "❌ Tidak ada"
		} else {
			cmdSuffix = fmt.Sprintf("<code>%s</code>", cmdSuffix)
		}

		text := fmt.Sprintf(
			"⚙️ <b><i>BOT SETTINGS &amp; CONFIGURATION</i></b>\n\n"+
				"┠ <b>Bot Name:</b> <code>%s</code>\n"+
				"┠ <b>Version:</b> <code>%s</code>\n"+
				"┠ <b>Owner ID:</b> <code>%d</code>\n"+
				"┠ <b>Download Dir:</b> <code>%s</code>\n"+
				"┠ <b>Port Health Check:</b> <code>%s</code>\n"+
				"┠ <b>Rclone Path:</b> %s\n"+
				"┠ <b>Command Suffix:</b> %s\n"+
				"┠ <b>Total Authorized:</b> <code>%d</code>\n"+
				"┖ <b>Total Sudo Users:</b> <code>%d</code>\n\n"+
				"💡 <i>Gunakan file <code>config.env</code> untuk mengubah konfigurasi permanen.</i>",
			bot.BotName, bot.Version, bot.ConfigDict.OwnerID,
			bot.ConfigDict.DownloadDir, bot.ConfigDict.Port,
			rclonePath, cmdSuffix,
			len(bot.ConfigDict.AuthorizedChats), len(bot.ConfigDict.SudoUsers),
		)

		markup := &tele.ReplyMarkup{}
		btnClose := markup.Data("❌ Tutup", "bs_close")
		markup.Inline(markup.Row(btnClose))

		return c.Send(text, markup, &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	b.Handle(&tele.Btn{Unique: "bs_close"}, func(c tele.Context) error {
		return c.Delete()
	})

	b.Handle("/bsetting", handleBotSettings)
	b.Handle("/bs", handleBotSettings)
}
