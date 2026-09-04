package bot

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken        string
	OwnerID         int64
	AuthorizedChats map[int64]bool
	SudoUsers       map[int64]bool
	DownloadDir     string
	RclonePath      string
	CmdSuffix       string
	Port            string
	mu              sync.RWMutex
}

var ConfigDict Config

func LoadConfig() {
	if err := godotenv.Load("config.env"); err != nil {
		log.Println("[INFO] config.env tidak ditemukan, membaca Environment Variables sistem...")
	}

	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("[FATAL] BOT_TOKEN wajib diisi!")
	}

	ownerID, _ := strconv.ParseInt(os.Getenv("OWNER_ID"), 10, 64)

	downloadDir := os.Getenv("DOWNLOAD_DIR")
	if downloadDir == "" {
		downloadDir = "downloads"
	}
	absDir, err := filepath.Abs(downloadDir)
	if err == nil {
		downloadDir = absDir
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ConfigDict = Config{
		BotToken:        botToken,
		OwnerID:         ownerID,
		AuthorizedChats: make(map[int64]bool),
		SudoUsers:       make(map[int64]bool),
		DownloadDir:     downloadDir,
		RclonePath:      os.Getenv("RCLONE_PATH"),
		CmdSuffix:       os.Getenv("CMD_SUFFIX"),
		Port:            port,
	}

	// Parse AUTHORIZED_CHATS
	if chats := os.Getenv("AUTHORIZED_CHATS"); chats != "" {
		for _, c := range strings.Fields(chats) {
			if id, err := strconv.ParseInt(c, 10, 64); err == nil {
				ConfigDict.AuthorizedChats[id] = true
			}
		}
	}

	// Parse SUDO_USERS
	if sudos := os.Getenv("SUDO_USERS"); sudos != "" {
		for _, s := range strings.Fields(sudos) {
			if id, err := strconv.ParseInt(s, 10, 64); err == nil {
				ConfigDict.SudoUsers[id] = true
			}
		}
	}

	os.MkdirAll(downloadDir, 0755)
	log.Printf("[INFO] Config berhasil dimuat. OwnerID: %d, DownloadDir: %s, Port: %s", ownerID, downloadDir, port)
}

func (c *Config) IsAuthorized(userID, chatID int64) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.OwnerID != 0 && userID == c.OwnerID {
		return true
	}
	if c.SudoUsers[userID] {
		return true
	}
	if c.AuthorizedChats[chatID] || c.AuthorizedChats[userID] {
		return true
	}
	if c.OwnerID == 0 && len(c.AuthorizedChats) == 0 {
		return true
	}
	return false
}

func (c *Config) IsOwnerOrSudo(userID int64) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return (c.OwnerID != 0 && userID == c.OwnerID) || c.SudoUsers[userID]
}

func (c *Config) AuthorizeChat(id int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.AuthorizedChats[id] = true
}

func (c *Config) UnauthorizeChat(id int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.AuthorizedChats, id)
}

func (c *Config) GetAllAuthorized() []int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	list := make([]int64, 0, len(c.AuthorizedChats))
	for id := range c.AuthorizedChats {
		list = append(list, id)
	}
	return list
}
