package main

import (
	"context"
	"log"

	"github.com/Tejasbankar/tasq/internal/config"
	"github.com/Tejasbankar/tasq/internal/queue"
	"github.com/Tejasbankar/tasq/internal/storage"
	"github.com/Tejasbankar/tasq/internal/worker"
)

func SendEmail(ctx context.Context, t queue.Task) error {
	return nil
}

func main() {
	cfg, err := config.Load()

	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	pool, err := storage.NewPostgres(cfg)

	if err != nil {
		log.Fatalf("Failed to initiate database connection: %v", err)
	}

	defer pool.Close()

	repo := storage.NewTaskRepository(pool)

	w, err := worker.New(repo)
	if err != nil {
		log.Fatalf("Failed to create a worker: %v", err)
	}

	if err := w.Register("send_email", SendEmail); err != nil {
		log.Fatalf("Failed to register send_email handler: %v", err)
	}

	w.Start(context.Background())
}
