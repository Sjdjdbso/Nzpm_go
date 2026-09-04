package ext_utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func IsArchive(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".zip", ".rar", ".7z", ".tar", ".gz", ".bz2", ".xz":
		return true
	}
	if strings.HasSuffix(strings.ToLower(path), ".tar.gz") || strings.HasSuffix(strings.ToLower(path), ".tar.xz") {
		return true
	}
	return false
}

func ExtractArchive(archivePath string) (string, error) {
	if !IsArchive(archivePath) {
		return archivePath, fmt.Errorf("file bukan format arsip: %s", filepath.Base(archivePath))
	}

	dir := filepath.Dir(archivePath)
	baseName := strings.TrimSuffix(filepath.Base(archivePath), filepath.Ext(archivePath))
	outDir := filepath.Join(dir, baseName)
	os.MkdirAll(outDir, 0755)

	log.Printf("[INFO] Mengekstrak %s ke %s...", archivePath, outDir)
	cmd := exec.Command("7z", "x", archivePath, "-o"+outDir, "-aoa")
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("gagal ekstrak: %s (%v)", string(output), err)
	}

	os.Remove(archivePath)
	return outDir, nil
}

func CompressToZip(sourcePath string) (string, error) {
	fi, err := os.Stat(sourcePath)
	if err != nil {
		return "", err
	}

	if !fi.IsDir() && strings.HasSuffix(strings.ToLower(sourcePath), ".zip") {
		return sourcePath, nil
	}

	dir := filepath.Dir(sourcePath)
	baseName := filepath.Base(sourcePath)
	tempZip := filepath.Join(dir, fmt.Sprintf("tmp_%d.zip", os.Getpid()))
	finalZip := filepath.Join(dir, strings.TrimSuffix(baseName, filepath.Ext(baseName))+".zip")

	log.Printf("[INFO] Mengompres %s menjadi %s...", sourcePath, finalZip)
	cmd := exec.Command("7z", "a", "-tzip", tempZip, sourcePath)
	if output, err := cmd.CombinedOutput(); err != nil {
		os.Remove(tempZip)
		return "", fmt.Errorf("gagal kompres zip: %s (%v)", string(output), err)
	}

	os.RemoveAll(sourcePath)
	if err := os.Rename(tempZip, finalZip); err != nil {
		return tempZip, nil
	}
	return finalZip, nil
}

func SplitArchive(filePath string) ([]string, error) {
	dir := filepath.Dir(filePath)
	baseName := filepath.Base(filePath)
	outPrefix := filepath.Join(dir, strings.TrimSuffix(baseName, filepath.Ext(baseName))+".part")

	log.Printf("[INFO] Memecah file %s (chunk 49MB)...", filePath)
	cmd := exec.Command("7z", "a", "-v49m", outPrefix+".zip", filePath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("gagal memecah: %s (%v)", string(output), err)
	}

	matches, err := filepath.Glob(outPrefix + ".*")
	if err != nil || len(matches) == 0 {
		return nil, fmt.Errorf("file split tidak ditemukan: %v", err)
	}

	os.Remove(filePath)
	return matches, nil
}

func CleanPath(path string) {
	if path == "" {
		return
	}
	clean := filepath.Clean(path)
	if clean == "." || clean == "/" || clean == "downloads" {
		return
	}
	os.RemoveAll(clean)
}
