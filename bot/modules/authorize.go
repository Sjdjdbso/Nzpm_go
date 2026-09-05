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
	parseTargetID := func(c tele.Context) int64 {
		args := c.Args()
		if len(args) > 0 {
			if id, err := strconv.ParseInt(args[0], 10, 64); err == nil {
				return id
			}
		}
		if c.Message().ReplyTo != nil && c.Message().ReplyTo.Sender != nil {
			return c.Message().ReplyTo.Sender.ID
		}
		return c.Chat().ID
	}

	// 1. /authorize, /a, /auth
	handleAuth := func(c tele.Context) error {
		if !bot.ConfigDict.IsOwnerOrSudo(c.Sender().ID) {
			return c.Send("⚠️ Hanya Owner atau Sudo yang dapat menggunakan perintah ini.")
		}
		targetID := parseTargetID(c)
		bot.ConfigDict.AuthorizeChat(targetID)
		return c.Send(fmt.Sprintf("✅ <b>ID <code>%d</code> berhasil diotorisasi!</b>", targetID), &tele.SendOptions{ParseMode: tele.ModeHTML})
	}
	for _, cmd := range telegram_helper.BotCommands.AuthorizeCommand {
		b.Handle(cmd, handleAuth)
	}

	// 2. /unauthorize, /ua, /unauth
	handleUnauth := func(c tele.Context) error {
		if !bot.ConfigDict.IsOwnerOrSudo(c.Sender().ID) {
			return c.Send("⚠️ Hanya Owner atau Sudo yang dapat menggunakan perintah ini.")
		}
		targetID := parseTargetID(c)
		bot.ConfigDict.UnauthorizeChat(targetID)
		return c.Send(fmt.Sprintf("🛑 <b>Otorisasi ID <code>%d</code> berhasil dicabut!</b>", targetID), &tele.SendOptions{ParseMode: tele.ModeHTML})
	}
	for _, cmd := range telegram_helper.BotCommands.UnAuthorizeCommand {
		b.Handle(cmd, handleUnauth)
	}

	// 3. /addsudo (Owner only)
	b.Handle(telegram_helper.BotCommands.AddSudoCommand, func(c tele.Context) error {
		if bot.ConfigDict.OwnerID != 0 && c.Sender().ID != bot.ConfigDict.OwnerID {
			return c.Send("⚠️ Hanya Pemilik Bot (Owner) yang dapat menambahkan pengguna Sudo!")
		}
		targetID := parseTargetID(c)
		bot.ConfigDict.AddSudoUser(targetID)
		return c.Send(fmt.Sprintf("👑 <b>Pengguna <code>%d</code> berhasil dipromosikan menjadi SUDO!</b>", targetID), &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	// 4. /rmsudo (Owner only)
	b.Handle(telegram_helper.BotCommands.RmSudoCommand, func(c tele.Context) error {
		if bot.ConfigDict.OwnerID != 0 && c.Sender().ID != bot.ConfigDict.OwnerID {
			return c.Send("⚠️ Hanya Pemilik Bot (Owner) yang dapat menghapus pengguna Sudo!")
		}
		targetID := parseTargetID(c)
		bot.ConfigDict.RemoveSudoUser(targetID)
		return c.Send(fmt.Sprintf("🛑 <b>Hak akses SUDO pengguna <code>%d</code> telah dicabut!</b>", targetID), &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	// 5. /blacklist, /bl
	handleBlacklist := func(c tele.Context) error {
		if !bot.ConfigDict.IsOwnerOrSudo(c.Sender().ID) {
			return c.Send("⚠️ Hanya Owner atau Sudo yang dapat mem-blacklist pengguna.")
		}
		targetID := parseTargetID(c)
		if bot.ConfigDict.OwnerID != 0 && targetID == bot.ConfigDict.OwnerID {
			return c.Send("❌ Tidak bisa mem-blacklist Owner!")
		}
		bot.ConfigDict.AddBlacklistUser(targetID)
		return c.Send(fmt.Sprintf("🚫 <b>Pengguna <code>%d</code> berhasil dimasukkan ke daftar BLACKLIST!</b>", targetID), &tele.SendOptions{ParseMode: tele.ModeHTML})
	}
	for _, cmd := range telegram_helper.BotCommands.AddBlackListCommand {
		b.Handle(cmd, handleBlacklist)
	}

	// 6. /rmblacklist, /rbl
	handleRmBlacklist := func(c tele.Context) error {
		if !bot.ConfigDict.IsOwnerOrSudo(c.Sender().ID) {
			return c.Send("⚠️ Hanya Owner atau Sudo yang dapat menghapus blacklist.")
		}
		targetID := parseTargetID(c)
		bot.ConfigDict.RemoveBlacklistUser(targetID)
		return c.Send(fmt.Sprintf("✅ <b>Pengguna <code>%d</code> telah dihapus dari BLACKLIST!</b>", targetID), &tele.SendOptions{ParseMode: tele.ModeHTML})
	}
	for _, cmd := range telegram_helper.BotCommands.RmBlackListCommand {
		b.Handle(cmd, handleRmBlacklist)
	}

	// 7. /authlist
	b.Handle(telegram_helper.BotCommands.AuthListCommand, func(c tele.Context) error {
		if !bot.ConfigDict.IsOwnerOrSudo(c.Sender().ID) {
			return c.Send("⚠️ Hanya Owner atau Sudo yang dapat melihat daftar otorisasi.")
		}
		authList := bot.ConfigDict.GetAllAuthorized()
		sudoList := bot.ConfigDict.GetAllSudo()
		blackList := bot.ConfigDict.GetAllBlacklist()

		var sb strings.Builder
		sb.WriteString("<b>📋 Ringkasan Otorisasi Bot:</b>\n\n")

		sb.WriteString(fmt.Sprintf("👑 <b>Owner:</b> <code>%d</code>\n\n", bot.ConfigDict.OwnerID))

		sb.WriteString("👥 <b>Sudo Users:</b>\n")
		if len(sudoList) == 0 {
			sb.WriteString("• <i>Tidak ada</i>\n")
		} else {
			for _, id := range sudoList {
				sb.WriteString(fmt.Sprintf("• <code>%d</code>\n", id))
			}
		}

		sb.WriteString("\n💬 <b>Authorized Chats:</b>\n")
		if len(authList) == 0 {
			sb.WriteString("• <i>Tidak ada</i>\n")
		} else {
			for _, id := range authList {
				sb.WriteString(fmt.Sprintf("• <code>%d</code>\n", id))
			}
		}

		sb.WriteString("\n🚫 <b>Blacklisted Users:</b>\n")
		if len(blackList) == 0 {
			sb.WriteString("• <i>Tidak ada</i>\n")
		} else {
			for _, id := range blackList {
				sb.WriteString(fmt.Sprintf("• <code>%d</code>\n", id))
			}
		}

		return c.Send(sb.String(), &tele.SendOptions{ParseMode: tele.ModeHTML})
	})
}
