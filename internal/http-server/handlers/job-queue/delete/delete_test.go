package delete_test

import (
	"context"
	"errors"
	del "job-queue/internal/http-server/handlers/job-queue/delete"
	"job-queue/internal/http-server/handlers/job-queue/delete/mocks"
	"job-queue/internal/lib/logger/handlers/slogdiscard"
	storage "job-queue/internal/repository"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func TestDeleteHandler(t *testing.T) {
	cases := []struct {
		name      string
		jobID     int
		status    int
		respError string
		setup     func(m *mocks.JobDelete)
	}{
		{
			name:   "Success",
			jobID:  1,
			status: http.StatusNoContent,
			setup: func(m *mocks.JobDelete) {
				m.On("DeleteJob", 1).
					Return(nil).
					Once()
			},
		},
		{
			name:   "Not found",
			jobID:  0,
			status: http.StatusNotFound,
			setup: func(m *mocks.JobDelete) {
				m.On("DeleteJob", 0).
					Return(storage.ErrJobNotFound).
					Once()
			},
		},
		{
			name:   "Internal error",
			jobID:  1,
			status: http.StatusInternalServerError,
			setup: func(m *mocks.JobDelete) {
				m.On("DeleteJob", 1).
					Return(errors.New("db error")).
					Once()
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			jobDeleteMock := mocks.NewJobDelete(t)

			if tc.setup != nil {
				tc.setup(jobDeleteMock)
			}

			handler := del.New(slogdiscard.NewDiscardLogger(), jobDeleteMock)

			req := newDeleteReq(tc.jobID)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.status, rr.Code)

			if tc.status == http.StatusNoContent {
				require.Empty(t, rr.Body.String())
			}
		})
	}
}

func newDeleteReq(id int) *http.Request {
	req := httptest.NewRequest(http.MethodDelete, "/job/"+strconv.Itoa(id), nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("jobID", strconv.Itoa(id))

	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}
