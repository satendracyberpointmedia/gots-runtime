package runtime

import (
	"context"
	"sync"
)

// Lifecycle manages the runtime process lifecycle
type Lifecycle struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	state  State
	mu     sync.RWMutex
}

// State represents the runtime state
type State int

const (
	StateInitialized State = iota
	StateRunning
	StateShuttingDown
	StateStopped
)

// NewLifecycle creates a new lifecycle manager
func NewLifecycle() *Lifecycle {
	ctx, cancel := context.WithCancel(context.Background())
	return &Lifecycle{
		ctx:    ctx,
		cancel: cancel,
		state:  StateInitialized,
	}
}

// Start initializes and starts the runtime
func (l *Lifecycle) Start() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.state != StateInitialized {
		return ErrInvalidState
	}

	l.state = StateRunning
	return nil
}

// Stop gracefully shuts down the runtime
func (l *Lifecycle) Stop() error {
	l.mu.Lock()
	if l.state != StateRunning {
		l.mu.Unlock()
		return ErrInvalidState
	}
	l.state = StateShuttingDown
	l.mu.Unlock()

	// Cancel context to signal shutdown
	l.cancel()

	// Wait for all goroutines to finish
	l.wg.Wait()

	l.mu.Lock()
	l.state = StateStopped
	l.mu.Unlock()

	return nil
}

// Context returns the lifecycle context
func (l *Lifecycle) Context() context.Context {
	return l.ctx
}

// State returns the current state
func (l *Lifecycle) State() State {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.state
}

// AddGoroutine tracks a goroutine for graceful shutdown
func (l *Lifecycle) AddGoroutine() {
	l.wg.Add(1)
}

// DoneGoroutine signals a goroutine has finished
func (l *Lifecycle) DoneGoroutine() {
	l.wg.Done()
}

// Errors
var (
	ErrInvalidState = &RuntimeError{Message: "invalid state transition"}
)

// RuntimeError represents a runtime error
type RuntimeError struct {
	Message string
}

func (e *RuntimeError) Error() string {
	return e.Message
}
