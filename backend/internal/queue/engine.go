package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gautammehendale/dispatch/internal/models"
	"github.com/gautammehendale/dispatch/internal/store"
)

type Engine struct {
	redis    *store.RedisStore
	postgres *store.PostgresStore
	hub      EventHub
}

// EventHub lets the engine broadcast events without importing the WS layer.
type EventHub interface {
	Broadcast(event *models.WSEvent)
}

func NewEngine(r *store.RedisStore, pg *store.PostgresStore, hub EventHub) *Engine {
	return &Engine{redis: r, postgres: pg, hub: hub}
}

func (e *Engine) Enqueue(ctx context.Context, req *models.EnqueueRequest) (*models.Job, error) {
	if req.Queue == "" {
		req.Queue = "default"
	}
	if req.MaxRetries == 0 {
		req.MaxRetries = 3
	}
	runAt := time.Now()
	if req.RunAt != nil {
		runAt = *req.RunAt
	}

	job := &models.Job{
		ID:         uuid.NewString(),
		Type:       req.Type,
		Payload:    req.Payload,
		Priority:   models.ParsePriority(req.Priority),
		Status:     models.StatusPending,
		Queue:      req.Queue,
		MaxRetries: req.MaxRetries,
		Attempts:   0,
		RunAt:      runAt,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Meta:       req.Meta,
	}

	if runAt.After(time.Now().Add(time.Second)) {
		if err := e.scheduleDelayed(ctx, job); err != nil {
			return nil, fmt.Errorf("schedule delayed: %w", err)
		}
	} else {
		if err := e.redis.Enqueue(ctx, job); err != nil {
			return nil, fmt.Errorf("redis enqueue: %w", err)
		}
	}

	if err := e.postgres.SaveJob(ctx, job); err != nil {
		return nil, fmt.Errorf("pg save: %w", err)
	}

	e.hub.Broadcast(&models.WSEvent{Type: "job.enqueued", Payload: job})
	return job, nil
}

func (e *Engine) scheduleDelayed(ctx context.Context, job *models.Job) error {
	job.Status = models.StatusPending
	return e.redis.RequeueWithBackoff(ctx, job, time.Until(job.RunAt))
}

func (e *Engine) Dequeue(ctx context.Context, queueName string) (*models.Job, error) {
	return e.redis.Dequeue(ctx, queueName)
}

func (e *Engine) MarkRunning(ctx context.Context, job *models.Job, workerID string) error {
	now := time.Now()
	job.Status = models.StatusRunning
	job.WorkerID = workerID
	job.StartedAt = &now
	job.Attempts++
	job.UpdatedAt = now
	if err := e.redis.UpdateJob(ctx, job); err != nil {
		return err
	}
	e.hub.Broadcast(&models.WSEvent{Type: "job.started", Payload: job})
	return e.postgres.SaveJob(ctx, job)
}

func (e *Engine) MarkCompleted(ctx context.Context, job *models.Job, exec *models.JobExecution) error {
	now := time.Now()
	job.Status = models.StatusCompleted
	job.CompletedAt = &now
	job.UpdatedAt = now
	e.redis.IncrCounter(ctx, "completed")
	e.redis.RecordLatency(ctx, exec.DurationMs)
	if err := e.redis.UpdateJob(ctx, job); err != nil {
		return err
	}
	_ = e.postgres.SaveJob(ctx, job)
	_ = e.postgres.SaveExecution(ctx, exec)
	e.hub.Broadcast(&models.WSEvent{Type: "job.completed", Payload: job})
	return nil
}

func (e *Engine) MarkFailed(ctx context.Context, job *models.Job, exec *models.JobExecution, err error) error {
	job.Error = err.Error()
	job.UpdatedAt = time.Now()

	_ = e.postgres.SaveExecution(ctx, exec)

	if job.Attempts >= job.MaxRetries {
		if rErr := e.redis.RequeueToDLQ(ctx, job); rErr != nil {
			return rErr
		}
		_ = e.postgres.SaveJob(ctx, job)
		e.hub.Broadcast(&models.WSEvent{Type: "job.dead", Payload: job})
		return nil
	}

	backoff := backoffDuration(job.Attempts)
	job.Status = models.StatusRetrying
	e.redis.IncrCounter(ctx, "failed")
	if rErr := e.redis.RequeueWithBackoff(ctx, job, backoff); rErr != nil {
		return rErr
	}
	_ = e.postgres.SaveJob(ctx, job)
	e.hub.Broadcast(&models.WSEvent{Type: "job.failed", Payload: job})
	return nil
}

func (e *Engine) CancelJob(ctx context.Context, id string) error {
	job, err := e.redis.GetJob(ctx, id)
	if err != nil {
		return fmt.Errorf("job not found: %w", err)
	}
	if job.Status != models.StatusPending {
		return fmt.Errorf("can only cancel pending jobs")
	}
	job.Status = models.StatusCancelled
	job.UpdatedAt = time.Now()
	_ = e.redis.UpdateJob(ctx, job)
	_ = e.postgres.SaveJob(ctx, job)
	e.hub.Broadcast(&models.WSEvent{Type: "job.cancelled", Payload: job})
	return nil
}

// backoffDuration computes exponential backoff: 5s, 25s, 125s, ...
func backoffDuration(attempts int) time.Duration {
	base := 3 * time.Second
	for i := 0; i < attempts; i++ {
		base *= 3
		if base > 30*time.Second {
			return 30 * time.Second
		}
	}
	return base
}
