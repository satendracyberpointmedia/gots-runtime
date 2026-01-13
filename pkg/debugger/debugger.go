package debugger

import (
	"context"
	"sync"
)

// Debugger represents a runtime debugger
type Debugger struct {
	breakpoints map[string][]int // file -> line numbers
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.RWMutex
}

// NewDebugger creates a new debugger
func NewDebugger(ctx context.Context) *Debugger {
	debugCtx, cancel := context.WithCancel(ctx)
	return &Debugger{
		breakpoints: make(map[string][]int),
		ctx:         debugCtx,
		cancel:      cancel,
	}
}

// SetBreakpoint sets a breakpoint at a line in a file
func (d *Debugger) SetBreakpoint(file string, line int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	if lines, ok := d.breakpoints[file]; ok {
		// Check if breakpoint already exists
		for _, l := range lines {
			if l == line {
				return
			}
		}
		d.breakpoints[file] = append(lines, line)
	} else {
		d.breakpoints[file] = []int{line}
	}
}

// RemoveBreakpoint removes a breakpoint
func (d *Debugger) RemoveBreakpoint(file string, line int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	lines, ok := d.breakpoints[file]
	if !ok {
		return
	}

	newLines := make([]int, 0, len(lines))
	for _, l := range lines {
		if l != line {
			newLines = append(newLines, l)
		}
	}
	d.breakpoints[file] = newLines
}

// HasBreakpoint checks if there's a breakpoint at a line
func (d *Debugger) HasBreakpoint(file string, line int) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	lines, ok := d.breakpoints[file]
	if !ok {
		return false
	}

	for _, l := range lines {
		if l == line {
			return true
		}
	}
	return false
}

// GetBreakpoints returns all breakpoints for a file
func (d *Debugger) GetBreakpoints(file string) []int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.breakpoints[file]
}

// ClearBreakpoints clears all breakpoints
func (d *Debugger) ClearBreakpoints() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.breakpoints = make(map[string][]int)
}

// Stop stops the debugger
func (d *Debugger) Stop() {
	d.cancel()
}

// DebuggerError represents a debugger error
type DebuggerError struct {
	Message string
}

func (e *DebuggerError) Error() string {
	return e.Message
}

