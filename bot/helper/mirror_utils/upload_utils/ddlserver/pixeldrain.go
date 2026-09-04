package ddlserver

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"go-mirror-bot/bot/helper/ext_utils"
)

type PixeldrainResponse struct {
	ID      string `json:"id"`
	Success *bool  `json:"success,omitempty"`
	Message string `json:"message,omitempty"`
	Name    string `json:"name,omitempty"`
	Size    int64  `json:"size,omitempty"`
}

type Pixeldrain struct {
	APIKey string
	Client *http.Client
}

func NewPixeldrain(apiKey string) *Pixeldrain {
	return &Pixeldrain{
		APIKey: apiKey,
		Client: &http.Client{Timeout: 30 * time.Minute},
	}
}

// IsPdApi memvalidasi API Key Pixeldrain
func (p *Pixeldrain) IsPdApi(apiKey string) bool {
	if apiKey == "" {
		return false
	}
	req, err := http.NewRequest("GET", "https://pixeldrain.com/api/user", nil)
	if err != nil {
		return false
	}
	token := base64.StdEncoding.EncodeToString([]byte(":" + apiKey))
	req.Header.Set("Authorization", "Basic "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		var res map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&res); err == nil {
			if _, ok := res["id"]; ok {
				return true
			}
			if canUpload, ok := res["can_upload"].(bool); ok && canUpload {
				return true
			}
		}
	}
	return false
}

// Upload mengunggah file atau folder ke Pixeldrain
func (p *Pixeldrain) Upload(filePath string) (string, error) {
	fi, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("file tidak ditemukan: %w", err)
	}

	targetFile := filePath
	needCleanup := false
	if fi.IsDir() {
		zipPath, err := ext_utils.CompressToZip(filePath)
		if err != nil {
			return "", fmt.Errorf("gagal kompres folder untuk pixeldrain: %w", err)
		}
		targetFile = zipPath
		needCleanup = true
	}
	defer func() {
		if needCleanup {
			os.Remove(targetFile)
		}
	}()

	f, err := os.Open(targetFile)
	if err != nil {
		return "", fmt.Errorf("gagal membuka file: %w", err)
	}
	defer f.Close()

	fileName := filepath.Base(targetFile)
	encodedName := url.PathEscape(fileName)
	apiURL := fmt.Sprintf("https://pixeldrain.com/api/file/%s", encodedName)

	req, err := http.NewRequest("PUT", apiURL, f)
	if err != nil {
		return "", fmt.Errorf("gagal membuat request: %w", err)
	}

	if p.APIKey != "" {
		token := base64.StdEncoding.EncodeToString([]byte(":" + p.APIKey))
		req.Header.Set("Authorization", "Basic "+token)
	}

	resp, err := p.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("gagal upload ke Pixeldrain: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("gagal membaca respon Pixeldrain: %w", err)
	}

	var res PixeldrainResponse
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return "", fmt.Errorf("respon bukan JSON: %s", string(bodyBytes))
	}

	if res.Success != nil && !*res.Success {
		return "", fmt.Errorf("pixeldrain error: %s", res.Message)
	}

	if res.ID == "" {
		return "", fmt.Errorf("pixeldrain tidak mengembalikan ID: %s", string(bodyBytes))
	}

	return fmt.Sprintf("https://pixeldrain.com/u/%s", res.ID), nil
}
