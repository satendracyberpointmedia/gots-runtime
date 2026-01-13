package runtime

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// MemoryIsolation provides per-module memory isolation
type MemoryIsolation struct {
	modules map[string]*ModuleMemory
	mu      sync.RWMutex
}

// ModuleMemory represents memory usage for a module
type ModuleMemory struct {
	ModuleID     string
	Allocated    uint64
	MaxAllocated uint64
	Allocations  map[uintptr]uint64
	mu           sync.RWMutex
}

// NewMemoryIsolation creates a new memory isolation manager
func NewMemoryIsolation() *MemoryIsolation {
	return &MemoryIsolation{
		modules: make(map[string]*ModuleMemory),
	}
}

// RegisterModule registers a module for memory tracking
func (mi *MemoryIsolation) RegisterModule(moduleID string) {
	mi.mu.Lock()
	defer mi.mu.Unlock()
	
	if _, exists := mi.modules[moduleID]; !exists {
		mi.modules[moduleID] = &ModuleMemory{
			ModuleID:    moduleID,
			Allocations: make(map[uintptr]uint64),
		}
	}
}

// TrackAllocation tracks a memory allocation for a module
func (mi *MemoryIsolation) TrackAllocation(moduleID string, ptr uintptr, size uint64) {
	mi.mu.RLock()
	module, ok := mi.modules[moduleID]
	mi.mu.RUnlock()
	
	if !ok {
		return
	}
	
	module.mu.Lock()
	defer module.mu.Unlock()
	
	module.Allocations[ptr] = size
	module.Allocated += size
	if module.Allocated > module.MaxAllocated {
		module.MaxAllocated = module.Allocated
	}
}

// TrackDeallocation tracks a memory deallocation for a module
func (mi *MemoryIsolation) TrackDeallocation(moduleID string, ptr uintptr) {
	mi.mu.RLock()
	module, ok := mi.modules[moduleID]
	mi.mu.RUnlock()
	
	if !ok {
		return
	}
	
	module.mu.Lock()
	defer module.mu.Unlock()
	
	if size, ok := module.Allocations[ptr]; ok {
		module.Allocated -= size
		delete(module.Allocations, ptr)
	}
}

// GetModuleMemory gets memory usage for a module
func (mi *MemoryIsolation) GetModuleMemory(moduleID string) (*ModuleMemory, error) {
	mi.mu.RLock()
	defer mi.mu.RUnlock()
	
	module, ok := mi.modules[moduleID]
	if !ok {
		return nil, fmt.Errorf("module not found: %s", moduleID)
	}
	
	return module, nil
}

// GetAllModulesMemory gets memory usage for all modules
func (mi *MemoryIsolation) GetAllModulesMemory() map[string]*ModuleMemory {
	mi.mu.RLock()
	defer mi.mu.RUnlock()
	
	result := make(map[string]*ModuleMemory)
	for k, v := range mi.modules {
		result[k] = v
	}
	return result
}

// MemoryLeakDetector detects memory leaks
type MemoryLeakDetector struct {
	isolation *MemoryIsolation
	threshold uint64
	interval  time.Duration
	stop      chan struct{}
	mu        sync.Mutex
}

// NewMemoryLeakDetector creates a new memory leak detector
func NewMemoryLeakDetector(isolation *MemoryIsolation, threshold uint64, interval time.Duration) *MemoryLeakDetector {
	return &MemoryLeakDetector{
		isolation: isolation,
		threshold: threshold,
		interval:  interval,
		stop:      make(chan struct{}),
	}
}

// Start starts the memory leak detector
func (mld *MemoryLeakDetector) Start() {
	mld.mu.Lock()
	defer mld.mu.Unlock()
	
	go mld.detect()
}

// Stop stops the memory leak detector
func (mld *MemoryLeakDetector) Stop() {
	mld.mu.Lock()
	defer mld.mu.Unlock()
	close(mld.stop)
}

// detect periodically checks for memory leaks
func (mld *MemoryLeakDetector) detect() {
	ticker := time.NewTicker(mld.interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			mld.checkLeaks()
		case <-mld.stop:
			return
		}
	}
}

// checkLeaks checks for memory leaks in all modules
func (mld *MemoryLeakDetector) checkLeaks() {
	modules := mld.isolation.GetAllModulesMemory()
	
	for moduleID, module := range modules {
		module.mu.RLock()
		allocated := module.Allocated
		allocationCount := len(module.Allocations)
		module.mu.RUnlock()
		
		if allocated > mld.threshold {
			// Trigger GC and check again
			runtime.GC()
			time.Sleep(100 * time.Millisecond)
			
			module.mu.RLock()
			afterGC := module.Allocated
			module.mu.RUnlock()
			
			if afterGC > mld.threshold {
				// Potential leak detected
				fmt.Printf("WARNING: Potential memory leak detected in module %s: %d bytes allocated, %d allocations\n",
					moduleID, afterGC, allocationCount)
			}
		}
	}
}

// CrashContainer provides crash containment for modules
type CrashContainer struct {
	modules map[string]*ModuleContainer
	mu      sync.RWMutex
}

// ModuleContainer contains a module and its crash recovery
type ModuleContainer struct {
	ModuleID      string
	RecoveryFunc  func(error)
	IsRecovering  bool
	CrashCount    int
	LastCrash     time.Time
	mu            sync.RWMutex
}

// NewCrashContainer creates a new crash container
func NewCrashContainer() *CrashContainer {
	return &CrashContainer{
		modules: make(map[string]*ModuleContainer),
	}
}

// RegisterModule registers a module for crash containment
func (cc *CrashContainer) RegisterModule(moduleID string, recoveryFunc func(error)) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	
	cc.modules[moduleID] = &ModuleContainer{
		ModuleID:     moduleID,
		RecoveryFunc: recoveryFunc,
		CrashCount:   0,
	}
}

// Execute executes a function with crash containment
func (cc *CrashContainer) Execute(moduleID string, fn func() error) error {
	cc.mu.RLock()
	container, ok := cc.modules[moduleID]
	cc.mu.RUnlock()
	
	if !ok {
		// No container, execute directly
		return fn()
	}
	
	// Execute with panic recovery
	defer func() {
		if r := recover(); r != nil {
			container.mu.Lock()
			container.CrashCount++
			container.LastCrash = time.Now()
			container.IsRecovering = true
			container.mu.Unlock()
			
			var err error
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("panic: %v", r)
			}
			
			// Call recovery function
			if container.RecoveryFunc != nil {
				container.RecoveryFunc(err)
			}
			
			container.mu.Lock()
			container.IsRecovering = false
			container.mu.Unlock()
		}
	}()
	
	return fn()
}

// GetModuleStatus gets the crash status for a module
func (cc *CrashContainer) GetModuleStatus(moduleID string) (*ModuleContainer, bool) {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	container, ok := cc.modules[moduleID]
	return container, ok
}

