package security

import (
	"fmt"
	"sync"
)

// Permission represents a permission type
type Permission string

const (
	PermissionFSRead   Permission = "fs:read"
	PermissionFSWrite Permission = "fs:write"
	PermissionNetDial  Permission = "net:dial"
	PermissionNetListen Permission = "net:listen"
	PermissionEnvRead  Permission = "env:read"
	PermissionEnvWrite Permission = "env:write"
	PermissionAll      Permission = "*"
)

// PermissionSet represents a set of permissions
type PermissionSet struct {
	permissions map[Permission]bool
	mu         sync.RWMutex
}

// NewPermissionSet creates a new permission set
func NewPermissionSet(permissions ...Permission) *PermissionSet {
	ps := &PermissionSet{
		permissions: make(map[Permission]bool),
	}
	for _, perm := range permissions {
		ps.permissions[perm] = true
	}
	return ps
}

// Has checks if a permission is granted
func (ps *PermissionSet) Has(permission Permission) bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	
	// Check for wildcard permission
	if ps.permissions[PermissionAll] {
		return true
	}
	
	return ps.permissions[permission]
}

// Add adds a permission
func (ps *PermissionSet) Add(permission Permission) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.permissions[permission] = true
}

// Remove removes a permission
func (ps *PermissionSet) Remove(permission Permission) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	delete(ps.permissions, permission)
}

// GetAll returns all permissions
func (ps *PermissionSet) GetAll() []Permission {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	
	perms := make([]Permission, 0, len(ps.permissions))
	for perm := range ps.permissions {
		perms = append(perms, perm)
	}
	return perms
}

// Policy represents a security policy
type Policy struct {
	moduleID    string
	permissions *PermissionSet
	restrictions map[string]interface{}
	mu          sync.RWMutex
}

// NewPolicy creates a new security policy
func NewPolicy(moduleID string) *Policy {
	return &Policy{
		moduleID:    moduleID,
		permissions: NewPermissionSet(),
		restrictions: make(map[string]interface{}),
	}
}

// Allow grants a permission
func (p *Policy) Allow(permission Permission) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.permissions.Add(permission)
}

// Deny denies a permission
func (p *Policy) Deny(permission Permission) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.permissions.Remove(permission)
}

// Check checks if a permission is allowed
func (p *Policy) Check(permission Permission) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.permissions.Has(permission)
}

// SetRestriction sets a restriction
func (p *Policy) SetRestriction(key string, value interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.restrictions[key] = value
}

// GetRestriction gets a restriction
func (p *Policy) GetRestriction(key string) (interface{}, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	value, ok := p.restrictions[key]
	return value, ok
}

// PermissionManager manages permissions for modules
type PermissionManager struct {
	policies map[string]*Policy
	mu       sync.RWMutex
}

// NewPermissionManager creates a new permission manager
func NewPermissionManager() *PermissionManager {
	return &PermissionManager{
		policies: make(map[string]*Policy),
	}
}

// RegisterPolicy registers a policy for a module
func (pm *PermissionManager) RegisterPolicy(moduleID string, policy *Policy) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.policies[moduleID] = policy
}

// GetPolicy gets a policy for a module
func (pm *PermissionManager) GetPolicy(moduleID string) (*Policy, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	policy, ok := pm.policies[moduleID]
	return policy, ok
}

// CheckPermission checks if a module has a permission
func (pm *PermissionManager) CheckPermission(moduleID string, permission Permission) error {
	pm.mu.RLock()
	policy, ok := pm.policies[moduleID]
	pm.mu.RUnlock()
	
	if !ok {
		// Default: deny all if no policy
		return &PermissionError{
			ModuleID:   moduleID,
			Permission: permission,
			Message:    "no policy found for module",
		}
	}
	
	if !policy.Check(permission) {
		return &PermissionError{
			ModuleID:   moduleID,
			Permission: permission,
			Message:    "permission denied",
		}
	}
	
	return nil
}

// PermissionError represents a permission error
type PermissionError struct {
	ModuleID   string
	Permission Permission
	Message    string
}

func (e *PermissionError) Error() string {
	return fmt.Sprintf("permission denied: module %s does not have permission %s: %s", 
		e.ModuleID, e.Permission, e.Message)
}

