-- +goose Up
CREATE TABLE tasks (
  id uuid PRIMARY KEY,
  type TEXT NOT NULL,
  payload JSONB NOT NULL,

  status TEXT NOT NULL CHECK(
    status IN (
      'pending',
      'processing',
      'completed',
      'failed'
    )
  ),

  retry_count INTEGER NOT NULL DEFAULT 0,

  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE tasks;
