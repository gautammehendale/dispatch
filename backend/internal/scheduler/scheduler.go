package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/gautammehendale/dispatch/internal/store"
)

type Scheduler struct {
	redis     *store.RedisStore
	queueName string
	interval  time.Duration
}

func New(redis *store.RedisStore, queueName string) *Scheduler {
	return &Scheduler{
		redis:     redis,
		queueName: queueName,
		interval:  time.Second,
	}
}

// Run polls for delayed jobs ready to be enqueued and moves them to the active queue.
func (s *Scheduler) Run(ctx context.Context) {
	tick := time.NewTicker(s.interval)
	defer tick.Stop()
	log.Printf("[scheduler] running, polling every %s", s.interval)
	for {
		select {
		case <-tick.C:
			if err := s.redis.PollScheduled(ctx, s.queueName); err != nil {
				log.Printf("[scheduler] poll error: %v", err)
			}
		case <-ctx.Done():
			log.Printf("[scheduler] stopped")
			return
		}
	}
}
