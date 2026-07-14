package worker

import "fmt"

type Registry struct {
	handlers map[string]Handler
}

func NewRegistry() *Registry {
	return &Registry{
		handlers: map[string]Handler{},
	}
}

func (r *Registry) Register(taskType string, handler Handler) error {
	if _, ok := r.handlers[taskType]; ok {
		return fmt.Errorf("handler already registered for task type %q", taskType)
	}

	r.handlers[taskType] = handler
	return nil
}

func (r *Registry) Get(taskType string) (Handler, bool) {
	h, ok := r.handlers[taskType]
	return h, ok
}

func (r *Registry) SupportedTypes() []string {
	keys := make([]string, 0, len(r.handlers))

	for k := range r.handlers {
		keys = append(keys, k)
	}

	return keys
}
