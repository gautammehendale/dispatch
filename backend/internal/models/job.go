package models

import (
	"time"
)

type Priority int

const (
	PriorityCritical Priority = 4
	PriorityHigh     Priority = 3
	PriorityNormal   Priority = 2
	PriorityLow      Priority = 1
)

func (p Priority) String() string {
	switch p {
	case PriorityCritical:
		return "CRITICAL"
	case PriorityHigh:
		return "HIGH"
	case PriorityNormal:
		return "NORMAL"
	case PriorityLow:
		return "LOW"
	default:
		return "NORMAL"
	}
}

func ParsePriority(s string) Priority {
	switch s {
	case "CRITICAL":
		return PriorityCritical
	case "HIGH":
		return PriorityHigh
	case "LOW":
		return PriorityLow
	default:
		return PriorityNormal
	}
}

type JobStatus string

const (
	StatusPending   JobStatus = "pending"
	StatusRunning   JobStatus = "running"
	StatusCompleted JobStatus = "completed"
	StatusFailed    JobStatus = "failed"
	StatusRetrying  JobStatus = "retrying"
	StatusDead      JobStatus = "dead"
	StatusCancelled JobStatus = "cancelled"
)

type Job struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Payload     map[string]any    `json:"payload"`
	Priority    Priority          `json:"priority"`
	Status      JobStatus         `json:"status"`
	Queue       string            `json:"queue"`
	MaxRetries  int               `json:"max_retries"`
	Attempts    int               `json:"attempts"`
	RunAt       time.Time         `json:"run_at"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	StartedAt   *time.Time        `json:"started_at,omitempty"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
	Error       string            `json:"error,omitempty"`
	WorkerID    string            `json:"worker_id,omitempty"`
	Meta        map[string]string `json:"meta,omitempty"`
}

type JobExecution struct {
	ID        string    `json:"id"`
	JobID     string    `json:"job_id"`
	WorkerID  string    `json:"worker_id"`
	Attempt   int       `json:"attempt"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at"`
	DurationMs int64    `json:"duration_ms"`
	Status    JobStatus `json:"status"`
	Error     string    `json:"error,omitempty"`
}

type EnqueueRequest struct {
	Type       string            `json:"type"`
	Payload    map[string]any    `json:"payload"`
	Priority   string            `json:"priority"`
	Queue      string            `json:"queue"`
	MaxRetries int               `json:"max_retries"`
	RunAt      *time.Time        `json:"run_at,omitempty"`
	Meta       map[string]string `json:"meta,omitempty"`
}

type QueueStats struct {
	Name    string `json:"name"`
	Depth   int64  `json:"depth"`
	Paused  bool   `json:"paused"`
}

type Metrics struct {
	TotalEnqueued  int64          `json:"total_enqueued"`
	TotalCompleted int64          `json:"total_completed"`
	TotalFailed    int64          `json:"total_failed"`
	TotalDead      int64          `json:"total_dead"`
	ActiveWorkers  int            `json:"active_workers"`
	Throughput     float64        `json:"throughput_per_sec"`
	AvgLatencyMs   float64        `json:"avg_latency_ms"`
	P99LatencyMs   float64        `json:"p99_latency_ms"`
	QueueStats     []QueueStats   `json:"queues"`
	WorkerStats    []WorkerStatus `json:"workers"`
}

type WorkerStatus struct {
	ID         string     `json:"id"`
	Status     string     `json:"status"`
	CurrentJob string     `json:"current_job,omitempty"`
	JobsRun    int64      `json:"jobs_run"`
	StartedAt  time.Time  `json:"started_at"`
	LastBeatAt time.Time  `json:"last_beat_at"`
}

type WSEvent struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}
