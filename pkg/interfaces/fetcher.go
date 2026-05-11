package interfaces

import (
	"go-catch/model"
)

type Fetcher interface {
	Fetch(cityName, cityCode string) ([]byte, error)
}

type JobFetcher interface {
	Fetcher
	FetchJobs(cityCode string, page int) ([]model.Job, error)
}