package delete

import (
	"errors"
	"job-queue/internal/domain/job"
	"job-queue/internal/logger/sl"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const op = "handlers.job-queue.delete.New"

func New(log *slog.Logger, deleteJob JobDelete) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		if err := deleteJob.DeleteJob(r.Context(), jobID); err != nil {
			if errors.Is(err, job.ErrJobNotFound) {
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
