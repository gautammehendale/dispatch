package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gautammehendale/dispatch/internal/models"
)

const (
	queueKeyFmt    = "dispatch:queue:%s:%d"
	jobKeyFmt      = "dispatch:job:%s"
	pausedKeyFmt   = "dispatch:queue:%s:paused"
	metricsKey     = "dispatch:metrics"
	latencyListKey = "dispatch:latency"
	pubsubChannel  = "dispatch:events"
)

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(addr string) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connect: %w", err)
	}
	return &RedisStore{client: client}, nil
}

func (r *RedisStore) Enqueue(ctx context.Context, job *models.Job) error {
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	pipe := r.client.Pipeline()
	jobKey := fmt.Sprintf(jobKeyFmt, job.ID)
	queueKey := fmt.Sprintf(queueKeyFmt, job.Queue, int(job.Priority))
	pipe.Set(ctx, jobKey, data, 24*time.Hour)
	pipe.LPush(ctx, queueKey, job.ID)
	pipe.Incr(ctx, "dispatch:counter:enqueued")
	_, err = pipe.Exec(ctx)
	return err
}

// Dequeue pops from the highest available priority queue for a given queue name.
func (r *RedisStore) Dequeue(ctx context.Context, queueName string) (*models.Job, error) {
	paused, _ := r.client.Get(ctx, fmt.Sprintf(pausedKeyFmt, queueName)).Bool()
	if paused {
		return nil, nil
	}

	priorities := []models.Priority{
		models.PriorityCritical,
		models.PriorityHigh,
		models.PriorityNormal,
		models.PriorityLow,
	}

	for _, p := range priorities {
		key := fmt.Sprintf(queueKeyFmt, queueName, int(p))
		result, err := r.client.RPop(ctx, key).Result()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			return nil, err
		}
		return r.GetJob(ctx, result)
	}
	return nil, nil
}

func (r *RedisStore) GetJob(ctx context.Context, id string) (*models.Job, error) {
	data, err := r.client.Get(ctx, fmt.Sprintf(jobKeyFmt, id)).Bytes()
	if err != nil {
		return nil, err
	}
	var job models.Job
	if err := json.Unmarshal(data, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *RedisStore) UpdateJob(ctx context.Context, job *models.Job) error {
	job.UpdatedAt = time.Now()
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, fmt.Sprintf(jobKeyFmt, job.ID), data, 24*time.Hour).Err()
}

func (r *RedisStore) RequeueToDLQ(ctx context.Context, job *models.Job) error {
	job.Status = models.StatusDead
	job.UpdatedAt = time.Now()
	data, _ := json.Marshal(job)
	pipe := r.client.Pipeline()
	pipe.Set(ctx, fmt.Sprintf(jobKeyFmt, job.ID), data, 7*24*time.Hour)
	pipe.LPush(ctx, "dispatch:dlq", job.ID)
	pipe.Incr(ctx, "dispatch:counter:dead")
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisStore) RequeueWithBackoff(ctx context.Context, job *models.Job, delay time.Duration) error {
	job.Status = models.StatusRetrying
	job.UpdatedAt = time.Now()
	data, _ := json.Marshal(job)
	pipe := r.client.Pipeline()
	pipe.Set(ctx, fmt.Sprintf(jobKeyFmt, job.ID), data, 24*time.Hour)
	pipe.ZAdd(ctx, "dispatch:scheduled", &redis.Z{
		Score:  float64(time.Now().Add(delay).Unix()),
		Member: job.ID,
	})
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisStore) PollScheduled(ctx context.Context, queueName string) error {
	now := float64(time.Now().Unix())
	ids, err := r.client.ZRangeByScore(ctx, "dispatch:scheduled", &redis.ZRangeBy{
		Min: "0",
		Max: fmt.Sprintf("%f", now),
	}).Result()
	if err != nil || len(ids) == 0 {
		return err
	}
	pipe := r.client.Pipeline()
	for _, id := range ids {
		pipe.ZRem(ctx, "dispatch:scheduled", id)
	}
	_, err = pipe.Exec(ctx)
	if err != nil {
		return err
	}
	for _, id := range ids {
		job, err := r.GetJob(ctx, id)
		if err != nil {
			continue
		}
		job.Status = models.StatusPending
		if err := r.Enqueue(ctx, job); err != nil {
			continue
		}
	}
	return nil
}

func (r *RedisStore) PauseQueue(ctx context.Context, name string) error {
	return r.client.Set(ctx, fmt.Sprintf(pausedKeyFmt, name), true, 0).Err()
}

func (r *RedisStore) ResumeQueue(ctx context.Context, name string) error {
	return r.client.Del(ctx, fmt.Sprintf(pausedKeyFmt, name)).Err()
}

func (r *RedisStore) IsQueuePaused(ctx context.Context, name string) (bool, error) {
	val, err := r.client.Get(ctx, fmt.Sprintf(pausedKeyFmt, name)).Bool()
	if err == redis.Nil {
		return false, nil
	}
	return val, err
}

func (r *RedisStore) QueueDepths(ctx context.Context, queueName string) (map[string]int64, error) {
	priorities := []struct {
		name  string
		value models.Priority
	}{
		{"CRITICAL", models.PriorityCritical},
		{"HIGH", models.PriorityHigh},
		{"NORMAL", models.PriorityNormal},
		{"LOW", models.PriorityLow},
	}
	result := make(map[string]int64)
	for _, p := range priorities {
		key := fmt.Sprintf(queueKeyFmt, queueName, int(p.value))
		n, _ := r.client.LLen(ctx, key).Result()
		result[p.name] = n
	}
	return result, nil
}

func (r *RedisStore) GetCounters(ctx context.Context) map[string]int64 {
	keys := []string{
		"dispatch:counter:enqueued",
		"dispatch:counter:completed",
		"dispatch:counter:failed",
		"dispatch:counter:dead",
	}
	result := make(map[string]int64)
	for _, k := range keys {
		val, _ := r.client.Get(ctx, k).Int64()
		result[k] = val
	}
	return result
}

func (r *RedisStore) IncrCounter(ctx context.Context, key string) {
	r.client.Incr(ctx, "dispatch:counter:"+key)
}

func (r *RedisStore) RecordLatency(ctx context.Context, ms int64) {
	r.client.LPush(ctx, latencyListKey, ms)
	r.client.LTrim(ctx, latencyListKey, 0, 999)
}

func (r *RedisStore) Publish(ctx context.Context, event *models.WSEvent) {
	data, _ := json.Marshal(event)
	r.client.Publish(ctx, pubsubChannel, data)
}

func (r *RedisStore) Subscribe(ctx context.Context) *redis.PubSub {
	return r.client.Subscribe(ctx, pubsubChannel)
}

func (r *RedisStore) DLQJobs(ctx context.Context) ([]*models.Job, error) {
	ids, err := r.client.LRange(ctx, "dispatch:dlq", 0, 99).Result()
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool)
	jobs := make([]*models.Job, 0, len(ids))
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true
		job, err := r.GetJob(ctx, id)
		if err == nil && job.Status == models.StatusDead {
			jobs = append(jobs, job)
		}
	}
	return jobs, nil
}

func (r *RedisStore) RemoveFromDLQ(ctx context.Context, id string) error {
	return r.client.LRem(ctx, "dispatch:dlq", 0, id).Err()
}

func (r *RedisStore) Client() *redis.Client {
	return r.client
}
