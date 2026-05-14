CREATE INDEX IF NOT EXISTS idx_jobs_ready
ON jobs (available_at)
WHERE job_status = 'pending';

CREATE INDEX IF NOT EXISTS idx_jobs_queue
ON jobs (queue);