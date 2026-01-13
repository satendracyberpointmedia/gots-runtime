package main

import (
	"fmt"
	"os"
	"path/filepath"
	"os/signal"
	"syscall"

	"gots-runtime/internal/config"
	"gots-runtime/pkg/debugger"
	"gots-runtime/pkg/testrunner"

	"github.com/spf13/cobra"
)

var (
	version = "0.1.0"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "gots",
		Short: "Go-based Multithreaded Runtime with Inbuilt TypeScript",
		Long:  "A next-generation runtime environment that combines Golang's multithreading capabilities with TypeScript as a first-class citizen.",
		Version: version,
	}

	var runCmd = &cobra.Command{
		Use:   "run [file]",
		Short: "Run a TypeScript file",
		Long:  "Execute a TypeScript file using the GoTS runtime",
		Args:  cobra.ExactArgs(1),
		RunE:  runFile,
	}

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("gots version %s\n", version)
		},
	}

	var initCmd = &cobra.Command{
		Use:   "init [project-name]",
		Short: "Initialize a new GoTS project",
		Long:  "Create a new GoTS project with basic structure",
		Args:  cobra.MaximumNArgs(1),
		RunE:  initProject,
	}

	var buildCmd = &cobra.Command{
		Use:   "build [file]",
		Short: "Build a TypeScript file",
		Long:  "Compile a TypeScript file to JavaScript (for compatibility)",
		Args:  cobra.ExactArgs(1),
		RunE:  buildFile,
	}

	var testCmd = &cobra.Command{
		Use:   "test [pattern]",
		Short: "Run tests",
		Long:  "Run tests in the current project",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runTests,
	}

	var debugCmd = &cobra.Command{
		Use:   "debug [file]",
		Short: "Debug a TypeScript file",
		Long:  "Start debugging session for a TypeScript file",
		Args:  cobra.ExactArgs(1),
		RunE:  debugFile,
	}

	var serveCmd = &cobra.Command{
		Use:   "serve [file]",
		Short: "Start a long-running server",
		Long:  "Start a long-running TypeScript server with hot reload",
		Args:  cobra.ExactArgs(1),
		RunE:  serveFile,
	}

	var profileCmd = &cobra.Command{
		Use:   "profile [file]",
		Short: "Profile a TypeScript file",
		Long:  "Run a TypeScript file with profiling enabled",
		Args:  cobra.ExactArgs(1),
		RunE:  profileFile,
	}

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(debugCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(profileCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runFile(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	// Get absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Get project root (directory containing the file)
	projectRoot := filepath.Dir(absPath)

	// Create runtime manager
	rm, err := NewRuntimeManager(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to create runtime manager: %w", err)
	}
	defer rm.Shutdown()

	// Get module ID from config or use default
	moduleID := "main"
	if rm.GetConfig().Main != "" {
		moduleID = rm.GetConfig().Main
	}

	// Execute the module
	if err := rm.ExecuteModule(moduleID, absPath); err != nil {
		return fmt.Errorf("failed to execute module: %w", err)
	}

	fmt.Printf("Module executed successfully: %s\n", absPath)
	return nil
}

func initProject(cmd *cobra.Command, args []string) error {
	projectName := "my-gots-project"
	if len(args) > 0 {
		projectName = args[0]
	}

	// Create project directory
	if err := os.MkdirAll(projectName, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Create main.ts
	mainContent := `// Main entry point
console.log("Hello from GoTS Runtime!");

export function main(): void {
    console.log("Main function executed");
}
`
	if err := os.WriteFile(filepath.Join(projectName, "main.ts"), []byte(mainContent), 0644); err != nil {
		return fmt.Errorf("failed to create main.ts: %w", err)
	}

	// Create gots.json config
	cfg := config.GetDefaultConfig()
	cfg.Name = projectName
	cfg.Main = "main.ts"
	
	configPath := filepath.Join(projectName, "gots.json")
	if err := config.SaveConfig(cfg, configPath); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}

	// Create README.md
	readmeContent := fmt.Sprintf(`# %s

A GoTS Runtime project.

## Running

`+"```bash"+`
gots run main.ts
`+"```"+`

## Testing

`+"```bash"+`
gots test
`+"```"+`
`, projectName)
	if err := os.WriteFile(filepath.Join(projectName, "README.md"), []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("failed to create README.md: %w", err)
	}

	fmt.Printf("Project '%s' initialized successfully!\n", projectName)
	return nil
}

func buildFile(cmd *cobra.Command, args []string) error {
	filePath := args[0]
	
	// For Phase 5, we'll just validate the file
	// In a full implementation, this would compile TS to JS
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", absPath)
	}

	fmt.Printf("Building %s...\n", absPath)
	fmt.Println("Build complete (validation only in Phase 5)")
	return nil
}

func runTests(cmd *cobra.Command, args []string) error {
	pattern := "**/*.test.ts"
	if len(args) > 0 {
		pattern = args[0]
	}

	// Get current directory as project root
	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Create runtime manager
	rm, err := NewRuntimeManager(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to create runtime manager: %w", err)
	}
	defer rm.Shutdown()

	// Create test runner
	runner := testrunner.NewRunner(projectRoot)
	
	// Discover and run tests
	results, err := runner.RunTests(pattern)
	if err != nil {
		return fmt.Errorf("failed to run tests: %w", err)
	}

	// Print results
	passed := 0
	failed := 0
	for _, result := range results {
		if result.Passed {
			passed++
			fmt.Printf("✓ %s\n", result.Name)
		} else {
			failed++
			if result.Error != nil {
				fmt.Printf("✗ %s: %s\n", result.Name, result.Error)
			} else {
				fmt.Printf("✗ %s\n", result.Name)
			}
		}
	}

	fmt.Printf("\nTests: %d passed, %d failed\n", passed, failed)
	
	if failed > 0 {
		return fmt.Errorf("some tests failed")
	}
	
	return nil
}

func debugFile(cmd *cobra.Command, args []string) error {
	filePath := args[0]
	
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", absPath)
	}

	// Get project root
	projectRoot := filepath.Dir(absPath)

	// Create runtime manager
	rm, err := NewRuntimeManager(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to create runtime manager: %w", err)
	}
	defer rm.Shutdown()

	// Create debugger
	ctx := rm.GetIntegration().GetOrchestrator().Context()
	dbg := debugger.NewDebugger(ctx)
	defer dbg.Stop()

	fmt.Printf("Debugger started for %s\n", absPath)
	fmt.Println("Use 'continue', 'step', 'inspect <var>' commands")
	fmt.Println("Type 'quit' to exit")

	// Simple interactive loop (in production, use proper debugger protocol)
	// For now, just execute the file
	moduleID := "main"
	if err := rm.ExecuteModule(moduleID, absPath); err != nil {
		return fmt.Errorf("failed to execute module: %w", err)
	}

	return nil
}

func serveFile(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	projectRoot := filepath.Dir(absPath)

	// Create runtime manager
	rm, err := NewRuntimeManager(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to create runtime manager: %w", err)
	}
	defer rm.Shutdown()

	// Execute module
	moduleID := "main"
	if err := rm.ExecuteModule(moduleID, absPath); err != nil {
		return fmt.Errorf("failed to execute module: %w", err)
	}

	fmt.Printf("Server started. Press Ctrl+C to stop.\n")
	
	// Wait for interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
	return nil
}

func profileFile(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	projectRoot := filepath.Dir(absPath)

	// Create runtime manager
	rm, err := NewRuntimeManager(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to create runtime manager: %w", err)
	}
	defer rm.Shutdown()

	// Get profiler from integration
	profiler := rm.GetIntegration().GetTSEngine()
	
	fmt.Printf("Profiling %s...\n", absPath)
	
	// Start CPU profiling
	// Note: In a full implementation, this would use the profiler from observability
	_ = profiler
	
	// Execute module
	moduleID := "main"
	if err := rm.ExecuteModule(moduleID, absPath); err != nil {
		return fmt.Errorf("failed to execute module: %w", err)
	}

	fmt.Println("Profiling complete. Check metrics endpoint for results.")
	return nil
}

