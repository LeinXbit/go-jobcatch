package notifier

import (
	"go-catch/model"
	"go-catch/logger"
	"strings"
)

type ConsoleNotifier struct{}

func NewConsoleNotifier() *ConsoleNotifier {
	return &ConsoleNotifier{}
}

func (n *ConsoleNotifier) NotifyNewJob(job model.Job) {
	separator := strings.Repeat("=", 50)
	logger.Info(separator)

	if job.Salary != "" {
		logger.Infof("salary range: %s\n", job.Salary)
	}else {
		logger.Infof("negotiable salary\n")
	}

	logger.Infof("company: %s\n", job.Company)

	if job.URL != "" {
		logger.Infof("job URL: %s\n", job.URL)
	}

	logger.Info(separator)
}

func (n *ConsoleNotifier) NotifyBatch(jobs []model.Job) {
	if len(jobs) == 0 {
		return
	}
	logger.Infof("Total new jobs found: %d\n", len(jobs))
	for _,job := range jobs {
		n.NotifyNewJob(job)
	}
}