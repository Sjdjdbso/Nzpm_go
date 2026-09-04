package modules

import (
	"fmt"
	"strings"
	"time"

	"go-mirror-bot/bot"
	"go-mirror-bot/bot/helper/telegram_helper"

	tele "gopkg.in/telebot.v3"
)

func InitBroadcast(b *tele.Bot) {
	handleBroadcast := telegram_helper.SudoGuard(func(c tele.Context) error {
		var msgToBroadcast string
		var photo *tele.Photo
		var doc *tele.Document

		reply := c.Message().ReplyTo
		if reply != nil {
			if reply.Text != "" {
				msgToBroadcast = reply.Text
			} else if reply.Caption != "" {
				msgToBroadcast = reply.Caption
			}
			if reply.Photo != nil {
				photo = reply.Photo
			}
			if reply.Document != nil {
				doc = reply.Document
			}
		} else {
			msgToBroadcast = strings.TrimSpace(strings.TrimPrefix(c.Message().Text, c.Message().Payload))
			if len(c.Args()) > 0 {
				msgToBroadcast = strings.Join(c.Args(), " ")
			}
		}

		if msgToBroadcast == "" && photo == nil && doc == nil {
			return c.Send("⚠️ Format: <code>/broadcast [pesan]</code> atau reply pesan/foto yang ingin disiarkan ke semua user yang diauthorisasi.", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}

		statusMsg, _ := c.Bot().Send(c.Recipient(), "📢 <i>Memulai siaran pesan...</i>", &tele.SendOptions{ParseMode: tele.ModeHTML})

		recipients := bot.ConfigDict.GetAllAuthorized()

		go func() {
			success := 0
			failed := 0

			for _, uid := range recipients {
				userTarget := &tele.User{ID: uid}
				var err error
				if photo != nil {
					p := &tele.Photo{File: photo.File, Caption: msgToBroadcast}
					_, err = c.Bot().Send(userTarget, p, &tele.SendOptions{ParseMode: tele.ModeHTML})
				} else if doc != nil {
					d := &tele.Document{File: doc.File, Caption: msgToBroadcast, FileName: doc.FileName}
					_, err = c.Bot().Send(userTarget, d, &tele.SendOptions{ParseMode: tele.ModeHTML})
				} else {
					_, err = c.Bot().Send(userTarget, msgToBroadcast, &tele.SendOptions{ParseMode: tele.ModeHTML})
				}

				if err == nil {
					success++
				} else {
					failed++
				}
				time.Sleep(100 * time.Millisecond) // hindari flood wait
			}

			report := fmt.Sprintf(
				"📢 <b><i>HASIL SIARAN (BROADCAST)</i></b>\n"+
					"┠ <b>Total Target:</b> <code>%d</code>\n"+
					"┠ <b>Berhasil Terkirim:</b> <code>%d</code>\n"+
					"┖ <b>Gagal/Blocked:</b> <code>%d</code>",
				len(recipients), success, failed,
			)
			c.Bot().Edit(statusMsg, report, &tele.SendOptions{ParseMode: tele.ModeHTML})
		}()

		return nil
	})

	b.Handle("/broadcast", handleBroadcast)
	b.Handle("/bc", handleBroadcast)
}
