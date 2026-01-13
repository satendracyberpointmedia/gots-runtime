package worker

import (
	"context"
	"sync"
	"time"
)

// Worker represents a worker goroutine
type Worker struct {
	id       int
	taskChan chan *Task
	resultChan chan *TaskResult
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	busy     bool
	mu       sync.RWMutex
}

// NewWorker creates a new worker
func NewWorker(id int, ctx context.Context) *Worker {
	workerCtx, cancel := context.WithCancel(ctx)
	return &Worker{
		id:         id,
		taskChan:   make(chan *Task, 1),
		resultChan: make(chan *TaskResult, 1),
		ctx:        workerCtx,
		cancel:     cancel,
		busy:       false,
	}
}

// Start starts the worker
func (w *Worker) Start() {
	w.wg.Add(1)
	go w.run()
}

// Stop stops the worker
func (w *Worker) Stop() {
	w.cancel()
	close(w.taskChan)
	w.wg.Wait()
}

// AssignTask assigns a task to the worker
func (w *Worker) AssignTask(task *Task) error {
	select {
	case w.taskChan <- task:
		return nil
	case <-w.ctx.Done():
		return w.ctx.Err()
	}
}

// IsBusy returns whether the worker is currently busy
func (w *Worker) IsBusy() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.busy
}

// ResultChan returns the result channel
func (w *Worker) ResultChan() <-chan *TaskResult {
	return w.resultChan
}

// run is the main worker loop
func (w *Worker) run() {
	defer w.wg.Done()

	for {
		select {
		case task, ok := <-w.taskChan:
			if !ok {
				return
			}
			w.executeTask(task)
		case <-w.ctx.Done():
			return
		}
	}
}

// executeTask executes a task
func (w *Worker) executeTask(task *Task) {
	w.mu.Lock()
	w.busy = true
	w.mu.Unlock()

	start := time.Now()
	err := task.Execute(w.ctx)
	duration := time.Since(start)

	result := &TaskResult{
		TaskID:   task.ID,
		Error:    err,
		Duration: duration,
	}

	select {
	case w.resultChan <- result:
	case <-w.ctx.Done():
	}

	w.mu.Lock()
	w.busy = false
	w.mu.Unlock()
}

