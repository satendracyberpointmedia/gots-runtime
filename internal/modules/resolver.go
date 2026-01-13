package modules

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gots-runtime/internal/tsengine"
)

// Resolver resolves module paths
type Resolver struct {
	basePath string
}

// NewResolver creates a new module resolver
func NewResolver(basePath string) *Resolver {
	return &Resolver{
		basePath: basePath,
	}
}

// Resolve resolves a module import path to a file path
func (r *Resolver) Resolve(importPath string, fromPath string) (string, error) {
	// Handle stdlib imports first
	if strings.HasPrefix(importPath, "gots/stdlib/") {
		return r.ResolveStdlib(importPath)
	}

	// Handle relative imports
	if strings.HasPrefix(importPath, "./") || strings.HasPrefix(importPath, "../") {
		baseDir := filepath.Dir(fromPath)
		resolved := filepath.Join(baseDir, importPath)
		
		// Try with .ts extension
		if !strings.HasSuffix(resolved, ".ts") && !strings.HasSuffix(resolved, ".tsx") {
			// Try .ts first
			tsPath := resolved + ".ts"
			if _, err := os.Stat(tsPath); err == nil {
				return filepath.Clean(tsPath), nil
			}
			// Try .tsx
			tsxPath := resolved + ".tsx"
			if _, err := os.Stat(tsxPath); err == nil {
				return filepath.Clean(tsxPath), nil
			}
			// Return .ts as default
			return filepath.Clean(tsPath), nil
		}
		
		return filepath.Clean(resolved), nil
	}

	// Handle absolute imports (starting with /)
	if strings.HasPrefix(importPath, "/") {
		resolved := filepath.Join(r.basePath, importPath)
		if !strings.HasSuffix(resolved, ".ts") && !strings.HasSuffix(resolved, ".tsx") {
			resolved += ".ts"
		}
		return filepath.Clean(resolved), nil
	}

	// Handle node_modules style imports (for future phases)
	// For Phase 1, we'll look in the base path
	resolved := filepath.Join(r.basePath, importPath)
	if !strings.HasSuffix(resolved, ".ts") && !strings.HasSuffix(resolved, ".tsx") {
		// Try to find the module
		tsPath := resolved + ".ts"
		if _, err := os.Stat(tsPath); err == nil {
			return filepath.Clean(tsPath), nil
		}
		tsxPath := resolved + ".tsx"
		if _, err := os.Stat(tsxPath); err == nil {
			return filepath.Clean(tsxPath), nil
		}
		return filepath.Clean(tsPath), nil
	}

	return filepath.Clean(resolved), nil
}

// ValidatePath validates that a resolved path exists and is a TypeScript file
func (r *Resolver) ValidatePath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("module not found: %s", path)
	}

	if info.IsDir() {
		return fmt.Errorf("module path is a directory: %s", path)
	}

	if !strings.HasSuffix(path, ".ts") && !strings.HasSuffix(path, ".tsx") {
		return fmt.Errorf("only TypeScript files (.ts, .tsx) are allowed: %s", path)
	}

	return nil
}

// ResolveStdlib resolves a stdlib import path
func (r *Resolver) ResolveStdlib(importPath string) (string, error) {
	// Use the stdlib resolver from tsengine
	stdlibPath, err := tsengine.GetStdlibPath(importPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve stdlib import: %w", err)
	}
	
	// Return a special marker path that indicates this is a stdlib module
	// The loader will handle this specially
	return "stdlib:" + stdlibPath, nil
}

