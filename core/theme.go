package core

import (
	"fmt"
	"strings"
	"time"
)

// Simbol Progress Bar khas WZML-X
const (
	ProgFull  = "■"
	ProgEmpty = "□"
)

// GenerateWZMLBar membuat bar: ■■■■■■□□□□
func GenerateWZMLBar(percent float64) string {
	const totalBlocks = 10
	filled := int((percent / 100.0) * totalBlocks)
	if filled > totalBlocks {
		filled = totalBlocks
	}
	if filled < 0 {
		filled = 0
	}
	empty := totalBlocks - filled
	return strings.Repeat(ProgFull, filled) + strings.Repeat(ProgEmpty, empty)
}

// FormatWZMLTaskStatus menyusun tampilan status progres per task mirip WZML-X
func FormatWZMLTaskStatus(name string, progress float64, completed, total, speed int64, eta, status, mode, user string, userID int64, gid string, startTime time.Time) string {
	bar := GenerateWZMLBar(progress)
	elapsed := formatDuration(time.Since(startTime))
	if eta == "" {
		eta = "-"
	}

	return fmt.Sprintf(
		"<b><i>%s</i></b>\n"+
			"┃ %s\n"+
			"┠ <b>Processed:</b> %s of %s (%.2f%%)\n"+
			"┠ <b>Status:</b> %s | <b>ETA:</b> %s\n"+
			"┠ <b>Speed:</b> %s | <b>Elapsed:</b> %s\n"+
			"┠ <b>Engine:</b> Aria2c\n"+
			"┠ <b>Mode:</b> %s\n"+
			"┠ <b>User:</b> %s | <b>ID:</b> <code>%d</code>\n"+
			"┖ <code>/cancel %s</code>",
		name,
		bar,
		FormatBytes(completed), FormatBytes(total), progress,
		status, eta,
		FormatSpeed(speed), elapsed,
		mode,
		user, userID,
		gid,
	)
}

// FormatWZMLFooter menyusun footer status keseluruhan
func FormatWZMLFooter(taskCount int, stats SysStats, totalSpeed int64) string {
	return fmt.Sprintf(
		"⌬ <b><i>Bot Stats</i></b>\n"+
			"┠ <b>Tasks:</b> %d\n"+
			"┠ <b>CPU:</b> %.1f%% | <b>F:</b> %s [%.1f%%]\n"+
			"┠ <b>RAM:</b> %.1f%% | <b>UPTIME:</b> %s\n"+
			"┖ <b>DL:</b> %s | <b>UL:</b> 0 B/s",
		taskCount,
		stats.CPUPercent, stats.DiskFree, 100-stats.DiskPercent,
		stats.RAMPercent, stats.Uptime,
		FormatSpeed(totalSpeed),
	)
}

// FormatWZMLComplete menyusun pesan sukses persis WZML-X
func FormatWZMLComplete(name string, size int64, elapsed string, mode string, destPath string, userTag string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b><i>%s</i></b>\n┃\n", name))
	sb.WriteString(fmt.Sprintf("┠ <b>Size:</b> %s\n", FormatBytes(size)))
	sb.WriteString(fmt.Sprintf("┠ <b>Elapsed:</b> %s\n", elapsed))
	sb.WriteString(fmt.Sprintf("┠ <b>Mode:</b> %s\n", mode))
	if destPath != "" {
		sb.WriteString(fmt.Sprintf("┠ <b>Path:</b> <code>%s</code>\n", destPath))
	}
	sb.WriteString(fmt.Sprintf("┖ <b>By:</b> %s", userTag))
	return sb.String()
}
