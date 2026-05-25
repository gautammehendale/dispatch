package worker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gautammehendale/dispatch/internal/models"
	"github.com/gautammehendale/dispatch/internal/queue"
	"github.com/gautammehendale/dispatch/internal/store"
)

type HandlerFunc func(ctx context.Context, job *models.Job) error

type Pool struct {
	id          string
	concurrency int
	queueName   string
	engine      *queue.Engine
	redis       *store.RedisStore
	handlers    map[string]HandlerFunc
	status      []*models.WorkerStatus
	mu          sync.RWMutex
	wg          sync.WaitGroup
	stopCh      chan struct{}
}

func NewPool(concurrency int, queueName string, engine *queue.Engine, redis *store.RedisStore) *Pool {
	statuses := make([]*models.WorkerStatus, concurrency)
	for i := 0; i < concurrency; i++ {
		statuses[i] = &models.WorkerStatus{
			ID:        fmt.Sprintf("worker-%s-%d", uuid.NewString()[:8], i+1),
			Status:    "idle",
			StartedAt: time.Now(),
		}
	}
	return &Pool{
		id:          uuid.NewString(),
		concurrency: concurrency,
		queueName:   queueName,
		engine:      engine,
		redis:       redis,
		handlers:    make(map[string]HandlerFunc),
		status:      statuses,
		stopCh:      make(chan struct{}),
	}
}

func (p *Pool) Register(jobType string, handler HandlerFunc) {
	p.handlers[jobType] = handler
}

func (p *Pool) Start(ctx context.Context) {
	log.Printf("[pool] starting %d workers on queue %q", p.concurrency, p.queueName)
	for i := 0; i < p.concurrency; i++ {
		p.wg.Add(1)
		go p.runWorker(ctx, p.status[i])
	}
	go p.heartbeat(ctx)
}

func (p *Pool) Stop() {
	close(p.stopCh)
	p.wg.Wait()
	log.Printf("[pool] all workers stopped")
}

func (p *Pool) Workers() []*models.WorkerStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()
	result := make([]*models.WorkerStatus, len(p.status))
	copy(result, p.status)
	return result
}

func (p *Pool) runWorker(ctx context.Context, ws *models.WorkerStatus) {
	defer p.wg.Done()
	for {
		select {
		case <-p.stopCh:
			p.setWorkerStatus(ws, "stopped", "")
			return
		case <-ctx.Done():
			p.setWorkerStatus(ws, "stopped", "")
			return
		default:
			job, err := p.engine.Dequeue(ctx, p.queueName)
			if err != nil {
				log.Printf("[worker %s] dequeue error: %v", ws.ID, err)
				time.Sleep(500 * time.Millisecond)
				continue
			}
			if job == nil {
				time.Sleep(200 * time.Millisecond)
				continue
			}
			p.processJob(ctx, ws, job)
		}
	}
}

func (p *Pool) processJob(ctx context.Context, ws *models.WorkerStatus, job *models.Job) {
	p.setWorkerStatus(ws, "busy", job.ID)

	if err := p.engine.MarkRunning(ctx, job, ws.ID); err != nil {
		log.Printf("[worker %s] mark running: %v", ws.ID, err)
	}

	startedAt := time.Now()
	handler, ok := p.handlers[job.Type]
	var execErr error
	if !ok {
		execErr = fmt.Errorf("no handler registered for job type %q", job.Type)
	} else {
		execErr = handler(ctx, job)
	}
	endedAt := time.Now()

	exec := &models.JobExecution{
		ID:         uuid.NewString(),
		JobID:      job.ID,
		WorkerID:   ws.ID,
		Attempt:    job.Attempts,
		StartedAt:  startedAt,
		EndedAt:    endedAt,
		DurationMs: endedAt.Sub(startedAt).Milliseconds(),
		Status:     models.StatusCompleted,
	}

	if execErr != nil {
		exec.Status = models.StatusFailed
		exec.Error = execErr.Error()
		if err := p.engine.MarkFailed(ctx, job, exec, execErr); err != nil {
			log.Printf("[worker %s] mark failed: %v", ws.ID, err)
		}
	} else {
		if err := p.engine.MarkCompleted(ctx, job, exec); err != nil {
			log.Printf("[worker %s] mark completed: %v", ws.ID, err)
		}
	}

	p.mu.Lock()
	ws.JobsRun++
	ws.LastBeatAt = time.Now()
	p.mu.Unlock()

	p.setWorkerStatus(ws, "idle", "")
}

func (p *Pool) setWorkerStatus(ws *models.WorkerStatus, status, currentJob string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	ws.Status = status
	ws.CurrentJob = currentJob
	ws.LastBeatAt = time.Now()
}

func (p *Pool) heartbeat(ctx context.Context) {
	tick := time.NewTicker(2 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			p.mu.Lock()
			for _, ws := range p.status {
				ws.LastBeatAt = time.Now()
			}
			p.mu.Unlock()
		case <-p.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}
