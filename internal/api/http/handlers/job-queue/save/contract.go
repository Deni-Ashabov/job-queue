package save

import (
	"context"
	"encoding/json"
	"job-queue/internal/domain/job"
	"job-queue/internal/models"
)

type JobSaver interface {
	SaveJob(ctx context.Context, queue models.QueueType, payload json.RawMessage) (job.Job, error)
}
