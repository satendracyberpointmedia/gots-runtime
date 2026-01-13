package runtime

import (
	"context"
	"sync"
)

// BasicScheduler is a simple scheduler implementation for Phase 1
type BasicScheduler struct {
	taskChan chan Task
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewBasicScheduler creates a new basic scheduler
func NewBasicScheduler(ctx context.Context) *BasicScheduler {
	schedCtx, cancel := context.WithCancel(ctx)
	return &BasicScheduler{
		taskChan: make(chan Task, 100),
		ctx:      schedCtx,
		cancel:   cancel,
	}
}

// Start starts the scheduler
func (s *BasicScheduler) Start() {
	s.wg.Add(1)
	go s.run()
}

// Schedule schedules a task for execution
func (s *BasicScheduler) Schedule(task Task) error {
	select {
	case s.taskChan <- task:
		return nil
	case <-s.ctx.Done():
		return s.ctx.Err()
	}
}

// Shutdown gracefully shuts down the scheduler
func (s *BasicScheduler) Shutdown() error {
	s.cancel()
	close(s.taskChan)
	s.wg.Wait()
	return nil
}

// run is the main scheduler loop
func (s *BasicScheduler) run() {
	defer s.wg.Done()

	for {
		select {
		case task, ok := <-s.taskChan:
			if !ok {
				return
			}
			// Execute task in a goroutine
			go func(t Task) {
				_ = t.Execute(s.ctx)
			}(task)
		case <-s.ctx.Done():
			return
		}
	}
}

