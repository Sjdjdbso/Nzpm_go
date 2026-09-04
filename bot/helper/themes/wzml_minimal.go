package themes

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"go-mirror-bot/bot/helper/ext_utils"
)

const (
	ProgFull  = "■"
	ProgEmpty = "□"
)

type SystemStats struct {
	Uptime      string
	RAMPercent  float64
	RAMUsed     string
	RAMFree     string
	RAMTotal    string
	DiskPercent float64
	DiskUsed    string
	DiskFree    string
	DiskTotal   string
	CPUPercent  float64
	OSArch      string
	GoVersion   string
}

func GenerateBar(percent float64) string {
	const total = 10
	filled := int((percent / 100.0) * total)
	if filled > total {
		filled = total
	}
	if filled < 0 {
		filled = 0
	}
	return strings.Repeat(ProgFull, filled) + strings.Repeat(ProgEmpty, total-filled)
}

func CollectStats(startTime time.Time) SystemStats {
	st := SystemStats{
		OSArch:    fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		GoVersion: runtime.Version(),
	}

	if data, err := os.ReadFile("/proc/uptime"); err == nil {
		fields := strings.Fields(string(data))
		if len(fields) > 0 {
			if sec, err := strconv.ParseFloat(fields[0], 64); err == nil {
				st.Uptime = ext_utils.FormatDuration(time.Duration(sec) * time.Second)
			}
		}
	}
	if st.Uptime == "" {
		st.Uptime = ext_utils.FormatDuration(time.Since(startTime))
	}

	if file, err := os.Open("/proc/meminfo"); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		var totalKb, availKb float64
		for scanner.Scan() {
			l := scanner.Text()
			if strings.HasPrefix(l, "MemTotal:") {
				f := strings.Fields(l)
				if len(f) >= 2 {
					totalKb, _ = strconv.ParseFloat(f[1], 64)
				}
			} else if strings.HasPrefix(l, "MemAvailable:") {
				f := strings.Fields(l)
				if len(f) >= 2 {
					availKb, _ = strconv.ParseFloat(f[1], 64)
				}
			}
		}
		if totalKb > 0 {
			usedKb := totalKb - availKb
			st.RAMPercent = (usedKb / totalKb) * 100
			st.RAMTotal = ext_utils.FormatBytes(int64(totalKb * 1024))
			st.RAMFree = ext_utils.FormatBytes(int64(availKb * 1024))
			st.RAMUsed = ext_utils.FormatBytes(int64(usedKb * 1024))
		}
	}

	var statfs syscall.Statfs_t
	if err := syscall.Statfs("/", &statfs); err == nil {
		total := statfs.Blocks * uint64(statfs.Bsize)
		free := statfs.Bavail * uint64(statfs.Bsize)
		used := total - free
		if total > 0 {
			st.DiskPercent = (float64(used) / float64(total)) * 100
			st.DiskTotal = ext_utils.FormatBytes(int64(total))
			st.DiskFree = ext_utils.FormatBytes(int64(free))
			st.DiskUsed = ext_utils.FormatBytes(int64(used))
		}
	}

	st.CPUPercent = float64(runtime.NumGoroutine()) * 0.4
	if st.CPUPercent > 99 {
		st.CPUPercent = 99
	}
	return st
}

// FormatStatusMsg menyusun tampilan live status persis WZML-X (wzml_minimal.py)
func FormatStatusMsg(t *ext_utils.Task) string {
	bar := GenerateBar(t.Progress)
	elapsed := ext_utils.FormatDuration(time.Since(t.StartTime))
	eta := t.ETA
	if eta == "" {
		eta = "-"
	}
	engine := t.Engine
	if engine == "" {
		engine = "Aria2c"
	}

	return fmt.Sprintf(
		"<b><i>%s</i></b>\n"+
			"┃ %s\n"+
			"┠ <b>Processed:</b> %s of %s (%.2f%%)\n"+
			"┠ <b>Status:</b> %s | <b>ETA:</b> %s\n"+
			"┠ <b>Speed:</b> %s | <b>Elapsed:</b> %s\n"+
			"┠ <b>Engine:</b> %s\n"+
			"┠ <b>Mode:</b> %s\n"+
			"┠ <b>User:</b> %s | <b>ID:</b> <code>%d</code>\n"+
			"┖ <code>/cancel %s</code>",
		t.Name,
		bar,
		ext_utils.FormatBytes(t.CompletedSize), ext_utils.FormatBytes(t.TotalSize), t.Progress,
		t.Status, eta,
		ext_utils.FormatSpeed(t.Speed), elapsed,
		engine,
		t.Mode,
		t.User, t.UserID,
		t.GID,
	)
}

// FormatFooter menyusun footer status WZML-X
func FormatFooter(taskCount int, stats SystemStats, totalDlSpeed int64) string {
	return fmt.Sprintf(
		"⌬ <b><i>Bot Stats</i></b>\n"+
			"┠ <b>Tasks:</b> %d\n"+
			"┠ <b>CPU:</b> %.1f%% | <b>F:</b> %s [%.1f%%]\n"+
			"┠ <b>RAM:</b> %.1f%% | <b>UPTIME:</b> %s\n"+
			"┖ <b>DL:</b> %s | <b>UL:</b> 0 B/s",
		taskCount,
		stats.CPUPercent, stats.DiskFree, 100-stats.DiskPercent,
		stats.RAMPercent, stats.Uptime,
		ext_utils.FormatSpeed(totalDlSpeed),
	)
}

// FormatCompleteMsg menyusun format selesai (onUploadComplete)
func FormatCompleteMsg(name string, size int64, mode string, destPath string, userTag string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b><i>%s</i></b>\n┃\n", name))
	sb.WriteString(fmt.Sprintf("┠ <b>Size:</b> %s\n", ext_utils.FormatBytes(size)))
	sb.WriteString(fmt.Sprintf("┠ <b>Mode:</b> %s\n", mode))
	if destPath != "" {
		sb.WriteString(fmt.Sprintf("┠ <b>Path:</b> <code>%s</code>\n", destPath))
	}
	sb.WriteString(fmt.Sprintf("┖ <b>By:</b> %s", userTag))
	return sb.String()
}
