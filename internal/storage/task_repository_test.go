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

func TestTaskRepository_ClaimTask(t *testing.T) {
	repo := setupTestRepository(t)
	ctx := context.Background()

	task := queue.Task{
		ID:         uuid.New(),
		Type:       "send_email",
		Payload:    json.RawMessage(`{"recipient":"test@example.com"}`),
		Status:     queue.StatusPending,
		RetryCount: 0,
		RunAt:      time.Now().UTC(),
	}

	if err := repo.Create(ctx, task); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	got, err := repo.ClaimTask(ctx, []string{"send_email"})
	if err != nil {
		t.Fatalf("failed to claim task: %v", err)
	}

	if got == nil {
		t.Fatal("expected task to be claimed, got nil")
	}

	if got.ID != task.ID {
		t.Errorf("expected task ID %s, got %s", task.ID, got.ID)
	}

	if got.Type != task.Type {
		t.Errorf("expected task type %q, got %q", task.Type, got.Type)
	}

	if got.Status != queue.StatusProcessing {
		t.Errorf("expected status %q, got %q", queue.StatusProcessing, got.Status)
	}

	if got.RetryCount != task.RetryCount {
		t.Errorf("expected retry count %d, got %d", task.RetryCount, got.RetryCount)
	}

	if !got.RunAt.Equal(task.RunAt.Truncate(time.Microsecond)) {
		t.Errorf(
			"expected run_at %v, got %v",
			task.RunAt.Truncate(time.Microsecond),
			got.RunAt,
		)
	}

	var status queue.Status

	err = repo.pool.QueryRow(
		ctx,
		"SELECT status FROM tasks WHERE id = $1",
		task.ID,
	).Scan(&status)

	if err != nil {
		t.Fatalf("failed to fetch claimed task: %v", err)
	}

	if status != queue.StatusProcessing {
		t.Errorf("expected persisted status %q, got %q", queue.StatusProcessing, status)
	}
}

