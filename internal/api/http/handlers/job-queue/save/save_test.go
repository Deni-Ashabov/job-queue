package save_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"job-queue/internal/api/http/dto"
	"job-queue/internal/api/http/handlers/job-queue/save"
	"job-queue/internal/api/http/handlers/job-queue/save/mocks"
	"job-queue/internal/domain/job"
	"job-queue/internal/logger/handlers/slogdiscard"
	"job-queue/internal/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSaveHandler(t *testing.T) {
	cases := []struct {
		name      string
		queue     string
		payload   json.RawMessage
		respError string
		status    int
		setup     func(m *mocks.JobSaver)
	}{
		{
			name:  "Success",
			queue: "emails",
			payload: json.RawMessage(`{
				"to": "user1",
				"subject": "Hello",
				"body": "Some content"
			}`),
			status: http.StatusOK,
			setup: func(m *mocks.JobSaver) {
				m.On("SaveJob",
					mock.MatchedBy(func(ctx context.Context) bool {
						return ctx != nil
					}),
					models.QueueType("emails"),
					mock.Anything).
					Return(job.Job{
						ID:        1,
						Queue:     "emails",
						JobStatus: "processing",
					}, nil).
					Once()
			},
		},
		{
			name:  "Wrong queue",
			queue: "emals",
			payload: json.RawMessage(`{
				"to": "user1",
				"subject": "Hello",
				"body": "Some content"
			}`),
			respError: "field Queue must be one of allowed values (emails, payment, notification)",
			status:    http.StatusBadRequest,
		},
		{
			name:      "Invalid JSON",
			queue:     "emails",
			payload:   json.RawMessage(`{"invalid":`),
			respError: "failed to decode request",
			status:    http.StatusBadRequest,
		},
		{
			name:  "SaveJob Error",
			queue: "emails",
			payload: json.RawMessage(`{
				"to": "user1",
				"subject": "Hello",
				"body": "Some content"
			}`),
			respError: "failed to add job",
			status:    http.StatusInternalServerError,
			setup: func(m *mocks.JobSaver) {
				m.
					On("SaveJob", mock.MatchedBy(func(ctx context.Context) bool {
						return ctx != nil
					}),
						models.QueueType("emails"),
						mock.Anything).
					Return(job.Job{}, errors.New("unexpected error")).
					Once()
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			jobServerMock := mocks.NewJobSaver(t)

			if tc.setup != nil {
				tc.setup(jobServerMock)
			}

			handler := save.New(slogdiscard.NoopLogger(), jobServerMock)

			inputBytes, _ := json.Marshal(map[string]any{
				"queue":   tc.queue,
				"payload": tc.payload,
			})

			req, err := http.NewRequest(http.MethodPost, "/job", bytes.NewReader([]byte(inputBytes)))

			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.status, rr.Code)

			body := rr.Body.String()

			var resp dto.SaveResponse

			require.NoError(t, json.Unmarshal([]byte(body), &resp))

			require.Equal(t, tc.respError, resp.Error)
		})
	}
}
