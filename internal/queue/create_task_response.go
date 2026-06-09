package queue

import (
	"github.com/google/uuid"
)

type CreateTaskResponse struct {
	ID uuid.UUID `json:"id"`
	Status Status `json:"status"`
}
