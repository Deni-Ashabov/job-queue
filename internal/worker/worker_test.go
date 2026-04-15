package worker_test

import (
	"encoding/json"
	"job-queue/internal/lib/logger/handlers/slogdiscard"
	"job-queue/internal/repository/postgres"
	"job-queue/internal/worker"
	"job-queue/internal/worker/mocks"
	"testing"
)

func TestProcessJob(t *testing.T) {
	cases := []struct {
		name   string
		job    postgres.Job
		status postgres.JobStatus
	}{
		{
			name: "email success",
			job: postgres.Job{
				ID:    1,
				Queue: "emails",
				Payload: json.RawMessage(`{
					"to":"a",
					"subject":"b",
					"body":"c"
				}`),
			},
			status: postgres.StatusDone,
		},
		{
			name: "email invalid json",
			job: postgres.Job{
				ID:      2,
				Queue:   "emails",
				Payload: json.RawMessage(`invalid json`),
			},
			status: postgres.StatusFailed,
		},
		{
			name: "payment success",
			job: postgres.Job{
				ID:    3,
				Queue: "payment",
				Payload: json.RawMessage(`{
					"user_id":"42",
					"amount":"100"
				}`),
			},
			status: postgres.StatusDone,
		},
		{
			name: "payment invalid json",
			job: postgres.Job{
				ID:      4,
				Queue:   "payment",
				Payload: json.RawMessage(`invalid json`),
			},
			status: postgres.StatusFailed,
		},
		{
			name: "notification success",
			job: postgres.Job{
				ID:    5,
				Queue: "notification",
				Payload: json.RawMessage(`{
					"user_id":"42",
					"text":"hello"
				}`),
			},
			status: postgres.StatusDone,
		},
		{
			name: "notification invalid json",
			job: postgres.Job{
				ID:      6,
				Queue:   "notification",
				Payload: json.RawMessage(`invalid json`),
			},
			status: postgres.StatusFailed,
		},
		{
			name: "unknown queue",
			job: postgres.Job{
				ID:      7,
				Queue:   "unknown",
				Payload: json.RawMessage(`{}`),
			},
			status: postgres.StatusFailed,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			storage := mocks.NewStorage(t)
			w := worker.New(storage, slogdiscard.NewDiscardLogger())

			storage.On("ChangeStatus",
				tc.job.ID,
				tc.status,
			).Return(nil).Once()

			w.ProcessJob(tc.job)

			storage.AssertExpectations(t)
		})
	}
}
