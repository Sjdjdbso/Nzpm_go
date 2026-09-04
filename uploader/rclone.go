package uploader

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// UploadFile mengupload file menggunakan rclone ke remote tujuan
func UploadFile(localPath string, remoteDestination string, onProgress func(line string)) error {
	if remoteDestination == "" {
		log.Println("[INFO] RCLONE_PATH belum diset, melewati tahap upload.")
		return nil
	}

	// Cek apakah file lokal ada
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return fmt.Errorf("file lokal tidak ditemukan: %s", localPath)
	}

	args := []string{
		"copy",
		localPath,
		remoteDestination,
		"--stats", "2s",
		"--stats-one-line",
		"-v",
	}

	// Jika ada file rclone.conf di root folder, gunakan secara otomatis
	if _, err := os.Stat("rclone.conf"); err == nil {
		args = append(args, "--config", "rclone.conf")
	}

	cmd := exec.Command("rclone", args...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("gagal memulai rclone: %w", err)
	}

	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		line := scanner.Text()
		if onProgress != nil && strings.TrimSpace(line) != "" {
			onProgress(line)
		}
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("rclone error: %w", err)
	}

	return nil
}

// CleanLocal menghapus file lokal setelah selesai di-upload untuk menghemat disk
func CleanLocal(path string) {
	if path == "" {
		return
	}
	// Pastikan tidak menghapus folder induk downloads
	cleanPath := filepath.Clean(path)
	if cleanPath == "." || cleanPath == "/" || cleanPath == "downloads" {
		return
	}
	if err := os.RemoveAll(cleanPath); err != nil {
		log.Printf("[WARN] Gagal menghapus file lokal %s: %v", cleanPath, err)
	} else {
		log.Printf("[INFO] Berhasil membersihkan file lokal: %s", cleanPath)
	}
}
