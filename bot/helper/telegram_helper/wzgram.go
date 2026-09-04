package telegram_helper

import (
	"log"

	"go-mirror-bot/bot"
)

// WZGramClientInfo mewakili identitas client WZGram pada porting Go
type WZGramClientInfo struct {
	Framework string
	APIID     int64
	APIHash   string
	IsUserBot bool
}

// GetWZGramInfo mengembalikan status konfigurasi client Telegram (wzgram)
func GetWZGramInfo() WZGramClientInfo {
	return WZGramClientInfo{
		Framework: "WZGram-Go (1:1 WZML-X Engine)",
		APIID:     bot.ConfigDict.TelegramAPI,
		APIHash:   bot.ConfigDict.TelegramHash,
		IsUserBot: bot.ConfigDict.UserSessionString != "",
	}
}

func LogWZGramStatus() {
	info := GetWZGramInfo()
	log.Printf("[WZGRAM] Framework: %s", info.Framework)
	if info.APIID != 0 {
		log.Printf("[WZGRAM] Telegram API ID: %d | Hash: [Tersedia]", info.APIID)
	} else {
		log.Printf("[WZGRAM] Menggunakan Bot API Client (Telebot)")
	}
	if info.IsUserBot {
		log.Printf("[WZGRAM] User session string terdeteksi.")
	}
}
