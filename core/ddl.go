package core

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var httpClient = &http.Client{
	Timeout: 15 * time.Second,
}

// ResolveDDL mendeteksi dan mengonversi link dari file hosting (GDrive, Mediafire, Pixeldrain, Dropbox, dll.) menjadi Direct Link
func ResolveDDL(rawURL string) (string, string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL, "", nil
	}

	host := strings.ToLower(u.Host)

	// 1. Google Drive (drive.google.com)
	if strings.Contains(host, "drive.google.com") {
		return resolveGoogleDrive(rawURL)
	}

	// 2. Mediafire (mediafire.com)
	if strings.Contains(host, "mediafire.com") {
		return resolveMediafire(rawURL)
	}

	// 3. Pixeldrain (pixeldrain.com)
	if strings.Contains(host, "pixeldrain.com") {
		return resolvePixeldrain(rawURL)
	}

	// 4. Dropbox (dropbox.com)
	if strings.Contains(host, "dropbox.com") {
		return resolveDropbox(rawURL)
	}

	// 5. Solidfiles (solidfiles.com)
	if strings.Contains(host, "solidfiles.com") {
		return resolveSolidfiles(rawURL)
	}

	// Fallback ke URL asli
	return rawURL, "", nil
}

// ── Resolver Google Drive ──────────────────────────────────────────────────
func resolveGoogleDrive(rawURL string) (string, string, error) {
	// Ambil File ID
	var fileID string
	if strings.Contains(rawURL, "/d/") {
		re := regexp.MustCompile(`/d/([a-zA-Z0-9_-]+)`)
		matches := re.FindStringSubmatch(rawURL)
		if len(matches) > 1 {
			fileID = matches[1]
		}
	} else if strings.Contains(rawURL, "id=") {
		u, _ := url.Parse(rawURL)
		fileID = u.Query().Get("id")
	}

	if fileID == "" {
		return rawURL, "", fmt.Errorf("tidak dapat menemukan Google Drive File ID")
	}

	// Gunakan endpoint direct download Google Drive
	directLink := fmt.Sprintf("https://drive.usercontent.google.com/download?id=%s&export=download&authuser=0&confirm=t", fileID)
	return directLink, "", nil
}

// ── Resolver Mediafire ─────────────────────────────────────────────────────
func resolveMediafire(rawURL string) (string, string, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return rawURL, "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	resp, err := httpClient.Do(req)
	if err != nil {
		return rawURL, "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return rawURL, "", err
	}
	html := string(bodyBytes)

	// Cari link di tombol download: href="https://download...mediafire.com/..."
	re := regexp.MustCompile(`aria-label="Download file"\s+href="([^"]+)"`)
	matches := re.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1], "", nil
	}

	// Regex fallback
	reFallback := regexp.MustCompile(`href="(https?://download[^"]+mediafire\.com/[^"]+)"`)
	matchesFallback := reFallback.FindStringSubmatch(html)
	if len(matchesFallback) > 1 {
		return matchesFallback[1], "", nil
	}

	return rawURL, "", fmt.Errorf("gagal mengekstrak direct link Mediafire")
}

// ── Resolver Pixeldrain ────────────────────────────────────────────────────
func resolvePixeldrain(rawURL string) (string, string, error) {
	re := regexp.MustCompile(`pixeldrain\.com/u/([a-zA-Z0-9]+)`)
	matches := re.FindStringSubmatch(rawURL)
	if len(matches) > 1 {
		id := matches[1]
		return fmt.Sprintf("https://pixeldrain.com/api/file/%s?download", id), "", nil
	}
	return rawURL, "", nil
}

// ── Resolver Dropbox ───────────────────────────────────────────────────────
func resolveDropbox(rawURL string) (string, string, error) {
	if strings.Contains(rawURL, "dl=0") {
		return strings.Replace(rawURL, "dl=0", "dl=1", 1), "", nil
	}
	if !strings.Contains(rawURL, "dl=1") {
		if strings.Contains(rawURL, "?") {
			return rawURL + "&dl=1", "", nil
		}
		return rawURL + "?dl=1", "", nil
	}
	return rawURL, "", nil
}

// ── Resolver Solidfiles ────────────────────────────────────────────────────
func resolveSolidfiles(rawURL string) (string, string, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return rawURL, "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return rawURL, "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	re := regexp.MustCompile(`"downloadUrl":"([^"]+)"`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) > 1 {
		return matches[1], "", nil
	}
	return rawURL, "", nil
}
