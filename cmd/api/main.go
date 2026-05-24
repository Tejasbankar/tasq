package main

import (
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

	log.Printf("starting api server on :%s", cfg.HTTPPort)

	if err := http.ListenAndServe(":"+cfg.HTTPPort, router); err != nil {
		log.Fatal(err)
	}
}
