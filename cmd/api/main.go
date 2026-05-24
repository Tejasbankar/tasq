package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Tejasbankar/tasq/internal/config"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg, err := config.Load()

	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	router := chi.NewRouter()

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
			log.Printf("failed to encode health response: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	log.Printf("starting api server on :%s", cfg.HTTPPort)

	if err := http.ListenAndServe(":"+cfg.HTTPPort, router); err != nil {
		log.Fatal(err)
	}
}
