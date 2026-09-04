package ext_utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func FormatSpeed(bytesPerSec int64) string {
	return fmt.Sprintf("%s/s", FormatBytes(bytesPerSec))
}

func FormatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60
	secs := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, mins, secs)
	}
	return fmt.Sprintf("%dm %ds", mins, secs)
}

func CalculateETA(total, completed, speed int64) string {
	if speed <= 0 || completed >= total {
		return "0s"
	}
	remainingBytes := total - completed
	seconds := remainingBytes / speed
	return FormatDuration(time.Duration(seconds) * time.Second)
}

func StringToInt64(s string) int64 {
	val, _ := strconv.ParseInt(s, 10, 64)
	return val
}

type MirrorArgs struct {
	Link         string
	CustomName   string
	CustomRemote string
	IsMagnet     bool
	IsZip        bool
	IsExtract    bool
}

// ArgParser mem-parsing command options persis wzv3
func ArgParser(rawText string) MirrorArgs {
	res := MirrorArgs{}
	words := strings.Fields(rawText)
	if len(words) > 0 && strings.HasPrefix(words[0], "/") {
		words = words[1:]
	}

	joined := strings.Join(words, " ")

	if strings.Contains(joined, "|") {
		parts := strings.SplitN(joined, "|", 2)
		res.Link = strings.TrimSpace(parts[0])
		res.CustomName = strings.TrimSpace(parts[1])
		if strings.HasPrefix(res.Link, "magnet:?") {
			res.IsMagnet = true
		}
		return res
	}

	var linkParts []string
	for i := 0; i < len(words); i++ {
		w := words[i]
		if w == "-n" && i+1 < len(words) {
			res.CustomName = words[i+1]
			i++
		} else if w == "-rc" && i+1 < len(words) {
			res.CustomRemote = words[i+1]
			i++
		} else if w == "-z" {
			res.IsZip = true
		} else if w == "-e" {
			res.IsExtract = true
		} else {
			linkParts = append(linkParts, w)
		}
	}

	res.Link = strings.TrimSpace(strings.Join(linkParts, " "))
	if strings.HasPrefix(res.Link, "magnet:?") {
		res.IsMagnet = true
	}

	return res
}
