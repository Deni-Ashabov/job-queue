package worker

import (
	"context"
	"encoding/json"
	"job-queue/internal/lib/logger/sl"
	"job-queue/internal/repository/postgres"
	"log/slog"
	"time"
)

type Storage interface {
	FetchPendingJobs() ([]postgres.Job, error)
	ChangeStatus(id int64, status postgres.JobStatus) error
}

type Worker struct {
	storage  Storage
	log      *slog.Logger
	handlers map[string]func(postgres.Job) error
}

func New(storage Storage, log *slog.Logger) *Worker {
	w := &Worker{
		storage: storage,
		log:     log,
	}

	w.handlers = map[string]func(postgres.Job) error{
		"emails":       w.ProcessEmailJob,
		"payment":      w.processPaymentJob,
		"notification": w.processNotificationJob,
	}

	return w
}

func (w *Worker) Run(ctx context.Context) {
	const workersCount = 5

	jobsChan := make(chan postgres.Job, 10)

	for i := 0; i < workersCount; i++ {
		go w.workerLoop(jobsChan)
	}

	w.log.Info("worker started")

	for {
		select {
		case <-ctx.Done():
			w.log.Info("stopping worker")
			return

		default:
			jobs, err := w.storage.FetchPendingJobs()
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

func (w *Worker) workerLoop(jobs <-chan postgres.Job) {
	for job := range jobs {
		w.ProcessJob(job)
	}
}

func (w *Worker) ProcessJob(job postgres.Job) {
	handler, ok := w.handlers[string(job.Queue)]
	if !ok {
		w.log.Error("wrong job type", slog.String("queue", string(job.Queue)))

		if err := w.storage.ChangeStatus(job.ID, postgres.StatusFailed); err != nil {
			w.log.Error("failed to change status", sl.Err(err))
		}
		return
	}

	err := handler(job)
	if err != nil {
		w.log.Error("job failed", sl.Err(err))

		if err := w.storage.ChangeStatus(job.ID, postgres.StatusFailed); err != nil {
			w.log.Error("failed to change status", sl.Err(err))
		}
		return
	}

	if err := w.storage.ChangeStatus(job.ID, postgres.StatusDone); err != nil {
		w.log.Error("failed to change status", sl.Err(err))
	}
}

func (w *Worker) ProcessEmailJob(job postgres.Job) error {
	var payload struct {
		To      string `json:"to"`
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}

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

func (w *Worker) processPaymentJob(job postgres.Job) error {
	var payload struct {
		UserID string `json:"user_id"`
		Amount string `json:"amount"`
	}

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

func (w *Worker) processNotificationJob(job postgres.Job) error {
	var payload struct {
		UserID string `json:"user_id"`
		Text   string `json:"text"`
	}

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
