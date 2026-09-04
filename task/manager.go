package task

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type Task struct {
	GID           string
	Name          string
	Status        string // "Downloading", "Uploading", "Completed", "Error"
	TotalSize     int64
	CompletedSize int64
	Speed         int64
	ETA           string
	Progress      float64
	FilePath      string
	User          string
	StartTime     time.Time
	ErrorMessage  string
}

type Manager struct {
	mu    sync.RWMutex
	tasks map[string]*Task
}

var TaskMgr = &Manager{
	tasks: make(map[string]*Task),
}

func (m *Manager) Add(t *Task) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasks[t.GID] = t
}

func (m *Manager) Get(gid string) *Task {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tasks[gid]
}

func (m *Manager) Remove(gid string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.tasks, gid)
}

func (m *Manager) All() []*Task {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]*Task, 0, len(m.tasks))
	for _, t := range m.tasks {
		list = append(list, t)
	}
	return list
}

// GenerateProgressBar membuat visual bar: [██████░░░░]
func GenerateProgressBar(percent float64) string {
	const totalBlocks = 10
	filled := int((percent / 100.0) * totalBlocks)
	if filled > totalBlocks {
		filled = totalBlocks
	}
	if filled < 0 {
		filled = 0
	}
	empty := totalBlocks - filled
	return fmt.Sprintf("[%s%s]", strings.Repeat("█", filled), strings.Repeat("░", empty))
}

// FormatStatusView menyusun ringkasan status aktif untuk Telegram
func (m *Manager) FormatStatusView() string {
	tasks := m.All()
	if len(tasks) == 0 {
		return "<b>Tidak ada proses aktif saat ini.</b>"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>⚡ Tugas Berjalan (%d):</b>\n\n", len(tasks)))

	for _, t := range tasks {
		bar := GenerateProgressBar(t.Progress)
		icon := "📥"
		if t.Status == "Uploading" {
			icon = "📤"
		} else if t.Status == "Error" {
			icon = "❌"
		}

		sb.WriteString(fmt.Sprintf("%s <b>%s</b>\n", icon, t.Name))
		sb.WriteString(fmt.Sprintf("├ <b>Status:</b> %s\n", t.Status))
		sb.WriteString(fmt.Sprintf("├ <b>Progres:</b> %s %.1f%%\n", bar, t.Progress))
		sb.WriteString(fmt.Sprintf("├ <b>Ukuran:</b> %s / %s\n", formatBytes(t.CompletedSize), formatBytes(t.TotalSize)))
		if t.Status == "Downloading" {
			sb.WriteString(fmt.Sprintf("├ <b>Speed:</b> %s/s | <b>ETA:</b> %s\n", formatBytes(t.Speed), t.ETA))
		}
		sb.WriteString(fmt.Sprintf("└ <b>Diminta oleh:</b> %s\n\n", t.User))
	}

	return sb.String()
}

func formatBytes(b int64) string {
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
