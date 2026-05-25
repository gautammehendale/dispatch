package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/gautammehendale/dispatch/internal/models"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(dsn string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres open: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("postgres ping: %w", err)
	}
	s := &PostgresStore{db: db}
	if err := s.migrate(ctx); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

func (p *PostgresStore) migrate(ctx context.Context) error {
	_, err := p.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS jobs (
			id           TEXT PRIMARY KEY,
			type         TEXT NOT NULL,
			payload      JSONB,
			priority     INTEGER NOT NULL DEFAULT 2,
			status       TEXT NOT NULL DEFAULT 'pending',
			queue        TEXT NOT NULL DEFAULT 'default',
			max_retries  INTEGER NOT NULL DEFAULT 3,
			attempts     INTEGER NOT NULL DEFAULT 0,
			run_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			started_at   TIMESTAMPTZ,
			completed_at TIMESTAMPTZ,
			error        TEXT,
			worker_id    TEXT,
			meta         JSONB
		);

		CREATE TABLE IF NOT EXISTS job_executions (
			id          TEXT PRIMARY KEY,
			job_id      TEXT NOT NULL REFERENCES jobs(id),
			worker_id   TEXT NOT NULL,
			attempt     INTEGER NOT NULL,
			started_at  TIMESTAMPTZ NOT NULL,
			ended_at    TIMESTAMPTZ NOT NULL,
			duration_ms BIGINT NOT NULL,
			status      TEXT NOT NULL,
			error       TEXT
		);

		CREATE INDEX IF NOT EXISTS idx_jobs_status   ON jobs(status);
		CREATE INDEX IF NOT EXISTS idx_jobs_queue    ON jobs(queue);
		CREATE INDEX IF NOT EXISTS idx_jobs_priority ON jobs(priority DESC);
		CREATE INDEX IF NOT EXISTS idx_jobs_run_at   ON jobs(run_at);
	`)
	return err
}

func (p *PostgresStore) SaveJob(ctx context.Context, job *models.Job) error {
	payload, _ := json.Marshal(job.Payload)
	meta, _ := json.Marshal(job.Meta)
	_, err := p.db.ExecContext(ctx, `
		INSERT INTO jobs (id, type, payload, priority, status, queue, max_retries,
		                  attempts, run_at, created_at, updated_at, error, worker_id, meta)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		ON CONFLICT (id) DO UPDATE SET
			status       = EXCLUDED.status,
			attempts     = EXCLUDED.attempts,
			updated_at   = EXCLUDED.updated_at,
			started_at   = EXCLUDED.started_at,
			completed_at = EXCLUDED.completed_at,
			error        = EXCLUDED.error,
			worker_id    = EXCLUDED.worker_id`,
		job.ID, job.Type, payload, int(job.Priority), string(job.Status),
		job.Queue, job.MaxRetries, job.Attempts, job.RunAt,
		job.CreatedAt, job.UpdatedAt, job.Error, job.WorkerID, meta,
	)
	return err
}

func (p *PostgresStore) SaveExecution(ctx context.Context, exec *models.JobExecution) error {
	_, err := p.db.ExecContext(ctx, `
		INSERT INTO job_executions (id, job_id, worker_id, attempt, started_at, ended_at, duration_ms, status, error)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		exec.ID, exec.JobID, exec.WorkerID, exec.Attempt,
		exec.StartedAt, exec.EndedAt, exec.DurationMs,
		string(exec.Status), exec.Error,
	)
	return err
}

func (p *PostgresStore) ListJobs(ctx context.Context, status, queue string, limit, offset int) ([]*models.Job, int, error) {
	args := []any{}
	where := "WHERE 1=1"
	i := 1
	if status != "" {
		where += fmt.Sprintf(" AND status = $%d", i)
		args = append(args, status)
		i++
	}
	if queue != "" {
		where += fmt.Sprintf(" AND queue = $%d", i)
		args = append(args, queue)
		i++
	}

	var total int
	countArgs := make([]any, len(args))
	copy(countArgs, args)
	row := p.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM jobs "+where, countArgs...)
	if err := row.Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	rows, err := p.db.QueryContext(ctx,
		"SELECT id, type, priority, status, queue, attempts, max_retries, created_at, updated_at, error, worker_id "+
			"FROM jobs "+where+" ORDER BY created_at DESC LIMIT $"+fmt.Sprint(i)+" OFFSET $"+fmt.Sprint(i+1),
		args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var jobs []*models.Job
	for rows.Next() {
		j := &models.Job{}
		if err := rows.Scan(&j.ID, &j.Type, &j.Priority, &j.Status, &j.Queue,
			&j.Attempts, &j.MaxRetries, &j.CreatedAt, &j.UpdatedAt, &j.Error, &j.WorkerID); err != nil {
			return nil, 0, err
		}
		jobs = append(jobs, j)
	}
	return jobs, total, nil
}

func (p *PostgresStore) GetJob(ctx context.Context, id string) (*models.Job, error) {
	row := p.db.QueryRowContext(ctx,
		"SELECT id, type, payload, priority, status, queue, max_retries, attempts, run_at, created_at, updated_at, error, worker_id FROM jobs WHERE id=$1", id)
	j := &models.Job{}
	var payload []byte
	err := row.Scan(&j.ID, &j.Type, &payload, &j.Priority, &j.Status, &j.Queue,
		&j.MaxRetries, &j.Attempts, &j.RunAt, &j.CreatedAt, &j.UpdatedAt, &j.Error, &j.WorkerID)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(payload, &j.Payload)
	return j, nil
}

func (p *PostgresStore) GetExecutions(ctx context.Context, jobID string) ([]*models.JobExecution, error) {
	rows, err := p.db.QueryContext(ctx,
		"SELECT id, job_id, worker_id, attempt, started_at, ended_at, duration_ms, status, error FROM job_executions WHERE job_id=$1 ORDER BY attempt", jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var execs []*models.JobExecution
	for rows.Next() {
		e := &models.JobExecution{}
		if err := rows.Scan(&e.ID, &e.JobID, &e.WorkerID, &e.Attempt, &e.StartedAt, &e.EndedAt, &e.DurationMs, &e.Status, &e.Error); err != nil {
			return nil, err
		}
		execs = append(execs, e)
	}
	return execs, nil
}
