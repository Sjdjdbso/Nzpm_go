package uploader

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CloneCloud menjalankan rclone copy langsung antar remote cloud (Google Drive, Mega, OneDrive)
func CloneCloud(srcRemote, dstRemote string, onProgress func(string)) error {
	args := []string{
		"copy",
		srcRemote,
		dstRemote,
		"--stats", "2s",
		"--stats-one-line",
		"-v",
	}

	if _, err := os.Stat("rclone.conf"); err == nil {
		args = append(args, "--config", "rclone.conf")
	}

	cmd := exec.Command("rclone", args...)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("gagal memulai clone rclone: %w", err)
	}

	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		line := scanner.Text()
		if onProgress != nil && strings.TrimSpace(line) != "" {
			onProgress(line)
		}
	}

	return cmd.Wait()
}

// CountRemote menghitung jumlah file dan total ukuran remote
func CountRemote(remotePath string) (string, error) {
	args := []string{"size", remotePath}
	if _, err := os.Stat("rclone.conf"); err == nil {
		args = append(args, "--config", "rclone.conf")
	}

	out, err := exec.Command("rclone", args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("gagal menghitung ukuran remote: %s", string(out))
	}

	return strings.TrimSpace(string(out)), nil
}
