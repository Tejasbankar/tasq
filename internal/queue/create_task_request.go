package queue

import (
	"encoding/json"
	"errors"
	"time"
)

type CreateTaskRequest struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
	RunAt   *time.Time      `json:"run_at,omitempty"`
}

func (r CreateTaskRequest) Validate() error {
	if r.Type == "" {
		return errors.New("type is required")
	}

	if len(r.Payload) == 0 {
		return errors.New("payload is required")
	}

	return nil
}
