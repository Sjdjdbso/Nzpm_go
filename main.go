package main

import (
	"log"
	"time"

	"go-mirror-bot/bot"
	"go-mirror-bot/config"
	"go-mirror-bot/core"

	tele "gopkg.in/telebot.v3"
)

func main() {
	log.Println("==================================================")
	log.Println("🚀 Memulai Go-Mirror-Bot (Lightweight Koyeb Ready)")
	log.Println("==================================================")

	// 1. Muat Konfigurasi
	config.LoadConfig()

	// 2. Jalankan HTTP Health Check Server untuk Koyeb
	bot.StartHealthServer(config.AppConfig.Port)

	// 3. Inisialisasi Aria2 Client & Pastikan Daemon Aktif
	core.InitAriaClient("")
	if err := core.EnsureAria2Daemon(); err != nil {
		log.Printf("[WARN] Tidak dapat memulai Aria2 Daemon otomatis: %v", err)
		log.Println("[INFO] Pastikan Aria2c dijalankan via aria.sh jika berjalan di Docker.")
	}

	// 4. Inisialisasi Bot Telegram
	pref := tele.Settings{
		Token:  config.AppConfig.BotToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatalf("[FATAL] Gagal membuat client bot Telegram: %v", err)
	}

	// 5. Daftarkan Handler Bot
	bot.RegisterHandlers(b)

	log.Printf("[INFO] Bot @%s berhasil aktif dan siap menerima perintah!", b.Me.Username)
	b.Start()
}
