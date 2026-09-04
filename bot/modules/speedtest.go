package modules

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"go-mirror-bot/bot/helper/ext_utils"
	"go-mirror-bot/bot/helper/telegram_helper"

	tele "gopkg.in/telebot.v3"
)

type SpeedtestResult struct {
	Download float64 `json:"download"`
	Upload   float64 `json:"upload"`
	Ping     float64 `json:"ping"`
	Share    string  `json:"share"`
	Server   struct {
		Name    string  `json:"name"`
		Country string  `json:"country"`
		Sponsor string  `json:"sponsor"`
		Latency float64 `json:"latency"`
	} `json:"server"`
	Client struct {
		IP  string `json:"ip"`
		ISP string `json:"isp"`
	} `json:"client"`
}

func InitSpeedtest(b *tele.Bot) {
	handleSpeed := telegram_helper.AuthGuard(func(c tele.Context) error {
		statusMsg, _ := c.Bot().Send(c.Recipient(), "<i>Initiating Speedtest... Mohon tunggu 15-30 detik...</i>", &tele.SendOptions{ParseMode: tele.ModeHTML})

		go func() {
			cmd := exec.Command("speedtest-cli", "--json", "--share")
			outBytes, err := cmd.Output()
			if err != nil {
				c.Bot().Edit(statusMsg, fmt.Sprintf("❌ <b>Speedtest Gagal:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
				return
			}

			var res SpeedtestResult
			if err := json.Unmarshal(outBytes, &res); err != nil {
				c.Bot().Edit(statusMsg, fmt.Sprintf("❌ <b>Gagal membaca hasil:</b> %v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
				return
			}

			// Format persis WZML-X
			msg := fmt.Sprintf(
				"➲ <b><i>SPEEDTEST INFO</i></b>\n"+
					"┠ <b>Upload:</b> <code>%s/s</code>\n"+
					"┠ <b>Download:</b> <code>%s/s</code>\n"+
					"┠ <b>Ping:</b> <code>%.2f ms</code>\n\n"+
					"➲ <b><i>SPEEDTEST SERVER</i></b>\n"+
					"┠ <b>Name:</b> <code>%s</code>\n"+
					"┠ <b>Country:</b> <code>%s</code>\n"+
					"┖ <b>Sponsor:</b> <code>%s</code>\n\n"+
					"➲ <b><i>CLIENT DETAILS</i></b>\n"+
					"┠ <b>IP Address:</b> <code>%s</code>\n"+
					"┖ <b>ISP:</b> <code>%s</code>",
				ext_utils.FormatBytes(int64(res.Upload/8)),
				ext_utils.FormatBytes(int64(res.Download/8)),
				res.Ping,
				res.Server.Name,
				res.Server.Country,
				res.Server.Sponsor,
				res.Client.IP,
				res.Client.ISP,
			)

			if res.Share != "" {
				c.Bot().Delete(statusMsg)
				photo := &tele.Photo{
					File:    tele.FromURL(res.Share),
					Caption: msg,
				}
				c.Bot().Send(c.Recipient(), photo, &tele.SendOptions{ParseMode: tele.ModeHTML})
			} else {
				c.Bot().Edit(statusMsg, msg, &tele.SendOptions{ParseMode: tele.ModeHTML})
			}
		}()
		return nil
	})

	b.Handle("/speedtest", handleSpeed)
	b.Handle("/sp", handleSpeed)
}
