package get_test

import (
	"context"
	"encoding/json"
	"errors"
	resp "job-queue/internal/api/http/dto"
	"job-queue/internal/api/http/handlers/job-queue/get"
	"job-queue/internal/api/http/handlers/job-queue/get/mocks"
	"job-queue/internal/domain/job"
	"job-queue/internal/logger/handlers/slogdiscard"
	"job-queue/internal/models"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetHandler(t *testing.T) {
	cases := []struct {
		name      string
		jobID     int
		status    int
		respError string
		setup     func(m *mocks.JobGet)
	}{
		{
			name:   "Success",
			jobID:  1,
			status: http.StatusOK,
			setup: func(m *mocks.JobGet) {
				m.On("GetJob",
					mock.MatchedBy(func(ctx context.Context) bool {
						return ctx != nil
					}),
					1).
					Return(job.Job{
						ID:        1,
						Queue:     "emails",
						JobStatus: "processing",
					}, nil).
					Once()
			},
		},
		{
			name:   "Not found",
			jobID:  0,
			status: http.StatusNotFound,
			setup: func(m *mocks.JobGet) {
				m.On("GetJob",
					mock.MatchedBy(func(ctx context.Context) bool {
						return ctx != nil
					}),
					0).
					Return(job.Job{}, job.ErrJobNotFound).
					Once()
			},
		},
		{
			name:   "Internal error",
			jobID:  1,
			status: http.StatusInternalServerError,
			setup: func(m *mocks.JobGet) {
				m.On("GetJob",
					mock.MatchedBy(func(ctx context.Context) bool {
						return ctx != nil
					}),
					1).
					Return(job.Job{}, errors.New("db error")).
					Once()
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			jobGetMock := mocks.NewJobGet(t)

			if tc.setup != nil {
				tc.setup(jobGetMock)
			}

			handler := get.New(slogdiscard.NoopLogger(), jobGetMock)

			req := newGetReq(tc.jobID)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.status, rr.Code)

			var resp resp.GetResponse
			require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

			if tc.status == http.StatusOK {
				require.Equal(t, int64(tc.jobID), resp.JobID)
				require.Equal(t, models.QueueType("emails"), resp.Queue)
				require.Equal(t, "processing", resp.JobStatus)
			}

			if tc.status != http.StatusOK {
				require.NotEmpty(t, resp.Error)
			}
		})
	}
}

func newGetReq(id int) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/job/"+strconv.Itoa(id), nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("jobID", strconv.Itoa(id))

	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}
