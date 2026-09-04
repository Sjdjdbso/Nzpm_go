package core

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// IsArchive memeriksa apakah file adalah format arsip yang didukung
func IsArchive(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".zip", ".rar", ".7z", ".tar", ".gz", ".bz2", ".xz":
		return true
	}
	// Cek .tar.gz dsb
	if strings.HasSuffix(strings.ToLower(path), ".tar.gz") || strings.HasSuffix(strings.ToLower(path), ".tar.xz") {
		return true
	}
	return false
}

// ExtractArchive mengekstrak arsip menggunakan 7z
func ExtractArchive(archivePath string) (string, error) {
	if !IsArchive(archivePath) {
		return archivePath, fmt.Errorf("file bukan format arsip yang didukung: %s", filepath.Base(archivePath))
	}

	dir := filepath.Dir(archivePath)
	baseName := strings.TrimSuffix(filepath.Base(archivePath), filepath.Ext(archivePath))
	outDir := filepath.Join(dir, baseName)

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return "", err
	}

	log.Printf("[INFO] Mengekstrak %s ke %s...", archivePath, outDir)
	cmd := exec.Command("7z", "x", archivePath, "-o"+outDir, "-aoa")
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("gagal mengekstrak arsip: %s (%v)", string(output), err)
	}

	// Hapus arsip asli setelah berhasil diekstrak
	os.Remove(archivePath)
	return outDir, nil
}

// CompressToZip mengompres file atau direktori menjadi file .zip
func CompressToZip(sourcePath string) (string, error) {
	dir := filepath.Dir(sourcePath)
	baseName := filepath.Base(sourcePath)
	zipName := strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ".zip"
	zipPath := filepath.Join(dir, zipName)

	log.Printf("[INFO] Mengompres %s menjadi %s...", sourcePath, zipPath)
	cmd := exec.Command("7z", "a", "-tzip", zipPath, sourcePath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("gagal mengompres zip: %s (%v)", string(output), err)
	}

	// Hapus sumber asli jika berbeda dari zipPath
	if sourcePath != zipPath {
		os.RemoveAll(sourcePath)
	}

	return zipPath, nil
}

// SplitArchive memecah file menjadi part-part maksimal 49MB (untuk batas Telegram Bot API)
func SplitArchive(filePath string) ([]string, error) {
	dir := filepath.Dir(filePath)
	baseName := filepath.Base(filePath)
	outPrefix := filepath.Join(dir, strings.TrimSuffix(baseName, filepath.Ext(baseName))+".part")

	log.Printf("[INFO] Memecah file %s menjadi chunk 49MB...", filePath)
	cmd := exec.Command("7z", "a", "-v49m", outPrefix+".zip", filePath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("gagal memecah file: %s (%v)", string(output), err)
	}

	// Cari semua file hasil split
	matches, err := filepath.Glob(outPrefix + ".*")
	if err != nil || len(matches) == 0 {
		return nil, fmt.Errorf("file split tidak ditemukan: %v", err)
	}

	// Hapus file asli
	os.Remove(filePath)
	return matches, nil
}
