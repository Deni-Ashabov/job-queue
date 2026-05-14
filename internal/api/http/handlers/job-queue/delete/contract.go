package delete

import "context"

type JobDelete interface {
	DeleteJob(ctx context.Context, jobID int) error
}
