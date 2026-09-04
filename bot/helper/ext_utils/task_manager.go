package ext_utils

import (
	"sync"
	"time"
)

type Task struct {
	UID           string
	GID           string
	Name          string
	Status        string // "Downloading", "Uploading", "Completed", "Error"
	Mode          string // "Mirror", "Leech"
	Engine        string // "Aria2c", "yt-dlp"
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

type TaskManager struct {
	mu           sync.RWMutex
	Tasks        map[string]*Task
	BotStartTime time.Time
}

var TaskMgr = &TaskManager{
	Tasks:        make(map[string]*Task),
	BotStartTime: time.Now(),
}

func (m *TaskManager) Add(t *Task) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Tasks[t.GID] = t
}

func (m *TaskManager) Get(gid string) *Task {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Tasks[gid]
}

func (m *TaskManager) Remove(gid string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.Tasks, gid)
}

func (m *TaskManager) All() []*Task {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]*Task, 0, len(m.Tasks))
	for _, t := range m.Tasks {
		list = append(list, t)
	}
	return list
}

func (m *TaskManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.Tasks)
}
