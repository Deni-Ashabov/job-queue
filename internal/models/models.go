package models

type JobStatus string

const (
	StatusPending    JobStatus = "pending"
	StatusProcessing JobStatus = "processing"
	StatusDone       JobStatus = "done"
	StatusFailed     JobStatus = "failed"
)

type QueueType string

const (
	QueueEmail        QueueType = "emails"
	QueuePayment      QueueType = "payment"
	QueueNotification QueueType = "notification"
)
