package download_utils

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
	Timeout: 20 * time.Second,
}

// DirectLinkGenerator mendeteksi dan mengonversi link hoster menjadi Direct Download Link
func DirectLinkGenerator(rawURL string) (string, string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL, "", nil
	}

	host := strings.ToLower(u.Host)

	// Google Drive
	if strings.Contains(host, "drive.google.com") || strings.Contains(host, "docs.google.com") {
		var fileID string
		if strings.Contains(rawURL, "/d/") {
			re := regexp.MustCompile(`/d/([a-zA-Z0-9_-]+)`)
			m := re.FindStringSubmatch(rawURL)
			if len(m) > 1 {
				fileID = m[1]
			}
		} else if strings.Contains(rawURL, "id=") {
			fileID = u.Query().Get("id")
		}

		if fileID != "" {
			directURL, fileName, gdErr := resolveGoogleDrive(fileID)
			if gdErr != nil {
				return "", "", gdErr
			}
			return directURL, fileName, nil
		}
	}

	// Mediafire
	if strings.Contains(host, "mediafire.com") {
		req, err := http.NewRequest("GET", rawURL, nil)
		if err == nil {
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
			if resp, err := httpClient.Do(req); err == nil {
				defer resp.Body.Close()
				bodyBytes, _ := io.ReadAll(resp.Body)
				html := string(bodyBytes)
				re := regexp.MustCompile(`aria-label="Download file"\s+href="([^"]+)"`)
				if m := re.FindStringSubmatch(html); len(m) > 1 {
					return m[1], "", nil
				}
				reFallback := regexp.MustCompile(`href="(https?://download[^"]+mediafire\.com/[^"]+)"`)
				if m := reFallback.FindStringSubmatch(html); len(m) > 1 {
					return m[1], "", nil
				}
			}
		}
	}

	// Pixeldrain (File /u/ & List /l/)
	if strings.Contains(host, "pixeldrain.com") {
		reList := regexp.MustCompile(`pixeldrain\.com/l/([a-zA-Z0-9]+)`)
		if m := reList.FindStringSubmatch(rawURL); len(m) > 1 {
			return fmt.Sprintf("https://pixeldrain.com/api/list/%s/zip", m[1]), fmt.Sprintf("pixeldrain_list_%s.zip", m[1]), nil
		}
		reFile := regexp.MustCompile(`pixeldrain\.com/u/([a-zA-Z0-9]+)`)
		if m := reFile.FindStringSubmatch(rawURL); len(m) > 1 {
			return fmt.Sprintf("https://pixeldrain.com/api/file/%s?download", m[1]), "", nil
		}
	}

	// Dropbox
	if strings.Contains(host, "dropbox.com") {
		if strings.Contains(rawURL, "dl=0") {
			return strings.Replace(rawURL, "dl=0", "dl=1", 1), "", nil
		}
		if !strings.Contains(rawURL, "dl=1") {
			if strings.Contains(rawURL, "?") {
				return rawURL + "&dl=1", "", nil
			}
			return rawURL + "?dl=1", "", nil
		}
	}

	// Solidfiles
	if strings.Contains(host, "solidfiles.com") {
		req, err := http.NewRequest("GET", rawURL, nil)
		if err == nil {
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
			if resp, err := httpClient.Do(req); err == nil {
				defer resp.Body.Close()
				bodyBytes, _ := io.ReadAll(resp.Body)
				re := regexp.MustCompile(`"downloadUrl":"([^"]+)"`)
				if m := re.FindStringSubmatch(string(bodyBytes)); len(m) > 1 {
					return m[1], "", nil
				}
			}
		}
	}

	return rawURL, "", nil
}

func resolveGoogleDrive(fileID string) (string, string, error) {
	initURL := fmt.Sprintf("https://drive.google.com/uc?id=%s&export=download", fileID)
	req, err := http.NewRequest("GET", initURL, nil)
	if err != nil {
		return fmt.Sprintf("https://drive.usercontent.google.com/download?id=%s&export=download&confirm=t", fileID), "", nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Sprintf("https://drive.usercontent.google.com/download?id=%s&export=download&confirm=t", fileID), "", nil
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	disposition := resp.Header.Get("Content-Disposition")
	var fileName string
	if disposition != "" {
		reName := regexp.MustCompile(`filename="?([^";]+)"?`)
		if m := reName.FindStringSubmatch(disposition); len(m) > 1 {
			fileName = m[1]
		}
	}

	// Jika langsung stream file biner
	if !strings.Contains(contentType, "text/html") && resp.StatusCode == 200 {
		return resp.Request.URL.String(), fileName, nil
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return "", "", err
	}
	html := string(bodyBytes)

	// Cek jika kuota unduhan habis (Quota exceeded)
	if strings.Contains(html, "Quota exceeded") || strings.Contains(html, "Too many users have viewed or downloaded this file recently") {
		return "", "", fmt.Errorf("Google Drive Download Quota Exceeded! File ini sedang dibatasi oleh Google karena melebihi batas kuota unduhan publik harian. Silakan coba lagi nanti atau gunakan link alternatif.")
	}

	// Cek jika butuh login / akun privat
	if strings.Contains(html, "accounts.google.com/signin") || strings.Contains(html, "Access denied") {
		return "", "", fmt.Errorf("Google Drive Access Denied! File ini bersifat privat dan memerlukan izin akses. Pastikan file dibagikan publik (Anyone with the link).")
	}

	// Ekstrak nama file dari teks halaman konfirmasi jika belum ada
	if fileName == "" {
		reNameInHtml := regexp.MustCompile(`<span class="uc-name-size"><a[^>]*>([^<]+)</a>`)
		if m := reNameInHtml.FindStringSubmatch(html); len(m) > 1 {
			fileName = strings.TrimSpace(m[1])
		}
	}

	// Ekstrak form UUID konfirmasi Google Drive
	reUUID := regexp.MustCompile(`name="uuid"\s+value="([^"]+)"`)
	mUUID := reUUID.FindStringSubmatch(html)
	if len(mUUID) > 1 {
		uuid := mUUID[1]
		finalURL := fmt.Sprintf("https://drive.usercontent.google.com/download?id=%s&export=download&confirm=t&uuid=%s", fileID, uuid)
		return finalURL, fileName, nil
	}

	// Fallback token confirm lama
	reConfirm := regexp.MustCompile(`confirm=([0-9A-Za-z_-]+)`)
	if m := reConfirm.FindStringSubmatch(html); len(m) > 1 {
		token := m[1]
		finalURL := fmt.Sprintf("https://drive.usercontent.google.com/download?id=%s&export=download&confirm=%s", fileID, token)
		return finalURL, fileName, nil
	}

	return fmt.Sprintf("https://drive.usercontent.google.com/download?id=%s&export=download&confirm=t", fileID), fileName, nil
}
