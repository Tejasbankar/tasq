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

func (r *TaskRepository) ClaimTask(ctx context.Context) (*queue.Task, error) {
	tx, err := r.pool.Begin(ctx)
	task := &queue.Task{}

	if err != nil {
		return task, err
	}

	defer tx.Rollback(ctx)

	const fetchQuery = `
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
	ORDER BY created_at
	FOR UPDATE SKIP LOCKED
	LIMIT 1
	`

	row := tx.QueryRow(ctx, fetchQuery)
	err = row.Scan(&task.ID, &task.Type, &task.Payload, &task.Status, &task.RetryCount, &task.CreatedAt, &task.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	const updateQuery = `
	Update tasks
	SET status='processing',
	updated_at=NOW()
	WHERE id=$1
	`
	_, err = tx.Exec(ctx, updateQuery, task.ID)

	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)

	if err != nil {
		return nil, err
	}

	task.Status = queue.StatusProcessing

	return task, err
}

func (r *TaskRepository) MarkCompleted(ctx context.Context, task queue.Task) error {
	const query = `Update tasks SET status='completed', updated_at=NOW() WHERE id=$1`

	if _, err := r.pool.Exec(ctx, query, task.ID); err != nil {
		return err
	}

	return nil
}
