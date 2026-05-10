package parser

import (
	"bytes"
	"go-catch/model"
	"strings"
	"strconv"
	"github.com/PuerkitoBio/goquery"
)

type JobResult struct {
	City string
	Jobs []model.Job
	Err error
}

func ParseJobs51(htmlData []byte, city string) ([]model.Job, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(htmlData))
	if err != nil {
		return nil, err
	}

	var jobs []model.Job

	doc.Find(".job").Each(func(i int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Find(".jname at").Text())
		company := strings.TrimSpace(s.Find(".cname at").Text())
		salary := strings.TrimSpace(s.Find(".sal").Text())
		jobID, exists := s.Attr("id")

		if title !="" && strings.Contains(strings.ToLower(title), "go") {
			id := jobID
			if !exists {
				id = strconv.Itoa(i) // Fallback to index if no ID
			}

			job := model.Job{
				ID:        id,
				Title:     title + " " + salary,
				Company:   company,
				City:      city,
			}
			jobs = append(jobs, job)
		}
	})

	return jobs, nil
}

