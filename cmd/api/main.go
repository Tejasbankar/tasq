package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Tejasbankar/tasq/internal/config"
	"github.com/Tejasbankar/tasq/internal/queue"
	"github.com/Tejasbankar/tasq/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

	repo := storage.NewTaskRepository(pool)

	router := chi.NewRouter()

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if err := pool.Ping(r.Context()); err != nil {
			log.Printf("failed to connect to db: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"status": "degraded"})
			return
		}

		if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
			log.Printf("failed to encode health response: %v", err)
			return
		}
	})

	router.Post("/tasks", func(w http.ResponseWriter, r *http.Request) {
		var req queue.CreateTaskRequest

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("failed to parse payload: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"status": "failed"})
			return
		}

		if err := req.Validate(); err != nil {
			log.Printf("invalid payload received: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"status": "failed"})
			return
		}

		task := queue.Task{
			ID:         uuid.New(),
			Type:       req.Type,
			Payload:    req.Payload,
			Status:     queue.StatusPending,
			RetryCount: 0,
		}

		if err := repo.Create(r.Context(), task); err != nil {
			log.Printf("could not create task: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"status": "failed"})
			return
		}

		res := queue.CreateTaskResponse{
			ID:     task.ID,
			Status: task.Status,
		}

		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(res); err != nil {
			log.Printf("failed to encode response: %v", err)
			return
		}
	})

	log.Printf("starting api server on :%s", cfg.HTTPPort)

	if err := http.ListenAndServe(":"+cfg.HTTPPort, router); err != nil {
		log.Fatal(err)
	}
}
