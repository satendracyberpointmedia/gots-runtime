package serverless

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Event represents a serverless event
type Event struct {
	Type      string
	Payload   json.RawMessage
	Timestamp time.Time
	Source    string
}

// Context represents serverless function context
type Context struct {
	RequestID   string
	FunctionName string
	Timeout     time.Duration
	MemoryLimit uint64
	Env         map[string]string
}

// Handler is a function that handles serverless events
type Handler func(ctx context.Context, event *Event, fnCtx *Context) (interface{}, error)

// Function represents a serverless function
type Function struct {
	Name        string
	Handler     Handler
	Timeout     time.Duration
	MemoryLimit uint64
	Runtime     string
	Env         map[string]string
	mu          sync.RWMutex
}

// NewFunction creates a new serverless function
func NewFunction(name string, handler Handler) *Function {
	return &Function{
		Name:        name,
		Handler:     handler,
		Timeout:     30 * time.Second,
		MemoryLimit: 128 * 1024 * 1024, // 128MB default
		Runtime:     "gots",
		Env:         make(map[string]string),
	}
}

// SetTimeout sets the function timeout
func (f *Function) SetTimeout(timeout time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Timeout = timeout
}

// SetMemoryLimit sets the memory limit
func (f *Function) SetMemoryLimit(limit uint64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.MemoryLimit = limit
}

// Invoke invokes the function
func (f *Function) Invoke(ctx context.Context, event *Event) (interface{}, error) {
	fnCtx := &Context{
		RequestID:    generateRequestID(),
		FunctionName: f.Name,
		Timeout:      f.Timeout,
		MemoryLimit:  f.MemoryLimit,
		Env:          f.Env,
	}
	
	// Create context with timeout
	invokeCtx, cancel := context.WithTimeout(ctx, f.Timeout)
	defer cancel()
	
	return f.Handler(invokeCtx, event, fnCtx)
}

// ServerlessRuntime provides serverless execution environment
type ServerlessRuntime struct {
	functions map[string]*Function
	mu       sync.RWMutex
}

// NewServerlessRuntime creates a new serverless runtime
func NewServerlessRuntime() *ServerlessRuntime {
	return &ServerlessRuntime{
		functions: make(map[string]*Function),
	}
}

// RegisterFunction registers a function
func (sr *ServerlessRuntime) RegisterFunction(function *Function) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.functions[function.Name] = function
}

// InvokeFunction invokes a function by name
func (sr *ServerlessRuntime) InvokeFunction(name string, event *Event) (interface{}, error) {
	sr.mu.RLock()
	function, ok := sr.functions[name]
	sr.mu.RUnlock()
	
	if !ok {
		return nil, fmt.Errorf("function not found: %s", name)
	}
	
	return function.Invoke(context.Background(), event)
}

// ListFunctions lists all registered functions
func (sr *ServerlessRuntime) ListFunctions() []string {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	
	names := make([]string, 0, len(sr.functions))
	for name := range sr.functions {
		names = append(names, name)
	}
	return names
}

var requestIDCounter uint64
var requestIDMu sync.Mutex

func generateRequestID() string {
	requestIDMu.Lock()
	defer requestIDMu.Unlock()
	requestIDCounter++
	return fmt.Sprintf("req-%d", requestIDCounter)
}

