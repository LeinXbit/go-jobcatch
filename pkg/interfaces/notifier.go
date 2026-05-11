package interfaces

import (
	"go-catch/model"
)

type Notifier interface {
	NotifyNewJob(job model.Job)
	NotifierBatch(jobs []model.Job)
}

type ConfigurableNotifier interface {
	Notifier
	SetWebhook(url string)
	SetRecipient(email string)
}