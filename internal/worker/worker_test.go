package worker_test

import (
	"context"
	"encoding/json"
	"job-queue/internal/domain/job"
	"job-queue/internal/logger/handlers/slogdiscard"
	"job-queue/internal/models"
	"job-queue/internal/worker"
	"job-queue/internal/worker/mocks"
	"testing"

	"github.com/stretchr/testify/mock"
)

func TestProcessJob(t *testing.T) {
	cases := []struct {
		name   string
		job    job.Job
		status models.JobStatus
	}{
		{
			name: "email success",
			job: job.Job{
				ID:    1,
				Queue: "emails",
				Payload: json.RawMessage(`{
					"to":"a",
					"subject":"b",
					"body":"c"
				}`),
			},
			status: models.StatusDone,
		},
		{
			name: "email invalid json",
			job: job.Job{
				ID:      2,
				Queue:   "emails",
				Payload: json.RawMessage(`invalid json`),
			},
			status: models.StatusFailed,
		},
		{
			name: "payment success",
			job: job.Job{
				ID:    3,
				Queue: "payment",
				Payload: json.RawMessage(`{
					"user_id":"42",
					"amount":"100"
				}`),
			},
			status: models.StatusDone,
		},
		{
			name: "payment invalid json",
			job: job.Job{
				ID:      4,
				Queue:   "payment",
				Payload: json.RawMessage(`invalid json`),
			},
			status: models.StatusFailed,
		},
		{
			name: "notification success",
			job: job.Job{
				ID:    5,
				Queue: "notification",
				Payload: json.RawMessage(`{
					"user_id":"42",
					"text":"hello"
				}`),
			},
			status: models.StatusDone,
		},
		{
			name: "notification invalid json",
			job: job.Job{
				ID:      6,
				Queue:   "notification",
				Payload: json.RawMessage(`invalid json`),
			},
			status: models.StatusFailed,
		},
		{
			name: "unknown queue",
			job: job.Job{
				ID:      7,
				Queue:   "unknown",
				Payload: json.RawMessage(`{}`),
			},
			status: models.StatusFailed,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			storage := mocks.NewStorage(t)
			w := worker.New(storage, slogdiscard.NoopLogger())

			storage.On("ChangeStatus",
				mock.MatchedBy(func(ctx context.Context) bool {
					return ctx != nil
				}),
				tc.job.ID,
				tc.status,
			).Return(nil).Once()

			w.ProcessJob(context.Background(), tc.job)

			storage.AssertExpectations(t)
		})
	}
}
