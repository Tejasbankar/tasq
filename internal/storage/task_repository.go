package storage

import (
	"context"
	"errors"

	"github.com/Tejasbankar/tasq/internal/queue"
	"github.com/google/uuid"
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

func (r *TaskRepository) ClaimTask(ctx context.Context, supportedTypes []string) (*queue.Task, error) {
	if len(supportedTypes) == 0 {
		return nil, errors.New("supported types are required to claim task")
	}

	tx, err := r.pool.Begin(ctx)

	if err != nil {
		return nil, err
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
	AND type = ANY($1)
	ORDER BY created_at
	FOR UPDATE SKIP LOCKED
	LIMIT 1
	`

	task := &queue.Task{}
	row := tx.QueryRow(ctx, fetchQuery, supportedTypes)
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

func (r *TaskRepository) MarkCompleted(ctx context.Context, taskID uuid.UUID) error {
	const query = `Update tasks SET status='completed', updated_at=NOW() WHERE id=$1`

	if _, err := r.pool.Exec(ctx, query, taskID); err != nil {
		return err
	}

	return nil
}

func (r *TaskRepository) MarkFailed(ctx context.Context, taskID uuid.UUID) error {
	const query = `Update tasks SET status='failed', updated_at=NOW() WHERE id=$1`

	if _, err := r.pool.Exec(ctx, query, taskID); err != nil {
		return err
	}

	return nil
}

func (r *TaskRepository) RetryTask(ctx context.Context, taskID uuid.UUID) error {
	const query = `Update tasks SET status='pending', retry_count=retry_count + 1, updated_at=NOW() WHERE id=$1`

	if _, err := r.pool.Exec(ctx, query, taskID); err != nil {
		return err
	}

	return nil
}
