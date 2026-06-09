package storage

import (
	"context"
	"errors"

	"github.com/Tejasbankar/tasq/internal/queue"
	"github.com/jackc/pgx/v5"
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

func (r *TaskRepository) GetPendingTask(ctx context.Context) (*queue.Task, error) {
	const query = `
	SELECT
		id,
		type,
		payload,
		status,
		retry_count,
		created_at,
		updated_at
	FROM tasks
	WHERE status='pending'
	LIMIT 1
	`

	row := r.pool.QueryRow(ctx, query)
	task := &queue.Task{}
	err := row.Scan(&task.ID, &task.Type, &task.Payload, &task.Status, &task.RetryCount, &task.CreatedAt, &task.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return &queue.Task{}, err
	}

	return task, nil
}
