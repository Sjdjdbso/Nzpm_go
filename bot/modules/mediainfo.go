package modules

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"go-mirror-bot/bot/helper/telegram_helper"

	tele "gopkg.in/telebot.v3"
)

func InitMediaInfo(b *tele.Bot) {
	handleMediaInfo := telegram_helper.AuthGuard(func(c tele.Context) error {
		statusMsg, _ := c.Bot().Send(c.Recipient(), "<i>Generating MediaInfo... Mohon tunggu...</i>", &tele.SendOptions{ParseMode: tele.ModeHTML})

		var targetFile string
		var needCleanup bool

		// 1. Cek jika mereply dokumen/video
		if c.Message().ReplyTo != nil {
			reply := c.Message().ReplyTo
			var file tele.File
			var fileName string

			if reply.Document != nil {
				file = reply.Document.File
				fileName = reply.Document.FileName
			} else if reply.Video != nil {
				file = reply.Video.File
				fileName = reply.Video.FileName
				if fileName == "" {
					fileName = "video.mp4"
				}
			} else if reply.Audio != nil {
				file = reply.Audio.File
				fileName = reply.Audio.FileName
				if fileName == "" {
					fileName = "audio.mp3"
				}
			}

			if file.FileID != "" {
				tmpPath := filepath.Join("downloads", fmt.Sprintf("mi_%d_%s", os.Getpid(), fileName))
				reader, err := c.Bot().File(&file)
				if err == nil {
					defer reader.Close()
					out, err := os.Create(tmpPath)
					if err == nil {
						io.CopyN(out, reader, 10*1024*1024) // Salin 10MB pertama cukup untuk MediaInfo
						out.Close()
						targetFile = tmpPath
						needCleanup = true
					}
				}
			}
		}

		// 2. Cek jika memberikan link langsung: /mediainfo https://...
		if targetFile == "" && len(c.Args()) > 0 {
			targetURL := c.Args()[0]
			resp, err := http.Get(targetURL)
			if err == nil {
				defer resp.Body.Close()
				tmpPath := filepath.Join("downloads", fmt.Sprintf("mi_url_%d.tmp", os.Getpid()))
				out, err := os.Create(tmpPath)
				if err == nil {
					io.CopyN(out, resp.Body, 10*1024*1024) // 10MB pertama untuk membaca header media
					out.Close()
					targetFile = tmpPath
					needCleanup = true
				}
			}
		}

		if targetFile == "" {
			_, err := c.Bot().Edit(statusMsg, "⚠️ Reply file video/audio atau masukkan URL direct media. Contoh:\n<code>/mediainfo https://domain.com/video.mp4</code>", &tele.SendOptions{ParseMode: tele.ModeHTML})
			return err
		}

		go func() {
			defer func() {
				if needCleanup {
					os.Remove(targetFile)
				}
			}()

			cmd := exec.Command("mediainfo", targetFile)
			outBytes, err := cmd.CombinedOutput()
			if err != nil {
				c.Bot().Edit(statusMsg, fmt.Sprintf("❌ <b>MediaInfo Error:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
				return
			}

			outputStr := strings.TrimSpace(string(outBytes))
			if len(outputStr) > 4000 {
				outputStr = outputStr[:4000] + "\n...(terpotong)"
			}

			c.Bot().Edit(statusMsg, fmt.Sprintf("📑 <b><i>MediaInfo Details:</i></b>\n<pre>%s</pre>", outputStr), &tele.SendOptions{ParseMode: tele.ModeHTML})
		}()

		return nil
	})

	b.Handle("/mediainfo", handleMediaInfo)
	b.Handle("/mi", handleMediaInfo)
}
