package security

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// SandboxMode represents the sandbox execution mode
type SandboxMode int

const (
	SandboxModeNone SandboxMode = iota
	SandboxModeStrict
	SandboxModeDeterministic
)

// Sandbox represents a sandboxed execution environment
type Sandbox struct {
	id          string
	mode        SandboxMode
	permissions *PermissionSet
	timeout     time.Duration
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.RWMutex
}

// NewSandbox creates a new sandbox
func NewSandbox(id string, mode SandboxMode) *Sandbox {
	ctx, cancel := context.WithCancel(context.Background())
	return &Sandbox{
		id:          id,
		mode:        mode,
		permissions: NewPermissionSet(),
		timeout:     0, // No timeout by default
		ctx:         ctx,
		cancel:      cancel,
	}
}

// SetTimeout sets the execution timeout
func (s *Sandbox) SetTimeout(timeout time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.timeout = timeout
}

// SetPermissions sets the permissions for the sandbox
func (s *Sandbox) SetPermissions(permissions *PermissionSet) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.permissions = permissions
}

// CheckPermission checks if the sandbox has a permission
func (s *Sandbox) CheckPermission(permission Permission) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.mode == SandboxModeNone {
		return nil // No restrictions
	}
	
	if !s.permissions.Has(permission) {
		return &SandboxError{
			SandboxID:  s.id,
			Permission: permission,
			Message:    "permission denied in sandbox",
		}
	}
	
	return nil
}

// Context returns the sandbox context
func (s *Sandbox) Context() context.Context {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ctx
}

// Stop stops the sandbox
func (s *Sandbox) Stop() {
	s.cancel()
}

// IsDeterministic returns true if sandbox is in deterministic mode
func (s *Sandbox) IsDeterministic() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mode == SandboxModeDeterministic
}

// SandboxError represents a sandbox error
type SandboxError struct {
	SandboxID  string
	Permission Permission
	Message    string
}

func (e *SandboxError) Error() string {
	return fmt.Sprintf("sandbox error: %s (sandbox: %s, permission: %s)", 
		e.Message, e.SandboxID, e.Permission)
}

// SandboxManager manages sandboxes
type SandboxManager struct {
	sandboxes map[string]*Sandbox
	mu        sync.RWMutex
}

// NewSandboxManager creates a new sandbox manager
func NewSandboxManager() *SandboxManager {
	return &SandboxManager{
		sandboxes: make(map[string]*Sandbox),
	}
}

// CreateSandbox creates a new sandbox
func (sm *SandboxManager) CreateSandbox(id string, mode SandboxMode) *Sandbox {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	sandbox := NewSandbox(id, mode)
	sm.sandboxes[id] = sandbox
	return sandbox
}

// GetSandbox gets a sandbox by ID
func (sm *SandboxManager) GetSandbox(id string) (*Sandbox, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	sandbox, ok := sm.sandboxes[id]
	return sandbox, ok
}

// RemoveSandbox removes a sandbox
func (sm *SandboxManager) RemoveSandbox(id string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if sandbox, ok := sm.sandboxes[id]; ok {
		sandbox.Stop()
		delete(sm.sandboxes, id)
	}
}

// ExecuteInSandbox executes code in a sandbox
func (sm *SandboxManager) ExecuteInSandbox(sandboxID string, fn func() error) error {
	sandbox, ok := sm.GetSandbox(sandboxID)
	if !ok {
		return fmt.Errorf("sandbox not found: %s", sandboxID)
	}
	
	ctx := sandbox.Context()
	if sandbox.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, sandbox.timeout)
		defer cancel()
	}
	
	// Execute in goroutine with context
	done := make(chan error, 1)
	go func() {
		done <- fn()
	}()
	
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("sandbox execution timeout: %s", sandboxID)
	}
}

