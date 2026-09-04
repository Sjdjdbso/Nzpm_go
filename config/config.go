package config

import (
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken        string
	OwnerID         int64
	AuthorizedChats string
	SudoUsers       string
	DownloadDir     string
	RclonePath      string
	Port            string
}

var AppConfig Config

func LoadConfig() {
	if err := godotenv.Load("config.env"); err != nil {
		log.Println("[INFO] File config.env tidak ditemukan, membaca Environment Variables sistem (Koyeb/Docker)...")
	}

	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("[FATAL] BOT_TOKEN wajib diisi!")
	}

	ownerIDStr := os.Getenv("OWNER_ID")
	var ownerID int64
	if ownerIDStr != "" {
		parsed, err := strconv.ParseInt(ownerIDStr, 10, 64)
		if err == nil {
			ownerID = parsed
		}
	}

	authChats := os.Getenv("AUTHORIZED_CHATS")
	sudoUsers := os.Getenv("SUDO_USERS")

	downloadDir := os.Getenv("DOWNLOAD_DIR")
	if downloadDir == "" {
		downloadDir = "downloads"
	}

	absDownloadDir, err := filepath.Abs(downloadDir)
	if err == nil {
		downloadDir = absDownloadDir
	}

	rclonePath := os.Getenv("RCLONE_PATH")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	AppConfig = Config{
		BotToken:        botToken,
		OwnerID:         ownerID,
		AuthorizedChats: authChats,
		SudoUsers:       sudoUsers,
		DownloadDir:     downloadDir,
		RclonePath:      rclonePath,
		Port:            port,
	}

	// Inisialisasi daftar otorisasi
	InitAuth(authChats, sudoUsers)

	// Pastikan folder download ada
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		log.Fatalf("[FATAL] Gagal membuat direktori download: %v", err)
	}

	log.Printf("[INFO] Config berhasil dimuat. OwnerID: %d, DownloadDir: %s, Koyeb Port: %s", ownerID, downloadDir, port)
}
