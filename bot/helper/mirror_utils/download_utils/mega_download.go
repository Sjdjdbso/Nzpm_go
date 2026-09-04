package download_utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type MegaProgress struct {
	DownloadedBytes int64
	TotalBytes      int64
	Speed           string
	Percent         float64
}

// MegaDownload downloads files or folders from mega.nz using megatools
func MegaDownload(targetURL, outDir string, onProgress func(p MegaProgress)) (string, error) {
	absDir, err := filepath.Abs(outDir)
	if err != nil {
		absDir = outDir
	}
	_ = os.MkdirAll(absDir, 0755)

	args := []string{
		"dl",
		"--path", absDir,
		"--print-names",
		targetURL,
	}

	cmd := exec.Command("megatools", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("gagal pipe stdout megatools: %w", err)
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("gagal menjalankan megatools: %w", err)
	}

	var downloadedNames []string
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		downloadedNames = append(downloadedNames, line)
		log.Printf("[MEGATOOLS] %s", line)
	}

	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("megatools download error: %w", err)
	}

	if len(downloadedNames) > 0 {
		first := downloadedNames[0]
		if filepath.IsAbs(first) {
			return first, nil
		}
		return filepath.Join(absDir, first), nil
	}

	entries, err := os.ReadDir(absDir)
	if err == nil && len(entries) > 0 {
		return filepath.Join(absDir, entries[0].Name()), nil
	}

	return absDir, nil
}
