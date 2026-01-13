package runtime

import (
	"context"
	"fmt"
	"sync"
)

// Orchestrator is the main runtime orchestrator that manages the entire runtime lifecycle
type Orchestrator struct {
	lifecycle *Lifecycle
	scheduler Scheduler
	mu        sync.RWMutex
}

// Scheduler interface for task scheduling
type Scheduler interface {
	Schedule(task Task) error
	Shutdown() error
}

// Task represents a unit of work
type Task interface {
	Execute(ctx context.Context) error
	IsCPUIntensive() bool
}

// NewOrchestrator creates a new runtime orchestrator
func NewOrchestrator() *Orchestrator {
	return &Orchestrator{
		lifecycle: NewLifecycle(),
	}
}

// SetScheduler sets the task scheduler
func (o *Orchestrator) SetScheduler(scheduler Scheduler) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.scheduler = scheduler
}

// Start initializes and starts the runtime
func (o *Orchestrator) Start() error {
	if err := o.lifecycle.Start(); err != nil {
		return fmt.Errorf("failed to start lifecycle: %w", err)
	}
	return nil
}

// Stop gracefully shuts down the runtime
func (o *Orchestrator) Stop() error {
	o.mu.Lock()
	if o.scheduler != nil {
		if err := o.scheduler.Shutdown(); err != nil {
			o.mu.Unlock()
			return fmt.Errorf("failed to shutdown scheduler: %w", err)
		}
	}
	o.mu.Unlock()

	if err := o.lifecycle.Stop(); err != nil {
		return fmt.Errorf("failed to stop lifecycle: %w", err)
	}
	return nil
}

// ExecuteTask schedules a task for execution
func (o *Orchestrator) ExecuteTask(task Task) error {
	o.mu.RLock()
	scheduler := o.scheduler
	o.mu.RUnlock()

	if scheduler == nil {
		// Fallback to direct execution if no scheduler
		o.lifecycle.AddGoroutine()
		go func() {
			defer o.lifecycle.DoneGoroutine()
			_ = task.Execute(o.lifecycle.Context())
		}()
		return nil
	}

	return scheduler.Schedule(task)
}

// Context returns the orchestrator context
func (o *Orchestrator) Context() context.Context {
	return o.lifecycle.Context()
}

// State returns the current runtime state
func (o *Orchestrator) State() State {
	return o.lifecycle.State()
}

// Lifecycle returns the lifecycle manager
func (o *Orchestrator) Lifecycle() *Lifecycle {
	return o.lifecycle
}

