package upload_utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"go-mirror-bot/bot"
	"go-mirror-bot/bot/helper/ext_utils"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

const (
	GDriveFolderMime = "application/vnd.google-apps.folder"
	GDriveDownloadURL = "https://drive.google.com/uc?id=%s&export=download"
	GDriveFolderURL   = "https://drive.google.com/drive/folders/%s"
	GDriveViewURL     = "https://drive.google.com/file/d/%s/view"
)

type TokenData struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	TokenURI     string   `json:"token_uri"`
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	Scopes       []string `json:"scopes"`
}

type GoogleDriveHelper struct {
	srv         *drive.Service
	ctx         context.Context
	saFiles     []string
	saIndex     int
	mu          sync.Mutex
	isCancelled bool
}

type ProgressReader struct {
	Reader   io.Reader
	Total    int64
	Current  int64
	Callback func(processed, total int64)
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	if n > 0 {
		pr.Current += int64(n)
		if pr.Callback != nil {
			pr.Callback(pr.Current, pr.Total)
		}
	}
	return n, err
}

type ProgressWriter struct {
	Writer   io.Writer
	Total    int64
	Current  int64
	Callback func(processed, total int64)
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n, err := pw.Writer.Write(p)
	if n > 0 {
		pw.Current += int64(n)
		if pw.Callback != nil {
			pw.Callback(pw.Current, pw.Total)
		}
	}
	return n, err
}

// ConvertTokenPickle converts Python token.pickle to token.json if needed
func ConvertTokenPickle(picklePath, jsonPath string) error {
	cmd := exec.Command("python3", "-c", fmt.Sprintf(`
import pickle, json

class Dummy:
    def __init__(self, *args, **kwargs): pass

class CustomUnpickler(pickle.Unpickler):
    def find_class(self, module, name):
        return Dummy

try:
    with open('%s', 'rb') as f:
        obj = CustomUnpickler(f).load()

    data = {
        'access_token': getattr(obj, 'token', ''),
        'refresh_token': getattr(obj, '_refresh_token', ''),
        'token_uri': getattr(obj, '_token_uri', 'https://oauth2.googleapis.com/token'),
        'client_id': getattr(obj, '_client_id', ''),
        'client_secret': getattr(obj, '_client_secret', ''),
        'scopes': getattr(obj, '_scopes', ['https://www.googleapis.com/auth/drive'])
    }

    with open('%s', 'w') as out:
        json.dump(data, out, indent=2)
except Exception as e:
    exit(1)
`, picklePath, jsonPath))
	return cmd.Run()
}

func NewGoogleDriveHelper() (*GoogleDriveHelper, error) {
	ctx := context.Background()
	helper := &GoogleDriveHelper{
		ctx: ctx,
	}

	// 1. Check Service Accounts
	accountsDir := "accounts"
	if bot.ConfigDict.UseServiceAccounts || dirExists(accountsDir) {
		entries, err := os.ReadDir(accountsDir)
		if err == nil {
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".json") {
					helper.saFiles = append(helper.saFiles, filepath.Join(accountsDir, e.Name()))
				}
			}
		}
	}

	if len(helper.saFiles) > 0 {
		helper.saIndex = rand.Intn(len(helper.saFiles))
		srv, err := drive.NewService(ctx, option.WithCredentialsFile(helper.saFiles[helper.saIndex]))
		if err == nil {
			helper.srv = srv
			log.Printf("[GDRIVE] Terotentikasi menggunakan Service Account: %s (%d/%d)",
				filepath.Base(helper.saFiles[helper.saIndex]), helper.saIndex+1, len(helper.saFiles))
			return helper, nil
		}
		log.Printf("[WARN] Gagal memuat Service Account %s: %v", helper.saFiles[helper.saIndex], err)
	}

	// 2. Check token.json
	tokenJsonPath := "token.json"
	if _, err := os.Stat(tokenJsonPath); os.IsNotExist(err) {
		// Try converting token.pickle if available
		if _, err := os.Stat("token.pickle"); err == nil {
			log.Println("[INFO] token.json tidak ditemukan, mengonversi token.pickle otomatis...")
			_ = ConvertTokenPickle("token.pickle", tokenJsonPath)
		}
	}

	f, err := os.Open(tokenJsonPath)
	if err == nil {
		defer f.Close()
		var td TokenData
		if err := json.NewDecoder(f).Decode(&td); err == nil && td.RefreshToken != "" {
			conf := &oauth2.Config{
				ClientID:     td.ClientID,
				ClientSecret: td.ClientSecret,
				Endpoint: oauth2.Endpoint{
					TokenURL: td.TokenURI,
				},
				Scopes: td.Scopes,
			}
			tok := &oauth2.Token{
				AccessToken:  td.AccessToken,
				RefreshToken: td.RefreshToken,
				TokenType:    "Bearer",
				Expiry:       time.Now().Add(-1 * time.Hour), // Force refresh on first call
			}
			client := conf.Client(ctx, tok)
			srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
			if err == nil {
				helper.srv = srv
				log.Println("[GDRIVE] Terotentikasi menggunakan token.json (OAuth2 User Credentials)")
				return helper, nil
			}
		}
	}

	return nil, errors.New("Google Drive credentials not found! Harap sediakan token.json / token.pickle atau Service Account di folder accounts/")
}

