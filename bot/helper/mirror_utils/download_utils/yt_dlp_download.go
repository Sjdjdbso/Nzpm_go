package download_utils

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
)

type YtProgress struct {
	Percent string
	Speed   string
	ETA     string
}

func YtDlpDownload(targetURL, outDir, customName string, onProgress func(p YtProgress)) (string, error) {
	outTemplate := "%(title)s.%(ext)s"
	if customName != "" {
		outTemplate = customName
	}
	absDir, _ := filepath.Abs(outDir)

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
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "PROGRESS_META:") {
			meta := strings.TrimPrefix(line, "PROGRESS_META:")
			parts := strings.Split(meta, "|")
			if len(parts) >= 3 && onProgress != nil {
				onProgress(YtProgress{
					Percent: strings.TrimSpace(parts[0]),
					Speed:   strings.TrimSpace(parts[1]),
					ETA:     strings.TrimSpace(parts[2]),
				})
			}
		} else if strings.Contains(line, "[download] Destination:") {
			lastFileName = strings.TrimSpace(strings.TrimPrefix(line, "[download] Destination:"))
		}
	}

	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("yt-dlp error: %w", err)
	}

	if lastFileName == "" {
		getNameCmd := exec.Command("yt-dlp", "--no-playlist", "--print", "filename", "-P", absDir, "-o", outTemplate, targetURL)
		if outBytes, getErr := getNameCmd.Output(); getErr == nil {
			lastFileName = strings.TrimSpace(string(outBytes))
		}
	}

	log.Printf("[INFO] yt-dlp unduhan selesai: %s", lastFileName)
	return lastFileName, nil
}
