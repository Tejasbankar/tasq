package worker

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/Tejasbankar/tasq/internal/queue"
	"github.com/Tejasbankar/tasq/internal/storage"
)

type Worker struct {
	repo     *storage.TaskRepository
	registry *Registry

	pollInterval time.Duration
}

const maxRetries = 3

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

func handleFailure(ctx context.Context, repo *storage.TaskRepository, task queue.Task) {
	if task.RetryCount >= maxRetries {
		if err := repo.MarkFailed(ctx, task.ID); err != nil {
			log.Printf("could not mark task %s as failed: %v", task.ID, err)
		}
	} else {
		if err := repo.RetryTask(ctx, task.ID); err != nil {
			log.Printf("could not mark task %s as pending: %v", task.ID, err)
		}
	}
}

func (w *Worker) Start(ctx context.Context) error {
	supportedTypes := w.registry.SupportedTypes()
	if len(supportedTypes) == 0 {
		return errors.New("at least one handler must be registered")
	}

	for {
		task, err := w.repo.ClaimTask(ctx, supportedTypes)

		if err != nil {
			log.Printf("could not claim a task: %v", err)
			time.Sleep(w.pollInterval)
			continue
		}

		if task == nil {
			time.Sleep(w.pollInterval)
			continue
		}

		log.Printf("claimed task with id: %s", task.ID)

		handler, ok := w.registry.Get(task.Type)
		if !ok {
			log.Printf("could not find a suitable handler for task %s type %q", task.ID, task.Type)

			handleFailure(ctx, w.repo, *task)

			time.Sleep(w.pollInterval)
			continue
		}

		if err := handler(ctx, *task); err != nil {
			log.Printf("task %s failed: %v", task.ID, err)

			handleFailure(ctx, w.repo, *task)

			time.Sleep(w.pollInterval)
			continue
		}

		if err := w.repo.MarkCompleted(ctx, task.ID); err != nil {
			log.Printf("could not mark task %s as completed: %v", task.ID, err)
			time.Sleep(w.pollInterval)
			continue
		}

		log.Printf("marked task %s as completed", task.ID)

		time.Sleep(w.pollInterval)
	}

}
