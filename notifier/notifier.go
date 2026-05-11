package notifier

import (
	"go-catch/model"
	"go-catch/logger"
)

type ConsoleNotifier struct{}

func NewConsoleNotifier() *ConsoleNotifier {
	return &ConsoleNotifier{}
}

func (n *ConsoleNotifier) NotifyNewJob(job model.Job) {
	logger.Infof("Find new job: %s at %s in %s",job.Title, job.Company, job.City)
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