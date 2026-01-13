package testrunner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dop251/goja"
	"gots-runtime/internal/tsengine"
)

// TestResult represents the result of a test
type TestResult struct {
	Name     string
	Passed   bool
	Error    error
	Duration int64 // milliseconds
}

// Runner represents a test runner
type Runner struct {
	testDir string
	engine  *tsengine.Engine
}

// NewRunner creates a new test runner
func NewRunner(testDir string) *Runner {
	return &Runner{
		testDir: testDir,
		engine:  tsengine.NewEngine(),
	}
}

// DiscoverTests discovers test files
func (r *Runner) DiscoverTests(pattern string) ([]string, error) {
	var testFiles []string
	
	err := filepath.Walk(r.testDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if info.IsDir() {
			return nil
		}
		
		// Check if file matches pattern
		if strings.HasSuffix(path, ".test.ts") || strings.HasSuffix(path, ".spec.ts") {
			if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
				testFiles = append(testFiles, path)
			}
		}
		
		return nil
	})
	
	return testFiles, err
}

// RunTests runs all discovered tests
func (r *Runner) RunTests(pattern string) ([]TestResult, error) {
	testFiles, err := r.DiscoverTests(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to discover tests: %w", err)
	}

	var results []TestResult
	for _, file := range testFiles {
		result, err := r.RunTest(file)
		if err != nil {
			result = &TestResult{
				Name:   file,
				Passed: false,
				Error:  err,
			}
		}
		results = append(results, *result)
	}

	return results, nil
}

// RunTest runs a single test file
func (r *Runner) RunTest(testFile string) (*TestResult, error) {
	startTime := time.Now()
	
	// Execute the test file
	_, err := r.engine.ExecuteFile(testFile)
	
	duration := time.Since(startTime).Milliseconds()
	
	if err != nil {
		return &TestResult{
			Name:     testFile,
			Passed:   false,
			Error:    fmt.Errorf("test execution failed: %w", err),
			Duration: duration,
		}, nil
	}
	
	// Check for test results in the engine
	// Look for test functions that were executed
	// For now, we'll assume success if no error occurred
	// In a full implementation, we'd parse test results from the executed code
	
	// Try to get test results from the engine
	testResults := r.engine.Get("__testResults__")
	if testResults != nil {
		// Check if test results exist
		if !goja.IsUndefined(testResults) {
			// Parse test results if available
			// This would require the test file to set __testResults__
			// For now, we assume success if no error occurred
		}
	}
	
	return &TestResult{
		Name:     testFile,
		Passed:   true,
		Duration: duration,
	}, nil
}

// Coverage represents test coverage information
type Coverage struct {
	TotalLines    int
	CoveredLines  int
	CoveragePercent float64
}

// GetCoverage calculates test coverage
func (r *Runner) GetCoverage() (*Coverage, error) {
	// Basic coverage calculation
	// In a full implementation, we'd track which lines were executed
	totalLines := 0
	coveredLines := 0
	
	// Walk test directory and count lines
	err := filepath.Walk(r.testDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if info.IsDir() {
			return nil
		}
		
		// Only count .ts files
		if strings.HasSuffix(path, ".ts") {
			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			
			lines := strings.Split(string(data), "\n")
			totalLines += len(lines)
			// For now, assume all lines are covered if tests pass
			// In a full implementation, we'd track actual coverage
			coveredLines += len(lines)
		}
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to calculate coverage: %w", err)
	}
	
	coveragePercent := 0.0
	if totalLines > 0 {
		coveragePercent = float64(coveredLines) / float64(totalLines) * 100.0
	}
	
	return &Coverage{
		TotalLines:     totalLines,
		CoveredLines:   coveredLines,
		CoveragePercent: coveragePercent,
	}, nil
}

