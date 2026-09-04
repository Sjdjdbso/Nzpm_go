package telegram_helper

import (
	"fmt"

	"go-mirror-bot/bot"

	tele "gopkg.in/telebot.v3"
)

func AuthGuard(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		if !bot.ConfigDict.IsAuthorized(c.Sender().ID, c.Chat().ID) {
			return c.Send("<i>You Are not authorized user! Deploy your own WZML-X Mirror-Leech bot</i>\n\n"+
				fmt.Sprintf("ℹ️ <b>Chat ID:</b> <code>%d</code> | <b>User ID:</b> <code>%d</code>", c.Chat().ID, c.Sender().ID),
				&tele.SendOptions{ParseMode: tele.ModeHTML})
		}
		return next(c)
	}
}

func SudoGuard(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		if !bot.ConfigDict.IsOwnerOrSudo(c.Sender().ID) {
			return c.Send("⚠️ <i>Perintah ini hanya dapat dijalankan oleh Bot Owner atau Sudo user!</i>",
				&tele.SendOptions{ParseMode: tele.ModeHTML})
		}
		return next(c)
	}
}
