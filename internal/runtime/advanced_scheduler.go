package runtime

import (
	"context"
	"fmt"
	"sync"

	"gots-runtime/internal/eventloop"
	"gots-runtime/internal/worker"
)

// AdvancedScheduler is an advanced scheduler with CPU/I/O-bound detection
type AdvancedScheduler struct {
	workerPool *worker.Pool
	eventLoop  *eventloop.Loop
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	mu         sync.RWMutex
}

// NewAdvancedScheduler creates a new advanced scheduler
func NewAdvancedScheduler(ctx context.Context, eventLoop *eventloop.Loop) *AdvancedScheduler {
	schedCtx, cancel := context.WithCancel(ctx)
	
	// Create worker pool with min 2, max 10 workers
	pool := worker.NewPool(schedCtx, 2, 10)
	
	return &AdvancedScheduler{
		workerPool: pool,
		eventLoop:  eventLoop,
		ctx:        schedCtx,
		cancel:     cancel,
	}
}

// Start starts the scheduler
func (s *AdvancedScheduler) Start() {
	s.workerPool.Start()
}

// Schedule schedules a task for execution
func (s *AdvancedScheduler) Schedule(task Task) error {
	// Detect if task is CPU-intensive or I/O-bound
	isCPUIntensive := task.IsCPUIntensive()
	
	if isCPUIntensive {
		// Schedule CPU-intensive tasks to worker pool
		workerTask := worker.NewTask(
			generateTaskID(),
			task.Execute,
			true,
			0,
		)
		return s.workerPool.Submit(workerTask)
	} else {
		// Schedule I/O-bound tasks to event loop
		event := eventloop.NewEvent(
			eventloop.EventIO,
			func() error {
				return task.Execute(s.ctx)
			},
			0,
		)
		return s.eventLoop.Enqueue(event)
	}
}

// Shutdown gracefully shuts down the scheduler
func (s *AdvancedScheduler) Shutdown() error {
	s.cancel()
	s.workerPool.Stop()
	s.wg.Wait()
	return nil
}

// WorkerPool returns the worker pool
func (s *AdvancedScheduler) WorkerPool() *worker.Pool {
	return s.workerPool
}

// generateTaskID generates a unique task ID
var taskIDCounter uint64
var taskIDMu sync.Mutex

func generateTaskID() string {
	taskIDMu.Lock()
	defer taskIDMu.Unlock()
	taskIDCounter++
	return fmt.Sprintf("task-%d", taskIDCounter)
}

