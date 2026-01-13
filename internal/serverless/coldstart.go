package serverless

import (
	"sync"
	"time"
)

// ColdStartOptimizer optimizes cold start times
type ColdStartOptimizer struct {
	warmPools    map[string]*WarmPool
	prewarmFuncs map[string]bool
	mu           sync.RWMutex
}

// WarmPool maintains a pool of warm function instances
type WarmPool struct {
	instances []*FunctionInstance
	maxSize   int
	mu        sync.Mutex
}

// FunctionInstance represents a warm function instance
type FunctionInstance struct {
	Function *Function
	LastUsed time.Time
	Ready    bool
}

// NewColdStartOptimizer creates a new cold start optimizer
func NewColdStartOptimizer() *ColdStartOptimizer {
	return &ColdStartOptimizer{
		warmPools:    make(map[string]*WarmPool),
		prewarmFuncs: make(map[string]bool),
	}
}

// Prewarm prewarms a function
func (cso *ColdStartOptimizer) Prewarm(functionName string, count int) {
	cso.mu.Lock()
	defer cso.mu.Unlock()
	
	cso.prewarmFuncs[functionName] = true
	
	pool := &WarmPool{
		instances: make([]*FunctionInstance, 0, count),
		maxSize:   count,
	}
	
	// Create warm instances
	for i := 0; i < count; i++ {
		pool.instances = append(pool.instances, &FunctionInstance{
			LastUsed: time.Now(),
			Ready:    true,
		})
	}
	
	cso.warmPools[functionName] = pool
}

// GetWarmInstance gets a warm instance from the pool
func (cso *ColdStartOptimizer) GetWarmInstance(functionName string) *FunctionInstance {
	cso.mu.RLock()
	pool, ok := cso.warmPools[functionName]
	cso.mu.RUnlock()
	
	if !ok {
		return nil
	}
	
	pool.mu.Lock()
	defer pool.mu.Unlock()
	
	if len(pool.instances) == 0 {
		return nil
	}
	
	// Get the most recently used instance
	instance := pool.instances[len(pool.instances)-1]
	pool.instances = pool.instances[:len(pool.instances)-1]
	instance.LastUsed = time.Now()
	
	return instance
}

// ReturnWarmInstance returns an instance to the pool
func (cso *ColdStartOptimizer) ReturnWarmInstance(functionName string, instance *FunctionInstance) {
	cso.mu.RLock()
	pool, ok := cso.warmPools[functionName]
	cso.mu.RUnlock()
	
	if !ok {
		return
	}
	
	pool.mu.Lock()
	defer pool.mu.Unlock()
	
	if len(pool.instances) < pool.maxSize {
		instance.LastUsed = time.Now()
		pool.instances = append(pool.instances, instance)
	}
}

// IsPrewarmed checks if a function is prewarmed
func (cso *ColdStartOptimizer) IsPrewarmed(functionName string) bool {
	cso.mu.RLock()
	defer cso.mu.RUnlock()
	return cso.prewarmFuncs[functionName]
}

