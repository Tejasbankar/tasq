-- +goose Up
ALTER TABLE tasks
ADD COLUMN run_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- +goose Down
ALTER TABLE tasks
DROP COLUMN run_at;
