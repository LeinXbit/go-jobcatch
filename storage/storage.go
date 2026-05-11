package storage

import (
	"go-catch/model"
	"sync"
)

type MemoryStorage struct {
	seen    map[string]bool
	newJobs []model.Job
	mutex   sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		seen:    make(map[string]bool),
	}
}

func (s *MemoryStorage) AddIfNew(job model.Job) bool {
	s.mutex.RLock()
	exists := s.seen[job.ID]
	s.mutex.RUnlock()

	if exists {
		return false
	}

	s.mutex.Lock()
	s.seen[job.ID] = true
	s.newJobs = append(s.newJobs, job)
	s.mutex.Unlock()
	return true
}

func (s *MemoryStorage) GetNewJobs() []model.Job {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.newJobs
}

func (s *MemoryStorage) ClearNewJobs() {
	s.mutex.Lock()
	s.newJobs = []model.Job{}
	s.mutex.Unlock()
}