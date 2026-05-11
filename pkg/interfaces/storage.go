package interfaces

import (
	"go-catch/model"
)

type Storage interface {
	AddIfNew(job model.Job) bool
	GetNewJobs() []model.Job
	ClearNewJobs()
}

type PersistentStorage interface {
	Storage
	Save(job model.Job) error
	Load() (map[string]bool, error)
}