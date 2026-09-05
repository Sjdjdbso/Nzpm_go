package listeners

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go-mirror-bot/bot"
	"go-mirror-bot/bot/helper/ext_utils"
	"go-mirror-bot/bot/helper/mirror_utils/download_utils"
	"go-mirror-bot/bot/helper/mirror_utils/upload_utils"
	"go-mirror-bot/bot/helper/mirror_utils/upload_utils/ddlserver"
	"go-mirror-bot/bot/helper/themes"

	tele "gopkg.in/telebot.v3"
)

type MirrorLeechListener struct {
	Bot        *tele.Bot
	Recipient  tele.Recipient
	StatusMsg  *tele.Message
	GID        string
	RcloneDest string
	IsLeech    bool
	IsZip      bool
	IsExtract  bool
	Markup     *tele.ReplyMarkup
}

func (l *MirrorLeechListener) ProcessCompletedDownload(filePath string, totalSize int64, t *ext_utils.Task) {
	fileName := filepath.Base(filePath)
	if totalSize <= 0 {
		if fi, err := os.Stat(filePath); err == nil {
			totalSize = fi.Size()
		}
	}
	t.TotalSize = totalSize
	t.CompletedSize = totalSize
	t.Progress = 100.0

	// Sanity Check: Tolak jika yang terunduh adalah halaman error HTML Google Drive
	if fi, err := os.Stat(filePath); err == nil && fi.Size() < 50000 {
		contentSample, _ := os.ReadFile(filePath)
		sampleStr := string(contentSample)
		if strings.Contains(sampleStr, "Google Drive - Quota exceeded") || strings.Contains(sampleStr, "Too many users have viewed or downloaded this file recently") {
			errMsg := "❌ <b>Unduhan Dibatalkan:</b> File yang terunduh adalah halaman error <i>Google Drive - Quota Exceeded</i>. Kuota download publik file ini telah habis dibatasi oleh Google."
			if l.StatusMsg != nil {
				l.Bot.Edit(l.StatusMsg, errMsg, &tele.SendOptions{ParseMode: tele.ModeHTML})
			} else {
				l.Bot.Send(l.Recipient, errMsg, &tele.SendOptions{ParseMode: tele.ModeHTML})
			}
			ext_utils.CleanPath(filePath)
			ext_utils.TaskMgr.Remove(l.GID)
			return
		}
	}

	if l.IsExtract && ext_utils.IsArchive(filePath) {
		if l.StatusMsg != nil {
			l.Bot.Edit(l.StatusMsg, "📦 <b>Mengekstrak arsip...</b>", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}
		if outDir, err := ext_utils.ExtractArchive(filePath); err == nil {
			filePath = outDir
			fileName = filepath.Base(filePath)
		}
	}

	if l.IsZip {
		if l.StatusMsg != nil {
			l.Bot.Edit(l.StatusMsg, "🗜 <b>Mengompres ke ZIP...</b>", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}
		if zipPath, err := ext_utils.CompressToZip(filePath); err == nil {
			filePath = zipPath
			fileName = filepath.Base(filePath)
		}
	}

	if l.IsLeech {
		t.Status = "Leeching"
		if l.StatusMsg != nil {
			l.Bot.Edit(l.StatusMsg, fmt.Sprintf("📤 <b>Unduhan Selesai!</b>\n📁 <code>%s</code>\n🚀 <i>Mengirim ke Telegram...</i>", fileName), &tele.SendOptions{ParseMode: tele.ModeHTML})
		}
		if err := upload_utils.TelegramUpload(l.Bot, l.Recipient, filePath, t.User, t.UserID); err != nil {
			l.Bot.Send(l.Recipient, fmt.Sprintf("❌ <b>Leech Gagal:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
		} else {
			l.Bot.Send(l.Recipient, themes.FormatCompleteMsg(fileName, totalSize, "Leech", "", t.User), &tele.SendOptions{ParseMode: tele.ModeHTML})
		}
	} else {
		destLower := strings.ToLower(l.RcloneDest)
		isDDL := destLower == "ddl" || destLower == "pixeldrain" || (destLower == "" && (bot.ConfigDict.DefaultUpload == "ddl" || bot.ConfigDict.DefaultUpload == "pixeldrain"))
		isGD := destLower == "gd" || destLower == "gdrive" || (destLower == "" && (bot.ConfigDict.DefaultUpload == "gd" || bot.ConfigDict.DefaultUpload == "gdrive"))

		if isGD {
			t.Status = "Uploading (GDrive)"
			if l.StatusMsg != nil {
				l.Bot.Edit(l.StatusMsg, fmt.Sprintf("📤 <b>Unduhan Selesai!</b>\n📁 <code>%s</code>\n🚀 <i>Mengunggah ke Google Drive...</i>", fileName), &tele.SendOptions{ParseMode: tele.ModeHTML})
			}

			gdHelper, err := upload_utils.NewGoogleDriveHelper()
			if err != nil {
				l.Bot.Send(l.Recipient, fmt.Sprintf("❌ <b>Google Drive Auth Gagal:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
			} else {
				upLink, upSize, mimeType, err := gdHelper.Upload(filePath, bot.ConfigDict.GdriveID, func(processed, total int64) {
					if total > 0 {
						t.CompletedSize = processed
						t.Progress = float64(processed) / float64(total) * 100
					}
				})
				if err != nil {
					l.Bot.Send(l.Recipient, fmt.Sprintf("❌ <b>GDrive Upload Gagal:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
				} else {
					markup := &tele.ReplyMarkup{}
					btnURL := markup.URL("🔗 Google Drive", upLink)
					var rows []tele.Row
					if bot.ConfigDict.IndexURL != "" {
						idxURL := strings.TrimRight(bot.ConfigDict.IndexURL, "/") + "/" + fileName
						btnIdx := markup.URL("⚡ Index Link", idxURL)
						rows = append(rows, markup.Row(btnURL, btnIdx))
					} else {
						rows = append(rows, markup.Row(btnURL))
					}
					markup.Inline(rows...)

					completeText := fmt.Sprintf(
						"<b><i>Mirror GDrive Selesai!</i></b>\n\n"+
							"➲ <b>File:</b> <code>%s</code>\n"+
							"┠ <b>Size:</b> <code>%s</code>\n"+
							"┠ <b>Type:</b> <code>%s</code>\n"+
							"┠ <b>Link:</b> <a href=\"%s\">Google Drive</a>\n"+
							"┖ <b>By:</b> %s",
						fileName, ext_utils.FormatBytes(upSize), mimeType, upLink, t.User,
					)
					l.Bot.Send(l.Recipient, completeText, markup, &tele.SendOptions{ParseMode: tele.ModeHTML})
				}
			}
		} else if isDDL {
			t.Status = "Uploading (DDL)"
			if l.StatusMsg != nil {
				l.Bot.Edit(l.StatusMsg, fmt.Sprintf("📤 <b>Unduhan Selesai!</b>\n📁 <code>%s</code>\n🚀 <i>Mengunggah ke Pixeldrain (DDL)...</i>", fileName), &tele.SendOptions{ParseMode: tele.ModeHTML})
			}

			apiKey := bot.ConfigDict.PixeldrainAPI
			if uCfg := ext_utils.UserStore.Get(t.UserID); uCfg != nil && uCfg.PixeldrainAPI != "" {
				apiKey = uCfg.PixeldrainAPI
			}

			pd := ddlserver.NewPixeldrain(apiKey)
			pdLink, err := pd.Upload(filePath)
			if err != nil {
				l.Bot.Send(l.Recipient, fmt.Sprintf("❌ <b>Pixeldrain Upload Gagal:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
			} else {
				completeText := fmt.Sprintf(
					"<b><i>Mirror DDL Selesai!</i></b>\n\n"+
						"➲ <b>File:</b> <code>%s</code>\n"+
						"┠ <b>Size:</b> <code>%s</code>\n"+
						"┠ <b>Server:</b> <code>Pixeldrain DDL</code>\n"+
						"┠ <b>Link:</b> <a href=\"%s\">%s</a>\n"+
						"┖ <b>By:</b> %s",
					fileName, ext_utils.FormatBytes(totalSize), pdLink, pdLink, t.User,
				)
				markup := &tele.ReplyMarkup{}
				btnURL := markup.URL("🔗 Unduh Pixeldrain", pdLink)
				markup.Inline(markup.Row(btnURL))
				l.Bot.Send(l.Recipient, completeText, markup, &tele.SendOptions{ParseMode: tele.ModeHTML})
			}
		} else {
			t.Status = "Uploading (Rclone)"
			if l.StatusMsg != nil {
				l.Bot.Edit(l.StatusMsg, fmt.Sprintf("📤 <b>Unduhan Selesai!</b>\n📁 <code>%s</code>\n🚀 <i>Mengunggah ke Cloud via Rclone...</i>", fileName), &tele.SendOptions{ParseMode: tele.ModeHTML})
			}
			lastUpdate := time.Now()
			userConf := upload_utils.GetUserRcloneConf(t.UserID)
			if err := upload_utils.RcloneTransfer(filePath, l.RcloneDest, userConf, func(p upload_utils.RcloneProgress) {
				t.Progress = p.Percentage
				t.ETA = p.ETA
				if time.Since(lastUpdate) >= 3*time.Second && l.StatusMsg != nil {
					lastUpdate = time.Now()
					l.Bot.Edit(l.StatusMsg, themes.FormatStatusMsg(t), l.Markup, &tele.SendOptions{ParseMode: tele.ModeHTML})
				}
			}); err != nil {
				l.Bot.Send(l.Recipient, fmt.Sprintf("❌ <b>Upload Gagal:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
			} else {
				dest := l.RcloneDest
				if dest == "" {
					dest = "Disimpan di server"
				}
				l.Bot.Send(l.Recipient, themes.FormatCompleteMsg(fileName, totalSize, "Mirror", dest, t.User), &tele.SendOptions{ParseMode: tele.ModeHTML})
			}
		}
	}

	ext_utils.CleanPath(filePath)
	ext_utils.TaskMgr.Remove(l.GID)
}

func (l *MirrorLeechListener) Start() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	var filePath, fileName, lastText string
	var totalSize int64
	activeGID := l.GID

	for range ticker.C {
		st, err := download_utils.Aria.TellStatus(activeGID)
		if err != nil {
			continue
		}

		if len(st.FollowedBy) > 0 {
			activeGID = st.FollowedBy[0]
			continue
		}

		t := ext_utils.TaskMgr.Get(activeGID)
		if t == nil {
			t = ext_utils.TaskMgr.Get(l.GID)
			if t == nil {
				return
			}
		}

		if len(st.Files) > 0 && st.Files[0].Path != "" {
			filePath = st.Files[0].Path
			fileName = filepath.Base(filePath)
			t.Name = fileName
			t.FilePath = filePath
		}

		total := ext_utils.StringToInt64(st.TotalLength)
		completed := ext_utils.StringToInt64(st.CompletedLength)
		speed := ext_utils.StringToInt64(st.DownloadSpeed)
		totalSize = total

		t.TotalSize = total
		t.CompletedSize = completed
		t.Speed = speed
		t.ETA = ext_utils.CalculateETA(total, completed, speed)
		if total > 0 {
			t.Progress = float64(completed) / float64(total) * 100.0
		}

		// Live Update
		if st.Status == "active" && total > 0 && l.StatusMsg != nil {
			newText := themes.FormatStatusMsg(t)
			if newText != lastText {
				lastText = newText
				l.Bot.Edit(l.StatusMsg, newText, l.Markup, &tele.SendOptions{ParseMode: tele.ModeHTML})
			}
		}

		// Selesai Download
		if st.Status == "complete" {
			log.Printf("[INFO] Aria2 GID %s selesai: %s", activeGID, filePath)
			l.ProcessCompletedDownload(filePath, totalSize, t)
			ext_utils.TaskMgr.Remove(activeGID)
			return
		}

		if st.Status == "error" {
			if l.StatusMsg != nil {
				l.Bot.Edit(l.StatusMsg, fmt.Sprintf("❌ <b>Unduhan Gagal (GID %s):</b> %s", activeGID, st.ErrorMessage), &tele.SendOptions{ParseMode: tele.ModeHTML})
			}
			ext_utils.TaskMgr.Remove(activeGID)
			ext_utils.TaskMgr.Remove(l.GID)
			return
		}

		if st.Status == "removed" {
			ext_utils.TaskMgr.Remove(activeGID)
			ext_utils.TaskMgr.Remove(l.GID)
			return
		}
	}
}
