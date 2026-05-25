package queue_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gautammehendale/dispatch/internal/models"
	"github.com/gautammehendale/dispatch/internal/store"
	"github.com/gautammehendale/dispatch/internal/queue"
)

// BenchmarkEnqueue measures raw enqueue throughput.
func BenchmarkEnqueue(b *testing.B) {
	r, err := store.NewRedisStore("localhost:6379")
	if err != nil {
		b.Skipf("redis not available: %v", err)
	}
	pg, err := store.NewPostgresStore("postgres://dispatch:dispatch@localhost:5432/dispatch?sslmode=disable")
	if err != nil {
		b.Skipf("postgres not available: %v", err)
	}
	eng := queue.NewEngine(r, pg, &mockHub{})
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = eng.Enqueue(ctx, &models.EnqueueRequest{
			Type:    "bench_job",
			Payload: map[string]any{"n": i},
		})
	}
}

// BenchmarkDequeue measures dequeue throughput from a pre-filled queue.
func BenchmarkDequeue(b *testing.B) {
	r, err := store.NewRedisStore("localhost:6379")
	if err != nil {
		b.Skipf("redis not available: %v", err)
	}
	pg, err := store.NewPostgresStore("postgres://dispatch:dispatch@localhost:5432/dispatch?sslmode=disable")
	if err != nil {
		b.Skipf("postgres not available: %v", err)
	}
	eng := queue.NewEngine(r, pg, &mockHub{})
	ctx := context.Background()

	// Pre-fill queue.
	for i := 0; i < b.N; i++ {
		_, _ = eng.Enqueue(ctx, &models.EnqueueRequest{Type: "dequeue_bench"})
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = eng.Dequeue(ctx, "default")
	}
}

// BenchmarkEnqueueParallel measures concurrent enqueue with 4 goroutines — simulates real worker load.
func BenchmarkEnqueueParallel(b *testing.B) {
	r, err := store.NewRedisStore("localhost:6379")
	if err != nil {
		b.Skipf("redis not available: %v", err)
	}
	pg, err := store.NewPostgresStore("postgres://dispatch:dispatch@localhost:5432/dispatch?sslmode=disable")
	if err != nil {
		b.Skipf("postgres not available: %v", err)
	}
	eng := queue.NewEngine(r, pg, &mockHub{})
	ctx := context.Background()

	b.SetParallelism(4)
	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = eng.Enqueue(ctx, &models.EnqueueRequest{Type: "parallel_bench"})
		}
	})
}

// TestThroughput measures end-to-end jobs/sec over a 5-second window with 4 workers.
// Run with: go test -v -run TestThroughput -timeout 30s
func TestThroughput(t *testing.T) {
	r, err := store.NewRedisStore("localhost:6379")
	if err != nil {
		t.Skipf("redis not available: %v", err)
	}
	pg, err := store.NewPostgresStore("postgres://dispatch:dispatch@localhost:5432/dispatch?sslmode=disable")
	if err != nil {
		t.Skipf("postgres not available: %v", err)
	}

	eng := queue.NewEngine(r, pg, &mockHub{})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var enqueued, processed int64
	var wg sync.WaitGroup
	numWorkers := 4

	// Producer goroutine.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				_, _ = eng.Enqueue(ctx, &models.EnqueueRequest{Type: "throughput_bench"})
				atomic.AddInt64(&enqueued, 1)
			}
		}
	}()

	// Worker goroutines.
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			workerID := fmt.Sprintf("bench-worker-%d", id)
			for {
				select {
				case <-ctx.Done():
					return
				default:
					job, err := eng.Dequeue(ctx, "default")
					if err != nil || job == nil {
						time.Sleep(time.Millisecond)
						continue
					}
					start := time.Now()
					_ = eng.MarkRunning(ctx, job, workerID)
					exec := &models.JobExecution{
						ID: fmt.Sprintf("exec-%d", atomic.AddInt64(&processed, 1)),
						JobID: job.ID, WorkerID: workerID,
						Attempt: job.Attempts, StartedAt: start, EndedAt: time.Now(),
						DurationMs: time.Since(start).Milliseconds(),
						Status: models.StatusCompleted,
					}
					_ = eng.MarkCompleted(ctx, job, exec)
				}
			}
		}(i)
	}

	start := time.Now()
	wg.Wait()
	elapsed := time.Since(start).Seconds()

	total := atomic.LoadInt64(&processed)
	throughput := float64(total) / elapsed
	t.Logf("─────────────────────────────────────────")
	t.Logf("  Throughput benchmark (4 workers, 5s)")
	t.Logf("  Jobs processed : %d", total)
	t.Logf("  Elapsed        : %.2fs", elapsed)
	t.Logf("  Jobs/sec       : %.0f", throughput)
	t.Logf("─────────────────────────────────────────")

	if throughput < 100 {
		t.Errorf("throughput too low: %.0f jobs/sec (expected > 100)", throughput)
	}
}

// TestLatency measures p50/p95/p99 job execution latency.
func TestLatency(t *testing.T) {
	r, err := store.NewRedisStore("localhost:6379")
	if err != nil {
		t.Skipf("redis not available: %v", err)
	}
	pg, err := store.NewPostgresStore("postgres://dispatch:dispatch@localhost:5432/dispatch?sslmode=disable")
	if err != nil {
		t.Skipf("postgres not available: %v", err)
	}

	eng := queue.NewEngine(r, pg, &mockHub{})
	ctx := context.Background()

	const n = 200
	latencies := make([]time.Duration, 0, n)
	var mu sync.Mutex

	for i := 0; i < n; i++ {
		job, _ := eng.Enqueue(ctx, &models.EnqueueRequest{Type: "latency_bench"})
		start := time.Now()
		dequeued, _ := eng.Dequeue(ctx, "default")
		if dequeued == nil {
			continue
		}
		_ = eng.MarkRunning(ctx, dequeued, "latency-worker")
		exec := &models.JobExecution{
			ID: fmt.Sprintf("lat-%d", i), JobID: dequeued.ID, WorkerID: "latency-worker",
			Attempt: 1, StartedAt: start, EndedAt: time.Now(),
			DurationMs: time.Since(start).Milliseconds(), Status: models.StatusCompleted,
		}
		_ = eng.MarkCompleted(ctx, dequeued, exec)
		mu.Lock()
		latencies = append(latencies, time.Since(start))
		mu.Unlock()
		_ = job
	}

	if len(latencies) == 0 {
		t.Skip("no latency data collected")
	}

	// Sort for percentiles.
	for i := 1; i < len(latencies); i++ {
		for j := i; j > 0 && latencies[j] < latencies[j-1]; j-- {
			latencies[j], latencies[j-1] = latencies[j-1], latencies[j]
		}
	}

	p50 := latencies[len(latencies)*50/100]
	p95 := latencies[len(latencies)*95/100]
	p99 := latencies[len(latencies)*99/100]

	t.Logf("─────────────────────────────────────")
	t.Logf("  Latency benchmark (%d jobs)", len(latencies))
	t.Logf("  p50 : %v", p50.Round(time.Microsecond))
	t.Logf("  p95 : %v", p95.Round(time.Microsecond))
	t.Logf("  p99 : %v", p99.Round(time.Microsecond))
	t.Logf("─────────────────────────────────────")
}
