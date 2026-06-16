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
		task, err := repo.ClaimTask(context.Background())

		if err != nil {
			log.Fatalf("could not claim a task: %v", err)
		}

		if task == nil {
			time.Sleep(2 * time.Second)
			continue
		}

		log.Printf("claimed task with id: %s", task.ID)

		if err := repo.MarkCompleted(context.Background(), *task); err != nil {
			log.Fatalf("could not update task status: %v", err)
		}

		log.Printf("marked task %s as completed", task.ID)

		time.Sleep(2 * time.Second)
	}
}
