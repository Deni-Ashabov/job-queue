package get

import (
	"context"
	"job-queue/internal/domain/job"
)

type JobGet interface {
	GetJob(ctx context.Context, jobID int) (job.Job, error)
}
