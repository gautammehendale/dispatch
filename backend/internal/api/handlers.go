package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/gautammehendale/dispatch/internal/models"
	"github.com/gautammehendale/dispatch/internal/queue"
	"github.com/gautammehendale/dispatch/internal/store"
	"github.com/gautammehendale/dispatch/internal/worker"
)

type Handler struct {
	engine   *queue.Engine
	redis    *store.RedisStore
	postgres *store.PostgresStore
	pool     *worker.Pool
}

func NewHandler(engine *queue.Engine, redis *store.RedisStore, postgres *store.PostgresStore, pool *worker.Pool) *Handler {
	return &Handler{engine: engine, redis: redis, postgres: postgres, pool: pool}
}

func (h *Handler) EnqueueJob(w http.ResponseWriter, r *http.Request) {
	var req models.EnqueueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Type == "" {
		writeError(w, http.StatusBadRequest, "job type is required")
		return
	}
	job, err := h.engine.Enqueue(r.Context(), &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, job)
}

func (h *Handler) ListJobs(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	queue := r.URL.Query().Get("queue")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit == 0 {
		limit = 50
	}
	jobs, total, err := h.postgres.ListJobs(r.Context(), status, queue, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"jobs": jobs, "total": total})
}

func (h *Handler) GetJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	job, err := h.postgres.GetJob(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "job not found")
		return
	}
	execs, _ := h.postgres.GetExecutions(r.Context(), id)
	writeJSON(w, http.StatusOK, map[string]any{"job": job, "executions": execs})
}

func (h *Handler) CancelJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.engine.CancelJob(r.Context(), id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

func (h *Handler) RetryJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	job, err := h.postgres.GetJob(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "job not found")
		return
	}
	job.Status = models.StatusPending
	job.Attempts = 0
	job.Error = ""
	_ = h.redis.RemoveFromDLQ(r.Context(), id)
	if err := h.redis.Enqueue(r.Context(), job); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func (h *Handler) ListQueues(w http.ResponseWriter, r *http.Request) {
	queues := []string{"default", "email", "notifications"}
	stats := make([]models.QueueStats, 0)
	for _, q := range queues {
		depths, _ := h.redis.QueueDepths(r.Context(), q)
		var total int64
		for _, d := range depths {
			total += d
		}
		stats = append(stats, models.QueueStats{Name: q, Depth: total})
	}
	writeJSON(w, http.StatusOK, stats)
}

func (h *Handler) PauseQueue(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	_ = h.redis.PauseQueue(r.Context(), name)
	writeJSON(w, http.StatusOK, map[string]string{"status": "paused"})
}

func (h *Handler) ResumeQueue(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	_ = h.redis.ResumeQueue(r.Context(), name)
	writeJSON(w, http.StatusOK, map[string]string{"status": "resumed"})
}

func (h *Handler) GetWorkers(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.pool.Workers())
}

func (h *Handler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	counters := h.redis.GetCounters(ctx)
	workers := h.pool.Workers()
	dlqJobs, _ := h.redis.DLQJobs(ctx)

	active := 0
	for _, w := range workers {
		if w.Status == "busy" {
			active++
		}
	}

	metrics := models.Metrics{
		TotalEnqueued:  counters["dispatch:counter:enqueued"],
		TotalCompleted: counters["dispatch:counter:completed"],
		TotalFailed:    counters["dispatch:counter:failed"],
		TotalDead:      int64(len(dlqJobs)),
		ActiveWorkers:  active,
		WorkerStats:    make([]models.WorkerStatus, len(workers)),
	}
	for i, w := range workers {
		metrics.WorkerStats[i] = *w
	}
	writeJSON(w, http.StatusOK, metrics)
}

func (h *Handler) GetDLQ(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.redis.DLQJobs(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"jobs": jobs, "total": len(jobs)})
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}
