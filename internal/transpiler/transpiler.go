package transpiler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Transpiler handles TypeScript to JavaScript conversion
type Transpiler struct {
	// Cache for transpiled code
	cache map[string]string
}

// New creates a new Transpiler instance
func New() *Transpiler {
	return &Transpiler{
		cache: make(map[string]string),
	}
}

// TranspileFile transpiles a TypeScript file to JavaScript
func (t *Transpiler) TranspileFile(tsFilePath string) (string, error) {
	// Check cache first
	if js, ok := t.cache[tsFilePath]; ok {
		return js, nil
	}

	// Read TypeScript file
	tsCode, err := os.ReadFile(tsFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Transpile
	jsCode, err := t.Transpile(string(tsCode), tsFilePath)
	if err != nil {
		return "", err
	}

	// Cache result
	t.cache[tsFilePath] = jsCode

	return jsCode, nil
}

// Transpile converts TypeScript code to JavaScript
func (t *Transpiler) Transpile(tsCode, filename string) (string, error) {
	// Try using esbuild first (fastest option)
	if js, err := t.transpileWithESBuild(tsCode, filename); err == nil {
		return js, nil
	}

	// Fallback to basic TypeScript stripping
	return t.basicTypeScriptStrip(tsCode), nil
}

// transpileWithESBuild uses esbuild for fast TypeScript transpilation
func (t *Transpiler) transpileWithESBuild(tsCode, filename string) (string, error) {
	// Check if esbuild is available
	esbuildPath, err := exec.LookPath("esbuild")
	if err != nil {
		return "", fmt.Errorf("esbuild not found: %w", err)
	}

	// Create temp file for input
	tmpDir := os.TempDir()
	inputFile := filepath.Join(tmpDir, "input.ts")
	outputFile := filepath.Join(tmpDir, "output.js")

	// Write TypeScript code to temp file
	if err := os.WriteFile(inputFile, []byte(tsCode), 0644); err != nil {
		return "", err
	}
	defer os.Remove(inputFile)

	// Run esbuild
	cmd := exec.Command(esbuildPath,
		inputFile,
		"--outfile="+outputFile,
		"--format=cjs",
		"--target=es2020",
		"--platform=node",
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("esbuild failed: %s", string(output))
	}

	// Read transpiled JavaScript
	jsCode, err := os.ReadFile(outputFile)
	if err != nil {
		return "", err
	}
	defer os.Remove(outputFile)

	return string(jsCode), nil
}

// basicTypeScriptStrip performs basic TypeScript syntax removal
// This is a fallback when esbuild is not available
func (t *Transpiler) basicTypeScriptStrip(tsCode string) string {
	lines := strings.Split(tsCode, "\n")
	result := make([]string, 0, len(lines))

	for _, line := range lines {
		// Skip type-only imports
		if strings.Contains(line, "import type") {
			continue
		}

		// Remove type annotations from variable declarations
		line = removeTypeAnnotations(line)

		// Remove interface declarations
		if strings.HasPrefix(strings.TrimSpace(line), "interface ") {
			continue
		}

		// Remove type declarations
		if strings.HasPrefix(strings.TrimSpace(line), "type ") {
			continue
		}

		// Remove 'as' type assertions
		line = removeTypeAssertions(line)

		// Convert 'export' to module.exports or exports
		line = convertExports(line)

		// Convert 'import' to require
		line = convertImports(line)

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// removeTypeAnnotations removes : Type annotations
func removeTypeAnnotations(line string) string {
	// Remove : Type from function parameters and variable declarations
	// This is a simple regex-like replacement

	// Handle function parameters: (param: Type) -> (param)
	if strings.Contains(line, ":") {
		parts := strings.Split(line, ":")
		if len(parts) > 1 {
			// Find closing ) or , or =
			for i := 1; i < len(parts); i++ {
				endIdx := strings.IndexAny(parts[i], "),=")
				if endIdx != -1 {
					parts[i] = parts[i][endIdx:]
				} else {
					parts[i] = ""
				}
			}
			line = strings.Join(parts, "")
		}
	}

	return line
}

// removeTypeAssertions removes 'as Type' assertions
func removeTypeAssertions(line string) string {
	if strings.Contains(line, " as ") {
		parts := strings.Split(line, " as ")
		if len(parts) > 1 {
			// Remove everything after 'as' until semicolon or end
			for i := 1; i < len(parts); i++ {
				endIdx := strings.IndexAny(parts[i], ";,)")
				if endIdx != -1 {
					parts[i] = parts[i][endIdx:]
				} else {
					parts[i] = ""
				}
			}
			line = strings.Join(parts, "")
		}
	}
	return line
}

// convertExports converts ES6 exports to CommonJS
func convertExports(line string) string {
	trimmed := strings.TrimSpace(line)

	// export default X -> module.exports = X
	if strings.HasPrefix(trimmed, "export default ") {
		return strings.Replace(line, "export default ", "module.exports = ", 1)
	}

	// export const X = Y -> exports.X = Y
	if strings.HasPrefix(trimmed, "export const ") {
		rest := strings.TrimPrefix(trimmed, "export const ")
		parts := strings.SplitN(rest, "=", 2)
		if len(parts) == 2 {
			varName := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			return fmt.Sprintf("const %s = %s\nexports.%s = %s", varName, value, varName, varName)
		}
	}

	// export function X() {} -> function X() {}\nexports.X = X
	if strings.HasPrefix(trimmed, "export function ") {
		rest := strings.TrimPrefix(trimmed, "export function ")
		funcName := strings.Split(rest, "(")[0]
		funcName = strings.TrimSpace(funcName)
		newLine := strings.Replace(line, "export function ", "function ", 1)
		return newLine + "\nexports." + funcName + " = " + funcName
	}

	// export { X, Y } -> exports.X = X; exports.Y = Y
	if strings.HasPrefix(trimmed, "export {") {
		content := strings.TrimPrefix(trimmed, "export {")
		content = strings.TrimSuffix(content, "}")
		exports := strings.Split(content, ",")

		var result []string
		for _, exp := range exports {
			exp = strings.TrimSpace(exp)
			if exp != "" {
				result = append(result, fmt.Sprintf("exports.%s = %s", exp, exp))
			}
		}
		return strings.Join(result, "; ")
	}

	return line
}

// convertImports converts ES6 imports to require
func convertImports(line string) string {
	trimmed := strings.TrimSpace(line)

	// import X from 'module' -> const X = require('module')
	if strings.HasPrefix(trimmed, "import ") && strings.Contains(trimmed, " from ") {
		parts := strings.Split(trimmed, " from ")
		if len(parts) == 2 {
			importPart := strings.TrimPrefix(parts[0], "import ")
			importPart = strings.TrimSpace(importPart)

			modulePath := strings.TrimSpace(parts[1])
			modulePath = strings.Trim(modulePath, "';\"")

			// Handle different import styles
			if strings.HasPrefix(importPart, "{") {
				// import { X, Y } from 'module' -> const { X, Y } = require('module')
				return fmt.Sprintf("const %s = require('%s')", importPart, modulePath)
			} else if strings.Contains(importPart, ",") {
				// import X, { Y } from 'module' -> handle mixed imports
				return fmt.Sprintf("const %s = require('%s')", importPart, modulePath)
			} else {
				// import X from 'module' -> const X = require('module')
				return fmt.Sprintf("const %s = require('%s')", importPart, modulePath)
			}
		}
	}

	return line
}

// ClearCache clears the transpilation cache
func (t *Transpiler) ClearCache() {
	t.cache = make(map[string]string)
}
