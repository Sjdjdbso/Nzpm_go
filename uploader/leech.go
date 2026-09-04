package uploader

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go-mirror-bot/core"

	tele "gopkg.in/telebot.v3"
)

// LeechToTelegram mengunggah file hasil download langsung ke obrolan Telegram
func LeechToTelegram(b *tele.Bot, recipient tele.Recipient, localPath string, userTag string) error {
	fi, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("file tidak ditemukan: %w", err)
	}

	// Jika berupa folder (misal dari torrent multi-file atau hasil ekstrak), jadikan zip terlebih dahulu
	targetPath := localPath
	if fi.IsDir() {
		zipPath, err := core.CompressToZip(localPath)
		if err != nil {
			return fmt.Errorf("gagal mengompres folder untuk leech: %w", err)
		}
		targetPath = zipPath
		fi, _ = os.Stat(targetPath)
	}

	const maxTgSize = 49 * 1024 * 1024 // 49 MB batas aman Telegram Bot API

	// Jika file melebihi batas 49 MB, lakukan auto-split
	if fi.Size() > maxTgSize {
		b.Send(recipient, fmt.Sprintf("⚠️ <b>Ukuran file (%s) melebihi batas Telegram Bot API (50 MB).</b>\n✂️ <i>Memecah file otomatis menjadi beberapa bagian 49MB...</i>",
			core.FormatBytes(fi.Size())), &tele.SendOptions{ParseMode: tele.ModeHTML})

		parts, err := core.SplitArchive(targetPath)
		if err != nil {
			return fmt.Errorf("gagal memecah file: %w", err)
		}

		sort.Strings(parts)
		totalParts := len(parts)

		for idx, part := range parts {
			partInfo, _ := os.Stat(part)
			partName := filepath.Base(part)

			caption := fmt.Sprintf(
				"📦 <b>Part [%d/%d]:</b> <code>%s</code>\n"+
					"📊 <b>Ukuran:</b> %s\n"+
					"👤 <b>Pengguna:</b> %s",
				idx+1, totalParts, partName,
				core.FormatBytes(partInfo.Size()),
				userTag,
			)

			doc := &tele.Document{
				File:     tele.FromDisk(part),
				FileName: partName,
				Caption:  caption,
			}

			_, err := b.Send(recipient, doc, &tele.SendOptions{ParseMode: tele.ModeHTML})
			if err != nil {
				log.Printf("[ERROR] Gagal mengirim part %s: %v", partName, err)
			}

			// Hapus part setelah dikirim
			os.Remove(part)
		}

		return nil
	}

	// File <= 49MB, kirim langsung
	fileName := filepath.Base(targetPath)
	caption := fmt.Sprintf(
		"📁 <b>File:</b> <code>%s</code>\n"+
			"📊 <b>Ukuran:</b> %s\n"+
			"👤 <b>Pengguna:</b> %s",
		fileName,
		core.FormatBytes(fi.Size()),
		userTag,
	)

	ext := strings.ToLower(filepath.Ext(targetPath))
	switch ext {
	case ".mp4", ".mkv", ".mov":
		video := &tele.Video{
			File:     tele.FromDisk(targetPath),
			Caption:  caption,
			FileName: fileName,
		}
		_, err = b.Send(recipient, video, &tele.SendOptions{ParseMode: tele.ModeHTML})
	case ".mp3", ".flac", ".wav", ".m4a":
		audio := &tele.Audio{
			File:     tele.FromDisk(targetPath),
			Caption:  caption,
			FileName: fileName,
		}
		_, err = b.Send(recipient, audio, &tele.SendOptions{ParseMode: tele.ModeHTML})
	default:
		doc := &tele.Document{
			File:     tele.FromDisk(targetPath),
			Caption:  caption,
			FileName: fileName,
		}
		_, err = b.Send(recipient, doc, &tele.SendOptions{ParseMode: tele.ModeHTML})
	}

	if err != nil {
		// Fallback ke Document jika send video/audio gagal
		doc := &tele.Document{
			File:     tele.FromDisk(targetPath),
			Caption:  caption,
			FileName: fileName,
		}
		_, err = b.Send(recipient, doc, &tele.SendOptions{ParseMode: tele.ModeHTML})
	}

	return err
}