func (g *GoogleDriveHelper) SwitchServiceAccount() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if len(g.saFiles) <= 1 {
		return errors.New("Tidak ada Service Account alternatif yang tersedia")
	}

	g.saIndex = (g.saIndex + 1) % len(g.saFiles)
	nextSA := g.saFiles[g.saIndex]
	srv, err := drive.NewService(g.ctx, option.WithCredentialsFile(nextSA))
	if err != nil {
		return err
	}
	g.srv = srv
	log.Printf("[GDRIVE] Switch ke Service Account: %s (%d/%d)", filepath.Base(nextSA), g.saIndex+1, len(g.saFiles))
	return nil
}

func (g *GoogleDriveHelper) Cancel() {
	g.isCancelled = true
}

func (g *GoogleDriveHelper) GetIdFromUrl(link string) (string, error) {
	link = strings.TrimSpace(link)
	if !strings.Contains(link, "/") {
		return link, nil
	}

	// Regex format standard GDrive
	r1 := regexp.MustCompile(`drive\.google\.com\/(?:drive\/folders\/|file\/d\/)([-\w]+)`)
	matches := r1.FindStringSubmatch(link)
	if len(matches) > 1 {
		return matches[1], nil
	}

	// Regex id parameter
	r2 := regexp.MustCompile(`[?&]id=([-\w]+)`)
	matches2 := r2.FindStringSubmatch(link)
	if len(matches2) > 1 {
		return matches2[1], nil
	}

	return "", fmt.Errorf("Google Drive ID tidak ditemukan pada link: %s", link)
}

func (g *GoogleDriveHelper) GetFileMetadata(fileId string) (*drive.File, error) {
	return g.srv.Files.Get(fileId).
		SupportsAllDrives(true).
		Fields("id, name, mimeType, size, parents, shortcutDetails").
		Do()
}

func (g *GoogleDriveHelper) GetFilesByFolderId(folderId string) ([]*drive.File, error) {
	var allFiles []*drive.File
	var pageToken string

	for {
		call := g.srv.Files.List().
			SupportsAllDrives(true).
			IncludeItemsFromAllDrives(true).
			Q(fmt.Sprintf("'%s' in parents and trashed = false", folderId)).
			Spaces("drive").
			PageSize(200).
			Fields("nextPageToken, files(id, name, mimeType, size, shortcutDetails)")

		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		res, err := call.Do()
		if err != nil {
			return nil, err
		}

		allFiles = append(allFiles, res.Files...)
		if res.NextPageToken == "" {
			break
		}
		pageToken = res.NextPageToken
	}
	return allFiles, nil
}

