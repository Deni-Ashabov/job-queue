package get

import (
	"errors"
	resp "job-queue/internal/api/http/dto"
	jobDomain "job-queue/internal/domain/job"
	"job-queue/internal/logger/sl"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

const op = "handlers.job-queue.get.New"

func New(log *slog.Logger, jobGet JobGet) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		jobIDParam := chi.URLParam(r, "jobID")
		if jobIDParam == "" {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("job id is required"))
			return
		}

		jobID, err := strconv.Atoi(jobIDParam)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("job id must be a number"))
			return
		}

		job, err := jobGet.GetJob(r.Context(), jobID)
		if errors.Is(err, jobDomain.ErrJobNotFound) {
			logger.Info("job not found", slog.String("jobID", jobIDParam))

			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, resp.Error("not found"))
			return
		}

		if err != nil {
			logger.Error("failed to get job", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))
			return
		}

		render.JSON(w, r, resp.GetResponse{
			Response:    resp.OK(),
			JobID:       job.ID,
			Queue:       job.Queue,
			Payload:     job.Payload,
			JobStatus:   job.JobStatus,
			AvailableAt: job.AvailableAt,
			CreatedAt:   job.CreatedAt,
		})
	}
}
