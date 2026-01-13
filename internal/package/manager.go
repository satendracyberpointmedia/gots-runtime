package pkg

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Package represents a package
type Package struct {
	Name        string
	Version     string
	Description string
	TypeScript  bool
	Audited     bool
	Dependencies map[string]string
}

// PackageManager manages packages
type PackageManager struct {
	packages    map[string]*Package
	packageDir  string
	registryURL string
	mu          sync.RWMutex
}

// NewPackageManager creates a new package manager
func NewPackageManager(packageDir, registryURL string) *PackageManager {
	return &PackageManager{
		packages:    make(map[string]*Package),
		packageDir:  packageDir,
		registryURL: registryURL,
	}
}

// Install installs a package
func (pm *PackageManager) Install(name, version string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	// Check if package is already installed
	if pkg, exists := pm.packages[name]; exists {
		if pkg.Version == version {
			return nil // Already installed
		}
	}
	
	// In a real implementation, this would:
	// 1. Fetch package from registry
	// 2. Validate it's TypeScript-only
	// 3. Run security audit
	// 4. Install dependencies
	
	pkg := &Package{
		Name:         name,
		Version:      version,
		TypeScript:   true, // Only TS packages allowed
		Audited:      true, // Would be set after audit
		Dependencies: make(map[string]string),
	}
	
	pm.packages[name] = pkg
	
	// Save package manifest
	return pm.savePackageManifest(pkg)
}

// Uninstall uninstalls a package
func (pm *PackageManager) Uninstall(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	if _, exists := pm.packages[name]; !exists {
		return fmt.Errorf("package not installed: %s", name)
	}
	
	delete(pm.packages, name)
	return nil
}

// List lists all installed packages
func (pm *PackageManager) List() []*Package {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	packages := make([]*Package, 0, len(pm.packages))
	for _, pkg := range pm.packages {
		packages = append(packages, pkg)
	}
	return packages
}

// GetPackage gets a package by name
func (pm *PackageManager) GetPackage(name string) (*Package, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	pkg, ok := pm.packages[name]
	return pkg, ok
}

// savePackageManifest saves a package manifest
func (pm *PackageManager) savePackageManifest(pkg *Package) error {
	manifestPath := filepath.Join(pm.packageDir, pkg.Name, "package.json")
	
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0755); err != nil {
		return fmt.Errorf("failed to create package directory: %w", err)
	}
	
	data, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}
	
	return os.WriteFile(manifestPath, data, 0644)
}

// Audit audits a package for security issues
func (pm *PackageManager) Audit(name string) (bool, error) {
	pkg, ok := pm.GetPackage(name)
	if !ok {
		return false, fmt.Errorf("package not found: %s", name)
	}
	
	// In a real implementation, this would run security checks
	// For now, we'll just mark it as audited
	pkg.Audited = true
	return true, nil
}

