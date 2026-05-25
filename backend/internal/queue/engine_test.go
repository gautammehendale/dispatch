package queue_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gautammehendale/dispatch/internal/models"
	"github.com/gautammehendale/dispatch/internal/queue"
	"github.com/gautammehendale/dispatch/internal/store"
)

// mockHub satisfies queue.EventHub without a real WebSocket server.
type mockHub struct{ events []string }

func (m *mockHub) Broadcast(e *models.WSEvent) { m.events = append(m.events, e.Type) }

func setupEngine(t *testing.T) (*queue.Engine, *store.RedisStore) {
	t.Helper()
	r, err := store.NewRedisStore("localhost:6379")
	if err != nil {
		t.Skipf("redis not available: %v", err)
	}
	pg, err := store.NewPostgresStore("postgres://dispatch:dispatch@localhost:5432/dispatch?sslmode=disable")
	if err != nil {
		t.Skipf("postgres not available: %v", err)
	}
	hub := &mockHub{}
	return queue.NewEngine(r, pg, hub), r
}

func TestEnqueue_BasicJob(t *testing.T) {
	eng, _ := setupEngine(t)
	ctx := context.Background()

	job, err := eng.Enqueue(ctx, &models.EnqueueRequest{
		Type:     "test_job",
		Priority: "HIGH",
		Payload:  map[string]any{"key": "value"},
	})
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	if job.ID == "" {
		t.Error("expected non-empty job ID")
	}
	if job.Status != models.StatusPending {
		t.Errorf("expected pending, got %s", job.Status)
	}
	if job.Priority != models.PriorityHigh {
		t.Errorf("expected HIGH priority, got %v", job.Priority)
	}
}

func TestEnqueue_DefaultsApplied(t *testing.T) {
	eng, _ := setupEngine(t)
	ctx := context.Background()

	job, err := eng.Enqueue(ctx, &models.EnqueueRequest{Type: "noop"})
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	if job.Queue != "default" {
		t.Errorf("expected queue=default, got %s", job.Queue)
	}
	if job.MaxRetries != 3 {
		t.Errorf("expected max_retries=3, got %d", job.MaxRetries)
	}
	if job.Priority != models.PriorityNormal {
		t.Errorf("expected NORMAL priority, got %v", job.Priority)
	}
}

func TestDequeue_PriorityOrdering(t *testing.T) {
	eng, _ := setupEngine(t)
	ctx := context.Background()

	// Enqueue LOW first, then CRITICAL — CRITICAL must come out first.
	_, _ = eng.Enqueue(ctx, &models.EnqueueRequest{Type: "low_job",      Priority: "LOW"})
	_, _ = eng.Enqueue(ctx, &models.EnqueueRequest{Type: "critical_job", Priority: "CRITICAL"})

	job, err := eng.Dequeue(ctx, "default")
	if err != nil {
		t.Fatalf("dequeue: %v", err)
	}
	if job == nil {
		t.Fatal("expected a job, got nil")
	}
	if job.Priority != models.PriorityCritical {
		t.Errorf("expected CRITICAL to dequeue first, got priority %v (type: %s)", job.Priority, job.Type)
	}
}

func TestMarkRunning_UpdatesState(t *testing.T) {
	eng, _ := setupEngine(t)
	ctx := context.Background()

	job, _ := eng.Enqueue(ctx, &models.EnqueueRequest{Type: "state_test"})
	dequeued, _ := eng.Dequeue(ctx, "default")
	if dequeued == nil {
		t.Skip("no job dequeued — queue may have had leftover jobs")
	}

	if err := eng.MarkRunning(ctx, dequeued, "worker-001"); err != nil {
		t.Fatalf("mark running: %v", err)
	}
	if dequeued.Status != models.StatusRunning {
		t.Errorf("expected running, got %s", dequeued.Status)
	}
	if dequeued.WorkerID != "worker-001" {
		t.Errorf("expected worker-001, got %s", dequeued.WorkerID)
	}
	if dequeued.Attempts != 1 {
		t.Errorf("expected attempts=1, got %d", dequeued.Attempts)
	}
	_ = job
}

func TestMarkFailed_RetryQueued(t *testing.T) {
	eng, _ := setupEngine(t)
	ctx := context.Background()

	job, _ := eng.Enqueue(ctx, &models.EnqueueRequest{Type: "fail_test", MaxRetries: 3})
	job.Attempts = 1

	exec := &models.JobExecution{
		ID: "exec-1", JobID: job.ID, WorkerID: "w1",
		Attempt: 1, StartedAt: time.Now(), EndedAt: time.Now(),
		DurationMs: 10, Status: models.StatusFailed, Error: "timeout",
	}

	if err := eng.MarkFailed(ctx, job, exec, fmt.Errorf("timeout")); err != nil {
		t.Fatalf("mark failed: %v", err)
	}
	if job.Status != models.StatusRetrying {
		t.Errorf("expected retrying, got %s", job.Status)
	}
}

func TestMarkFailed_ExhaustsToDeadLetter(t *testing.T) {
	eng, _ := setupEngine(t)
	ctx := context.Background()

	job, _ := eng.Enqueue(ctx, &models.EnqueueRequest{Type: "dlq_test", MaxRetries: 2})
	job.Attempts = 2 // already at max

	exec := &models.JobExecution{
		ID: "exec-dlq", JobID: job.ID, WorkerID: "w1",
		Attempt: 2, StartedAt: time.Now(), EndedAt: time.Now(),
		DurationMs: 5, Status: models.StatusFailed,
	}
	if err := eng.MarkFailed(ctx, job, exec, fmt.Errorf("permanent error")); err != nil {
		t.Fatalf("mark failed: %v", err)
	}
	if job.Status != models.StatusDead {
		t.Errorf("expected dead, got %s", job.Status)
	}
}

func TestCancelJob_PendingOnly(t *testing.T) {
	eng, _ := setupEngine(t)
	ctx := context.Background()

	job, _ := eng.Enqueue(ctx, &models.EnqueueRequest{Type: "cancel_test"})
	if err := eng.CancelJob(ctx, job.ID); err != nil {
		t.Fatalf("cancel: %v", err)
	}

	// Cancelling a non-pending job should error.
	job2, _ := eng.Enqueue(ctx, &models.EnqueueRequest{Type: "cancel_running"})
	dequeued, _ := eng.Dequeue(ctx, "default")
	if dequeued != nil && dequeued.ID == job2.ID {
		_ = eng.MarkRunning(ctx, dequeued, "w1")
		if err := eng.CancelJob(ctx, dequeued.ID); err == nil {
			t.Error("expected error cancelling running job")
		}
	}
}
