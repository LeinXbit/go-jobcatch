package parser

import (
	"bytes"
	"go-catch/model"
	"strings"
	"strconv"
	"github.com/PuerkitoBio/goquery"
)

type Job51Parser struct{
	Keywords []string
}

func NewJob51Parser(keywords []string) *Job51Parser {
	return &Job51Parser{Keywords: keywords}
}

func (p *Job51Parser) Parse(data []byte, city string) ([]model.Job, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	var jobs []model.Job

	doc.Find(".job").Each(func(i int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Find(".jname at").Text())
		company := strings.TrimSpace(s.Find(".cname at").Text())
		salary := strings.TrimSpace(s.Find(".sal").Text())
		jobID, exists := s.Attr("id")

		if title !="" && p.matchKeywords(title) {
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

func (p *Job51Parser) matchKeywords(title string) bool {
	lowerTitle := strings.ToLower(title)
	for _, keyword := range p.Keywords {
		if strings.Contains(lowerTitle, strings.ToLower(keyword)) {
			return true
		}
	}
	return false	
}