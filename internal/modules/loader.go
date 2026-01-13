package modules

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gots-runtime/internal/tsengine"
)

// Loader loads and executes modules
type Loader struct {
	resolver *Resolver
	cache    *Cache
	engines  map[string]*tsengine.Engine
	mu       sync.RWMutex
}

// NewLoader creates a new module loader
func NewLoader(basePath string) *Loader {
	return &Loader{
		resolver: NewResolver(basePath),
		cache:    NewCache(),
		engines:  make(map[string]*tsengine.Engine),
	}
}

// Load loads a module from a file path
func (l *Loader) Load(modulePath string) (*tsengine.Module, error) {
	// Check cache first
	if cached, ok := l.cache.Get(modulePath); ok {
		// Create a new module from cache
		engine := l.getOrCreateEngine(modulePath)
		module := tsengine.NewModule(cached.Path, engine)
		for k, v := range cached.Exports {
			module.SetExport(k, v)
		}
		return module, nil
	}

	// Validate path
	absPath, err := filepath.Abs(modulePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	if err := l.resolver.ValidatePath(absPath); err != nil {
		return nil, err
	}

	// Get or create engine for this module
	engine := l.getOrCreateEngine(absPath)

	// Load and execute the module
	module := tsengine.NewModule(absPath, engine)
	
	// Read and compile the module
	code, err := engine.Compiler().Compile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to compile module: %w", err)
	}

	// Set up module exports object
	exports := make(map[string]interface{})
	engine.Set("exports", exports)
	engine.Set("module", map[string]interface{}{
		"exports": exports,
	})

	// Execute the module code
	_, err = engine.Execute(code)
	if err != nil {
		return nil, fmt.Errorf("failed to execute module: %w", err)
	}

	// Extract exports
	exportsValue := engine.Get("exports")
	if exportsValue != nil {
		if exportsObj, ok := exportsValue.Export().(map[string]interface{}); ok {
			for k, v := range exportsObj {
				module.SetExport(k, v)
				exports[k] = v
			}
		}
	}

	// Cache the module
	info, _ := os.Stat(absPath)
	timestamp := int64(0)
	if info != nil {
		timestamp = info.ModTime().Unix()
	}
	
	l.cache.Set(absPath, &CachedModule{
		Path:      absPath,
		Exports:   exports,
		Timestamp: timestamp,
	})

	return module, nil
}

// getOrCreateEngine gets or creates an engine for a module path
func (l *Loader) getOrCreateEngine(modulePath string) *tsengine.Engine {
	l.mu.Lock()
	defer l.mu.Unlock()

	if engine, ok := l.engines[modulePath]; ok {
		return engine
	}

	engine := tsengine.NewEngine()
	l.engines[modulePath] = engine
	return engine
}

// Resolve resolves a module import path
func (l *Loader) Resolve(importPath string, fromPath string) (string, error) {
	return l.resolver.Resolve(importPath, fromPath)
}

// ClearCache clears the module cache
func (l *Loader) ClearCache() {
	l.cache.Clear()
}

