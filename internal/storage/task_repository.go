package storage

import (
	"context"

	"github.com/Tejasbankar/tasq/internal/queue"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskRepository struct {
	pool *pgxpool.Pool
}

func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{
		pool: pool,
	}
}

func (r *TaskRepository) Create(ctx context.Context, task queue.Task) error {
	const query = "INSERT INTO tasks(id, type, payload, status, retry_count) VALUES($1, $2, $3, $4, $5)"

	if _, err := r.pool.Exec(
		ctx,
		query,
		task.ID,
		task.Type,
		task.Payload,
		task.Status,
		task.RetryCount,
	); err != nil {
		return err
	}

	return nil
}
