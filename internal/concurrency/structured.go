package concurrency

import (
	"context"
	"fmt"
	"sync"
)

// TaskGroup represents a group of related tasks
type TaskGroup struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	errCh  chan error
	mu     sync.Mutex
	errors []error
}

// NewTaskGroup creates a new task group
func NewTaskGroup(ctx context.Context) *TaskGroup {
	taskCtx, cancel := context.WithCancel(ctx)
	return &TaskGroup{
		ctx:    taskCtx,
		cancel: cancel,
		errCh:  make(chan error, 10),
	}
}

// Go runs a function in the task group
func (tg *TaskGroup) Go(fn func() error) {
	tg.wg.Add(1)
	go func() {
		defer tg.wg.Done()
		
		if err := fn(); err != nil {
			select {
			case tg.errCh <- err:
			default:
			}
		}
	}()
}

// Wait waits for all tasks to complete
func (tg *TaskGroup) Wait() error {
	done := make(chan struct{})
	go func() {
		tg.wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		// Collect all errors
		close(tg.errCh)
		for err := range tg.errCh {
			tg.mu.Lock()
			tg.errors = append(tg.errors, err)
			tg.mu.Unlock()
		}
		
		if len(tg.errors) > 0 {
			return fmt.Errorf("task group errors: %v", tg.errors)
		}
		return nil
	case <-tg.ctx.Done():
		tg.cancel()
		return tg.ctx.Err()
	}
}

// Cancel cancels all tasks
func (tg *TaskGroup) Cancel() {
	tg.cancel()
}

// Context returns the task group context
func (tg *TaskGroup) Context() context.Context {
	return tg.ctx
}

// Supervisor supervises a group of workers
type Supervisor struct {
	workers []Worker
	mu      sync.RWMutex
}

// Worker represents a supervised worker
type Worker interface {
	Start(ctx context.Context) error
	Stop() error
	Name() string
}

// NewSupervisor creates a new supervisor
func NewSupervisor() *Supervisor {
	return &Supervisor{
		workers: make([]Worker, 0),
	}
}

// AddWorker adds a worker
func (s *Supervisor) AddWorker(worker Worker) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.workers = append(s.workers, worker)
}

// Start starts all workers
func (s *Supervisor) Start(ctx context.Context) error {
	s.mu.RLock()
	workers := make([]Worker, len(s.workers))
	copy(workers, s.workers)
	s.mu.RUnlock()
	
	for _, worker := range workers {
		if err := worker.Start(ctx); err != nil {
			return fmt.Errorf("failed to start worker %s: %w", worker.Name(), err)
		}
	}
	
	return nil
}

// Stop stops all workers
func (s *Supervisor) Stop() error {
	s.mu.RLock()
	workers := make([]Worker, len(s.workers))
	copy(workers, s.workers)
	s.mu.RUnlock()
	
	var firstErr error
	for _, worker := range workers {
		if err := worker.Stop(); err != nil {
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	
	return firstErr
}

// Scope represents a concurrency scope
type Scope struct {
	ctx    context.Context
	cancel context.CancelFunc
	tasks  []Task
	mu     sync.Mutex
}

// Task represents a task in a scope
type Task struct {
	Name string
	Fn   func(context.Context) error
}

// NewScope creates a new scope
func NewScope(ctx context.Context) *Scope {
	scopeCtx, cancel := context.WithCancel(ctx)
	return &Scope{
		ctx:    scopeCtx,
		cancel: cancel,
		tasks:  make([]Task, 0),
	}
}

// Spawn spawns a task in the scope
func (sc *Scope) Spawn(name string, fn func(context.Context) error) {
	sc.mu.Lock()
	sc.tasks = append(sc.tasks, Task{Name: name, Fn: fn})
	sc.mu.Unlock()
}

// Run runs all tasks in the scope
func (sc *Scope) Run() error {
	sc.mu.Lock()
	tasks := make([]Task, len(sc.tasks))
	copy(tasks, sc.tasks)
	sc.mu.Unlock()
	
	group := NewTaskGroup(sc.ctx)
	for _, task := range tasks {
		task := task
		group.Go(func() error {
			return task.Fn(sc.ctx)
		})
	}
	
	return group.Wait()
}

// Cancel cancels the scope
func (sc *Scope) Cancel() {
	sc.cancel()
}

// Context returns the scope context
func (sc *Scope) Context() context.Context {
	return sc.ctx
}

