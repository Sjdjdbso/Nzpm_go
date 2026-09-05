package modules

import (
	"fmt"
	"strings"

	"go-mirror-bot/bot/helper/mirror_utils/upload_utils"
	"go-mirror-bot/bot/helper/telegram_helper"

	tele "gopkg.in/telebot.v3"
)

func InitRcloneList(b *tele.Bot) {
	// Handler /rcl untuk membuka interactive Rclone remote & folder explorer
	b.Handle("/rcl", telegram_helper.AuthGuard(func(c tele.Context) error {
		userID := c.Sender().ID
		confPath := upload_utils.GetUserRcloneConf(userID)

		remotes, err := upload_utils.RcloneListRemotes(confPath)
		if err != nil || len(remotes) == 0 {
			return c.Send("⚠️ Tidak ada remote Rclone yang ditemukan! Pastikan Anda telah mengunggah <code>rclone.conf</code> dengan perintah <code>/setrclone</code>.", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}

		markup := &tele.ReplyMarkup{}
		var rows []tele.Row

		for _, rem := range remotes {
			btn := markup.Data("🌐 "+rem, "rcl_nav", rem, "")
			rows = append(rows, markup.Row(btn))
		}

		btnClose := markup.Data("❌ Tutup", "rcl_close")
		rows = append(rows, markup.Row(btnClose))
		markup.Inline(rows...)

		header := "📁 <b><i>RCLONE CLOUD EXPLORER</i></b>\n\nSilakan pilih remote penyimpanan cloud yang ingin dijelajahi:"
		return c.Send(header, markup, &tele.SendOptions{ParseMode: tele.ModeHTML})
	}))

	// Callback navigasi folder rcl
	b.Handle(&tele.Btn{Unique: "rcl_nav"}, func(c tele.Context) error {
		parts := strings.Split(c.Data(), "|")
		remote := parts[0]
		currentPath := ""
		if len(parts) > 1 {
			currentPath = parts[1]
		}

		userID := c.Sender().ID
		confPath := upload_utils.GetUserRcloneConf(userID)

		fullRemotePath := remote
		if currentPath != "" {
			fullRemotePath = strings.TrimSuffix(remote, ":") + ":" + currentPath
		}

		dirs, _ := upload_utils.RcloneListDirs(fullRemotePath, confPath)

		markup := &tele.ReplyMarkup{}
		var rows []tele.Row

		// Tampilkan folder maksimal 10
		maxDirs := 10
		if len(dirs) < maxDirs {
			maxDirs = len(dirs)
		}

		for i := 0; i < maxDirs; i++ {
			d := dirs[i]
			subPath := d
			if currentPath != "" {
				subPath = currentPath + "/" + d
			}
			btn := markup.Data("📁 "+d, "rcl_nav", fmt.Sprintf("%s|%s", remote, subPath))
			rows = append(rows, markup.Row(btn))
		}

		// Tombol aksi
		var navRow []tele.Btn
		if currentPath != "" {
			parentPath := ""
			if idx := strings.LastIndex(currentPath, "/"); idx != -1 {
				parentPath = currentPath[:idx]
			}
			btnBack := markup.Data("⬅️ Kembali", "rcl_nav", fmt.Sprintf("%s|%s", remote, parentPath))
			navRow = append(navRow, btnBack)
		} else {
			btnRemotes := markup.Data("⬅️ Daftar Remote", "rcl_remotes")
			navRow = append(navRow, btnRemotes)
		}

		btnSelect := markup.Data("✅ Salin Path", "rcl_sel", fullRemotePath)
		navRow = append(navRow, btnSelect)
		rows = append(rows, markup.Row(navRow...))

		btnClose := markup.Data("❌ Tutup", "rcl_close")
		rows = append(rows, markup.Row(btnClose))
		markup.Inline(rows...)

		header := fmt.Sprintf("📁 <b><i>RCLONE EXPLORER</i></b>\n\n➲ <b>Lokasi:</b> <code>%s</code>\n➲ <b>Folder Ditemukan:</b> %d", fullRemotePath, len(dirs))
		return c.Edit(header, markup, &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	// Callback kembali ke daftar remote
	b.Handle(&tele.Btn{Unique: "rcl_remotes"}, func(c tele.Context) error {
		userID := c.Sender().ID
		confPath := upload_utils.GetUserRcloneConf(userID)

		remotes, err := upload_utils.RcloneListRemotes(confPath)
		if err != nil || len(remotes) == 0 {
			return c.Edit("⚠️ Tidak ada remote Rclone yang ditemukan.", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}

		markup := &tele.ReplyMarkup{}
		var rows []tele.Row

		for _, rem := range remotes {
			btn := markup.Data("🌐 "+rem, "rcl_nav", rem)
			rows = append(rows, markup.Row(btn))
		}

		btnClose := markup.Data("❌ Tutup", "rcl_close")
		rows = append(rows, markup.Row(btnClose))
		markup.Inline(rows...)

		header := "📁 <b><i>RCLONE CLOUD EXPLORER</i></b>\n\nSilakan pilih remote penyimpanan cloud yang ingin dijelajahi:"
		return c.Edit(header, markup, &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	// Callback pilih path
	b.Handle(&tele.Btn{Unique: "rcl_sel"}, func(c tele.Context) error {
		selectedPath := c.Data()
		return c.Send(fmt.Sprintf("📋 <b>Path Rclone Terpilih:</b>\n<code>%s</code>\n\n💡 <i>Gunakan path ini pada perintah mirror atau clone:\n<code>/mirror &lt;link&gt; -rc %s</code></i>", selectedPath, selectedPath), &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	// Callback tutup
	b.Handle(&tele.Btn{Unique: "rcl_close"}, func(c tele.Context) error {
		return c.Delete()
	})
}
