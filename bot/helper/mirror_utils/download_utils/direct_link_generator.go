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
	Timeout: 15 * time.Second,
}

// DirectLinkGenerator mendeteksi dan mengonversi link hoster menjadi Direct Download Link
func DirectLinkGenerator(rawURL string) (string, string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL, "", nil
	}

	host := strings.ToLower(u.Host)

	// Google Drive
	if strings.Contains(host, "drive.google.com") {
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
			return fmt.Sprintf("https://drive.usercontent.google.com/download?id=%s&export=download&authuser=0&confirm=t", fileID), "", nil
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
			req.Header.Set("User-Agent", "Mozilla/5.0")
			if resp, err := httpClient.Do(req); err == nil {
				defer resp.Body.Close()
				body, _ := io.ReadAll(resp.Body)
				re := regexp.MustCompile(`"downloadUrl":"([^"]+)"`)
				if m := re.FindStringSubmatch(string(body)); len(m) > 1 {
					return m[1], "", nil
				}
			}
		}
	}

	return rawURL, "", nil
}
