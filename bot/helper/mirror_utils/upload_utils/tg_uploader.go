package upload_utils

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go-mirror-bot/bot/helper/ext_utils"

	tele "gopkg.in/telebot.v3"
)

func TelegramUpload(b *tele.Bot, recipient tele.Recipient, localPath string, userTag string) error {
	fi, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("file tidak ditemukan: %w", err)
	}

	targetPath := localPath
	if fi.IsDir() {
		zipPath, err := ext_utils.CompressToZip(localPath)
		if err != nil {
			return fmt.Errorf("gagal mengompres folder leech: %w", err)
		}
		targetPath = zipPath
		fi, _ = os.Stat(targetPath)
	}

	const maxTgSize = 49 * 1024 * 1024

	if fi.Size() > maxTgSize {
		b.Send(recipient, fmt.Sprintf("⚠️ <b>Ukuran (%s) &gt; 50MB (Batas Bot API).</b>\n✂️ <i>Memecah file otomatis 49MB...</i>",
			ext_utils.FormatBytes(fi.Size())), &tele.SendOptions{ParseMode: tele.ModeHTML})

		parts, err := ext_utils.SplitArchive(targetPath)
		if err != nil {
			return err
		}

		sort.Strings(parts)
		for idx, part := range parts {
			partInfo, _ := os.Stat(part)
			partName := filepath.Base(part)
			caption := fmt.Sprintf("📦 <b>Part [%d/%d]:</b> <code>%s</code>\n📊 <b>Ukuran:</b> %s\n👤 <b>By:</b> %s",
				idx+1, len(parts), partName, ext_utils.FormatBytes(partInfo.Size()), userTag)

			doc := &tele.Document{
				File:     tele.FromDisk(part),
				FileName: partName,
				Caption:  caption,
			}
			b.Send(recipient, doc, &tele.SendOptions{ParseMode: tele.ModeHTML})
			os.Remove(part)
		}
		return nil
	}

	fileName := filepath.Base(targetPath)
	caption := fmt.Sprintf("📁 <b>File:</b> <code>%s</code>\n📊 <b>Ukuran:</b> %s\n👤 <b>By:</b> %s",
		fileName, ext_utils.FormatBytes(fi.Size()), userTag)

	ext := strings.ToLower(filepath.Ext(targetPath))
	switch ext {
	case ".mp4", ".mkv", ".mov":
		video := &tele.Video{File: tele.FromDisk(targetPath), Caption: caption, FileName: fileName}
		_, err = b.Send(recipient, video, &tele.SendOptions{ParseMode: tele.ModeHTML})
	case ".mp3", ".flac", ".wav", ".m4a":
		audio := &tele.Audio{File: tele.FromDisk(targetPath), Caption: caption, FileName: fileName}
		_, err = b.Send(recipient, audio, &tele.SendOptions{ParseMode: tele.ModeHTML})
	default:
		doc := &tele.Document{File: tele.FromDisk(targetPath), Caption: caption, FileName: fileName}
		_, err = b.Send(recipient, doc, &tele.SendOptions{ParseMode: tele.ModeHTML})
	}

	if err != nil {
		doc := &tele.Document{File: tele.FromDisk(targetPath), Caption: caption, FileName: fileName}
		_, err = b.Send(recipient, doc, &tele.SendOptions{ParseMode: tele.ModeHTML})
	}
	return err
}