func (g *GoogleDriveHelper) SetPermission(fileId string) error {
	perm := &drive.Permission{
		Role: "reader",
		Type: "anyone",
	}
	_, err := g.srv.Permissions.Create(fileId, perm).SupportsAllDrives(true).Do()
	return err
}

func (g *GoogleDriveHelper) CreateDirectory(name, parentId string) (string, error) {
	folderMetadata := &drive.File{
		Name:        name,
		Description: bot.ConfigDict.GDInfo,
		MimeType:    GDriveFolderMime,
	}
	if parentId != "" {
		folderMetadata.Parents = []string{parentId}
	}

	folder, err := g.srv.Files.Create(folderMetadata).SupportsAllDrives(true).Do()
	if err != nil {
		return "", err
	}

	if !bot.ConfigDict.IsTeamDrive {
		_ = g.SetPermission(folder.Id)
	}
	return folder.Id, nil
}

// Upload mengunggah file tunggal atau folder secara rekursif ke Google Drive
func (g *GoogleDriveHelper) Upload(filePath, parentId string, progressCb func(processed, total int64)) (string, int64, string, error) {
	if g.srv == nil {
		return "", 0, "", errors.New("Google Drive service tidak aktif")
	}

	if parentId == "" {
		parentId = bot.ConfigDict.GdriveID
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return "", 0, "", err
	}

	if !info.IsDir() {
		return g.uploadSingleFile(filePath, parentId, progressCb)
	}

	// Folder upload
	folderName := filepath.Base(filePath)
	dirId, err := g.CreateDirectory(folderName, parentId)
	if err != nil {
		return "", 0, "", fmt.Errorf("gagal membuat direktori root di GDrive: %w", err)
	}

	totalSize, err := g.uploadDirRecursive(filePath, dirId, progressCb, 0)
	if err != nil {
		return "", 0, "", err
	}

	link := fmt.Sprintf(GDriveFolderURL, dirId)
	return link, totalSize, "Folder", nil
}

func (g *GoogleDriveHelper) uploadSingleFile(filePath, parentId string, progressCb func(processed, total int64)) (string, int64, string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", 0, "", err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return "", 0, "", err
	}
	totalSize := stat.Size()
	fileName := filepath.Base(filePath)

	mimeType := mime.TypeByExtension(filepath.Ext(fileName))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	fileMetadata := &drive.File{
		Name:        fileName,
		Description: bot.ConfigDict.GDInfo,
		MimeType:    mimeType,
	}
	if parentId != "" {
		fileMetadata.Parents = []string{parentId}
	}

	pr := &ProgressReader{
		Reader:   f,
		Total:    totalSize,
		Callback: progressCb,
	}

	res, err := g.srv.Files.Create(fileMetadata).
		Media(pr).
		SupportsAllDrives(true).
		Do()

	if err != nil {
		// Attempt service account switch if rate limited
		if isQuotaOrRateLimitError(err) && len(g.saFiles) > 1 {
			if swErr := g.SwitchServiceAccount(); swErr == nil {
				return g.uploadSingleFile(filePath, parentId, progressCb)
			}
		}
		return "", 0, "", err
	}

	if !bot.ConfigDict.IsTeamDrive {
		_ = g.SetPermission(res.Id)
	}

	link := fmt.Sprintf(GDriveDownloadURL, res.Id)
	return link, totalSize, mimeType, nil
}

func (g *GoogleDriveHelper) uploadDirRecursive(dirPath, parentId string, progressCb func(processed, total int64), currentProcessed int64) (int64, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return 0, err
	}

	var totalDirSize int64
	for _, entry := range entries {
		if g.isCancelled {
			return 0, errors.New("upload dibatalkan oleh pengguna")
		}

		fullPath := filepath.Join(dirPath, entry.Name())
		if entry.IsDir() {
			subDirId, err := g.CreateDirectory(entry.Name(), parentId)
			if err != nil {
				return 0, err
			}
			subSize, err := g.uploadDirRecursive(fullPath, subDirId, progressCb, currentProcessed+totalDirSize)
			if err != nil {
				return 0, err
			}
			totalDirSize += subSize
		} else {
			_, fSize, _, err := g.uploadSingleFile(fullPath, parentId, func(p, t int64) {
				if progressCb != nil {
					progressCb(currentProcessed+totalDirSize+p, 0)
				}
			})
			if err != nil {
				return 0, err
			}
			totalDirSize += fSize
		}
	}
	return totalDirSize, nil
}

