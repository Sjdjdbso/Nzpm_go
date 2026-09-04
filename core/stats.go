package core

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type SysStats struct {
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

func GetSystemStats(botStartTime time.Time) SysStats {
	stats := SysStats{
		OSArch:    fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		GoVersion: runtime.Version(),
	}

	// 1. Uptime
	uptimeBytes, err := os.ReadFile("/proc/uptime")
	if err == nil {
		parts := strings.Fields(string(uptimeBytes))
		if len(parts) > 0 {
			if sec, err := strconv.ParseFloat(parts[0], 64); err == nil {
				stats.Uptime = formatDuration(time.Duration(sec) * time.Second)
			}
		}
	}
	if stats.Uptime == "" {
		stats.Uptime = formatDuration(time.Since(botStartTime))
	}

	// 2. RAM dari /proc/meminfo
	file, err := os.Open("/proc/meminfo")
	if err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		var totalKb, availKb float64
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "MemTotal:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					totalKb, _ = strconv.ParseFloat(fields[1], 64)
				}
			} else if strings.HasPrefix(line, "MemAvailable:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					availKb, _ = strconv.ParseFloat(fields[1], 64)
				}
			}
		}
		if totalKb > 0 {
			usedKb := totalKb - availKb
			stats.RAMPercent = (usedKb / totalKb) * 100
			stats.RAMTotal = FormatBytes(int64(totalKb * 1024))
			stats.RAMFree = FormatBytes(int64(availKb * 1024))
			stats.RAMUsed = FormatBytes(int64(usedKb * 1024))
		}
	}

	// 3. Disk Stats
	var stat syscall.Statfs_t
	wd, err := os.Getwd()
	if err != nil {
		wd = "/"
	}
	if err := syscall.Statfs(wd, &stat); err == nil {
		total := stat.Blocks * uint64(stat.Bsize)
		free := stat.Bavail * uint64(stat.Bsize)
		used := total - free
		if total > 0 {
			stats.DiskPercent = (float64(used) / float64(total)) * 100
			stats.DiskTotal = FormatBytes(int64(total))
			stats.DiskFree = FormatBytes(int64(free))
			stats.DiskUsed = FormatBytes(int64(used))
		}
	}

	// 4. CPU usage (perkiraan beban sederhana)
	stats.CPUPercent = float64(runtime.NumGoroutine()) * 0.5
	if stats.CPUPercent > 99 {
		stats.CPUPercent = 99
	}

	return stats
}

func formatDuration(d time.Duration) string {
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
