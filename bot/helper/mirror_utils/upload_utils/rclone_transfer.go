package upload_utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func RcloneTransfer(localPath, remoteDest string, onProgress func(string)) error {
	if remoteDest == "" {
		log.Println("[INFO] RCLONE_PATH kosong, file disimpan di server.")
		return nil
	}

	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return fmt.Errorf("file lokal tidak ada: %s", localPath)
	}

	args := []string{"copy", localPath, remoteDest, "--stats", "2s", "--stats-one-line", "-v"}
	if _, err := os.Stat("rclone.conf"); err == nil {
		args = append(args, "--config", "rclone.conf")
	}

	cmd := exec.Command("rclone", args...)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		l := scanner.Text()
		if onProgress != nil && strings.TrimSpace(l) != "" {
			onProgress(l)
		}
	}
	return cmd.Wait()
}

func RcloneClone(src, dst string) error {
	args := []string{"copy", src, dst, "--stats", "2s", "--stats-one-line", "-v"}
	if _, err := os.Stat("rclone.conf"); err == nil {
		args = append(args, "--config", "rclone.conf")
	}
	return exec.Command("rclone", args...).Run()
}

func RcloneCount(remotePath string) (string, error) {
	args := []string{"size", remotePath}
	if _, err := os.Stat("rclone.conf"); err == nil {
		args = append(args, "--config", "rclone.conf")
	}
	out, err := exec.Command("rclone", args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("gagal menghitung size: %s", string(out))
	}
	return strings.TrimSpace(string(out)), nil
}
