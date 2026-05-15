package storage

import (
    "sync"

    "go-catch/model"
)

// MemoryStorage 内存存储
type MemoryStorage struct {
    seen    map[string]bool
    newJobs []model.Job
    mu      sync.RWMutex
}

// NewMemoryStorage 创建内存存储
func NewMemoryStorage() *MemoryStorage {
    return &MemoryStorage{
        seen:    make(map[string]bool),
        newJobs: make([]model.Job, 0),
    }
}

// AddIfNew 添加新岗位
func (s *MemoryStorage) AddIfNew(job model.Job) bool {
    s.mu.Lock()
    defer s.mu.Unlock()

    if s.seen[job.JobID] {
        return false
    }

    s.seen[job.JobID] = true
    s.newJobs = append(s.newJobs, job)
    return true
}

// GetNewJobs 获取新增岗位
func (s *MemoryStorage) GetNewJobs() []model.Job {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.newJobs
}

// ClearNewJobs 清空新增列表
func (s *MemoryStorage) ClearNewJobs() {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.newJobs = make([]model.Job, 0)
}