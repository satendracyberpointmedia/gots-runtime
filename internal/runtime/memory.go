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
	isolation  *MemoryIsolation
	threshold  uint64
	interval   time.Duration
	stop       chan struct{}
	leaks      map[string]LeakReport
	mu         sync.RWMutex
	autoRepair bool
}

// LeakReport contains information about a detected leak
type LeakReport struct {
	ModuleID        string
	AllocatedBytes  uint64
	AllocationCount int
	Timestamp       time.Time
	StackTrace      string
	Severity        string // "warning", "critical"
}

// NewMemoryLeakDetector creates a new memory leak detector
func NewMemoryLeakDetector(isolation *MemoryIsolation, threshold uint64, interval time.Duration) *MemoryLeakDetector {
	return &MemoryLeakDetector{
		isolation:  isolation,
		threshold:  threshold,
		interval:   interval,
		stop:       make(chan struct{}),
		leaks:      make(map[string]LeakReport),
		autoRepair: true,
	}
}

// SetAutoRepair enables or disables automatic repair
func (mld *MemoryLeakDetector) SetAutoRepair(enabled bool) {
	mld.mu.Lock()
	defer mld.mu.Unlock()
	mld.autoRepair = enabled
}

// Start starts the memory leak detector
func (mld *MemoryLeakDetector) Start() {
	go mld.detect()
}

// Stop stops the memory leak detector
func (mld *MemoryLeakDetector) Stop() {
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
				var severity string
				if afterGC > mld.threshold*2 {
					severity = "critical"
				} else {
					severity = "warning"
				}

				report := LeakReport{
					ModuleID:        moduleID,
					AllocatedBytes:  afterGC,
					AllocationCount: allocationCount,
					Timestamp:       time.Now(),
					Severity:        severity,
				}

				mld.mu.Lock()
				mld.leaks[moduleID] = report
				mld.mu.Unlock()

				// Auto-repair if enabled and critical
				if mld.autoRepair && severity == "critical" {
					mld.attemptRepair(moduleID)
				}
			}
		}
	}
}

// attemptRepair attempts to repair a memory leak
func (mld *MemoryLeakDetector) attemptRepair(moduleID string) {
	// Force garbage collection
	runtime.GC()

	// Clear old allocations if possible
	module, err := mld.isolation.GetModuleMemory(moduleID)
	if err != nil {
		return
	}

	module.mu.Lock()
	defer module.mu.Unlock()

	// Log attempt
	fmt.Printf("REPAIR: Attempting automatic repair for module %s\n", moduleID)
}

// GetLeakReports returns all detected leak reports
func (mld *MemoryLeakDetector) GetLeakReports() map[string]LeakReport {
	mld.mu.RLock()
	defer mld.mu.RUnlock()

	result := make(map[string]LeakReport)
	for k, v := range mld.leaks {
		result[k] = v
	}
	return result
}

// GetLeakReport returns the leak report for a specific module
func (mld *MemoryLeakDetector) GetLeakReport(moduleID string) (LeakReport, bool) {
	mld.mu.RLock()
	defer mld.mu.RUnlock()
	report, ok := mld.leaks[moduleID]
	return report, ok
}

// ClearLeakReports clears all leak reports
func (mld *MemoryLeakDetector) ClearLeakReports() {
	mld.mu.Lock()
	defer mld.mu.Unlock()
	mld.leaks = make(map[string]LeakReport)
}

// CrashContainer provides crash containment for modules
type CrashContainer struct {
	modules       map[string]*ModuleContainer
	maxCrashes    int
	recoveryDelay time.Duration
	mu            sync.RWMutex
}

// ModuleContainer contains a module and its crash recovery
type ModuleContainer struct {
	ModuleID     string
	RecoveryFunc func(error)
	IsRecovering bool
	CrashCount   int
	LastCrash    time.Time
	Crashes      []CrashEvent
	MaxCrashes   int
	mu           sync.RWMutex
}

// CrashEvent represents a crash event
type CrashEvent struct {
	Timestamp  time.Time
	Error      error
	StackTrace string
}

// NewCrashContainer creates a new crash container
func NewCrashContainer() *CrashContainer {
	return &CrashContainer{
		modules:       make(map[string]*ModuleContainer),
		maxCrashes:    10, // Keep last 10 crashes
		recoveryDelay: 1 * time.Second,
	}
}

// SetMaxCrashes sets the maximum number of crashes to track
func (cc *CrashContainer) SetMaxCrashes(max int) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.maxCrashes = max
}

// SetRecoveryDelay sets the recovery delay
func (cc *CrashContainer) SetRecoveryDelay(delay time.Duration) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.recoveryDelay = delay
}

// RegisterModule registers a module for crash containment
func (cc *CrashContainer) RegisterModule(moduleID string, recoveryFunc func(error)) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.modules[moduleID] = &ModuleContainer{
		ModuleID:     moduleID,
		RecoveryFunc: recoveryFunc,
		CrashCount:   0,
		Crashes:      make([]CrashEvent, 0),
		MaxCrashes:   cc.maxCrashes,
	}
}

// UnregisterModule unregisters a module
func (cc *CrashContainer) UnregisterModule(moduleID string) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	delete(cc.modules, moduleID)
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

			// Record crash event
			container.mu.Lock()
			container.Crashes = append(container.Crashes, CrashEvent{
				Timestamp:  time.Now(),
				Error:      err,
				StackTrace: getStackTrace(),
			})

			// Keep only the last MaxCrashes
			if len(container.Crashes) > container.MaxCrashes {
				container.Crashes = container.Crashes[len(container.Crashes)-container.MaxCrashes:]
			}
			container.mu.Unlock()

			// Delay recovery
			time.Sleep(cc.recoveryDelay)

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

	if container, ok := cc.modules[moduleID]; ok {
		// Return a copy
		container.mu.RLock()
		defer container.mu.RUnlock()

		crashes := make([]CrashEvent, len(container.Crashes))
		copy(crashes, container.Crashes)

		return &ModuleContainer{
			ModuleID:     container.ModuleID,
			IsRecovering: container.IsRecovering,
			CrashCount:   container.CrashCount,
			LastCrash:    container.LastCrash,
			Crashes:      crashes,
		}, true
	}

	return nil, false
}

// IsCritical checks if a module is in critical state
func (cc *CrashContainer) IsCritical(moduleID string) bool {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	if container, ok := cc.modules[moduleID]; ok {
		container.mu.RLock()
		defer container.mu.RUnlock()
		// Critical if more than 5 crashes in last minute
		count := 0
		now := time.Now()
		for _, crash := range container.Crashes {
			if now.Sub(crash.Timestamp) < time.Minute {
				count++
			}
		}
		return count > 5
	}

	return false
}

// ClearCrashes clears crash history for a module
func (cc *CrashContainer) ClearCrashes(moduleID string) {
	cc.mu.RLock()
	container, ok := cc.modules[moduleID]
	cc.mu.RUnlock()

	if ok {
		container.mu.Lock()
		defer container.mu.Unlock()
		container.Crashes = make([]CrashEvent, 0)
		container.CrashCount = 0
	}
}

// Helper function to get stack trace
func getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}
