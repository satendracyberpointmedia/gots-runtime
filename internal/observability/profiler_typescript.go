package observability

import (
	"encoding/base64"
	"os"
	"sync"

	"github.com/dop251/goja"
)

// TypeScriptProfiler wraps Profiler for TypeScript
type TypeScriptProfiler struct {
	profiler *Profiler
	engine   *goja.Runtime
	results  map[string]string // profile type -> base64 encoded data
	mu       sync.RWMutex
}

// NewTypeScriptProfiler creates a new TypeScript-wrapped profiler
func NewTypeScriptProfiler(engine *goja.Runtime) *TypeScriptProfiler {
	return &TypeScriptProfiler{
		profiler: NewProfiler(),
		engine:   engine,
		results:  make(map[string]string),
	}
}

// ToJSObject converts the profiler to a JavaScript object
func (tsp *TypeScriptProfiler) ToJSObject() *goja.Object {
	obj := tsp.engine.NewObject()
	
	// StartCPU method
	obj.Set("startCPU", func(outputPath goja.Value) *goja.Promise {
		promise, resolve, reject := tsp.engine.NewPromise()
		
		go func() {
			path := ""
			if outputPath != nil && !goja.IsUndefined(outputPath) {
				path = outputPath.String()
			}
			
			if err := tsp.profiler.StartCPUProfile(path); err != nil {
				reject(tsp.engine.ToValue(err.Error()))
			} else {
				resolve(tsp.engine.ToValue(true))
			}
		}()
		
		return promise
	})
	
	// StartMemory method
	obj.Set("startMemory", func(outputPath goja.Value) *goja.Promise {
		promise, resolve, reject := tsp.engine.NewPromise()
		
		go func() {
			path := ""
			if outputPath != nil && !goja.IsUndefined(outputPath) {
				path = outputPath.String()
			}
			
			if err := tsp.profiler.WriteHeapProfile(path); err != nil {
				reject(tsp.engine.ToValue(err.Error()))
			} else {
				// Read and encode the profile
				if path != "" {
					data, err := os.ReadFile(path)
					if err == nil {
						tsp.mu.Lock()
						tsp.results["memory"] = base64.StdEncoding.EncodeToString(data)
						tsp.mu.Unlock()
					}
				}
				resolve(tsp.engine.ToValue(true))
			}
		}()
		
		return promise
	})
	
	// StartGoroutine method
	obj.Set("startGoroutine", func(outputPath goja.Value) *goja.Promise {
		promise, resolve, reject := tsp.engine.NewPromise()
		
		go func() {
			path := ""
			if outputPath != nil && !goja.IsUndefined(outputPath) {
				path = outputPath.String()
			}
			
			if err := tsp.profiler.WriteGoroutineProfile(path); err != nil {
				reject(tsp.engine.ToValue(err.Error()))
			} else {
				// Read and encode the profile
				if path != "" {
					data, err := os.ReadFile(path)
					if err == nil {
						tsp.mu.Lock()
						tsp.results["goroutine"] = base64.StdEncoding.EncodeToString(data)
						tsp.mu.Unlock()
					}
				}
				resolve(tsp.engine.ToValue(true))
			}
		}()
		
		return promise
	})
	
	// StartBlock method
	obj.Set("startBlock", func(outputPath goja.Value) *goja.Promise {
		promise, resolve, reject := tsp.engine.NewPromise()
		
		go func() {
			path := ""
			if outputPath != nil && !goja.IsUndefined(outputPath) {
				path = outputPath.String()
			}
			
			if err := tsp.profiler.WriteBlockProfile(path); err != nil {
				reject(tsp.engine.ToValue(err.Error()))
			} else {
				// Read and encode the profile
				if path != "" {
					data, err := os.ReadFile(path)
					if err == nil {
						tsp.mu.Lock()
						tsp.results["block"] = base64.StdEncoding.EncodeToString(data)
						tsp.mu.Unlock()
					}
				}
				resolve(tsp.engine.ToValue(true))
			}
		}()
		
		return promise
	})
	
	// StartMutex method
	obj.Set("startMutex", func(outputPath goja.Value) *goja.Promise {
		promise, resolve, reject := tsp.engine.NewPromise()
		
		go func() {
			path := ""
			if outputPath != nil && !goja.IsUndefined(outputPath) {
				path = outputPath.String()
			}
			
			if err := tsp.profiler.WriteMutexProfile(path); err != nil {
				reject(tsp.engine.ToValue(err.Error()))
			} else {
				// Read and encode the profile
				if path != "" {
					data, err := os.ReadFile(path)
					if err == nil {
						tsp.mu.Lock()
						tsp.results["mutex"] = base64.StdEncoding.EncodeToString(data)
						tsp.mu.Unlock()
					}
				}
				resolve(tsp.engine.ToValue(true))
			}
		}()
		
		return promise
	})
	
	// Stop method
	obj.Set("stop", func() *goja.Promise {
		promise, resolve, _ := tsp.engine.NewPromise()
		
		go func() {
			// Stop CPU profiling if active
			if err := tsp.profiler.StopCPUProfile(); err != nil {
				// Not an error if CPU profiling wasn't active
				_ = err
			}
			
			// Read CPU profile if it exists
			tsp.mu.Lock()
			results := make(map[string]interface{})
			for k, v := range tsp.results {
				results[k] = v
			}
			tsp.mu.Unlock()
			
			resultObj := tsp.engine.NewObject()
			if cpu, ok := results["cpu"]; ok {
				resultObj.Set("cpu", cpu)
			}
			if memory, ok := results["memory"]; ok {
				resultObj.Set("memory", memory)
			}
			if goroutine, ok := results["goroutine"]; ok {
				resultObj.Set("goroutine", goroutine)
			}
			if block, ok := results["block"]; ok {
				resultObj.Set("block", block)
			}
			if mutex, ok := results["mutex"]; ok {
				resultObj.Set("mutex", mutex)
			}
			
			resolve(resultObj)
		}()
		
		return promise
	})
	
	// GetResults method
	obj.Set("getResults", func() map[string]interface{} {
		tsp.mu.RLock()
		defer tsp.mu.RUnlock()
		
		results := make(map[string]interface{})
		for k, v := range tsp.results {
			results[k] = v
		}
		return results
	})
	
	return obj
}

