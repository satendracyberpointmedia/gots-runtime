package debugger

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
)

// BreakpointInfo stores breakpoint information
type BreakpointInfo struct {
	ID       int
	File     string
	Line     int
	Enabled  bool
	HitCount int
}

// WatchExpression represents a watched variable
type WatchExpression struct {
	ID         int
	Expression string
	LastValue  interface{}
}

// Debugger represents a runtime debugger
type Debugger struct {
	breakpoints map[string][]int // file -> line numbers
	watches     map[int]*WatchExpression
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.RWMutex
	IsRunning   bool
	CurrentLine int
	CurrentFile string
	NextBPID    int
	NextWatchID int
	Variables   map[string]interface{}
}

// NewDebugger creates a new debugger
func NewDebugger(ctx context.Context) *Debugger {
	debugCtx, cancel := context.WithCancel(ctx)
	return &Debugger{
		breakpoints: make(map[string][]int),
		watches:     make(map[int]*WatchExpression),
		ctx:         debugCtx,
		cancel:      cancel,
		Variables:   make(map[string]interface{}),
		NextBPID:    1,
		NextWatchID: 1,
	}
}

// SetBreakpoint sets a breakpoint at a line in a file
func (d *Debugger) SetBreakpoint(file string, line int) int {
	d.mu.Lock()
	defer d.mu.Unlock()

	if lines, ok := d.breakpoints[file]; ok {
		// Check if breakpoint already exists
		for _, l := range lines {
			if l == line {
				return 0
			}
		}
		d.breakpoints[file] = append(lines, line)
	} else {
		d.breakpoints[file] = []int{line}
	}
	return d.NextBPID
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

// AddWatch adds a watch expression
func (d *Debugger) AddWatch(expr string) int {
	d.mu.Lock()
	defer d.mu.Unlock()

	watchID := d.NextWatchID
	d.watches[watchID] = &WatchExpression{
		ID:         watchID,
		Expression: expr,
	}
	d.NextWatchID++
	return watchID
}

// RemoveWatch removes a watch expression
func (d *Debugger) RemoveWatch(id int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.watches, id)
}

// GetWatches returns all watch expressions
func (d *Debugger) GetWatches() map[int]*WatchExpression {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.watches
}

// SetVariable sets a variable for inspection
func (d *Debugger) SetVariable(name string, value interface{}) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Variables[name] = value
}

// GetVariable gets a variable value
func (d *Debugger) GetVariable(name string) (interface{}, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	val, ok := d.Variables[name]
	return val, ok
}

// InteractiveMode starts interactive debugging
func (d *Debugger) InteractiveMode() error {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("(gdb) ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmd := parts[0]

		switch cmd {
		case "continue", "c":
			fmt.Println("Continuing execution...")
			d.IsRunning = true
			return nil
		case "step", "s":
			d.CurrentLine++
			fmt.Printf("Stepped to line %d\n", d.CurrentLine)
		case "break", "b":
			if len(parts) < 2 {
				fmt.Println("Usage: break <line>")
				continue
			}
			var lineNum int
			fmt.Sscanf(parts[1], "%d", &lineNum)
			bpID := d.SetBreakpoint(d.CurrentFile, lineNum)
			fmt.Printf("Breakpoint %d set at line %d\n", bpID, lineNum)
		case "watch", "w":
			if len(parts) < 2 {
				fmt.Println("Usage: watch <variable>")
				continue
			}
			watchID := d.AddWatch(parts[1])
			fmt.Printf("Watch %d added for '%s'\n", watchID, parts[1])
		case "delete", "d":
			if len(parts) < 2 {
				fmt.Println("Usage: delete <breakpoint_id>")
				continue
			}
			// Simple implementation - remove all breakpoints
			d.ClearBreakpoints()
			fmt.Println("All breakpoints deleted")
		case "info":
			if len(parts) > 1 && parts[1] == "break" {
				d.printBreakpoints()
			} else {
				d.printInfo()
			}
		case "print", "p":
			if len(parts) < 2 {
				fmt.Println("Usage: print <variable>")
				continue
			}
			if val, ok := d.GetVariable(parts[1]); ok {
				fmt.Printf("%s = %v\n", parts[1], val)
			} else {
				fmt.Printf("Variable '%s' not found\n", parts[1])
			}
		case "quit", "q", "exit":
			fmt.Println("Exiting debugger")
			return fmt.Errorf("debugger exited")
		case "help", "h":
			d.printHelp()
		default:
			fmt.Printf("Unknown command: %s\n", cmd)
		}
	}

	return nil
}

func (d *Debugger) printBreakpoints() {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if len(d.breakpoints) == 0 {
		fmt.Println("No breakpoints set")
		return
	}

	for file, lines := range d.breakpoints {
		for _, line := range lines {
			fmt.Printf("  Breakpoint at %s:%d\n", file, line)
		}
	}
}

func (d *Debugger) printInfo() {
	fmt.Printf("File: %s\n", d.CurrentFile)
	fmt.Printf("Line: %d\n", d.CurrentLine)
	fmt.Printf("Running: %v\n", d.IsRunning)
}

func (d *Debugger) printHelp() {
	fmt.Println(`
Debugger Commands:
  continue (c)    - Continue execution
  step (s)        - Execute next line
  break (b) <n>   - Set breakpoint at line n
  watch (w) <var> - Watch variable
  delete (d) <id> - Delete breakpoint
  info            - Show debug info
  print (p) <var> - Print variable
  quit (q)        - Exit debugger
  help (h)        - Show this help`)
}

// Stop stops the debugger
func (d *Debugger) Stop() {
	d.cancel()
}
