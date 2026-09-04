package core

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
)

type YtDlpProgress struct {
	Percent string
	Speed   string
	ETA     string
}

// DownloadYtDlp mengunduh video/audio menggunakan binary yt-dlp
func DownloadYtDlp(targetURL string, outDir string, customName string, onProgress func(prog YtDlpProgress)) (string, error) {
	outTemplate := "%(title)s.%(ext)s"
	if customName != "" {
		outTemplate = customName
	}

	absDir, err := filepath.Abs(outDir)
	if err != nil {
		absDir = outDir
	}

	args := []string{
		"--newline",
		"--no-warnings",
		"--no-playlist",
		"-P", absDir,
		"-o", outTemplate,
		"--progress-template", "PROGRESS_META:%(progress._percent_str)s|%(progress._speed_str)s|%(progress._eta_str)s",
		targetURL,
	}

	cmd := exec.Command("yt-dlp", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("gagal menjalankan yt-dlp: %w", err)
	}

	var lastFileName string
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "PROGRESS_META:") {
			meta := strings.TrimPrefix(line, "PROGRESS_META:")
			parts := strings.Split(meta, "|")
			if len(parts) >= 3 && onProgress != nil {
				onProgress(YtDlpProgress{
					Percent: strings.TrimSpace(parts[0]),
					Speed:   strings.TrimSpace(parts[1]),
					ETA:     strings.TrimSpace(parts[2]),
				})
			}
		} else if strings.Contains(line, "[download] Destination:") {
			lastFileName = strings.TrimSpace(strings.TrimPrefix(line, "[download] Destination:"))
		} else if strings.Contains(line, "has already been downloaded") {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				lastFileName = parts[1]
			}
		}
	}

	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("yt-dlp error: %w", err)
	}

	// Jika nama file tidak tertangkap di log, cari file terbaru di outDir
	if lastFileName == "" {
		// Dapatkan nama file yang diunduh via yt-dlp --print filename
		getNameCmd := exec.Command("yt-dlp", "--no-playlist", "--print", "filename", "-P", absDir, "-o", outTemplate, targetURL)
		outBytes, getErr := getNameCmd.Output()
		if getErr == nil {
			lastFileName = strings.TrimSpace(string(outBytes))
		}
	}

	log.Printf("[INFO] Unduhan yt-dlp selesai: %s", lastFileName)
	return lastFileName, nil
}
