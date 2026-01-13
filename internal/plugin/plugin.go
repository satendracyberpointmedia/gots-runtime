package plugin

import (
	"fmt"
	"sync"
)

// Plugin represents a runtime plugin
type Plugin interface {
	Name() string
	Version() string
	Initialize(ctx *PluginContext) error
	Execute(ctx *PluginContext, args map[string]interface{}) (interface{}, error)
	Shutdown() error
}

// PluginContext provides context for plugin execution
type PluginContext struct {
	RuntimeID string
	Config    map[string]interface{}
	Logger    Logger
}

// Logger interface for plugins
type Logger interface {
	Info(format string, args ...interface{})
	Error(format string, args ...interface{})
}

// PluginManager manages plugins
type PluginManager struct {
	plugins map[string]Plugin
	mu      sync.RWMutex
}

// NewPluginManager creates a new plugin manager
func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make(map[string]Plugin),
	}
}

// Register registers a plugin
func (pm *PluginManager) Register(plugin Plugin) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	name := plugin.Name()
	if _, exists := pm.plugins[name]; exists {
		return fmt.Errorf("plugin already registered: %s", name)
	}
	
	pm.plugins[name] = plugin
	return nil
}

// Unregister unregisters a plugin
func (pm *PluginManager) Unregister(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	plugin, ok := pm.plugins[name]
	if !ok {
		return fmt.Errorf("plugin not found: %s", name)
	}
	
	if err := plugin.Shutdown(); err != nil {
		return fmt.Errorf("failed to shutdown plugin: %w", err)
	}
	
	delete(pm.plugins, name)
	return nil
}

// GetPlugin gets a plugin by name
func (pm *PluginManager) GetPlugin(name string) (Plugin, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	plugin, ok := pm.plugins[name]
	return plugin, ok
}

// ListPlugins lists all registered plugins
func (pm *PluginManager) ListPlugins() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	names := make([]string, 0, len(pm.plugins))
	for name := range pm.plugins {
		names = append(names, name)
	}
	return names
}

// Execute executes a plugin
func (pm *PluginManager) Execute(name string, ctx *PluginContext, args map[string]interface{}) (interface{}, error) {
	plugin, ok := pm.GetPlugin(name)
	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}
	
	return plugin.Execute(ctx, args)
}

// InitializeAll initializes all plugins
func (pm *PluginManager) InitializeAll(ctx *PluginContext) error {
	pm.mu.RLock()
	plugins := make([]Plugin, 0, len(pm.plugins))
	for _, plugin := range pm.plugins {
		plugins = append(plugins, plugin)
	}
	pm.mu.RUnlock()
	
	for _, plugin := range plugins {
		if err := plugin.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize plugin %s: %w", plugin.Name(), err)
		}
	}
	
	return nil
}

// ShutdownAll shuts down all plugins
func (pm *PluginManager) ShutdownAll() error {
	pm.mu.RLock()
	plugins := make([]Plugin, 0, len(pm.plugins))
	for _, plugin := range pm.plugins {
		plugins = append(plugins, plugin)
	}
	pm.mu.RUnlock()
	
	var firstErr error
	for _, plugin := range plugins {
		if err := plugin.Shutdown(); err != nil {
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	
	return firstErr
}

