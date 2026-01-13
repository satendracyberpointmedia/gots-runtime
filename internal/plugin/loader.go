package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// PluginManifest represents a plugin manifest
type PluginManifest struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	EntryPoint  string                 `json:"entryPoint"`
	Config      map[string]interface{} `json:"config"`
}

// PluginLoader loads plugins from filesystem
type PluginLoader struct {
	pluginDir string
}

// NewPluginLoader creates a new plugin loader
func NewPluginLoader(pluginDir string) *PluginLoader {
	return &PluginLoader{
		pluginDir: pluginDir,
	}
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

