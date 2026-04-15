package get_test

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"job-queue/internal/http-server/handlers/job-queue/get"
	"job-queue/internal/http-server/handlers/job-queue/get/mocks"
	"job-queue/internal/lib/logger/handlers/slogdiscard"
	storage "job-queue/internal/repository"
	"job-queue/internal/repository/postgres"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
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
				m.On("GetJob", 1).
					Return(postgres.Job{
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
				m.On("GetJob", 0).
					Return(postgres.Job{}, storage.ErrJobNotFound).
					Once()
			},
		},
		{
			name:   "Internal error",
			jobID:  1,
			status: http.StatusInternalServerError,
			setup: func(m *mocks.JobGet) {
				m.On("GetJob", 1).
					Return(postgres.Job{}, errors.New("db error")).
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

			handler := get.New(slogdiscard.NewDiscardLogger(), jobGetMock)

			req := newGetReq(tc.jobID)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.status, rr.Code)

			var resp get.Response
			require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

			if tc.status == http.StatusOK {
				require.Equal(t, int64(tc.jobID), resp.JobID)
				require.Equal(t, postgres.QueueType("emails"), resp.Queue)
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
