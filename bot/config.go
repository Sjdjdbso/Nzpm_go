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
	BotToken          string
	TelegramAPI       int64  // TELEGRAM_API / API_ID
	TelegramHash      string // TELEGRAM_HASH / API_HASH
	UserSessionString string // USER_SESSION_STRING
	OwnerID           int64
	AuthorizedChats   map[int64]bool
	SudoUsers         map[int64]bool
	DownloadDir       string
	RclonePath        string
	DefaultUpload     string // DEFAULT_UPLOAD ("rc", "ddl", "pixeldrain", "gd")
	PixeldrainAPI     string // PIXELDRAIN_API
	GdriveID          string // GDRIVE_ID
	IndexURL          string // INDEX_URL
	UseServiceAccounts bool   // USE_SERVICE_ACCOUNTS
	IsTeamDrive       bool   // IS_TEAM_DRIVE
	StopDuplicate     bool   // STOP_DUPLICATE
	DisableDriveLink  bool   // DISABLE_DRIVE_LINK
	GDInfo            string // GD_INFO
	CmdSuffix         string
	Port              string
	mu                sync.RWMutex
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

	// Telegram API ID & Hash (Persis wzv3 / wzgram)
	apiIDStr := os.Getenv("TELEGRAM_API")
	if apiIDStr == "" {
		apiIDStr = os.Getenv("API_ID")
	}
	telegramAPI, _ := strconv.ParseInt(apiIDStr, 10, 64)

	telegramHash := os.Getenv("TELEGRAM_HASH")
	if telegramHash == "" {
		telegramHash = os.Getenv("API_HASH")
	}

	userSession := os.Getenv("USER_SESSION_STRING")

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

	defaultUpload := strings.ToLower(os.Getenv("DEFAULT_UPLOAD"))
	if defaultUpload == "" {
		defaultUpload = "rc"
	}

	pixeldrainAPI := os.Getenv("PIXELDRAIN_API")
	gdriveID := os.Getenv("GDRIVE_ID")
	indexURL := os.Getenv("INDEX_URL")
	useServiceAccounts, _ := strconv.ParseBool(os.Getenv("USE_SERVICE_ACCOUNTS"))
	isTeamDrive, _ := strconv.ParseBool(os.Getenv("IS_TEAM_DRIVE"))
	stopDuplicate, _ := strconv.ParseBool(os.Getenv("STOP_DUPLICATE"))
	disableDriveLink, _ := strconv.ParseBool(os.Getenv("DISABLE_DRIVE_LINK"))
	gdInfo := os.Getenv("GD_INFO")
	if gdInfo == "" {
		gdInfo = "Uploaded by WZML-X Go"
	}

	ConfigDict = Config{
		BotToken:           botToken,
		TelegramAPI:        telegramAPI,
		TelegramHash:       telegramHash,
		UserSessionString:  userSession,
		OwnerID:            ownerID,
		AuthorizedChats:    make(map[int64]bool),
		SudoUsers:          make(map[int64]bool),
		DownloadDir:        downloadDir,
		RclonePath:         os.Getenv("RCLONE_PATH"),
		DefaultUpload:      defaultUpload,
		PixeldrainAPI:      pixeldrainAPI,
		GdriveID:           gdriveID,
		IndexURL:           indexURL,
		UseServiceAccounts: useServiceAccounts,
		IsTeamDrive:        isTeamDrive,
		StopDuplicate:      stopDuplicate,
		DisableDriveLink:   disableDriveLink,
		GDInfo:             gdInfo,
		CmdSuffix:          os.Getenv("CMD_SUFFIX"),
		Port:               port,
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
	log.Printf("[INFO] Config berhasil dimuat. OwnerID: %d, TelegramAPI: %d, DefaultUpload: %s, Port: %s",
		ownerID, telegramAPI, defaultUpload, port)
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
