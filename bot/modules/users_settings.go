package modules

import (
	"fmt"
	"strings"

	"go-mirror-bot/bot/helper/ext_utils"
	"go-mirror-bot/bot/helper/mirror_utils/upload_utils/ddlserver"
	"go-mirror-bot/bot/helper/telegram_helper"

	tele "gopkg.in/telebot.v3"
)

func InitUsersSettings(b *tele.Bot) {
	// 1. Dashboard User Settings
	handleUserSettings := telegram_helper.AuthGuard(func(c tele.Context) error {
		userID := c.Sender().ID
		u := ext_utils.UserStore.Get(userID)

		thumbStatus := "❌ Belum disetel (Default)"
		if u.HasThumbnail {
			thumbStatus = "✅ Kustom (Aktif)"
		}

		captionStatus := "❌ Default"
		if u.CustomCaption != "" {
			captionStatus = fmt.Sprintf("✅ <code>%s</code>", u.CustomCaption)
		}

		prefixStatus := "❌ Tidak ada"
		if u.LeechPrefix != "" {
			prefixStatus = fmt.Sprintf("✅ <code>%s</code>", u.LeechPrefix)
		}

		suffixStatus := "❌ Tidak ada"
		if u.LeechSuffix != "" {
			suffixStatus = fmt.Sprintf("✅ <code>%s</code>", u.LeechSuffix)
		}

		pdStatus := "❌ Default Bot"
		if u.PixeldrainAPI != "" {
			pdStatus = "✅ Kustom Pribadi"
		}

		text := fmt.Sprintf(
			"⚙️ <b><i>PENGATURAN PENGGUNA (WZML-X)</i></b>\n"+
				"┠ <b>User:</b> @%s (<code>%d</code>)\n"+
				"┠ <b>Custom Thumbnail:</b> %s\n"+
				"┠ <b>Custom Caption:</b> %s\n"+
				"┠ <b>Leech Prefix:</b> %s\n"+
				"┠ <b>Leech Suffix:</b> %s\n"+
				"┖ <b>Pixeldrain API:</b> %s\n\n"+
				"💡 <i>Gunakan tombol di bawah atau perintah langsung:\n"+
				"• Reply foto dengan /setthumb\n"+
				"• /setcaption &lt;teks&gt;\n"+
				"• /setprefix &lt;teks&gt; | /setsuffix &lt;teks&gt;\n"+
				"• /setpdapi &lt;api_key&gt; | /delpdapi</i>",
			c.Sender().Username, userID, thumbStatus, captionStatus, prefixStatus, suffixStatus, pdStatus,
		)

		markup := &tele.ReplyMarkup{}
		btnDelThumb := markup.Data("🗑 Hapus Thumbnail", "us_del_thumb")
		btnDelCap := markup.Data("🗑 Reset Caption", "us_del_cap")
		btnDelPd := markup.Data("🗑 Reset Pixeldrain API", "us_del_pd")
		btnClose := markup.Data("❌ Tutup", "us_close")

		markup.Inline(
			markup.Row(btnDelThumb, btnDelCap),
			markup.Row(btnDelPd),
			markup.Row(btnClose),
		)

		return c.Send(text, markup, &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	// 2. /setthumb (Reply foto)
	handleSetThumb := telegram_helper.AuthGuard(func(c tele.Context) error {
		userID := c.Sender().ID
		var photo *tele.Photo

		if c.Message().ReplyTo != nil && c.Message().ReplyTo.Photo != nil {
			photo = c.Message().ReplyTo.Photo
		} else if c.Message().Photo != nil {
			photo = c.Message().Photo
		}

		if photo == nil {
			return c.Send("⚠️ Silakan kirim atau reply foto dengan perintah <code>/setthumb</code> untuk dijadikan thumbnail kustom.", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}

		reader, err := c.Bot().File(&photo.File)
		if err != nil {
			return c.Send(fmt.Sprintf("❌ Gagal mengunduh foto: %v", err))
		}
		defer reader.Close()

		if err := ext_utils.UserStore.SaveThumbnail(userID, reader); err != nil {
			return c.Send(fmt.Sprintf("❌ Gagal menyimpan thumbnail: %v", err))
		}

		return c.Send("✅ <b>Custom thumbnail berhasil disimpan!</b> Thumbnail ini akan otomatis ditempel pada setiap file leech Anda.", &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	// 3. /delthumb
	handleDelThumb := telegram_helper.AuthGuard(func(c tele.Context) error {
		userID := c.Sender().ID
		_ = ext_utils.UserStore.DeleteThumbnail(userID)
		return c.Send("🗑 <b>Custom thumbnail berhasil dihapus.</b>", &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	// 4. /mythumb
	handleMyThumb := telegram_helper.AuthGuard(func(c tele.Context) error {
		userID := c.Sender().ID
		thumbPath := ext_utils.UserStore.GetThumbnailPath(userID)
		if thumbPath == "" {
			return c.Send("ℹ️ Anda belum memiliki custom thumbnail. Reply foto dengan <code>/setthumb</code>.", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}

		photo := &tele.Photo{
			File:    tele.FromDisk(thumbPath),
			Caption: "🖼 <b>Custom Thumbnail Anda</b>",
		}
		return c.Send(photo, &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	// 5. /setcaption
	handleSetCaption := telegram_helper.AuthGuard(func(c tele.Context) error {
		userID := c.Sender().ID
		args := c.Args()
		caption := strings.Join(args, " ")
		if caption == "" && c.Message().ReplyTo != nil {
			caption = c.Message().ReplyTo.Text
		}

		if caption == "" {
			return c.Send("⚠️ Format: <code>/setcaption [teks caption kustom]</code>\nVariabel yang didukung: <code>{filename}</code>, <code>{size}</code>, <code>{user}</code>", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}

		ext_utils.UserStore.SetCaption(userID, caption)
		return c.Send(fmt.Sprintf("✅ <b>Custom caption disimpan:</b>\n<code>%s</code>", caption), &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	// 6. /delcaption
	handleDelCaption := telegram_helper.AuthGuard(func(c tele.Context) error {
		userID := c.Sender().ID
		ext_utils.UserStore.SetCaption(userID, "")
		return c.Send("🗑 <b>Custom caption berhasil dihapus (kembali ke default).</b>", &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	// 7. /setprefix
	handleSetPrefix := telegram_helper.AuthGuard(func(c tele.Context) error {
		userID := c.Sender().ID
		prefix := strings.Join(c.Args(), " ")
		ext_utils.UserStore.SetPrefix(userID, prefix)
		if prefix == "" {
			return c.Send("🗑 <b>Leech prefix dinonaktifkan.</b>", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}
		return c.Send(fmt.Sprintf("✅ <b>Leech prefix disimpan:</b> <code>%s</code>", prefix), &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	// 8. /setsuffix
	handleSetSuffix := telegram_helper.AuthGuard(func(c tele.Context) error {
		userID := c.Sender().ID
		suffix := strings.Join(c.Args(), " ")
		ext_utils.UserStore.SetSuffix(userID, suffix)
		if suffix == "" {
			return c.Send("🗑 <b>Leech suffix dinonaktifkan.</b>", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}
		return c.Send(fmt.Sprintf("✅ <b>Leech suffix disimpan:</b> <code>%s</code>", suffix), &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	// 9. /setpdapi (Pixeldrain API)
	handleSetPdApi := telegram_helper.AuthGuard(func(c tele.Context) error {
		userID := c.Sender().ID
		if len(c.Args()) == 0 {
			return c.Send("⚠️ Format: <code>/setpdapi &lt;API_KEY&gt;</code>\nDapatkan API key di https://pixeldrain.com/user/api_keys", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}
		apiKey := c.Args()[0]
		pd := ddlserver.NewPixeldrain(apiKey)
		if !pd.IsPdApi(apiKey) {
			return c.Send("❌ <b>API Key Pixeldrain tidak valid!</b> Periksa kembali API key akun Anda.", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}

		ext_utils.UserStore.SetPixeldrainAPI(userID, apiKey)
		return c.Send("✅ <b>Pixeldrain API Key berhasil disimpan!</b> Unggahan DDL sekarang akan menggunakan akun pribadi Anda.", &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	// 10. /delpdapi
	handleDelPdApi := telegram_helper.AuthGuard(func(c tele.Context) error {
		userID := c.Sender().ID
		ext_utils.UserStore.SetPixeldrainAPI(userID, "")
		return c.Send("🗑 <b>Pixeldrain API Key berhasil dihapus.</b>", &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	// 11. /mypdapi
	handleMyPdApi := telegram_helper.AuthGuard(func(c tele.Context) error {
		userID := c.Sender().ID
		u := ext_utils.UserStore.Get(userID)
		if u.PixeldrainAPI == "" {
			return c.Send("ℹ️ Anda belum menyetel API Key Pixeldrain kustom. Gunakan <code>/setpdapi &lt;key&gt;</code>.", &tele.SendOptions{ParseMode: tele.ModeHTML})
		}
		return c.Send(fmt.Sprintf("🔑 <b>Pixeldrain API Key Anda:</b> <code>%s</code>", u.PixeldrainAPI), &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	// Button Callbacks
	b.Handle(&tele.Btn{Unique: "us_del_thumb"}, func(c tele.Context) error {
		userID := c.Sender().ID
		_ = ext_utils.UserStore.DeleteThumbnail(userID)
		_ = c.Respond(&tele.CallbackResponse{Text: "Thumbnail dihapus!"})
		return c.Edit("🗑 <b>Custom thumbnail berhasil dihapus.</b>", &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	b.Handle(&tele.Btn{Unique: "us_del_cap"}, func(c tele.Context) error {
		userID := c.Sender().ID
		ext_utils.UserStore.SetCaption(userID, "")
		_ = c.Respond(&tele.CallbackResponse{Text: "Caption direset!"})
		return c.Edit("🗑 <b>Custom caption berhasil direset.</b>", &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	b.Handle(&tele.Btn{Unique: "us_del_pd"}, func(c tele.Context) error {
		userID := c.Sender().ID
		ext_utils.UserStore.SetPixeldrainAPI(userID, "")
		_ = c.Respond(&tele.CallbackResponse{Text: "Pixeldrain API dihapus!"})
		return c.Edit("🗑 <b>Pixeldrain API Key berhasil direset.</b>", &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	b.Handle(&tele.Btn{Unique: "us_close"}, func(c tele.Context) error {
		return c.Delete()
	})

	// Registrasi Commands
	b.Handle("/usersettings", handleUserSettings)
	b.Handle("/usetting", handleUserSettings)
	b.Handle("/us", handleUserSettings)

	b.Handle("/setthumb", handleSetThumb)
	b.Handle("/delthumb", handleDelThumb)
	b.Handle("/remthumb", handleDelThumb)
	b.Handle("/mythumb", handleMyThumb)

	b.Handle("/setcaption", handleSetCaption)
	b.Handle("/delcaption", handleDelCaption)
	b.Handle("/setprefix", handleSetPrefix)
	b.Handle("/setsuffix", handleSetSuffix)

	b.Handle("/setpdapi", handleSetPdApi)
	b.Handle("/delpdapi", handleDelPdApi)
	b.Handle("/mypdapi", handleMyPdApi)
}
