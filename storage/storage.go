package storage

import (
	"go-catch/model"
	"sync"
)

var (
	seen    = make(map[string]bool)
	mutex   sync.RWMutex
	newJobs []model.Job
)

func AddIfNew(job model.Job) bool {
	mutex.RLock()
	exist := seen[job.ID]
	mutex.RUnlock()

	if exist {
		return false
	}

	mutex.Lock()
	seen[job.ID] = true
	newJobs = append(newJobs, job)
	mutex.Unlock()
	return true
}

func GetNewJobs() []model.Job {
	mutex.RLock()
	defer mutex.RUnlock()
	return newJobs
}

func ClearNewJobs() {
	mutex.Lock()
	newJobs = []model.Job{}
	mutex.Unlock()
}