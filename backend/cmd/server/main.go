package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gautammehendale/dispatch/internal/api"
	"github.com/gautammehendale/dispatch/internal/models"
	"github.com/gautammehendale/dispatch/internal/queue"
	"github.com/gautammehendale/dispatch/internal/scheduler"
	"github.com/gautammehendale/dispatch/internal/store"
	"github.com/gautammehendale/dispatch/internal/worker"
)

func main() {
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	postgresDSN := getEnv("POSTGRES_DSN", "postgres://dispatch:dispatch@localhost:5432/dispatch?sslmode=disable")
	port := getEnv("PORT", "8080")
	concurrency := 4

	log.Println("[dispatch] starting server...")

	redisStore, err := store.NewRedisStore(redisAddr)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	log.Println("[dispatch] redis connected")

	pgStore, err := store.NewPostgresStore(postgresDSN)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	log.Println("[dispatch] postgres connected, migrations applied")

	hub := api.NewHub()
	engine := queue.NewEngine(redisStore, pgStore, hub)

	pool := worker.NewPool(concurrency, "default", engine, redisStore)

	// Register demo handlers — in production, these come from your application code.
	pool.Register("send_email", func(ctx context.Context, job *models.Job) error {
		log.Printf("[handler] send_email: to=%v", job.Payload["to"])
		time.Sleep(50 * time.Millisecond)
		return nil
	})
	pool.Register("process_payment", func(ctx context.Context, job *models.Job) error {
		log.Printf("[handler] process_payment: amount=%v", job.Payload["amount"])
		time.Sleep(100 * time.Millisecond)
		return nil
	})
	pool.Register("resize_image", func(ctx context.Context, job *models.Job) error {
		log.Printf("[handler] resize_image: url=%v", job.Payload["url"])
		time.Sleep(200 * time.Millisecond)
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool.Start(ctx)
	go scheduler.New(redisStore, "default").Run(ctx)

	h := api.NewHandler(engine, redisStore, pgStore, pool)
	router := api.NewRouter(h, hub)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("[dispatch] listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("[dispatch] shutting down gracefully...")

	cancel()
	pool.Stop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("[dispatch] forced shutdown: %v", err)
	}
	log.Println("[dispatch] stopped")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
