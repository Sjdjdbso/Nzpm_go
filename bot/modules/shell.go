package modules

import (
	"fmt"
	"os/exec"
	"strings"

	"go-mirror-bot/bot"
	"go-mirror-bot/bot/helper/telegram_helper"

	tele "gopkg.in/telebot.v3"
)

func InitShell(b *tele.Bot) {
	b.Handle(telegram_helper.BotCommands.ShellCommand, func(c tele.Context) error {
		if !bot.ConfigDict.IsOwnerOrSudo(c.Sender().ID) {
			return c.Send("⚠️ Hanya Owner yang diizinkan mengeksekusi shell.")
		}
		cmdStr := strings.TrimSpace(strings.TrimPrefix(c.Text(), "/shell"))
		if cmdStr == "" {
			return c.Send("⚠️ Masukkan perintah shell. Contoh: <code>/shell df -h</code>", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}
		out, err := exec.Command("bash", "-c", cmdStr).CombinedOutput()
		resText := string(out)
		if err != nil {
			resText += fmt.Sprintf("\n[Error: %v]", err)
		}
		if len(resText) > 4000 {
			resText = resText[:4000] + "\n...(terpotong)"
		}
		return c.Send(fmt.Sprintf("<b>Terminal Output:</b>\n<pre>%s</pre>", resText), &tele.SendOptions{ParseMode: tele.ModeHTML})
	})
}
