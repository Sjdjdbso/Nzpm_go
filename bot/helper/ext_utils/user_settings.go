package ext_utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type UserConfig struct {
	UserID        int64  `json:"user_id"`
	CustomCaption string `json:"custom_caption"`
	LeechPrefix   string `json:"leech_prefix"`
	LeechSuffix   string `json:"leech_suffix"`
	HasThumbnail  bool   `json:"has_thumbnail"`
	PixeldrainAPI string `json:"pixeldrain_api"`
}

type UserSettingsStore struct {
	sync.RWMutex
	Users    map[int64]*UserConfig `json:"users"`
	filePath string
}

var UserStore = &UserSettingsStore{
	Users:    make(map[int64]*UserConfig),
	filePath: "users_settings.json",
}

func init() {
	UserStore.load()
}

func (s *UserSettingsStore) load() {
	s.Lock()
	defer s.Unlock()

	_ = os.MkdirAll("thumbnails", 0755)

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return
	}

	var loaded map[int64]*UserConfig
	if err := json.Unmarshal(data, &loaded); err == nil {
		s.Users = loaded
	}
}

func (s *UserSettingsStore) save() {
	data, err := json.MarshalIndent(s.Users, "", "  ")
	if err != nil {
		log.Printf("[ERROR] Gagal marshal user settings: %v", err)
		return
	}
	_ = os.WriteFile(s.filePath, data, 0644)
}

func (s *UserSettingsStore) Get(userID int64) *UserConfig {
	s.RLock()
	defer s.RUnlock()

	if u, ok := s.Users[userID]; ok {
		thumbPath := filepath.Join("thumbnails", fmt.Sprintf("%d.jpg", userID))
		if _, err := os.Stat(thumbPath); err == nil {
			u.HasThumbnail = true
		} else {
			u.HasThumbnail = false
		}
		return u
	}

	thumbPath := filepath.Join("thumbnails", fmt.Sprintf("%d.jpg", userID))
	hasThumb := false
	if _, err := os.Stat(thumbPath); err == nil {
		hasThumb = true
	}

	return &UserConfig{
		UserID:       userID,
		HasThumbnail: hasThumb,
	}
}

func (s *UserSettingsStore) SetCaption(userID int64, caption string) {
	s.Lock()
	defer s.Unlock()

	u, ok := s.Users[userID]
	if !ok {
		u = &UserConfig{UserID: userID}
		s.Users[userID] = u
	}
	u.CustomCaption = caption
	s.save()
}

func (s *UserSettingsStore) SetPrefix(userID int64, prefix string) {
	s.Lock()
	defer s.Unlock()

	u, ok := s.Users[userID]
	if !ok {
		u = &UserConfig{UserID: userID}
		s.Users[userID] = u
	}
	u.LeechPrefix = prefix
	s.save()
}

func (s *UserSettingsStore) SetSuffix(userID int64, suffix string) {
	s.Lock()
	defer s.Unlock()

	u, ok := s.Users[userID]
	if !ok {
		u = &UserConfig{UserID: userID}
		s.Users[userID] = u
	}
	u.LeechSuffix = suffix
	s.save()
}

func (s *UserSettingsStore) SetPixeldrainAPI(userID int64, apiKey string) {
	s.Lock()
	defer s.Unlock()

	u, ok := s.Users[userID]
	if !ok {
		u = &UserConfig{UserID: userID}
		s.Users[userID] = u
	}
	u.PixeldrainAPI = apiKey
	s.save()
}

func (s *UserSettingsStore) SaveThumbnail(userID int64, src io.Reader) error {
	s.Lock()
	defer s.Unlock()

	_ = os.MkdirAll("thumbnails", 0755)
	destPath := filepath.Join("thumbnails", fmt.Sprintf("%d.jpg", userID))

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, src); err != nil {
		return err
	}

	u, ok := s.Users[userID]
	if !ok {
		u = &UserConfig{UserID: userID}
		s.Users[userID] = u
	}
	u.HasThumbnail = true
	s.save()
	return nil
}

func (s *UserSettingsStore) DeleteThumbnail(userID int64) error {
	s.Lock()
	defer s.Unlock()

	destPath := filepath.Join("thumbnails", fmt.Sprintf("%d.jpg", userID))
	_ = os.Remove(destPath)

	if u, ok := s.Users[userID]; ok {
		u.HasThumbnail = false
		s.save()
	}
	return nil
}

func (s *UserSettingsStore) GetThumbnailPath(userID int64) string {
	destPath := filepath.Join("thumbnails", fmt.Sprintf("%d.jpg", userID))
	if _, err := os.Stat(destPath); err == nil {
		return destPath
	}
	return ""
}