func TestTaskRepository_ClaimTask_NoTaskAvailable(t *testing.T) {
	repo := setupTestRepository(t)
	ctx := context.Background()

	got, err := repo.ClaimTask(ctx, []string{"send_email"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if got != nil {
		t.Fatalf("expected nil task, got: %+v", got)
	}
}

func TestTaskRepository_ClaimTask_IgnoresFutureTask(t *testing.T) {
	repo := setupTestRepository(t)
	ctx := context.Background()

	task := queue.Task{
		ID:         uuid.New(),
		Type:       "send_email",
		Payload:    json.RawMessage(`{"recipient":"test@example.com"}`),
		Status:     queue.StatusPending,
		RetryCount: 0,
		RunAt:      time.Now().UTC().Add(time.Hour),
	}

	if err := repo.Create(ctx, task); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	got, err := repo.ClaimTask(ctx, []string{"send_email"})
	if err != nil {
		t.Fatalf("failed to claim task: %v", err)
	}

	if got != nil {
		t.Fatalf("expected no task, got: %+v", got)
	}
}

func TestTaskRepository_ClaimTask_RespectsSupportedTypes(t *testing.T) {
	repo := setupTestRepository(t)
	ctx := context.Background()

	task := queue.Task{
		ID:         uuid.New(),
		Type:       "generate_report",
		Payload:    json.RawMessage(`{}`),
		Status:     queue.StatusPending,
		RetryCount: 0,
		RunAt:      time.Now().UTC(),
	}

	if err := repo.Create(ctx, task); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	got, err := repo.ClaimTask(ctx, []string{"send_email"})
	if err != nil {
		t.Fatalf("failed to claim task: %v", err)
	}

	if got != nil {
		t.Fatalf("expected no task, got: %+v", got)
	}
}

func TestTaskRepository_ClaimTask_ClaimsEarliestRunnableTask(t *testing.T) {
	repo := setupTestRepository(t)
	ctx := context.Background()

	now := time.Now().UTC()

	laterTask := queue.Task{
		ID:         uuid.New(),
		Type:       "send_email",
		Payload:    json.RawMessage(`{}`),
		Status:     queue.StatusPending,
		RetryCount: 0,
		RunAt:      now.Add(-5 * time.Minute),
	}

	earlierTask := queue.Task{
		ID:         uuid.New(),
		Type:       "send_email",
		Payload:    json.RawMessage(`{}`),
		Status:     queue.StatusPending,
		RetryCount: 0,
		RunAt:      now.Add(-10 * time.Minute),
	}

	if err := repo.Create(ctx, laterTask); err != nil {
		t.Fatalf("failed to create later task: %v", err)
	}

	if err := repo.Create(ctx, earlierTask); err != nil {
		t.Fatalf("failed to create earlier task: %v", err)
	}

	got, err := repo.ClaimTask(ctx, []string{"send_email"})
	if err != nil {
		t.Fatalf("failed to claim task: %v", err)
	}

	if got == nil {
		t.Fatal("expected task, got nil")
	}

	if got.ID != earlierTask.ID {
		t.Errorf("expected task %s, got %s", earlierTask.ID, got.ID)
	}
}

func TestTaskRepository_ClaimTask_RequiresSupportedTypes(t *testing.T) {
	repo := setupTestRepository(t)
	ctx := context.Background()

	got, err := repo.ClaimTask(ctx, []string{})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if got != nil {
		t.Fatalf("expected nil task, got: %+v", got)
	}
}

func TestTaskRepository_MarkCompleted(t *testing.T) {
	repo := setupTestRepository(t)
	ctx := context.Background()

	task := queue.Task{
		ID:         uuid.New(),
		Type:       "send_email",
		Payload:    json.RawMessage(`{}`),
		Status:     queue.StatusPending,
		RetryCount: 0,
		RunAt:      time.Now().UTC(),
	}

	if err := repo.Create(ctx, task); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	if _, err := repo.ClaimTask(ctx, []string{"send_email"}); err != nil {
		t.Fatalf("failed to claim task: %v", err)
	}

	if err := repo.MarkCompleted(ctx, task.ID); err != nil {
		t.Fatalf("failed to mark task completed: %v", err)
	}

	var status queue.Status

	if err := repo.pool.QueryRow(
		ctx,
		"SELECT status FROM tasks WHERE id = $1",
		task.ID,
	).Scan(&status); err != nil {
		t.Fatalf("failed to fetch task: %v", err)
	}

	if status != queue.StatusCompleted {
		t.Errorf("expected status %q, got %q", queue.StatusCompleted, status)
	}
}

func TestTaskRepository_MarkFailed(t *testing.T) {
	repo := setupTestRepository(t)
	ctx := context.Background()

	task := queue.Task{
		ID:         uuid.New(),
		Type:       "send_email",
		Payload:    json.RawMessage(`{}`),
		Status:     queue.StatusPending,
		RetryCount: 0,
		RunAt:      time.Now().UTC(),
	}

	if err := repo.Create(ctx, task); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	if _, err := repo.ClaimTask(ctx, []string{"send_email"}); err != nil {
		t.Fatalf("failed to claim task: %v", err)
	}

	if err := repo.MarkFailed(ctx, task.ID); err != nil {
		t.Fatalf("failed to mark task failed: %v", err)
	}

	var status queue.Status

	if err := repo.pool.QueryRow(
		ctx,
		"SELECT status FROM tasks WHERE id = $1",
		task.ID,
	).Scan(&status); err != nil {
		t.Fatalf("failed to fetch task: %v", err)
	}

	if status != queue.StatusFailed {
		t.Errorf("expected status %q, got %q", queue.StatusFailed, status)
	}
}

func TestTaskRepository_RetryTask(t *testing.T) {
	repo := setupTestRepository(t)
	ctx := context.Background()

	task := queue.Task{
		ID:         uuid.New(),
		Type:       "send_email",
		Payload:    json.RawMessage(`{}`),
		Status:     queue.StatusPending,
		RetryCount: 0,
		RunAt:      time.Now().UTC(),
	}

	if err := repo.Create(ctx, task); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	if _, err := repo.ClaimTask(ctx, []string{"send_email"}); err != nil {
		t.Fatalf("failed to claim task: %v", err)
	}

	if err := repo.RetryTask(ctx, task.ID); err != nil {
		t.Fatalf("failed to retry task: %v", err)
	}

	var (
		status     queue.Status
		retryCount int
	)

	if err := repo.pool.QueryRow(
		ctx,
		"SELECT status, retry_count FROM tasks WHERE id = $1",
		task.ID,
	).Scan(&status, &retryCount); err != nil {
		t.Fatalf("failed to fetch task: %v", err)
	}

	if status != queue.StatusPending {
		t.Errorf("expected status %q, got %q", queue.StatusPending, status)
	}

	if retryCount != 1 {
		t.Errorf("expected retry count 1, got %d", retryCount)
	}
}
