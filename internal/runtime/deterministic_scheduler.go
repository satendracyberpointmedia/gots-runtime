package runtime

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// DeterministicScheduler provides deterministic scheduling for debug/prod parity
type DeterministicScheduler struct {
	tasks              []TaskExecution
	execOrder          []string
	taskIDGen          uint64
	mu                 sync.RWMutex
	deterministic      bool
	seed               int64
	rng                *rand.Rand
	executionLog       []ExecutionRecord
	taskCompletionChan map[string]chan TaskResult
	waitGroup          sync.WaitGroup
	stopped            bool
}

// TaskExecution represents a task with metadata
type TaskExecution struct {
	ID        string
	Task      Task
	Priority  int
	Timestamp time.Time
	Context   context.Context
	Cancel    context.CancelFunc
}

// ExecutionRecord represents a recorded execution
type ExecutionRecord struct {
	TaskID    string
	Status    string // "pending", "running", "completed", "failed"
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	Error     error
}

// TaskResult represents the result of task execution
type TaskResult struct {
	TaskID string
	Error  error
	Result any
}

// NewDeterministicScheduler creates a new deterministic scheduler
func NewDeterministicScheduler(seed int64) *DeterministicScheduler {
	return &DeterministicScheduler{
		tasks:              make([]TaskExecution, 0),
		execOrder:          make([]string, 0),
		deterministic:      true,
		seed:               seed,
		rng:                rand.New(rand.NewSource(seed)),
		executionLog:       make([]ExecutionRecord, 0),
		taskCompletionChan: make(map[string]chan TaskResult),
		stopped:            false,
	}
}

// Schedule schedules a task deterministically
func (ds *DeterministicScheduler) Schedule(task Task) error {
	return ds.ScheduleWithPriority(task, 0)
}

// ScheduleWithPriority schedules a task with a priority (higher = more important)
func (ds *DeterministicScheduler) ScheduleWithPriority(task Task, priority int) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if ds.stopped {
		return fmt.Errorf("scheduler is stopped")
	}

	taskID := ds.generateTaskID()
	ctx, cancel := context.WithCancel(context.Background())

	execution := TaskExecution{
		ID:        taskID,
		Task:      task,
		Priority:  priority,
		Timestamp: time.Now(),
		Context:   ctx,
		Cancel:    cancel,
	}

	ds.tasks = append(ds.tasks, execution)
	ds.execOrder = append(ds.execOrder, taskID)
	ds.taskCompletionChan[taskID] = make(chan TaskResult, 1)

	// Record execution start
	ds.logExecution(ExecutionRecord{
		TaskID:    taskID,
		Status:    "pending",
		StartTime: time.Now(),
	})

	// In deterministic mode, execute tasks in strict order
	if ds.deterministic {
		ds.waitGroup.Add(1)
		go ds.executeTaskDeterministic(execution)
	} else {
		ds.waitGroup.Add(1)
		go ds.executeTaskConcurrent(execution)
	}

	return nil
}

// executeTaskDeterministic executes a task in deterministic order
func (ds *DeterministicScheduler) executeTaskDeterministic(execution TaskExecution) {
	defer ds.waitGroup.Done()

	ds.updateExecutionStatus(execution.ID, "running")
	startTime := time.Now()

	err := execution.Task.Execute(execution.Context)

	duration := time.Since(startTime)

	// Record result
	ds.mu.Lock()
	ds.executionLog = append(ds.executionLog, ExecutionRecord{
		TaskID:    execution.ID,
		Status:    "completed",
		StartTime: startTime,
		EndTime:   time.Now(),
		Duration:  duration,
		Error:     err,
	})
	ds.mu.Unlock()

	// Send result to channel
	ds.taskCompletionChan[execution.ID] <- TaskResult{
		TaskID: execution.ID,
		Error:  err,
	}
}

// executeTaskConcurrent executes a task concurrently
func (ds *DeterministicScheduler) executeTaskConcurrent(execution TaskExecution) {
	defer ds.waitGroup.Done()

	ds.updateExecutionStatus(execution.ID, "running")
	startTime := time.Now()

	err := execution.Task.Execute(execution.Context)

	duration := time.Since(startTime)

	// Record result
	ds.mu.Lock()
	ds.executionLog = append(ds.executionLog, ExecutionRecord{
		TaskID:    execution.ID,
		Status:    "completed",
		StartTime: startTime,
		EndTime:   time.Now(),
		Duration:  duration,
		Error:     err,
	})
	ds.mu.Unlock()

	// Send result to channel
	ds.taskCompletionChan[execution.ID] <- TaskResult{
		TaskID: execution.ID,
		Error:  err,
	}
}

// updateExecutionStatus updates the execution status in the log
func (ds *DeterministicScheduler) updateExecutionStatus(taskID, status string) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	for i := range ds.executionLog {
		if ds.executionLog[i].TaskID == taskID {
			ds.executionLog[i].Status = status
			break
		}
	}
}

// logExecution logs an execution record
func (ds *DeterministicScheduler) logExecution(record ExecutionRecord) {
	ds.executionLog = append(ds.executionLog, record)
}

// Shutdown shuts down the scheduler and waits for all tasks
func (ds *DeterministicScheduler) Shutdown() error {
	ds.mu.Lock()
	ds.stopped = true
	ds.mu.Unlock()

	// Wait for all tasks to complete
	ds.waitGroup.Wait()

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

// GetExecutionLog returns the execution log
func (ds *DeterministicScheduler) GetExecutionLog() []ExecutionRecord {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	result := make([]ExecutionRecord, len(ds.executionLog))
	copy(result, ds.executionLog)
	return result
}

// GetTaskResult waits for and returns the result of a task
func (ds *DeterministicScheduler) GetTaskResult(taskID string, timeout time.Duration) (TaskResult, error) {
	ds.mu.RLock()
	ch, ok := ds.taskCompletionChan[taskID]
	ds.mu.RUnlock()

	if !ok {
		return TaskResult{}, fmt.Errorf("task not found: %s", taskID)
	}

	select {
	case result := <-ch:
		return result, nil
	case <-time.After(timeout):
		return TaskResult{}, fmt.Errorf("task timeout: %s", taskID)
	}
}

// CancelTask cancels a scheduled task
func (ds *DeterministicScheduler) CancelTask(taskID string) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	for i := range ds.tasks {
		if ds.tasks[i].ID == taskID {
			ds.tasks[i].Cancel()
			return nil
		}
	}

	return fmt.Errorf("task not found: %s", taskID)
}

// generateTaskID generates a deterministic task ID
func (ds *DeterministicScheduler) generateTaskID() string {
	ds.taskIDGen++
	return fmt.Sprintf("task-%d-%d", ds.seed, ds.taskIDGen)
}

// GetStats returns scheduler statistics
func (ds *DeterministicScheduler) GetStats() map[string]interface{} {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	completed := 0
	failed := 0
	totalDuration := time.Duration(0)

	for _, record := range ds.executionLog {
		if record.Status == "completed" {
			completed++
			if record.Error != nil {
				failed++
			}
			totalDuration += record.Duration
		}
	}

	return map[string]interface{}{
		"total_tasks":     len(ds.execOrder),
		"completed_tasks": completed,
		"failed_tasks":    failed,
		"pending_tasks":   len(ds.tasks) - completed,
		"deterministic":   ds.deterministic,
		"total_duration":  totalDuration,
		"avg_task_time":   totalDuration / time.Duration(completed+1),
		"seed":            ds.seed,
	}
}
