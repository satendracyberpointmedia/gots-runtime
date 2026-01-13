package observability

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"
)

// Profiler provides production-grade profiling
type Profiler struct {
	enabled     bool
	profileType ProfileType
	outputPath  string
	mu          sync.RWMutex
}

// ProfileType represents the type of profile
type ProfileType int

const (
	ProfileTypeCPU ProfileType = iota
	ProfileTypeMemory
	ProfileTypeGoroutine
	ProfileTypeBlock
	ProfileTypeMutex
)

// NewProfiler creates a new profiler
func NewProfiler() *Profiler {
	return &Profiler{
		enabled: false,
	}
}

// StartCPUProfile starts CPU profiling
func (p *Profiler) StartCPUProfile(outputPath string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.enabled {
		return fmt.Errorf("profiling already in progress")
	}
	
	file, err := createProfileFile(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create profile file: %w", err)
	}
	
	if err := pprof.StartCPUProfile(file); err != nil {
		file.Close()
		return fmt.Errorf("failed to start CPU profile: %w", err)
	}
	
	p.enabled = true
	p.profileType = ProfileTypeCPU
	p.outputPath = outputPath
	return nil
}

// StopCPUProfile stops CPU profiling
func (p *Profiler) StopCPUProfile() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if !p.enabled || p.profileType != ProfileTypeCPU {
		return fmt.Errorf("CPU profiling not active")
	}
	
	pprof.StopCPUProfile()
	p.enabled = false
	return nil
}

// WriteHeapProfile writes a heap profile
func (p *Profiler) WriteHeapProfile(outputPath string) error {
	file, err := createProfileFile(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create profile file: %w", err)
	}
	defer file.Close()
	
	runtime.GC() // Force GC before profiling
	return pprof.WriteHeapProfile(file)
}

// WriteGoroutineProfile writes a goroutine profile
func (p *Profiler) WriteGoroutineProfile(outputPath string) error {
	return p.writeProfile("goroutine", outputPath)
}

// WriteBlockProfile writes a block profile
func (p *Profiler) WriteBlockProfile(outputPath string) error {
	return p.writeProfile("block", outputPath)
}

// WriteMutexProfile writes a mutex profile
func (p *Profiler) WriteMutexProfile(outputPath string) error {
	return p.writeProfile("mutex", outputPath)
}

// writeProfile writes a named profile
func (p *Profiler) writeProfile(name, outputPath string) error {
	file, err := createProfileFile(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create profile file: %w", err)
	}
	defer file.Close()
	
	profile := pprof.Lookup(name)
	if profile == nil {
		return fmt.Errorf("profile not found: %s", name)
	}
	
	return profile.WriteTo(file, 0)
}

// ProfileSnapshot represents a profiling snapshot
type ProfileSnapshot struct {
	Timestamp    time.Time
	CPUUsage     float64
	MemoryUsage  uint64
	GoroutineCount int
	HeapAlloc    uint64
	HeapSys      uint64
}

// TakeSnapshot takes a profiling snapshot
func (p *Profiler) TakeSnapshot() *ProfileSnapshot {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return &ProfileSnapshot{
		Timestamp:     time.Now(),
		MemoryUsage:   m.Alloc,
		GoroutineCount: runtime.NumGoroutine(),
		HeapAlloc:      m.HeapAlloc,
		HeapSys:        m.HeapSys,
	}
}

// ContinuousProfiler continuously profiles the application
type ContinuousProfiler struct {
	profiler    *Profiler
	interval    time.Duration
	snapshots   []*ProfileSnapshot
	maxSnapshots int
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	mu          sync.RWMutex
}

// NewContinuousProfiler creates a new continuous profiler
func NewContinuousProfiler(profiler *Profiler, interval time.Duration, maxSnapshots int) *ContinuousProfiler {
	ctx, cancel := context.WithCancel(context.Background())
	return &ContinuousProfiler{
		profiler:     profiler,
		interval:     interval,
		snapshots:    make([]*ProfileSnapshot, 0),
		maxSnapshots: maxSnapshots,
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Start starts continuous profiling
func (cp *ContinuousProfiler) Start() {
	cp.wg.Add(1)
	go cp.profile()
}

// Stop stops continuous profiling
func (cp *ContinuousProfiler) Stop() {
	cp.cancel()
	cp.wg.Wait()
}

// profile continuously takes snapshots
func (cp *ContinuousProfiler) profile() {
	defer cp.wg.Done()
	
	ticker := time.NewTicker(cp.interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			snapshot := cp.profiler.TakeSnapshot()
			cp.addSnapshot(snapshot)
		case <-cp.ctx.Done():
			return
		}
	}
}

// addSnapshot adds a snapshot
func (cp *ContinuousProfiler) addSnapshot(snapshot *ProfileSnapshot) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	
	cp.snapshots = append(cp.snapshots, snapshot)
	if len(cp.snapshots) > cp.maxSnapshots {
		cp.snapshots = cp.snapshots[1:]
	}
}

// GetSnapshots returns all snapshots
func (cp *ContinuousProfiler) GetSnapshots() []*ProfileSnapshot {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	
	result := make([]*ProfileSnapshot, len(cp.snapshots))
	copy(result, cp.snapshots)
	return result
}

// GetStats returns profiling statistics
func (cp *ContinuousProfiler) GetStats() *ProfilingStats {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	
	if len(cp.snapshots) == 0 {
		return &ProfilingStats{}
	}
	
	stats := &ProfilingStats{
		SnapshotCount: len(cp.snapshots),
	}
	
	var totalMemory, totalHeapAlloc, totalHeapSys uint64
	for _, snap := range cp.snapshots {
		totalMemory += snap.MemoryUsage
		totalHeapAlloc += snap.HeapAlloc
		totalHeapSys += snap.HeapSys
		if snap.GoroutineCount > stats.MaxGoroutines {
			stats.MaxGoroutines = snap.GoroutineCount
		}
	}
	
	count := uint64(len(cp.snapshots))
	stats.AvgMemoryUsage = totalMemory / count
	stats.AvgHeapAlloc = totalHeapAlloc / count
	stats.AvgHeapSys = totalHeapSys / count
	
	return stats
}

// ProfilingStats represents profiling statistics
type ProfilingStats struct {
	SnapshotCount  int
	MaxGoroutines  int
	AvgMemoryUsage uint64
	AvgHeapAlloc   uint64
	AvgHeapSys     uint64
}

// Helper function to create profile file
func createProfileFile(path string) (*os.File, error) {
	return os.Create(path)
}

