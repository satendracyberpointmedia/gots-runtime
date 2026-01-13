package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gots-runtime/internal/config"
	"gots-runtime/internal/observability"
	"gots-runtime/internal/runtime"
	"gots-runtime/internal/security"
)

// RuntimeManager manages the runtime integration for CLI
type RuntimeManager struct {
	integration *runtime.RuntimeIntegration
	config      *config.ProjectConfig
	autoConfig  *observability.AutoConfig
	projectRoot string
}

// NewRuntimeManager creates a new runtime manager
func NewRuntimeManager(projectRoot string) (*RuntimeManager, error) {
	// Try to load config
	var cfg *config.ProjectConfig
	configPath, err := config.FindConfig(projectRoot)
	if err == nil {
		cfg, err = config.LoadConfig(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	} else {
		// Use default config
		cfg = config.GetDefaultConfig()
	}
	
	// Create runtime integration
	integration := runtime.NewRuntimeIntegration()
	
	// Initialize runtime
	if err := integration.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize runtime: %w", err)
	}
	
	// Create auto-config for observability
	autoConfig := observability.NewAutoConfig()
	if cfg.Observability != nil && cfg.Observability.Enabled {
		if err := autoConfig.Setup(); err != nil {
			return nil, fmt.Errorf("failed to setup observability: %w", err)
		}
		
		// Start health server if configured
		if cfg.Observability.HealthPort > 0 {
			addr := fmt.Sprintf(":%d", cfg.Observability.HealthPort)
			if err := autoConfig.StartHealthServer(addr); err != nil {
				return nil, fmt.Errorf("failed to start health server: %w", err)
			}
		}
	}
	
	// Register modules with permissions
	if err := registerModules(integration, cfg); err != nil {
		return nil, fmt.Errorf("failed to register modules: %w", err)
	}
	
	return &RuntimeManager{
		integration: integration,
		config:      cfg,
		autoConfig:  autoConfig,
		projectRoot: projectRoot,
	}, nil
}

// registerModules registers modules with their permissions
func registerModules(integration *runtime.RuntimeIntegration, cfg *config.ProjectConfig) error {
	// Register permissions from config
	for _, permConfig := range cfg.Permissions {
		perms := permConfig.ToSecurityPermissions()
		if err := integration.RegisterModule(permConfig.Module, perms...); err != nil {
			return fmt.Errorf("failed to register module %s: %w", permConfig.Module, err)
		}
	}
	
	// Register modules from config
	for _, modConfig := range cfg.Modules {
		var perms []security.Permission
		for _, p := range modConfig.Permissions {
			perms = append(perms, security.Permission(p))
		}
		
		if err := integration.RegisterModule(modConfig.ID, perms...); err != nil {
			return fmt.Errorf("failed to register module %s: %w", modConfig.ID, err)
		}
	}
	
	return nil
}

// GetIntegration returns the runtime integration
func (rm *RuntimeManager) GetIntegration() *runtime.RuntimeIntegration {
	return rm.integration
}

// GetConfig returns the project config
func (rm *RuntimeManager) GetConfig() *config.ProjectConfig {
	return rm.config
}

// ExecuteModule executes a TypeScript module
func (rm *RuntimeManager) ExecuteModule(moduleID, filePath string) error {
	// Get absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}
	
	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", absPath)
	}
	
	// Execute module
	return rm.integration.ExecuteModule(moduleID, absPath)
}

// Shutdown shuts down the runtime
func (rm *RuntimeManager) Shutdown() error {
	if rm.autoConfig != nil {
		if err := rm.autoConfig.Stop(); err != nil {
			return fmt.Errorf("failed to stop observability: %w", err)
		}
	}
	
	if rm.integration != nil {
		if err := rm.integration.Shutdown(); err != nil {
			return fmt.Errorf("failed to shutdown runtime: %w", err)
		}
	}
	
	return nil
}

// GetProjectRoot returns the project root directory
func (rm *RuntimeManager) GetProjectRoot() string {
	return rm.projectRoot
}

