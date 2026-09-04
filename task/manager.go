package task

import (
	"strings"
	"sync"
	"time"

	"go-mirror-bot/core"
)

type Task struct {
	GID           string
	Name          string
	Status        string // "Downloading", "Uploading", "Completed", "Error"
	Mode          string // "Mirror", "Leech"
	TotalSize     int64
	CompletedSize int64
	Speed         int64
	ETA           string
	Progress      float64
	FilePath      string
	User          string
	UserID        int64
	StartTime     time.Time
	ErrorMessage  string
}

type Manager struct {
	mu           sync.RWMutex
	tasks        map[string]*Task
	BotStartTime time.Time
}

var TaskMgr = &Manager{
	tasks:        make(map[string]*Task),
	BotStartTime: time.Now(),
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

func (m *Manager) CancelAll() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for gid := range m.tasks {
		core.Aria.Remove(gid)
		delete(m.tasks, gid)
		count++
	}
	return count
}

// FormatStatusView menyusun tampilan /status persis tema WZML-X
func (m *Manager) FormatStatusView() string {
	tasks := m.All()
	stats := core.GetSystemStats(m.BotStartTime)

	if len(tasks) == 0 {
		return "<i>No Active Downloads!</i>\n\n" + core.FormatWZMLFooter(0, stats, 0)
	}

	var sb strings.Builder
	var totalSpeed int64

	for _, t := range tasks {
		totalSpeed += t.Speed
		sb.WriteString(core.FormatWZMLTaskStatus(
			t.Name,
			t.Progress,
			t.CompletedSize,
			t.TotalSize,
			t.Speed,
			t.ETA,
			t.Status,
			t.Mode,
			t.User,
			t.UserID,
			t.GID,
			t.StartTime,
		))
		sb.WriteString("\n\n")
	}

	sb.WriteString(core.FormatWZMLFooter(len(tasks), stats, totalSpeed))
	return sb.String()
}
