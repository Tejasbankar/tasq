package worker

import (
	"context"

	"github.com/Tejasbankar/tasq/internal/queue"
)

type Handler func(context.Context, queue.Task) error
