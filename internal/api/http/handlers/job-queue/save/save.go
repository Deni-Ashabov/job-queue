package save

import (
	"encoding/json"
	resp "job-queue/internal/api/http/dto"
	"job-queue/internal/logger/sl"
	"job-queue/internal/models"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	Queue   models.QueueType `json:"queue" validate:"oneof=emails payment notification"`
	Payload json.RawMessage  `json:"payload" validate:"required,json"`
}

type Response struct {
	resp.Response
	JobID       int64            `json:"id"`
	Queue       models.QueueType `json:"queue"`
	JobStatus   string           `json:"job_status"`
	AvailableAt time.Time        `json:"available_at"`
}

const op = "handlers.job-queue.save.New"

func New(log *slog.Logger, jobSaver JobSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)

		if err != nil {
			logger.Error("failed to decode request body", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}

		logger.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			logger.Error("invalid request", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}

		job, err := jobSaver.SaveJob(r.Context(), req.Queue, req.Payload)

		if err != nil {
			logger.Error("failed to add job", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to add job"))
			return
		}

		logger.Info("job added", slog.Int64("id", job.ID))

		render.JSON(w, r, Response{
			Response:    resp.OK(),
			JobID:       job.ID,
			Queue:       job.Queue,
			JobStatus:   job.JobStatus,
			AvailableAt: job.AvailableAt,
		})
	}
}
