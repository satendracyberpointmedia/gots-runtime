package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gots-runtime/internal/transpiler"

	"github.com/dop251/goja"
)

// Runtime manages the JavaScript/TypeScript execution environment
type Runtime struct {
	vm         *goja.Runtime
	transpiler *transpiler.Transpiler
	stdlibPath string
	modules    map[string]interface{}
}

// New creates a new Runtime instance
func New(stdlibPath string) (*Runtime, error) {
	r := &Runtime{
		vm:         goja.New(),
		transpiler: transpiler.New(),
		stdlibPath: stdlibPath,
		modules:    make(map[string]interface{}),
	}

	// Initialize built-in objects
	if err := r.initializeBuiltins(); err != nil {
		return nil, fmt.Errorf("failed to initialize builtins: %w", err)
	}

	// Load stdlib if path is provided
	if stdlibPath != "" {
		if err := r.loadStdlib(); err != nil {
			return nil, fmt.Errorf("failed to load stdlib: %w", err)
		}
	}

	return r, nil
}

// initializeBuiltins sets up built-in objects and functions
func (r *Runtime) initializeBuiltins() error {
	// Add console object
	console := r.vm.NewObject()
	console.Set("log", func(args ...interface{}) {
		fmt.Println(args...)
	})
	console.Set("error", func(args ...interface{}) {
		fmt.Fprintln(os.Stderr, args...)
	})
	console.Set("warn", func(args ...interface{}) {
		fmt.Fprintln(os.Stderr, "Warning:")
	})
	r.vm.Set("console", console)

	// Add require function
	r.vm.Set("require", r.requireFunction())

	// Add global object
	r.vm.Set("global", r.vm.GlobalObject())

	return nil
}

// requireFunction creates a CommonJS-style require function
func (r *Runtime) requireFunction() func(string) interface{} {
	return func(modulePath string) interface{} {
		// Check if already loaded
		if mod, ok := r.modules[modulePath]; ok {
			return mod
		}

		// Try to load the module
		mod, err := r.loadModule(modulePath)
		if err != nil {
			panic(r.vm.ToValue(fmt.Sprintf("Cannot find module '%s': %v", modulePath, err)))
		}

		// Cache the module
		r.modules[modulePath] = mod
		return mod
	}
}

// loadModule loads a module by path
func (r *Runtime) loadModule(modulePath string) (interface{}, error) {
	// Resolve module path
	resolvedPath, err := r.resolveModulePath(modulePath)
	if err != nil {
		return nil, err
	}

	// Check if it's a TypeScript or JavaScript file
	var code string
	if strings.HasSuffix(resolvedPath, ".ts") {
		// Transpile TypeScript to JavaScript
		code, err = r.transpiler.TranspileFile(resolvedPath)
		if err != nil {
			return nil, fmt.Errorf("transpilation failed: %w", err)
		}
	} else {
		// Read JavaScript directly
		content, err := os.ReadFile(resolvedPath)
		if err != nil {
			return nil, err
		}
		code = string(content)
	}

	// Create module context
	moduleObj := r.vm.NewObject()
	exportsObj := r.vm.NewObject()
	moduleObj.Set("exports", exportsObj)

	// Set module and exports in scope
	r.vm.Set("module", moduleObj)
	r.vm.Set("exports", exportsObj)

	// Execute the module code
	_, err = r.vm.RunString(code)
	if err != nil {
		return nil, fmt.Errorf("module execution failed: %w", err)
	}

	// Get the exports
	moduleExports := moduleObj.Get("exports")

	return moduleExports, nil
}

// resolveModulePath resolves a module path to an actual file path
func (r *Runtime) resolveModulePath(modulePath string) (string, error) {
	// If it's a relative path, resolve it
	if strings.HasPrefix(modulePath, "./") || strings.HasPrefix(modulePath, "../") {
		// This would need the current module's directory context
		// For now, just check if file exists
		if _, err := os.Stat(modulePath); err == nil {
			return modulePath, nil
		}

		// Try with .ts extension
		tsPath := modulePath + ".ts"
		if _, err := os.Stat(tsPath); err == nil {
			return tsPath, nil
		}

		// Try with .js extension
		jsPath := modulePath + ".js"
		if _, err := os.Stat(jsPath); err == nil {
			return jsPath, nil
		}
	}

	// Check in stdlib
	if r.stdlibPath != "" {
		stdlibModulePath := filepath.Join(r.stdlibPath, modulePath)

		// Try as-is
		if _, err := os.Stat(stdlibModulePath); err == nil {
			return stdlibModulePath, nil
		}

		// Try with .ts
		if _, err := os.Stat(stdlibModulePath + ".ts"); err == nil {
			return stdlibModulePath + ".ts", nil
		}

		// Try with .js
		if _, err := os.Stat(stdlibModulePath + ".js"); err == nil {
			return stdlibModulePath + ".js", nil
		}

		// Try index.ts
		indexPath := filepath.Join(stdlibModulePath, "index.ts")
		if _, err := os.Stat(indexPath); err == nil {
			return indexPath, nil
		}

		// Try index.js
		indexPath = filepath.Join(stdlibModulePath, "index.js")
		if _, err := os.Stat(indexPath); err == nil {
			return indexPath, nil
		}
	}

	return "", fmt.Errorf("module not found: %s", modulePath)
}

// loadStdlib loads all stdlib modules
func (r *Runtime) loadStdlib() error {
	if r.stdlibPath == "" {
		return nil
	}

	// Check if stdlib directory exists
	if _, err := os.Stat(r.stdlibPath); os.IsNotExist(err) {
		return fmt.Errorf("stdlib directory not found: %s", r.stdlibPath)
	}

	// Don't preload stdlib modules - load them on demand via require()
	// This avoids errors with incomplete stdlib files during development
	return nil
}

// ExecuteFile executes a TypeScript or JavaScript file
func (r *Runtime) ExecuteFile(filePath string) (goja.Value, error) {
	var code string
	var err error

	if strings.HasSuffix(filePath, ".ts") {
		// Transpile TypeScript
		code, err = r.transpiler.TranspileFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("transpilation failed: %w", err)
		}
	} else {
		// Read JavaScript
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		code = string(content)
	}

	// Execute code
	return r.vm.RunString(code)
}

// ExecuteString executes TypeScript or JavaScript code from a string
func (r *Runtime) ExecuteString(code string, isTypeScript bool) (goja.Value, error) {
	if isTypeScript {
		// Transpile first
		js, err := r.transpiler.Transpile(code, "<string>")
		if err != nil {
			return nil, fmt.Errorf("transpilation failed: %w", err)
		}
		code = js
	}

	return r.vm.RunString(code)
}

// GetVM returns the underlying Goja VM
func (r *Runtime) GetVM() *goja.Runtime {
	return r.vm
}
