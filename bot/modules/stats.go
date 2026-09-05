package modules

import (
	"fmt"
	"time"

	"go-mirror-bot/bot/helper/ext_utils"
	"go-mirror-bot/bot/helper/telegram_helper"
	"go-mirror-bot/bot/helper/themes"

	tele "gopkg.in/telebot.v3"
)

func InitStats(b *tele.Bot) {
	b.Handle(telegram_helper.BotCommands.StartCommand, func(c tele.Context) error {
		if c.Chat().Type == tele.ChatPrivate && ext_utils.DB != nil {
			go ext_utils.DB.AddPMUser(c.Sender().ID)
		}
		msg := "<i>This bot can mirror all your links|files|torrents to Google Drive or any rclone cloud or to telegram.</i>\n\n" +
			"<b>Type /help to get a list of available commands</b>"
		return c.Send(msg, &tele.SendOptions{ParseMode: tele.ModeHTML})
	})

	handleHelp := func(c tele.Context) error {
		helpText := "㊂ <b><i>WZML-X Go Help Guide Menu!</i></b>\n\n" +
			"<b>Aria2, DDL & GDrive Mirror / Leech:</b>\n" +
			"• <code>/mirror</code>, <code>/m</code> - Direct Link, GDrive, Mega, MediaFire, Magnet, Torrent ke Cloud\n" +
			"• <code>/leech</code>, <code>/l</code> - Download & kirim langsung ke Telegram\n" +
			"• <code>/zipmirror</code>, <code>/zm</code> | <code>/zipleech</code>, <code>/zl</code> - Kompres Zip\n" +
			"• <code>/unzipmirror</code>, <code>/uzm</code> | <code>/unzipleech</code>, <code>/uzl</code> - Ekstrak Arsip\n\n" +
			"<b>YT-DLP (YouTube & Video Sites):</b>\n" +
			"• <code>/ytdl</code>, <code>/y</code> - Unduh video/audio ke Cloud\n" +
			"• <code>/ytdlleech</code>, <code>/yl</code> - Unduh video/audio ke Telegram\n" +
			"• <code>/ytdlzip</code>, <code>/yz</code> | <code>/ytdlzipleech</code>, <code>/yzl</code>\n\n" +
			"<b>Cloud & Google Drive Tools:</b>\n" +
			"• <code>/clone &lt;src&gt; &lt;dst&gt;</code>, <code>/c</code> - Salin GDrive-to-GDrive atau Remote Cloud\n" +
			"• <code>/count &lt;link/remote&gt;</code> - Hitung total file & ukuran GDrive/Remote\n" +
			"• <code>/del &lt;gdrive_link&gt;</code> - Hapus file/folder Google Drive\n" +
			"• <code>/list &lt;query&gt;</code> - Cari file/folder di Google Drive\n" +
			"• <code>/gdclean</code> - Bersihkan folder/sampah Google Drive (Sudo)\n\n" +
			"<b>Pengaturan & Admin:</b>\n" +
			"• <code>/usersettings</code>, <code>/us</code> - Kelola thumbnail, caption, prefix, API\n" +
			"• <code>/authorize</code> | <code>/unauthorize</code> | <code>/authlist</code>\n" +
			"• <code>/addsudo</code> | <code>/rmsudo</code> | <code>/blacklist</code> | <code>/rmblacklist</code>\n" +
			"• <code>/status</code>, <code>/s</code> | <code>/cancel &lt;gid&gt;</code> | <code>/cancelall</code>\n" +
			"• <code>/stats</code>, <code>/st</code> | <code>/ping</code>, <code>/p</code> | <code>/shell &lt;cmd&gt;</code>"
		return c.Send(helpText, &tele.SendOptions{ParseMode: tele.ModeHTML})
	}
	for _, cmd := range telegram_helper.BotCommands.HelpCommand {
		b.Handle(cmd, handleHelp)
	}

	handlePing := func(c tele.Context) error {
		start := time.Now()
		latency := time.Since(start).Milliseconds()
		return c.Send(fmt.Sprintf("<b>Pong</b>\n<code>%d ms..</code>", latency), &tele.SendOptions{ParseMode: tele.ModeHTML})
	}
	for _, cmd := range telegram_helper.BotCommands.PingCommand {
		b.Handle(cmd, handlePing)
	}

	handleStats := func(c tele.Context) error {
		st := themes.CollectStats(ext_utils.TaskMgr.BotStartTime)
		statsMsg := fmt.Sprintf(
			"⌬ <b><i>BOT STATISTICS :</i></b>\n"+
				"┖ <b>Bot Uptime :</b> %s\n\n"+
				"┎ <b><i>RAM ( MEMORY ) :</i></b>\n"+
				"┃ %s %.1f%%\n"+
				"┖ <b>U :</b> %s | <b>F :</b> %s | <b>T :</b> %s\n\n"+
				"┎ <b><i>DISK :</i></b>\n"+
				"┃ %s %.1f%%\n"+
				"┖ <b>U :</b> %s | <b>F :</b> %s | <b>T :</b> %s\n\n"+
				"⌬ <b><i>OS SYSTEM :</i></b>\n"+
				"┠ <b>OS Arch :</b> %s\n"+
				"┖ <b>Go Runtime :</b> %s",
			st.Uptime,
			themes.GenerateBar(st.RAMPercent), st.RAMPercent,
			st.RAMUsed, st.RAMFree, st.RAMTotal,
			themes.GenerateBar(st.DiskPercent), st.DiskPercent,
			st.DiskUsed, st.DiskFree, st.DiskTotal,
			st.OSArch,
			st.GoVersion,
		)
		return c.Send(statsMsg, &tele.SendOptions{ParseMode: tele.ModeHTML})
	}
	for _, cmd := range telegram_helper.BotCommands.StatsCommand {
		b.Handle(cmd, handleStats)
	}
}
