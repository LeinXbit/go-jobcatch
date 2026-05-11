package interfaces

import (
	"go-catch/model"
)

type Parser interface {
	Parse(data []byte, cityName string) ([]model.Job, error)
}

type FilteredParser interface {
	Parser
	SetKeywords(keyword []string)
	GetKeywords() []string
}