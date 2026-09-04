package modules

import (
	"os"
	"os/exec"

	"go-mirror-bot/bot/helper/telegram_helper"

	tele "gopkg.in/telebot.v3"
)

func InitRestart(b *tele.Bot) {
	// /restart & /r (Owner/Sudo only)
	handleRestart := telegram_helper.SudoGuard(func(c tele.Context) error {
		_ = c.Send("🔄 <b>Memulai ulang bot...</b>\n<i>Mohon tunggu beberapa saat hingga bot aktif kembali.</i>", &tele.SendOptions{ParseMode: tele.ModeHTML})
		go func() {
			executable, err := os.Executable()
			if err != nil {
				executable = "./go-mirror-bot"
			}
			cmd := exec.Command(executable)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			_ = cmd.Start()
			os.Exit(0)
		}()
		return nil
	})

	// /log (Owner/Sudo only)
	handleLog := telegram_helper.SudoGuard(func(c tele.Context) error {
		candidates := []string{"log.txt", "bot.log"}
		var foundLog string
		for _, f := range candidates {
			if _, err := os.Stat(f); err == nil {
				foundLog = f
				break
			}
		}

		if foundLog != "" {
			doc := &tele.Document{
				File:     tele.FromDisk(foundLog),
				FileName: foundLog,
				Caption:  "📄 <b>File Log Bot</b>",
			}
			return c.Send(doc, &tele.SendOptions{ParseMode: tele.ModeHTML})
		}

		return c.Send("ℹ️ Log saat ini diarahkan ke stdout (Docker/Console).", &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	b.Handle("/restart", handleRestart)
	b.Handle("/r", handleRestart)
	b.Handle("/restartall", handleRestart)
	b.Handle("/log", handleLog)
}
