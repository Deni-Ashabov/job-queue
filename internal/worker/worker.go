package worker

import (
	"context"
	"encoding/json"
	"job-queue/internal/config"
	"job-queue/internal/domain/job"
	"job-queue/internal/dto"
	"job-queue/internal/logger/sl"
	"job-queue/internal/models"
	"log/slog"
	"time"
)

type Storage interface {
	FetchAndLockJobs(ctx context.Context, fromStatus models.JobStatus, toStatus models.JobStatus, limit int) ([]job.Job, error)
	ChangeStatus(ctx context.Context, jobID int64, status models.JobStatus) error
}

type Worker struct {
	storage  Storage
	log      *slog.Logger
	handlers map[string]func(job.Job) error
}

func New(storage Storage, log *slog.Logger) *Worker {
	w := &Worker{
		storage: storage,
		log:     log,
	}

	w.handlers = map[string]func(job.Job) error{
		"emails":       w.ProcessEmailJob,
		"payment":      w.processPaymentJob,
		"notification": w.processNotificationJob,
	}

	return w
}

func (w *Worker) Run(ctx context.Context, cfg *config.Config) {
	jobsChan := make(chan job.Job, cfg.JobsBufferSize)

	for i := 0; i < cfg.WorkersCount; i++ {
		go w.workerLoop(ctx, jobsChan)
	}

	w.log.Info("worker started")

	for {
		select {
		case <-ctx.Done():
			w.log.Info("stopping worker")
			return

		default:
			jobs, err := w.storage.FetchAndLockJobs(ctx, models.StatusProcessing, models.StatusPending, cfg.JobsBufferSize)
			if err != nil {
				w.log.Error("failed to fetch jobs", sl.Err(err))
				time.Sleep(time.Second)
				continue
			}

			for _, job := range jobs {
				select {
				case jobsChan <- job:
				case <-ctx.Done():
					return
				}
			}

			time.Sleep(1 * time.Second)
		}
	}
}

func (w *Worker) workerLoop(ctx context.Context, jobs <-chan job.Job) {
	for job := range jobs {
		w.ProcessJob(ctx, job)
	}
}

func (w *Worker) ProcessJob(ctx context.Context, job job.Job) {
	handler, ok := w.handlers[string(job.Queue)]
	if !ok {
		w.log.Error("wrong job type", slog.String("queue", string(job.Queue)))

		if err := w.storage.ChangeStatus(ctx, job.ID, models.StatusFailed); err != nil {
			w.log.Error("failed to change status", sl.Err(err))
		}
		return
	}

	err := handler(job)
	if err != nil {
		w.log.Error("job failed", sl.Err(err))

		if err := w.storage.ChangeStatus(ctx, job.ID, models.StatusFailed); err != nil {
			w.log.Error("failed to change status", sl.Err(err))
		}
		return
	}

	if err := w.storage.ChangeStatus(ctx, job.ID, models.StatusDone); err != nil {
		w.log.Error("failed to change status", sl.Err(err))
	}
}

func (w *Worker) ProcessEmailJob(job job.Job) error {
	var payload dto.EmailJob

	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		w.log.Error("invalid email payload", sl.Err(err))
		return err
	}

	w.log.Info("email job processing",
		slog.String("to", payload.To),
		slog.String("subject", payload.To),
	)

	// TODO: send email
	w.log.Info("email job done (stub)")

	return nil
}

func (w *Worker) processPaymentJob(job job.Job) error {
	var payload dto.PaymentJob

	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		w.log.Error("invalid payment payload", sl.Err(err))
		return err
	}

	w.log.Info("payment job processing",
		slog.String("user_id", payload.UserID),
		slog.String("amount", payload.Amount),
	)

	// TODO: charge payment
	w.log.Info("payment job done (stub)")

	return nil
}

func (w *Worker) processNotificationJob(job job.Job) error {
	var payload dto.NotificationJob

	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		w.log.Error("invalid notification payload", sl.Err(err))
		return err
	}

	w.log.Info("notification job processing",
		slog.String("user_id", payload.UserID),
		slog.String("amount", payload.Text),
	)

	// TODO: push notification
	w.log.Info("notification job done (stub)")

	return nil
}
