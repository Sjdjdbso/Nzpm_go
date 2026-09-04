package listeners

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"go-mirror-bot/bot/helper/ext_utils"
	"go-mirror-bot/bot/helper/mirror_utils/download_utils"
	"go-mirror-bot/bot/helper/mirror_utils/upload_utils"
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
			t.Progress = 100.0

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
				if err := upload_utils.TelegramUpload(l.Bot, l.Recipient, filePath, t.User); err != nil {
					l.Bot.Send(l.Recipient, fmt.Sprintf("❌ <b>Leech Gagal:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
				} else {
					l.Bot.Send(l.Recipient, themes.FormatCompleteMsg(fileName, totalSize, "Leech", "", t.User), &tele.SendOptions{ParseMode: tele.ModeHTML})
				}
			} else {
				t.Status = "Uploading"
				if l.StatusMsg != nil {
					l.Bot.Edit(l.StatusMsg, fmt.Sprintf("📤 <b>Unduhan Selesai!</b>\n📁 <code>%s</code>\n🚀 <i>Mengunggah ke Cloud...</i>", fileName), &tele.SendOptions{ParseMode: tele.ModeHTML})
				}
				if err := upload_utils.RcloneTransfer(filePath, l.RcloneDest, nil); err != nil {
					l.Bot.Send(l.Recipient, fmt.Sprintf("❌ <b>Upload Gagal:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
				} else {
					dest := l.RcloneDest
					if dest == "" {
						dest = "Disimpan di server"
					}
					l.Bot.Send(l.Recipient, themes.FormatCompleteMsg(fileName, totalSize, "Mirror", dest, t.User), &tele.SendOptions{ParseMode: tele.ModeHTML})
				}
			}

			ext_utils.CleanPath(filePath)
			ext_utils.TaskMgr.Remove(activeGID)
			ext_utils.TaskMgr.Remove(l.GID)
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
