package main

import (
	"log"
	"time"

	"go-mirror-bot/bot"
	"go-mirror-bot/bot/helper/mirror_utils/download_utils"
	"go-mirror-bot/bot/helper/telegram_helper"
	"go-mirror-bot/bot/modules"

	tele "gopkg.in/telebot.v3"
)

func main() {
	log.Println("==================================================")
	log.Printf("🚀 Memulai %s (%s) [Koyeb Ready]\n", bot.BotName, bot.Version)
	log.Println("==================================================")

	// 1. Muat Konfigurasi (bot/config.py)
	bot.LoadConfig()

	// 2. Jalankan Health Server & Aria2c (bot/startup.py)
	bot.LaunchHealthServer(bot.ConfigDict.Port)
	download_utils.InitAriaClient("")
	if err := bot.LaunchAria2c(); err != nil {
		log.Printf("[WARN] Gagal auto-start Aria2c: %v (Akan menggunakan aria.sh di Docker)", err)
	}

	// 3. Inisialisasi Bot Telegram
	pref := tele.Settings{
		Token:  bot.ConfigDict.BotToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatalf("[FATAL] Gagal membuat client bot: %v", err)
	}
	telegram_helper.LogWZGramStatus()

	// 4. Daftarkan Semua Modul (bot/modules/)
	modules.InitStats(b)
	modules.InitAuthorize(b)
	modules.InitCancel(b)
	modules.InitStatus(b)
	modules.InitClone(b)
	modules.InitMirrorLeech(b)
	modules.InitYtDlp(b)
	modules.InitShell(b)
	modules.InitSpeedtest(b)
	modules.InitMediaInfo(b)
	modules.InitUsersSettings(b)
	modules.InitBroadcast(b)
	modules.InitRestart(b)
	modules.InitBotSettings(b)

	log.Printf("[INFO] Bot @%s aktif dan siap menerima perintah!\n", b.Me.Username)
	b.Start()
}