// Clone menduplikasi file atau folder di Google Drive secara cloud server-side
func (g *GoogleDriveHelper) Clone(link, destId string, progressCb func(processed, total int64, files, folders int)) (string, int64, string, int, int, error) {
	if g.srv == nil {
		return "", 0, "", 0, 0, errors.New("Google Drive service tidak aktif")
	}
	if destId == "" {
		destId = bot.ConfigDict.GdriveID
	}

	fileId, err := g.GetIdFromUrl(link)
	if err != nil {
		return "", 0, "", 0, 0, err
	}

	meta, err := g.GetFileMetadata(fileId)
	if err != nil {
		return "", 0, "", 0, 0, err
	}

	if meta.MimeType == GDriveFolderMime {
		dirId, err := g.CreateDirectory(meta.Name, destId)
		if err != nil {
			return "", 0, "", 0, 0, err
		}

		totalBytes, totalFiles, totalFolders, err := g.cloneFolderRecursive(meta.Id, dirId, progressCb, 0, 0, 0)
		if err != nil {
			return "", 0, "", 0, 0, err
		}
		link := fmt.Sprintf(GDriveFolderURL, dirId)
		return link, totalBytes, "Folder", totalFiles, totalFolders, nil
	}

	// Clone single file
	copiedFile, err := g.srv.Files.Copy(fileId, &drive.File{
		Name:    meta.Name,
		Parents: []string{destId},
	}).SupportsAllDrives(true).Do()

	if err != nil {
		if isQuotaOrRateLimitError(err) && len(g.saFiles) > 1 {
			if swErr := g.SwitchServiceAccount(); swErr == nil {
				return g.Clone(link, destId, progressCb)
			}
		}
		return "", 0, "", 0, 0, err
	}

	if !bot.ConfigDict.IsTeamDrive {
		_ = g.SetPermission(copiedFile.Id)
	}

	linkRes := fmt.Sprintf(GDriveViewURL, copiedFile.Id)
	return linkRes, meta.Size, meta.MimeType, 1, 0, nil
}

func (g *GoogleDriveHelper) cloneFolderRecursive(srcFolderId, destFolderId string, progressCb func(p, t int64, f, fo int), curBytes int64, curFiles, curFolders int) (int64, int, int, error) {
	items, err := g.GetFilesByFolderId(srcFolderId)
	if err != nil {
		return 0, 0, 0, err
	}

	totalBytes := curBytes
	totalFiles := curFiles
	totalFolders := curFolders

	for _, item := range items {
		if g.isCancelled {
			return 0, 0, 0, errors.New("clone dibatalkan")
		}

		if item.MimeType == GDriveFolderMime {
			totalFolders++
			subFolderId, err := g.CreateDirectory(item.Name, destFolderId)
			if err != nil {
				return 0, 0, 0, err
			}
			b, f, fo, err := g.cloneFolderRecursive(item.Id, subFolderId, progressCb, totalBytes, totalFiles, totalFolders)
			if err != nil {
				return 0, 0, 0, err
			}
			totalBytes = b
			totalFiles = f
			totalFolders = fo
		} else {
			totalFiles++
			totalBytes += item.Size
			_, err := g.srv.Files.Copy(item.Id, &drive.File{
				Name:    item.Name,
				Parents: []string{destFolderId},
			}).SupportsAllDrives(true).Do()

			if err != nil {
				if isQuotaOrRateLimitError(err) && len(g.saFiles) > 1 {
					_ = g.SwitchServiceAccount()
				}
			}

			if progressCb != nil {
				progressCb(totalBytes, 0, totalFiles, totalFolders)
			}
		}
	}
	return totalBytes, totalFiles, totalFolders, nil
}

