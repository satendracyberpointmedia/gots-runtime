package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// PluginManifest represents a plugin manifest
type PluginManifest struct {
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Description  string                 `json:"description"`
	Author       string                 `json:"author"`
	License      string                 `json:"license"`
	EntryPoint   string                 `json:"entryPoint"`
	Config       map[string]interface{} `json:"config"`
	Hooks        []string               `json:"hooks"`
	Capabilities []string               `json:"capabilities"`
}

// LoadedPlugin represents a loaded plugin with metadata
type LoadedPlugin struct {
	Name      string
	Path      string
	Manifest  *PluginManifest
	Plugin    Plugin
	Loaded    bool
	LoadedAt  time.Time
	LastError error
}

// PluginLoader loads plugins from filesystem
type PluginLoader struct {
	pluginDir  string
	plugins    map[string]*LoadedPlugin
	mu         sync.RWMutex
	searchPath []string
}

// NewPluginLoader creates a new plugin loader
func NewPluginLoader(pluginDir string) *PluginLoader {
	return &PluginLoader{
		pluginDir:  pluginDir,
		plugins:    make(map[string]*LoadedPlugin),
		searchPath: []string{pluginDir},
	}
}

// AddSearchPath adds a search path for plugins
func (pl *PluginLoader) AddSearchPath(path string) {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	pl.searchPath = append(pl.searchPath, path)
}

// LoadManifest loads a plugin manifest
func (pl *PluginLoader) LoadManifest(pluginPath string) (*PluginManifest, error) {
	manifestPath := filepath.Join(pluginPath, "plugin.json")

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest PluginManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return &manifest, nil
}

// DiscoverPlugins discovers plugins in the plugin directory
func (pl *PluginLoader) DiscoverPlugins() ([]string, error) {
	var plugins []string

	err := filepath.Walk(pl.pluginDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			manifestPath := filepath.Join(path, "plugin.json")
			if _, err := os.Stat(manifestPath); err == nil {
				plugins = append(plugins, path)
			}
		}

		return nil
	})

	return plugins, err
}

// LoadPlugin loads a plugin by name
func (pl *PluginLoader) LoadPlugin(name string) (*LoadedPlugin, error) {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	// Check if already loaded
	if loaded, ok := pl.plugins[name]; ok && loaded.Loaded {
		return loaded, nil
	}

	// Find plugin in search paths
	var pluginPath string
	for _, searchPath := range pl.searchPath {
		candidate := filepath.Join(searchPath, name)
		if _, err := os.Stat(candidate); err == nil {
			pluginPath = candidate
			break
		}
	}

	if pluginPath == "" {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}

	// Load manifest
	manifest, err := pl.LoadManifest(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load manifest for %s: %w", name, err)
	}

	// Create loaded plugin entry
	loadedPlugin := &LoadedPlugin{
		Name:     name,
		Path:     pluginPath,
		Manifest: manifest,
		Loaded:   true,
		LoadedAt: time.Now(),
	}

	pl.plugins[name] = loadedPlugin
	return loadedPlugin, nil
}

// UnloadPlugin unloads a plugin
func (pl *PluginLoader) UnloadPlugin(name string) error {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	loaded, ok := pl.plugins[name]
	if !ok {
		return fmt.Errorf("plugin not loaded: %s", name)
	}

	if loaded.Plugin != nil {
		if err := loaded.Plugin.Shutdown(); err != nil {
			return fmt.Errorf("failed to shutdown plugin: %w", err)
		}
	}

	delete(pl.plugins, name)
	return nil
}

// ListLoadedPlugins lists all loaded plugins
func (pl *PluginLoader) ListLoadedPlugins() []*LoadedPlugin {
	pl.mu.RLock()
	defer pl.mu.RUnlock()

	plugins := make([]*LoadedPlugin, 0, len(pl.plugins))
	for _, plugin := range pl.plugins {
		if plugin.Loaded {
			plugins = append(plugins, plugin)
		}
	}
	return plugins
}

// GetLoadedPlugin gets a loaded plugin
func (pl *PluginLoader) GetLoadedPlugin(name string) (*LoadedPlugin, error) {
	pl.mu.RLock()
	defer pl.mu.RUnlock()

	loaded, ok := pl.plugins[name]
	if !ok || !loaded.Loaded {
		return nil, fmt.Errorf("plugin not loaded: %s", name)
	}

	return loaded, nil
}

// ValidatePlugin validates a plugin manifest
func (pl *PluginLoader) ValidatePlugin(manifest *PluginManifest) error {
	if manifest.Name == "" {
		return fmt.Errorf("plugin name is required")
	}

	if manifest.Version == "" {
		return fmt.Errorf("plugin version is required")
	}

	if manifest.EntryPoint == "" {
		return fmt.Errorf("plugin entry point is required")
	}

	return nil
}
