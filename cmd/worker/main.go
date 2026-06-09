package main

import (
	"context"
	"log"
	"time"

	"github.com/Tejasbankar/tasq/internal/config"
	"github.com/Tejasbankar/tasq/internal/storage"
)

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

	for {
		task, err := repo.GetPendingTask(context.Background())

		if err != nil {
			log.Fatalf("could not fetch task: %v", err)
		}

		if task == nil {
			time.Sleep(2 * time.Second)
			continue
		}

		log.Printf("found task id: %s status: %s", task.ID, task.Status)
		time.Sleep(2 * time.Second)
	}
}
