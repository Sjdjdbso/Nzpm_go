package modules

import (
	"go-mirror-bot/bot/helper/mirror_utils/status_utils"
	"go-mirror-bot/bot/helper/telegram_helper"

	tele "gopkg.in/telebot.v3"
)

func InitStatus(b *tele.Bot) {
	handleStatus := telegram_helper.AuthGuard(func(c tele.Context) error {
		inlineMarkup := &tele.ReplyMarkup{}
		btnRefresh := inlineMarkup.Data("🔄 Refresh", "refresh_status")
		inlineMarkup.Inline(inlineMarkup.Row(btnRefresh))
		return c.Send(status_utils.GetStatusMessage(), inlineMarkup, &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	for _, cmd := range telegram_helper.BotCommands.StatusCommand {
		b.Handle(cmd, handleStatus)
	}

	b.Handle(tele.OnCallback, func(c tele.Context) error {
		if c.Callback().Data == "refresh_status" {
			inlineMarkup := &tele.ReplyMarkup{}
			btnRefresh := inlineMarkup.Data("🔄 Refresh", "refresh_status")
			inlineMarkup.Inline(inlineMarkup.Row(btnRefresh))
			c.Edit(status_utils.GetStatusMessage(), inlineMarkup, &tele.SendOptions{ParseMode: tele.ModeHTML})
			return c.Respond(&tele.CallbackResponse{Text: "Status diperbarui!"})
		}
		return nil
	})
}
