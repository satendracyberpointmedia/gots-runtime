package api

import (
	"gots-runtime/internal/security"
)

// SecureEnv provides environment variable operations with security
type SecureEnv struct {
	env         *Env
	permManager *security.PermissionManager
	moduleID    string
}

// NewSecureEnv creates a new secure environment API
func NewSecureEnv(permManager *security.PermissionManager, moduleID string) *SecureEnv {
	return &SecureEnv{
		env:         NewEnv(),
		permManager: permManager,
		moduleID:    moduleID,
	}
}

// Get gets an environment variable with permission check
func (se *SecureEnv) Get(key string) (string, error) {
	// Check permission
	if err := se.permManager.CheckPermission(se.moduleID, security.PermissionEnvRead); err != nil {
		return "", err
	}
	
	return se.env.Get(key), nil
}

// Set sets an environment variable with permission check
func (se *SecureEnv) Set(key, value string) error {
	// Check permission
	if err := se.permManager.CheckPermission(se.moduleID, security.PermissionEnvWrite); err != nil {
		return err
	}
	
	return se.env.Set(key, value)
}

// Unset unsets an environment variable with permission check
func (se *SecureEnv) Unset(key string) error {
	// Check permission
	if err := se.permManager.CheckPermission(se.moduleID, security.PermissionEnvWrite); err != nil {
		return err
	}
	
	return se.env.Unset(key)
}

// LookupEnv looks up an environment variable with permission check
func (se *SecureEnv) LookupEnv(key string) (string, bool, error) {
	// Check permission
	if err := se.permManager.CheckPermission(se.moduleID, security.PermissionEnvRead); err != nil {
		return "", false, err
	}
	
	value, ok := se.env.LookupEnv(key)
	return value, ok, nil
}

