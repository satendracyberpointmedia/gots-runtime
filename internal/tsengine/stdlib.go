package tsengine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// StdlibLoader loads and registers standard library modules
type StdlibLoader struct {
	engine  *Engine
	modules map[string]string // module path -> TypeScript code
}

// NewStdlibLoader creates a new stdlib loader
func NewStdlibLoader(engine *Engine) *StdlibLoader {
	return &StdlibLoader{
		engine:  engine,
		modules: make(map[string]string),
	}
}

// Load loads all standard library modules
func (sl *StdlibLoader) Load() error {
	stdlibPath, err := resolveStdlibPath()
	if err != nil {
		return err
	}

	// Walk the stdlib directory
	err = filepath.Walk(stdlibPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Only load .ts files
		if !strings.HasSuffix(path, ".ts") {
			return nil
		}

		// Read the file
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read stdlib file %s: %w", path, err)
		}

		// Convert path to module name (e.g., stdlib/fs/index.ts -> gots/stdlib/fs)
		modulePath := sl.pathToModulePath(path)
		sl.modules[modulePath] = string(data)

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk stdlib directory: %w", err)
	}

	return nil
}

// pathToModulePath converts a file system path to a module import path
func (sl *StdlibLoader) pathToModulePath(path string) string {
	// Normalize path separators
	path = filepath.ToSlash(path)

	// Remove stdlib/ prefix and .ts extension
	path = strings.TrimPrefix(path, "stdlib/")
	path = strings.TrimPrefix(path, "../../stdlib/")
	path = strings.TrimSuffix(path, ".ts")

	// Handle index.ts files
	if strings.HasSuffix(path, "/index") {
		path = strings.TrimSuffix(path, "/index")
	}

	// Convert to gots/stdlib/* format
	return "gots/stdlib/" + path
}

// Register registers stdlib modules in the TypeScript engine
func (sl *StdlibLoader) Register() error {
	// Create a module registry in the engine
	vm := sl.engine.VM()

	// Create stdlib namespace
	stdlibObj := vm.NewObject()
	vm.Set("__stdlib__", stdlibObj)

	// Register each module
	for modulePath, code := range sl.modules {
		// Create module exports object
		exports := vm.NewObject()
		moduleObj := vm.NewObject()
		moduleObj.Set("exports", exports)

		// Set up module and exports in the VM
		vm.Set("module", moduleObj)
		vm.Set("exports", exports)

		// Execute the module code
		_, err := vm.RunString(code)
		if err != nil {
			return fmt.Errorf("failed to execute stdlib module %s: %w", modulePath, err)
		}

		// Get the exports
		exportsValue := vm.Get("exports")
		if exportsValue != nil {
			// Store in stdlib namespace
			parts := strings.Split(modulePath, "/")
			if len(parts) >= 3 {
				// gots/stdlib/fs -> fs
				moduleName := parts[2]
				stdlibObj.Set(moduleName, exportsValue)
			}
		}
	}

	return nil
}

// GetModuleCode returns the TypeScript code for a module path
func (sl *StdlibLoader) GetModuleCode(modulePath string) (string, bool) {
	code, ok := sl.modules[modulePath]
	return code, ok
}

// ResolveStdlib resolves a stdlib import path to the actual module path
func ResolveStdlib(importPath string) (string, error) {
	// Handle gots/stdlib/* imports
	if !strings.HasPrefix(importPath, "gots/stdlib/") {
		return "", fmt.Errorf("not a stdlib import: %s", importPath)
	}

	// Extract module name (e.g., gots/stdlib/fs -> fs)
	moduleName := strings.TrimPrefix(importPath, "gots/stdlib/")

	// Try to find the module file
	// First try index.ts
	possiblePaths := []string{
		fmt.Sprintf("stdlib/%s/index.ts", moduleName),
		fmt.Sprintf("stdlib/%s.ts", moduleName),
		filepath.Join("stdlib", moduleName, "index.ts"),
		filepath.Join("stdlib", moduleName+".ts"),
	}

	for _, path := range possiblePaths {
		// Check if file exists
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("stdlib module not found: %s", importPath)
}

// GetStdlibPath returns the file system path for a stdlib module
func GetStdlibPath(modulePath string) (string, error) {
	// Convert gots/stdlib/fs to ../../stdlib/fs/index.ts or ../../stdlib/fs.ts
	if !strings.HasPrefix(modulePath, "gots/stdlib/") {
		return "", fmt.Errorf("not a stdlib import: %s", modulePath)
	}

	moduleName := strings.TrimPrefix(modulePath, "gots/stdlib/")

	stdlibRoot, err := resolveStdlibPath()
	if err != nil {
		return "", err
	}

	// Try index.ts first
	indexPath := filepath.Join(stdlibRoot, moduleName, "index.ts")
	if _, err := os.Stat(indexPath); err == nil {
		return indexPath, nil
	}

	// Try direct .ts file
	tsPath := filepath.Join(stdlibRoot, moduleName+".ts")
	if _, err := os.Stat(tsPath); err == nil {
		return tsPath, nil
	}

	return "", fmt.Errorf("stdlib module not found: %s (root: %s)", modulePath, stdlibRoot)
}

// resolveStdlibPath determines the stdlib directory location.
// Priority:
// 1) GOTS_STDLIB_PATH environment variable, if set and exists
// 2) Directory named "stdlib" next to the running executable
// 3) "stdlib" in current working directory (for development)
// 4) "../../stdlib" (legacy/dev fallback)
func resolveStdlibPath() (string, error) {
	// 1) Environment override
	if envPath := os.Getenv("GOTS_STDLIB_PATH"); envPath != "" {
		if info, err := os.Stat(envPath); err == nil && info.IsDir() {
			return envPath, nil
		}
	}

	// 2) Next to the executable
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		exeStdlib := filepath.Join(exeDir, "stdlib")
		if info, err := os.Stat(exeStdlib); err == nil && info.IsDir() {
			return exeStdlib, nil
		}
	}

	// 3) Current working directory
	cwdStdlib := "stdlib"
	if info, err := os.Stat(cwdStdlib); err == nil && info.IsDir() {
		return cwdStdlib, nil
	}

	// 4) Legacy relative source path
	legacyStdlib := "../../stdlib"
	if info, err := os.Stat(legacyStdlib); err == nil && info.IsDir() {
		return legacyStdlib, nil
	}

	return "", fmt.Errorf("stdlib directory not found; tried GOTS_STDLIB_PATH, executable-relative stdlib, ./stdlib, ../../stdlib")
}
