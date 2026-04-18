package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	jobDomain "job-queue/internal/domain/job"
	"job-queue/internal/models"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	OpNew              = "repository.pgx.New"
	OpSaveJob          = "repository.pgx.SaveJob"
	OpGetJob           = "repository.pgx.GetJob"
	OpDeleteJob        = "repository.pgx.DeleteJob"
	OpChangeStatus     = "repository.pgx.ChangeStatus"
	OpFetchPendingJobs = "repository.pgx.FetchPendingJobs"
)

type Storage struct {
	db *pgxpool.Pool
}

func New(ctx context.Context, connStr string) (*Storage, error) {
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", OpNew, err)
	}

	const maxRetries = 5
	const delay = 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		if err := pool.Ping(ctx); err == nil {
			return &Storage{db: pool}, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}

	return nil, fmt.Errorf("%s: db is not reachable after retries", OpNew)
}

func (s *Storage) Close() {
	s.db.Close()
}

func (s *Storage) SaveJob(ctx context.Context, queue models.QueueType, payload json.RawMessage) (jobDomain.Job, error) {
	var job jobDomain.Job

	err := s.db.QueryRow(ctx, `
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
		return job, fmt.Errorf("%s: %w", OpSaveJob, err)
	}

	return job, nil
}

func (s *Storage) FetchPendingJobs(ctx context.Context) ([]jobDomain.Job, error) {
	rows, err := s.db.Query(ctx, `
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
		return nil, fmt.Errorf("%s: %w", OpFetchPendingJobs, err)
	}

	defer rows.Close()

	jobs := make([]jobDomain.Job, 0)

	for rows.Next() {
		var job jobDomain.Job

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

func (s *Storage) ChangeStatus(ctx context.Context, jobID int64, status models.JobStatus) error {
	res, err := s.db.Exec(ctx, `
		UPDATE jobs
		SET job_status = $1
		WHERE id = $2;
	`, status, jobID)
	if err != nil {
		return fmt.Errorf("%s: %w", OpChangeStatus, err)
	}

	rows := res.RowsAffected()

	if rows == 0 {
		return fmt.Errorf("%s: job not found", OpChangeStatus)
	}

	return nil
}

func (s *Storage) GetJob(ctx context.Context, jobID int) (jobDomain.Job, error) {
	var job jobDomain.Job

	err := s.db.QueryRow(ctx, `
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
			return jobDomain.Job{}, jobDomain.ErrJobNotFound
		}
		return jobDomain.Job{}, fmt.Errorf("%s: %w", OpGetJob, err)
	}

	return job, nil
}

func (s *Storage) DeleteJob(ctx context.Context, jobID int) error {
	res, err := s.db.Exec(ctx, `
		DELETE FROM jobs WHERE id = $1
	`, jobID)
	if err != nil {
		return fmt.Errorf("%s: %w", OpDeleteJob, err)
	}

	rows := res.RowsAffected()

	if rows == 0 {
		return jobDomain.ErrJobNotFound
	}

	return nil
}
