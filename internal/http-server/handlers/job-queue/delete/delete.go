package delete

import (
	"errors"
	"job-queue/internal/lib/logger/sl"
	storage "job-queue/internal/repository"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type JobDelete interface {
	DeleteJob(jobID int) error
}

func New(log *slog.Logger, deleteJob JobDelete) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.job-queue.delete.New"

		logger := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		jobIDParam := chi.URLParam(r, "jobID")
		if jobIDParam == "" {
			http.Error(w, "job id is required", http.StatusBadRequest)
			return
		}

		jobID, err := strconv.Atoi(jobIDParam)
		if err != nil {
			http.Error(w, "job id must be a number", http.StatusBadRequest)
			return
		}

		if err := deleteJob.DeleteJob(jobID); err != nil {
			if errors.Is(err, storage.ErrJobNotFound) {
				logger.Info("job not found", slog.String("jobID", jobIDParam))
				http.Error(w, "not found", http.StatusNotFound)
				return
			}

			logger.Error("failed to delete job", sl.Err(err))
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