// Count menghitung ukuran total, jumlah file, dan folder
func (g *GoogleDriveHelper) Count(link string) (name, mimeType string, totalBytes int64, totalFiles, totalFolders int, err error) {
	if g.srv == nil {
		return "", "", 0, 0, 0, errors.New("Google Drive service tidak aktif")
	}

	fileId, err := g.GetIdFromUrl(link)
	if err != nil {
		return "", "", 0, 0, 0, err
	}

	meta, err := g.GetFileMetadata(fileId)
	if err != nil {
		return "", "", 0, 0, 0, err
	}

	name = meta.Name
	mimeType = meta.MimeType

	if meta.MimeType != GDriveFolderMime {
		return name, "File", meta.Size, 1, 0, nil
	}

	mimeType = "Folder"
	totalBytes, totalFiles, totalFolders, err = g.countFolderRecursive(meta.Id)
	return name, mimeType, totalBytes, totalFiles, totalFolders, err
}

func (g *GoogleDriveHelper) countFolderRecursive(folderId string) (int64, int, int, error) {
	files, err := g.GetFilesByFolderId(folderId)
	if err != nil {
		return 0, 0, 0, err
	}

	var totalBytes int64
	var totalFiles, totalFolders int

	for _, f := range files {
		if f.MimeType == GDriveFolderMime {
			totalFolders++
			subBytes, subFiles, subFolders, err := g.countFolderRecursive(f.Id)
			if err == nil {
				totalBytes += subBytes
				totalFiles += subFiles
				totalFolders += subFolders
			}
		} else {
			totalFiles++
			totalBytes += f.Size
		}
	}
	return totalBytes, totalFiles, totalFolders, nil
}

// DeleteFile menghapus file atau folder di Google Drive
func (g *GoogleDriveHelper) DeleteFile(link string) (string, error) {
	if g.srv == nil {
		return "", errors.New("Google Drive service tidak aktif")
	}

	fileId, err := g.GetIdFromUrl(link)
	if err != nil {
		return "", err
	}

	err = g.srv.Files.Delete(fileId).SupportsAllDrives(true).Do()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("✅ <b>File/Folder berhasil dihapus!</b>\n🆔 <code>%s</code>", fileId), nil
}

// DriveClean mengosongkan folder atau drive
func (g *GoogleDriveHelper) DriveClean(driveId string, trash bool) (string, error) {
	if g.srv == nil {
		return "", errors.New("Google Drive service tidak aktif")
	}

	if driveId == "" {
		driveId = bot.ConfigDict.GdriveID
	}
	if driveId == "" {
		return "", errors.New("GDRIVE_ID tidak dikonfigurasi")
	}

	files, err := g.GetFilesByFolderId(driveId)
	if err != nil {
		return "", err
	}

	var count int
	var totalSize int64
	for _, f := range files {
		count++
		totalSize += f.Size
		if trash {
			_, _ = g.srv.Files.Update(f.Id, &drive.File{Trashed: true}).SupportsAllDrives(true).Do()
		} else {
			_ = g.srv.Files.Delete(f.Id).SupportsAllDrives(true).Do()
		}
	}

	action := "Dibersihkan permanen"
	if trash {
		action = "Dipindahkan ke Sampah"
	}
	return fmt.Sprintf("⌬ <b><i>Hasil Clean GDrive (%s):</i></b>\n\n<b>Total Item:</b> <code>%d</code>\n<b>Total Ukuran:</b> <code>%s</code>",
		action, count, ext_utils.FormatBytes(totalSize)), nil
}

