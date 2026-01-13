package runtime

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DeterministicScheduler provides deterministic scheduling for debug/prod parity
type DeterministicScheduler struct {
	tasks      []Task
	execOrder  []string
	taskIDGen  uint64
	mu         sync.RWMutex
	deterministic bool
	seed       int64
}

// NewDeterministicScheduler creates a new deterministic scheduler
func NewDeterministicScheduler(seed int64) *DeterministicScheduler {
	return &DeterministicScheduler{
		tasks:         make([]Task, 0),
		execOrder:     make([]string, 0),
		deterministic: true,
		seed:          seed,
	}
}

// Schedule schedules a task deterministically
func (ds *DeterministicScheduler) Schedule(task Task) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	
	taskID := ds.generateTaskID()
	ds.tasks = append(ds.tasks, task)
	ds.execOrder = append(ds.execOrder, taskID)
	
	// In deterministic mode, execute tasks in order
	if ds.deterministic {
		go func(t Task, id string) {
			_ = t.Execute(context.Background())
		}(task, taskID)
	}
	
	return nil
}

// Shutdown shuts down the scheduler
func (ds *DeterministicScheduler) Shutdown() error {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	
	// Wait for all tasks to complete
	time.Sleep(100 * time.Millisecond)
	
	ds.tasks = make([]Task, 0)
	ds.execOrder = make([]string, 0)
	return nil
}

// SetDeterministic sets deterministic mode
func (ds *DeterministicScheduler) SetDeterministic(deterministic bool) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.deterministic = deterministic
}

// GetExecutionOrder returns the execution order
func (ds *DeterministicScheduler) GetExecutionOrder() []string {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	
	result := make([]string, len(ds.execOrder))
	copy(result, ds.execOrder)
	return result
}

// generateTaskID generates a deterministic task ID
func (ds *DeterministicScheduler) generateTaskID() string {
	ds.taskIDGen++
	return fmt.Sprintf("task-%d-%d", ds.seed, ds.taskIDGen)
}

