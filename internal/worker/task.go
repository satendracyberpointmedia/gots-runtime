package worker

import (
	"context"
	"time"
)

// Task represents a unit of work
type Task struct {
	ID            string
	Handler       func(ctx context.Context) error
	IsCPUIntensive bool
	Priority      int
	CreatedAt     time.Time
}

// NewTask creates a new task
func NewTask(id string, handler func(ctx context.Context) error, isCPUIntensive bool, priority int) *Task {
	return &Task{
		ID:             id,
		Handler:        handler,
		IsCPUIntensive: isCPUIntensive,
		Priority:       priority,
		CreatedAt:      time.Now(),
	}
}

// Execute executes the task
func (t *Task) Execute(ctx context.Context) error {
	if t.Handler == nil {
		return nil
	}
	return t.Handler(ctx)
}

// TaskResult represents the result of a task execution
type TaskResult struct {
	TaskID  string
	Error   error
	Duration time.Duration
}

