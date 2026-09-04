package status_utils

import (
	"strings"

	"go-mirror-bot/bot/helper/ext_utils"
	"go-mirror-bot/bot/helper/themes"
)

func GetStatusMessage() string {
	tasks := ext_utils.TaskMgr.All()
	stats := themes.CollectStats(ext_utils.TaskMgr.BotStartTime)

	if len(tasks) == 0 {
		return "<i>No Active Downloads!</i>\n\n" + themes.FormatFooter(0, stats, 0)
	}

	var sb strings.Builder
	var totalSpeed int64
	for _, t := range tasks {
		totalSpeed += t.Speed
		sb.WriteString(themes.FormatStatusMsg(t))
		sb.WriteString("\n\n")
	}

	sb.WriteString(themes.FormatFooter(len(tasks), stats, totalSpeed))
	return sb.String()
}
