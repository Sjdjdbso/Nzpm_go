package config

import (
	"strconv"
	"strings"
	"sync"
)

type AuthManager struct {
	mu              sync.RWMutex
	authorizedChats map[int64]bool
	sudoUsers       map[int64]bool
}

var Auth = &AuthManager{
	authorizedChats: make(map[int64]bool),
	sudoUsers:       make(map[int64]bool),
}

// InitAuth memuat daftar auth awal dari environment variables
func InitAuth(authorizedChatsStr, sudoUsersStr string) {
	Auth.mu.Lock()
	defer Auth.mu.Unlock()

	// Parse AUTHORIZED_CHATS
	if authorizedChatsStr != "" {
		parts := strings.Fields(authorizedChatsStr)
		for _, p := range parts {
			if id, err := strconv.ParseInt(p, 10, 64); err == nil {
				Auth.authorizedChats[id] = true
			}
		}
	}

	// Parse SUDO_USERS
	if sudoUsersStr != "" {
		parts := strings.Fields(sudoUsersStr)
		for _, p := range parts {
			if id, err := strconv.ParseInt(p, 10, 64); err == nil {
				Auth.sudoUsers[id] = true
			}
		}
	}
}

// IsAuthorized mengecek apakah user atau chat diizinkan
func (a *AuthManager) IsAuthorized(userID, chatID int64) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// 1. Owner selalu diizinkan
	if AppConfig.OwnerID != 0 && userID == AppConfig.OwnerID {
		return true
	}

	// 2. Sudo user selalu diizinkan
	if a.sudoUsers[userID] {
		return true
	}

	// 3. Chat atau grup yang diotorisasi
	if a.authorizedChats[chatID] || a.authorizedChats[userID] {
		return true
	}

	// Jika OWNER_ID tidak diset dan belum ada auth, default izinkan untuk kemudahan setup awal
	if AppConfig.OwnerID == 0 && len(a.authorizedChats) == 0 {
		return true
	}

	return false
}

// IsOwnerOrSudo mengecek apakah user memiliki hak administratif
func (a *AuthManager) IsOwnerOrSudo(userID int64) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return (AppConfig.OwnerID != 0 && userID == AppConfig.OwnerID) || a.sudoUsers[userID]
}

func (a *AuthManager) AuthorizeChat(id int64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.authorizedChats[id] = true
}

func (a *AuthManager) UnauthorizeChat(id int64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.authorizedChats, id)
}

func (a *AuthManager) GetAllAuthorized() []int64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	list := make([]int64, 0, len(a.authorizedChats))
	for id := range a.authorizedChats {
		list = append(list, id)
	}
	return list
}
