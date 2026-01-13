package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gots-runtime/internal/security"
)

// ProjectConfig represents the project configuration
type ProjectConfig struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Main        string                 `json:"main,omitempty"`
	Permissions []PermissionConfig     `json:"permissions,omitempty"`
	Observability *ObservabilityConfig `json:"observability,omitempty"`
	Runtime     *RuntimeConfig         `json:"runtime,omitempty"`
	Modules     []ModuleConfig         `json:"modules,omitempty"`
}

// PermissionConfig represents module permissions
type PermissionConfig struct {
	Module      string   `json:"module"`
	Permissions []string `json:"permissions"`
}

// ObservabilityConfig represents observability settings
type ObservabilityConfig struct {
	Enabled      bool   `json:"enabled"`
	HealthPort   int    `json:"healthPort,omitempty"`
	MetricsPort  int    `json:"metricsPort,omitempty"`
	LogLevel     string `json:"logLevel,omitempty"`
	EnableTracing bool  `json:"enableTracing,omitempty"`
}

// RuntimeConfig represents runtime settings
type RuntimeConfig struct {
	SandboxMode      string `json:"sandboxMode,omitempty"`
	MaxWorkers       int    `json:"maxWorkers,omitempty"`
	EventQueueSize   int    `json:"eventQueueSize,omitempty"`
	EnableHotReload  bool   `json:"enableHotReload,omitempty"`
	TypeEnforcement  bool   `json:"typeEnforcement,omitempty"`
}

// ModuleConfig represents module configuration
type ModuleConfig struct {
	ID          string   `json:"id"`
	Path        string   `json:"path"`
	Permissions []string `json:"permissions,omitempty"`
	Sandbox     bool     `json:"sandbox,omitempty"`
}

// LoadConfig loads configuration from a file
func LoadConfig(configPath string) (*ProjectConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	var config ProjectConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	
	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	
	return &config, nil
}

// FindConfig searches for config file in directory and parent directories
func FindConfig(startDir string) (string, error) {
	dir := startDir
	for {
		configPath := filepath.Join(dir, "gots.json")
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}
		
		parent := filepath.Dir(dir)
		if parent == dir {
			break // Reached root
		}
		dir = parent
	}
	
	return "", fmt.Errorf("config file not found")
}

// SaveConfig saves configuration to a file
func SaveConfig(config *ProjectConfig, configPath string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}

// Validate validates the configuration
func (c *ProjectConfig) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("project name is required")
	}
	
	// Validate permissions
	for _, perm := range c.Permissions {
		if perm.Module == "" {
			return fmt.Errorf("permission module name is required")
		}
		for _, p := range perm.Permissions {
			if !isValidPermission(p) {
				return fmt.Errorf("invalid permission: %s", p)
			}
		}
	}
	
	// Validate modules
	for _, mod := range c.Modules {
		if mod.ID == "" {
			return fmt.Errorf("module ID is required")
		}
		if mod.Path == "" {
			return fmt.Errorf("module path is required")
		}
	}
	
	return nil
}

// GetDefaultConfig returns a default configuration
func GetDefaultConfig() *ProjectConfig {
	return &ProjectConfig{
		Name:    "my-gots-project",
		Version: "0.1.0",
		Main:    "main.ts",
		Observability: &ObservabilityConfig{
			Enabled:      true,
			HealthPort:   8080,
			MetricsPort:  9090,
			LogLevel:     "info",
			EnableTracing: true,
		},
		Runtime: &RuntimeConfig{
			SandboxMode:     "none",
			MaxWorkers:      10,
			EventQueueSize:  1000,
			EnableHotReload: false,
			TypeEnforcement: true,
		},
	}
}

// ToSecurityPermissions converts permission strings to security.Permission
func (pc *PermissionConfig) ToSecurityPermissions() []security.Permission {
	perms := make([]security.Permission, 0, len(pc.Permissions))
	for _, p := range pc.Permissions {
		perms = append(perms, security.Permission(p))
	}
	return perms
}

// isValidPermission checks if a permission string is valid
func isValidPermission(perm string) bool {
	validPerms := []string{
		string(security.PermissionFSRead),
		string(security.PermissionFSWrite),
		string(security.PermissionNetDial),
		string(security.PermissionNetListen),
		string(security.PermissionEnvRead),
		string(security.PermissionEnvWrite),
		string(security.PermissionAll),
	}
	
	for _, vp := range validPerms {
		if perm == vp {
			return true
		}
	}
	return false
}