// DriveList mencari file/folder di Google Drive
func (g *GoogleDriveHelper) DriveList(key string, isRecursive bool, itemType string) ([]*drive.File, error) {
	if g.srv == nil {
		return nil, errors.New("Google Drive service tidak aktif")
	}

	query := fmt.Sprintf("name contains '%s' and trashed = false", key)
	if itemType == "files" {
		query += fmt.Sprintf(" and mimeType != '%s'", GDriveFolderMime)
	} else if itemType == "folders" {
		query += fmt.Sprintf(" and mimeType = '%s'", GDriveFolderMime)
	}

	call := g.srv.Files.List().
		SupportsAllDrives(true).
		IncludeItemsFromAllDrives(true).
		Q(query).
		Spaces("drive").
		PageSize(50).
		Fields("files(id, name, mimeType, size)")

	res, err := call.Do()
	if err != nil {
		return nil, err
	}
	return res.Files, nil
}

// Download mengunduh file atau direktori dari Google Drive ke disk lokal
func (g *GoogleDriveHelper) Download(link, destDir string, progressCb func(processed, total int64)) (string, error) {
	if g.srv == nil {
		return "", errors.New("Google Drive service tidak aktif")
	}

	fileId, err := g.GetIdFromUrl(link)
	if err != nil {
		return "", err
	}

	meta, err := g.GetFileMetadata(fileId)
	if err != nil {
		return "", err
	}

	_ = os.MkdirAll(destDir, 0755)

	if meta.MimeType == GDriveFolderMime {
		targetFolder := filepath.Join(destDir, meta.Name)
		_ = os.MkdirAll(targetFolder, 0755)
		err := g.downloadFolderRecursive(meta.Id, targetFolder, progressCb, 0)
		return targetFolder, err
	}

	targetFilePath := filepath.Join(destDir, meta.Name)
	err = g.downloadSingleFile(meta.Id, targetFilePath, meta.Size, progressCb)
	return targetFilePath, err
}

func (g *GoogleDriveHelper) downloadSingleFile(fileId, targetPath string, totalSize int64, progressCb func(processed, total int64)) error {
	resp, err := g.srv.Files.Get(fileId).SupportsAllDrives(true).Download()
	if err != nil {
		if isQuotaOrRateLimitError(err) && len(g.saFiles) > 1 {
			if swErr := g.SwitchServiceAccount(); swErr == nil {
				return g.downloadSingleFile(fileId, targetPath, totalSize, progressCb)
			}
		}
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer out.Close()

	pw := &ProgressWriter{
		Writer:   out,
		Total:    totalSize,
		Callback: progressCb,
	}

	_, err = io.Copy(pw, resp.Body)
	return err
}

func (g *GoogleDriveHelper) downloadFolderRecursive(folderId, targetDir string, progressCb func(processed, total int64), curProcessed int64) error {
	files, err := g.GetFilesByFolderId(folderId)
	if err != nil {
		return err
	}

	for _, f := range files {
		if g.isCancelled {
			return errors.New("unduhan dibatalkan")
		}

		if f.MimeType == GDriveFolderMime {
			subDir := filepath.Join(targetDir, f.Name)
			_ = os.MkdirAll(subDir, 0755)
			_ = g.downloadFolderRecursive(f.Id, subDir, progressCb, curProcessed)
		} else {
			filePath := filepath.Join(targetDir, f.Name)
			_ = g.downloadSingleFile(f.Id, filePath, f.Size, func(p, t int64) {
				if progressCb != nil {
					progressCb(curProcessed+p, 0)
				}
			})
			curProcessed += f.Size
		}
	}
	return nil
}

func isQuotaOrRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	var gErr *googleapi.Error
	if errors.As(err, &gErr) {
		if gErr.Code == http.StatusTooManyRequests || gErr.Code == 403 {
			for _, item := range gErr.Errors {
				if item.Reason == "userRateLimitExceeded" ||
					item.Reason == "dailyLimitExceeded" ||
					item.Reason == "rateLimitExceeded" ||
					item.Reason == "downloadQuotaExceeded" {
					return true
				}
			}
		}
	}
	str := strings.ToLower(err.Error())
	return strings.Contains(str, "quota") || strings.Contains(str, "ratelimit")
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
