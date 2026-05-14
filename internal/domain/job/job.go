package job

import (
	"encoding/json"
	"job-queue/internal/models"
	"time"
)

type Job struct {
	ID          int64
	Queue       models.QueueType
	Payload     json.RawMessage
	JobStatus   string
	AvailableAt time.Time
	CreatedAt   time.Time
}
