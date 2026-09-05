package upload_utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type RcloneProgress struct {
	Transferred string
	Total       string
	Percentage  float64
	Speed       string
	ETA         string
}

var rcloneProgressRegex = regexp.MustCompile(`Transferred:\s+([0-9.]+\s*[A-Za-z]+)\s+/\s+([0-9.]+\s*[A-Za-z]+),\s+([0-9.]+)%,\s+([0-9.]+\s*[A-Za-z/]+),\s+ETA\s+(\S+)`)

// GetUserRcloneConf mengembalikan path rclone.conf milik user jika ada, atau fallback ke default
func GetUserRcloneConf(userID int64) string {
	if userID != 0 {
		userConf := fmt.Sprintf("rclone/%d.conf", userID)
		if _, err := os.Stat(userConf); err == nil {
			return userConf
		}
	}
	if _, err := os.Stat("rclone.conf"); err == nil {
		return "rclone.conf"
	}
	return ""
}

// RcloneTransfer mengunggah file atau folder ke remote cloud via Rclone dengan parsing live progress
func RcloneTransfer(localPath, remoteDest, confPath string, onProgress func(p RcloneProgress)) error {
	if remoteDest == "" {
		log.Println("[INFO] RCLONE_PATH kosong, file disimpan di server.")
		return nil
	}

	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return fmt.Errorf("file lokal tidak ditemukan: %s", localPath)
	}

	args := []string{"copy", localPath, remoteDest, "--stats", "1s", "--stats-one-line", "-v"}
	if confPath == "" {
		confPath = GetUserRcloneConf(0)
	}
	if confPath != "" {
		args = append(args, "--config", confPath)
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
		matches := rcloneProgressRegex.FindStringSubmatch(line)
		if len(matches) >= 6 && onProgress != nil {
			pct, _ := strconv.ParseFloat(matches[3], 64)
			onProgress(RcloneProgress{
				Transferred: matches[1],
				Total:       matches[2],
				Percentage:  pct,
				Speed:       matches[4],
				ETA:         matches[5],
			})
		}
	}

	return cmd.Wait()
}

// RcloneClone menyalin file atau folder antar remote cloud secara server-side
func RcloneClone(src, dst, confPath string, onProgress func(p RcloneProgress)) error {
	args := []string{"copy", src, dst, "--stats", "1s", "--stats-one-line", "-v"}
	if confPath == "" {
		confPath = GetUserRcloneConf(0)
	}
	if confPath != "" {
		args = append(args, "--config", confPath)
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
		line := scanner.Text()
		matches := rcloneProgressRegex.FindStringSubmatch(line)
		if len(matches) >= 6 && onProgress != nil {
			pct, _ := strconv.ParseFloat(matches[3], 64)
			onProgress(RcloneProgress{
				Transferred: matches[1],
				Total:       matches[2],
				Percentage:  pct,
				Speed:       matches[4],
				ETA:         matches[5],
			})
		}
	}

	return cmd.Wait()
}

// RcloneCount menghitung ukuran total dan jumlah file remote path
func RcloneCount(remotePath, confPath string) (string, error) {
	args := []string{"size", remotePath}
	if confPath == "" {
		confPath = GetUserRcloneConf(0)
	}
	if confPath != "" {
		args = append(args, "--config", confPath)
	}

	out, err := exec.Command("rclone", args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("gagal menghitung size: %s", string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

// RcloneListRemotes membaca daftar remote yang terdaftar di rclone.conf
func RcloneListRemotes(confPath string) ([]string, error) {
	if confPath == "" {
		confPath = GetUserRcloneConf(0)
	}
	if confPath == "" {
		return nil, fmt.Errorf("file rclone.conf tidak ditemukan")
	}

	args := []string{"listremotes", "--config", confPath}
	out, err := exec.Command("rclone", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("gagal membaca list remotes: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var remotes []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			remotes = append(remotes, l)
		}
	}
	return remotes, nil
}

// RcloneListDirs membaca daftar subfolder dalam suatu remote:path
func RcloneListDirs(remotePath, confPath string) ([]string, error) {
	if confPath == "" {
		confPath = GetUserRcloneConf(0)
	}

	args := []string{"lsf", "--dirs-only", remotePath}
	if confPath != "" {
		args = append(args, "--config", confPath)
	}

	out, err := exec.Command("rclone", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("gagal membaca folder remote: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var dirs []string
	for _, l := range lines {
		l = strings.TrimSpace(strings.TrimSuffix(l, "/"))
		if l != "" {
			dirs = append(dirs, l)
		}
	}
	return dirs, nil
}
