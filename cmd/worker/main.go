package main

import (
	"log"

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

	_ = storage.NewTaskRepository(pool)

	log.Print("worker started")
}
