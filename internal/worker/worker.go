package worker

import (
	"errors"
	"time"

	"github.com/Tejasbankar/tasq/internal/storage"
)

type Worker struct {
	repo     *storage.TaskRepository
	registry *Registry

	pollInterval time.Duration
}

func New(repo *storage.TaskRepository) (*Worker, error) {
	if repo == nil {
		return nil, errors.New("task repository is required")
	}

	return &Worker{
		repo:         repo,
		registry:     NewRegistry(),
		pollInterval: 2 * time.Second,
	}, nil
}

func (w *Worker) Register(taskType string, handler Handler) error {
	return w.registry.Register(taskType, handler)
}
