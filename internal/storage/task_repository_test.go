package storage

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/Tejasbankar/tasq/internal/config"
	"github.com/Tejasbankar/tasq/internal/queue"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func setupTestRepository(t *testing.T) *TaskRepository {
	t.Helper()

	if err := godotenv.Overload("../../.env.test"); err != nil {
		t.Fatalf("failed to load test environment: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.DBName != "tasq_test" {
		t.Fatalf("tests must use tasq_test database, got %q", cfg.DBName)
	}

	pool, err := NewPostgres(cfg)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
	})

	_, err = pool.Exec(context.Background(), "TRUNCATE TABLE tasks")
	if err != nil {
		t.Fatalf("failed to clean test database: %v", err)
	}

	return NewTaskRepository(pool)
}

func TestTaskRepository_Create(t *testing.T) {
	repo := setupTestRepository(t)
	ctx := context.Background()

	taskID := uuid.New()
	payload := json.RawMessage(`{"recipient":"test@example.com"}`)
	runAt := time.Now().UTC().Add(5 * time.Minute)

	task := queue.Task{
		ID:         taskID,
		Type:       "send_email",
		Payload:    payload,
		Status:     queue.StatusPending,
		RetryCount: 0,
		RunAt:      runAt,
	}

	if err := repo.Create(ctx, task); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	var (
		gotID         uuid.UUID
		gotType       string
		gotPayload    []byte
		gotStatus     string
		gotRetryCount int
		gotRunAt      time.Time
	)

	const query = `
		SELECT
			id,
			type,
			payload,
			status,
			retry_count,
			run_at
		FROM tasks
		WHERE id = $1
	`

	err := repo.pool.QueryRow(ctx, query, taskID).Scan(
		&gotID,
		&gotType,
		&gotPayload,
		&gotStatus,
		&gotRetryCount,
		&gotRunAt,
	)

	if err != nil {
		t.Fatalf("failed to fetch created task: %v", err)
	}

	if gotID != task.ID {
		t.Errorf("expected id %s, got %s", task.ID, gotID)
	}

	if gotType != task.Type {
		t.Errorf("expected type %q, got %q", task.Type, gotType)
	}

	var expectedPayload any
	var actualPayload any

	if err := json.Unmarshal(task.Payload, &expectedPayload); err != nil {
		t.Fatalf("failed to decode expected payload: %v", err)
	}

	if err := json.Unmarshal(gotPayload, &actualPayload); err != nil {
		t.Fatalf("failed to decode actual payload: %v", err)
	}

	if !reflect.DeepEqual(expectedPayload, actualPayload) {
		t.Errorf("expected payload %v, got %v", expectedPayload, actualPayload)
	}
	if gotStatus != string(task.Status) {
		t.Errorf("expected status %q, got %q", task.Status, gotStatus)
	}

	if gotRetryCount != task.RetryCount {
		t.Errorf("expected retry count %d, got %d", task.RetryCount, gotRetryCount)
	}

	expectedRunAt := runAt.Truncate(time.Microsecond)

	if !gotRunAt.Equal(expectedRunAt) {
		t.Errorf("expected run_at %v, got %v", expectedRunAt, gotRunAt)
	}
}
