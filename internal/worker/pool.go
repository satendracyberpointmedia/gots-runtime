package worker

import (
	"context"
	"sync"
	"time"
)

// Pool represents a pool of workers
type Pool struct {
	workers     []*Worker
	taskQueue   chan *Task
	resultChan  chan *TaskResult
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	minWorkers  int
	maxWorkers  int
	currentWorkers int
	mu          sync.RWMutex
}

// NewPool creates a new worker pool
func NewPool(ctx context.Context, minWorkers, maxWorkers int) *Pool {
	poolCtx, cancel := context.WithCancel(ctx)
	return &Pool{
		workers:        make([]*Worker, 0),
		taskQueue:       make(chan *Task, 100),
		resultChan:      make(chan *TaskResult, 100),
		ctx:             poolCtx,
		cancel:          cancel,
		minWorkers:      minWorkers,
		maxWorkers:      maxWorkers,
		currentWorkers:  0,
	}
}

// Start starts the worker pool
func (p *Pool) Start() {
	// Start minimum number of workers
	for i := 0; i < p.minWorkers; i++ {
		p.addWorker()
	}

	// Start task dispatcher
	p.wg.Add(1)
	go p.dispatch()

	// Start worker scaler
	p.wg.Add(1)
	go p.scale()
}

// Stop stops the worker pool
func (p *Pool) Stop() {
	p.cancel()
	close(p.taskQueue)

	p.mu.Lock()
	for _, worker := range p.workers {
		worker.Stop()
	}
	p.mu.Unlock()

	p.wg.Wait()
}

// Submit submits a task to the pool
func (p *Pool) Submit(task *Task) error {
	select {
	case p.taskQueue <- task:
		return nil
	case <-p.ctx.Done():
		return p.ctx.Err()
	}
}

// ResultChan returns the result channel
func (p *Pool) ResultChan() <-chan *TaskResult {
	return p.resultChan
}

// addWorker adds a new worker to the pool
func (p *Pool) addWorker() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.currentWorkers >= p.maxWorkers {
		return
	}

	worker := NewWorker(p.currentWorkers, p.ctx)
	worker.Start()

	// Forward results to pool result channel
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		for result := range worker.ResultChan() {
			select {
			case p.resultChan <- result:
			case <-p.ctx.Done():
				return
			}
		}
	}()

	p.workers = append(p.workers, worker)
	p.currentWorkers++
}

// removeWorker removes a worker from the pool
func (p *Pool) removeWorker() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.currentWorkers <= p.minWorkers {
		return
	}

	if len(p.workers) == 0 {
		return
	}

	// Remove the last worker
	worker := p.workers[len(p.workers)-1]
	p.workers = p.workers[:len(p.workers)-1]
	worker.Stop()
	p.currentWorkers--
}

// dispatch dispatches tasks to workers
func (p *Pool) dispatch() {
	defer p.wg.Done()

	for {
		select {
		case task, ok := <-p.taskQueue:
			if !ok {
				return
			}
			p.assignTask(task)
		case <-p.ctx.Done():
			return
		}
	}
}

// assignTask assigns a task to an available worker
func (p *Pool) assignTask(task *Task) {
	p.mu.RLock()
	workers := make([]*Worker, len(p.workers))
	copy(workers, p.workers)
	p.mu.RUnlock()

	// Try to find an idle worker
	for _, worker := range workers {
		if !worker.IsBusy() {
			_ = worker.AssignTask(task)
			return
		}
	}

	// All workers are busy, assign to first available or create new worker
	if p.currentWorkers < p.maxWorkers {
		p.addWorker()
		// Try again with new worker
		p.mu.RLock()
		if len(p.workers) > 0 {
			_ = p.workers[len(p.workers)-1].AssignTask(task)
		}
		p.mu.RUnlock()
	} else {
		// Assign to first worker (round-robin could be improved)
		if len(workers) > 0 {
			_ = workers[0].AssignTask(task)
		}
	}
}

// scale periodically adjusts the number of workers
func (p *Pool) scale() {
	defer p.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.adjustWorkers()
		case <-p.ctx.Done():
			return
		}
	}
}

// adjustWorkers adjusts the number of workers based on load
func (p *Pool) adjustWorkers() {
	p.mu.RLock()
	busyCount := 0
	for _, worker := range p.workers {
		if worker.IsBusy() {
			busyCount++
		}
	}
	queueSize := len(p.taskQueue)
	p.mu.RUnlock()

	// If all workers are busy and queue has tasks, add workers
	if busyCount == p.currentWorkers && queueSize > 0 && p.currentWorkers < p.maxWorkers {
		p.addWorker()
	}

	// If many workers are idle, remove some
	if busyCount < p.currentWorkers/2 && p.currentWorkers > p.minWorkers {
		p.removeWorker()
	}
}

// Stats returns pool statistics
type Stats struct {
	CurrentWorkers int
	BusyWorkers    int
	QueueSize      int
	MinWorkers     int
	MaxWorkers     int
}

// GetStats returns current pool statistics
func (p *Pool) GetStats() Stats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	busyCount := 0
	for _, worker := range p.workers {
		if worker.IsBusy() {
			busyCount++
		}
	}

	return Stats{
		CurrentWorkers: p.currentWorkers,
		BusyWorkers:    busyCount,
		QueueSize:      len(p.taskQueue),
		MinWorkers:     p.minWorkers,
		MaxWorkers:     p.maxWorkers,
	}
}

