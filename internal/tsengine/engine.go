package tsengine

import (
	"fmt"
	"sync"

	"github.com/dop251/goja"
)

// Engine represents the TypeScript execution engine
type Engine struct {
	vm       *goja.Runtime
	compiler *Compiler
	mu       sync.RWMutex
}

// NewEngine creates a new TypeScript execution engine
func NewEngine() *Engine {
	vm := goja.New()
	return &Engine{
		vm:       vm,
		compiler: NewCompiler(),
	}
}

// ExecuteFile executes a TypeScript file
func (e *Engine) ExecuteFile(filePath string) (goja.Value, error) {
	// Compile TypeScript to JavaScript
	jsCode, err := e.compiler.Compile(filePath)
	if err != nil {
		return nil, fmt.Errorf("compilation failed: %w", err)
	}

	// Execute the compiled JavaScript
	return e.Execute(jsCode)
}

// Execute executes JavaScript code
func (e *Engine) Execute(code string) (goja.Value, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	value, err := e.vm.RunString(code)
	if err != nil {
		return nil, fmt.Errorf("execution failed: %w", err)
	}

	return value, nil
}

// Set sets a value in the JavaScript runtime
func (e *Engine) Set(name string, value interface{}) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.vm.Set(name, value)
}

// Get gets a value from the JavaScript runtime
func (e *Engine) Get(name string) goja.Value {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.vm.Get(name)
}

// RegisterFunction registers a Go function in the JavaScript runtime
func (e *Engine) RegisterFunction(name string, fn interface{}) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.vm.Set(name, fn)
}

// VM returns the underlying goja runtime
func (e *Engine) VM() *goja.Runtime {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.vm
}

// Compiler returns the TypeScript compiler
func (e *Engine) Compiler() *Compiler {
	return e.compiler
}

