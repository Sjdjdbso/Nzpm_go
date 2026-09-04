package modules

import (
	"fmt"
	"strconv"
	"strings"

	"go-mirror-bot/bot"
	"go-mirror-bot/bot/helper/telegram_helper"

	tele "gopkg.in/telebot.v3"
)

func InitAuthorize(b *tele.Bot) {
	handleAuth := func(c tele.Context) error {
		if !bot.ConfigDict.IsOwnerOrSudo(c.Sender().ID) {
			return c.Send("⚠️ Hanya Owner atau Sudo yang dapat menggunakan perintah ini.")
		}
		var targetID int64
		args := c.Args()
		if len(args) > 0 {
			if id, err := strconv.ParseInt(args[0], 10, 64); err == nil {
				targetID = id
			}
		} else if c.Message().ReplyTo != nil && c.Message().ReplyTo.Sender != nil {
			targetID = c.Message().ReplyTo.Sender.ID
		} else {
			targetID = c.Chat().ID
		}

		bot.ConfigDict.AuthorizeChat(targetID)
		return c.Send(fmt.Sprintf("✅ <b>ID <code>%d</code> berhasil diotorisasi!</b>", targetID), &tele.SendOptions{ParseMode: tele.ModeHTML})
	}

	for _, cmd := range telegram_helper.BotCommands.AuthorizeCommand {
		b.Handle(cmd, handleAuth)
	}

	handleUnauth := func(c tele.Context) error {
		if !bot.ConfigDict.IsOwnerOrSudo(c.Sender().ID) {
			return c.Send("⚠️ Hanya Owner atau Sudo yang dapat menggunakan perintah ini.")
		}
		var targetID int64
		args := c.Args()
		if len(args) > 0 {
			if id, err := strconv.ParseInt(args[0], 10, 64); err == nil {
				targetID = id
			}
		} else if c.Message().ReplyTo != nil && c.Message().ReplyTo.Sender != nil {
			targetID = c.Message().ReplyTo.Sender.ID
		} else {
			targetID = c.Chat().ID
		}

		bot.ConfigDict.UnauthorizeChat(targetID)
		return c.Send(fmt.Sprintf("🛑 <b>Otorisasi ID <code>%d</code> berhasil dicabut!</b>", targetID), &tele.SendOptions{ParseMode: tele.ModeHTML})
	}

	for _, cmd := range telegram_helper.BotCommands.UnAuthorizeCommand {
		b.Handle(cmd, handleUnauth)
	}

	b.Handle(telegram_helper.BotCommands.AuthListCommand, func(c tele.Context) error {
		if !bot.ConfigDict.IsOwnerOrSudo(c.Sender().ID) {
			return c.Send("⚠️ Hanya Owner atau Sudo yang dapat melihat daftar otorisasi.")
		}
		list := bot.ConfigDict.GetAllAuthorized()
		var sb strings.Builder
		sb.WriteString("<b>📋 Daftar ID Terotorisasi:</b>\n")
		for _, id := range list {
			sb.WriteString(fmt.Sprintf("• <code>%d</code>\n", id))
		}
		return c.Send(sb.String(), &tele.SendOptions{ParseMode: tele.ModeHTML})
	})
}
