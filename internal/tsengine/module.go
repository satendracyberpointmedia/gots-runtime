package tsengine

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Module represents a loaded module
type Module struct {
	Path     string
	Exports  map[string]interface{}
	Engine   *Engine
}

// NewModule creates a new module
func NewModule(path string, engine *Engine) *Module {
	return &Module{
		Path:    path,
		Exports: make(map[string]interface{}),
		Engine:  engine,
	}
}

// SetExport sets an export value
func (m *Module) SetExport(name string, value interface{}) {
	m.Exports[name] = value
}

// GetExport gets an export value
func (m *Module) GetExport(name string) (interface{}, error) {
	value, ok := m.Exports[name]
	if !ok {
		return nil, fmt.Errorf("export '%s' not found in module '%s'", name, m.Path)
	}
	return value, nil
}

// ResolveModulePath resolves a module path relative to the current module
func ResolveModulePath(importPath string, currentModulePath string) (string, error) {
	// Handle relative imports
	if strings.HasPrefix(importPath, "./") || strings.HasPrefix(importPath, "../") {
		baseDir := filepath.Dir(currentModulePath)
		resolved := filepath.Join(baseDir, importPath)
		// Add .ts extension if not present
		if !strings.HasSuffix(resolved, ".ts") && !strings.HasSuffix(resolved, ".tsx") {
			resolved += ".ts"
		}
		return filepath.Clean(resolved), nil
	}

	// Handle absolute imports (node_modules style)
	// For Phase 1, we'll just return the path as-is
	// In later phases, we'll implement proper module resolution
	if !strings.HasSuffix(importPath, ".ts") && !strings.HasSuffix(importPath, ".tsx") {
		importPath += ".ts"
	}

	return importPath, nil
}

