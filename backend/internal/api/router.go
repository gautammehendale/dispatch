package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(h *Handler, hub *Hub) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	}))

	r.Get("/health", h.Health)
	r.Get("/ws", hub.ServeWS)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/jobs", func(r chi.Router) {
			r.Post("/", h.EnqueueJob)
			r.Get("/", h.ListJobs)
			r.Get("/{id}", h.GetJob)
			r.Delete("/{id}", h.CancelJob)
			r.Post("/{id}/cancel", h.CancelJob)
			r.Post("/{id}/retry", h.RetryJob)
		})

		r.Route("/queues", func(r chi.Router) {
			r.Get("/", h.ListQueues)
			r.Post("/{name}/pause", h.PauseQueue)
			r.Post("/{name}/resume", h.ResumeQueue)
		})

		r.Get("/workers", h.GetWorkers)
		r.Get("/metrics", h.GetMetrics)
		r.Get("/dlq", h.GetDLQ)
	})

	return r
}
