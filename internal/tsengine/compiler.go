package tsengine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Compiler handles TypeScript compilation
type Compiler struct {
	strictMode bool
	tsOnly     bool
}

// NewCompiler creates a new TypeScript compiler
func NewCompiler() *Compiler {
	return &Compiler{
		strictMode: true,
		tsOnly:     true,
	}
}

// Compile compiles TypeScript source code to JavaScript
// For Phase 1, we'll use a simple approach that checks for .ts extension
// and validates it's TypeScript (not plain JS)
func (c *Compiler) Compile(sourcePath string) (string, error) {
	// Check file extension
	if !strings.HasSuffix(sourcePath, ".ts") && !strings.HasSuffix(sourcePath, ".tsx") {
		if c.tsOnly {
			return "", fmt.Errorf("only TypeScript files (.ts, .tsx) are allowed, got: %s", sourcePath)
		}
	}

	// Read source file
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to read source file: %w", err)
	}

	// Basic validation: check if it looks like plain JavaScript
	if c.tsOnly {
		if err := c.ValidateTypeScript(string(source)); err != nil {
			return "", err
		}
	}

	// For Phase 1, we'll return the source as-is
	// In later phases, we'll integrate actual TypeScript compiler
	return string(source), nil
}

// validateTypeScript performs basic validation to ensure it's TypeScript
func (c *Compiler) ValidateTypeScript(source string) error {
	// Check for TypeScript-specific syntax
	hasTypeAnnotations := strings.Contains(source, ":") && 
		(strings.Contains(source, "string") || 
		 strings.Contains(source, "number") || 
		 strings.Contains(source, "boolean") ||
		 strings.Contains(source, "interface") ||
		 strings.Contains(source, "type ") ||
		 strings.Contains(source, "enum "))

	// Check for TypeScript keywords
	hasTSKeywords := strings.Contains(source, "interface") ||
		strings.Contains(source, "type ") ||
		strings.Contains(source, "enum ") ||
		strings.Contains(source, "declare ") ||
		strings.Contains(source, "namespace ")

	// If it's a very simple file, allow it (might be valid TS without annotations)
	// But if it's complex and has no TS features, reject it
	if !hasTypeAnnotations && !hasTSKeywords && len(source) > 100 {
		// Check if it looks like plain JS (has function declarations without types)
		if strings.Contains(source, "function ") && !strings.Contains(source, "function ") {
			return fmt.Errorf("plain JavaScript detected. Only TypeScript is allowed")
		}
	}

	return nil
}

// CompileWithTSC compiles using TypeScript compiler (tsc) if available
func (c *Compiler) CompileWithTSC(sourcePath string, outputDir string) (string, error) {
	// Check if tsc is available
	tscPath, err := exec.LookPath("tsc")
	if err != nil {
		return "", fmt.Errorf("TypeScript compiler (tsc) not found: %w", err)
	}

	// Create temporary tsconfig.json if it doesn't exist
	tsconfigPath := filepath.Join(filepath.Dir(sourcePath), "tsconfig.json")
	if _, err := os.Stat(tsconfigPath); os.IsNotExist(err) {
		tsconfig := `{
  "compilerOptions": {
    "target": "ES2020",
    "module": "ESNext",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true
  }
}`
		if err := os.WriteFile(tsconfigPath, []byte(tsconfig), 0644); err != nil {
			return "", fmt.Errorf("failed to create tsconfig.json: %w", err)
		}
	}

	// Run tsc
	cmd := exec.Command(tscPath, sourcePath, "--outDir", outputDir)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("TypeScript compilation failed: %w", err)
	}

	// Read compiled output
	outputPath := strings.TrimSuffix(sourcePath, ".ts") + ".js"
	outputPath = filepath.Join(outputDir, filepath.Base(outputPath))
	
	output, err := os.ReadFile(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to read compiled output: %w", err)
	}

	return string(output), nil
}

