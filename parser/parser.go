package parser

import (
	"encoding/json"
	"go-catch/model"
)

func ParseJobs(data []byte) ([]model.Job, error) {
	var jobs []model.Job
	err := json.Unmarshal(data, &jobs)
	return jobs, err
}

type JobResult struct {
	City string
	Jobs []model.Job
	Err error
}