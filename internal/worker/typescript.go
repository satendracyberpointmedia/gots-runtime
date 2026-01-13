package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/dop251/goja"
)

// TypeScriptWorker provides TypeScript bindings for worker pool
type TypeScriptWorker struct {
	pool    *Pool
	engine  *goja.Runtime
	ctx     context.Context
	cancel  context.CancelFunc
	mu      sync.RWMutex
}

// NewTypeScriptWorker creates a new TypeScript worker wrapper
func NewTypeScriptWorker(ctx context.Context, engine *goja.Runtime, minWorkers, maxWorkers int) *TypeScriptWorker {
	workerCtx, cancel := context.WithCancel(ctx)
	pool := NewPool(workerCtx, minWorkers, maxWorkers)
	pool.Start()
	
	return &TypeScriptWorker{
		pool:   pool,
		engine: engine,
		ctx:    workerCtx,
		cancel: cancel,
	}
}

// Spawn executes a task in a worker and returns a promise
func (tw *TypeScriptWorker) Spawn(taskID string, handler goja.Callable, data goja.Value) *goja.Promise {
	promise, resolve, reject := tw.engine.NewPromise()
	
	go func() {
		// Create a task that executes the handler
		task := NewTask(
			taskID,
			func(ctx context.Context) error {
				// Call the TypeScript handler with the data
				result, err := handler(nil, data)
				if err != nil {
					return fmt.Errorf("handler error: %w", err)
				}
				
				// Store result in task context (we'll need to modify this)
				_ = result
				return nil
			},
			true, // CPU intensive
			0,    // default priority
		)
		
		// Submit task to pool
		if err := tw.pool.Submit(task); err != nil {
			reject(tw.engine.ToValue(err.Error()))
			return
		}
		
		// Wait for result
		select {
		case result := <-tw.pool.ResultChan():
			if result.Error != nil {
				reject(tw.engine.ToValue(result.Error.Error()))
			} else {
				// Create result object
				resultObj := tw.engine.NewObject()
				resultObj.Set("id", result.TaskID)
				resultObj.Set("data", nil) // We need to store the actual result
				resultObj.Set("duration", result.Duration.Milliseconds())
				resolve(resultObj)
			}
		case <-tw.ctx.Done():
			reject(tw.engine.ToValue("worker pool closed"))
		case <-time.After(30 * time.Second):
			reject(tw.engine.ToValue("task timeout"))
		}
	}()
	
	return promise
}

// SpawnBatch executes multiple tasks in parallel
func (tw *TypeScriptWorker) SpawnBatch(tasks []interface{}) *goja.Promise {
	promise, resolve, reject := tw.engine.NewPromise()
	
	go func() {
		results := make([]interface{}, 0, len(tasks))
		
		for i, taskVal := range tasks {
			taskObj, ok := taskVal.(*goja.Object)
			if !ok {
				reject(tw.engine.ToValue(fmt.Sprintf("task %d is not an object", i)))
				return
			}
			
			taskID := taskObj.Get("id").String()
			handlerVal := taskObj.Get("handler")
			handler, ok := goja.AssertFunction(handlerVal)
			if !ok {
				reject(tw.engine.ToValue(fmt.Sprintf("task %d handler is not a function", i)))
				return
			}
			data := taskObj.Get("data")
			
			// Create task
			task := NewTask(
				taskID,
				func(ctx context.Context) error {
					_, err := handler(nil, data)
					return err
				},
				true,
				0,
			)
			
			// Submit task
			if err := tw.pool.Submit(task); err != nil {
				reject(tw.engine.ToValue(fmt.Sprintf("failed to submit task %d: %v", i, err)))
				return
			}
		}
		
		// Collect results
		for i := 0; i < len(tasks); i++ {
			select {
			case result := <-tw.pool.ResultChan():
				resultObj := tw.engine.NewObject()
				resultObj.Set("id", result.TaskID)
				resultObj.Set("data", nil)
				resultObj.Set("duration", result.Duration.Milliseconds())
				if result.Error != nil {
					resultObj.Set("error", tw.engine.ToValue(result.Error.Error()))
				}
				results = append(results, resultObj)
			case <-tw.ctx.Done():
				reject(tw.engine.ToValue("worker pool closed"))
				return
			case <-time.After(30 * time.Second):
				reject(tw.engine.ToValue("task timeout"))
				return
			}
		}
		
		resolve(tw.engine.ToValue(results))
	}()
	
	return promise
}

// GetStats returns worker pool statistics
func (tw *TypeScriptWorker) GetStats() map[string]interface{} {
	tw.mu.RLock()
	defer tw.mu.RUnlock()
	
	busy := 0
	for _, w := range tw.pool.workers {
		if w.IsBusy() {
			busy++
		}
	}
	
	return map[string]interface{}{
		"totalWorkers": tw.pool.currentWorkers,
		"busyWorkers":  busy,
		"idleWorkers":  tw.pool.currentWorkers - busy,
		"queuedTasks":  len(tw.pool.taskQueue),
	}
}

// Close closes the worker pool
func (tw *TypeScriptWorker) Close() error {
	tw.cancel()
	tw.pool.Stop()
	return nil
}

// SpawnWorker is a convenience function to spawn a single worker task
func SpawnWorker(ctx context.Context, engine *goja.Runtime, taskID string, handler goja.Callable, data goja.Value) *goja.Promise {
	worker := NewTypeScriptWorker(ctx, engine, 1, 1)
	return worker.Spawn(taskID, handler, data)
}

// Helper function to serialize/deserialize data for worker tasks
func serializeData(data goja.Value) (string, error) {
	jsonData, err := json.Marshal(data.Export())
	if err != nil {
		return "", fmt.Errorf("failed to serialize data: %w", err)
	}
	return string(jsonData), nil
}

func deserializeData(jsonStr string) (interface{}, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("failed to deserialize data: %w", err)
	}
	return data, nil
}

