package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	storage "job-queue/internal/repository"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type JobStatus string

const (
	StatusPending    JobStatus = "pending"
	StatusProcessing JobStatus = "processing"
	StatusDone       JobStatus = "done"
	StatusFailed     JobStatus = "failed"
)

type QueueType string

const (
	QueueEmail        QueueType = "emails"
	QueuePayment      QueueType = "payment"
	QueueNotification QueueType = "notification"
)

type Storage struct {
	db *pgxpool.Pool
}

type Job struct {
	ID          int64
	Queue       QueueType
	Payload     json.RawMessage
	JobStatus   string
	AvailableAt time.Time
	CreatedAt   time.Time
}

func New(connStr string) (*Storage, error) {
	const op = "repository.pgx.New"

	pool, err := pgxpool.New(context.Background(), connStr)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: pool}, nil
}

func (s *Storage) Close() {
	s.db.Close()
}

func (s *Storage) SaveJob(queue QueueType, payload json.RawMessage) (Job, error) {
	const op = "repository.pgx.SaveJob"

	var job Job

	err := s.db.QueryRow(context.Background(), `
		INSERT INTO jobs (queue, payload)
		VALUES ($1, $2)
		RETURNING id, queue, job_status, available_at, created_at
	`, queue, payload).Scan(
		&job.ID,
		&job.Queue,
		&job.JobStatus,
		&job.AvailableAt,
		&job.CreatedAt,
	)

	if err != nil {
		return job, fmt.Errorf("%s: %w", op, err)
	}

	return job, nil
}

func (s *Storage) FetchPendingJobs() ([]Job, error) {
	const op = "repository.pgx.FetchPendingJobs"

	rows, err := s.db.Query(context.Background(), `
		UPDATE jobs
		SET job_status = 'processing'
		WHERE id IN (
			SELECT id FROM jobs
			WHERE job_status = 'pending'
			AND available_at <= now()
			LIMIT 10
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, queue, payload, job_status, available_at;
	`)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer rows.Close()

	jobs := make([]Job, 0)

	for rows.Next() {
		var job Job

		err := rows.Scan(
			&job.ID,
			&job.Queue,
			&job.Payload,
			&job.JobStatus,
			&job.AvailableAt,
		)

		if err != nil {
			return nil, err
		}

		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return jobs, nil
}

func (s *Storage) ChangeStatus(jobID int64, status JobStatus) error {
	const op = "repository.pgx.ChangeStatus"

	res, err := s.db.Exec(context.Background(), `
		UPDATE jobs
		SET job_status = $1
		WHERE id = $2;
	`, status, jobID)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rows := res.RowsAffected()

	if rows == 0 {
		return fmt.Errorf("%s: job not found", op)
	}

	return nil
}

func (s *Storage) GetJob(jobID int) (Job, error) {
	const op = "repository.pgx.GetJob"

	var job Job

	err := s.db.QueryRow(context.Background(), `
		SELECT id, queue, payload, job_status, available_at, created_at
		FROM jobs
		WHERE id = $1
	`, jobID).Scan(
		&job.ID,
		&job.Queue,
		&job.Payload,
		&job.JobStatus,
		&job.AvailableAt,
		&job.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Job{}, storage.ErrJobNotFound
		}
		return Job{}, fmt.Errorf("%s: %w", op, err)
	}

	return job, nil
}

func (s *Storage) DeleteJob(jobID int) error {
	const op = "repository.pgx.DeleteJob"

	res, err := s.db.Exec(context.Background(), `
		DELETE FROM jobs WHERE id = $1
	`, jobID)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rows := res.RowsAffected()

	if rows == 0 {
		return storage.ErrJobNotFound
	}

	return nil
}
