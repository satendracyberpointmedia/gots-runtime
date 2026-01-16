package hotreload

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileChangeEvent represents a file system change
type FileChangeEvent struct {
	Path      string
	EventType string // "create", "write", "remove", "rename", "chmod"
	Time      time.Time
}

// HotReloadConfig contains configuration for hot reload
type HotReloadConfig struct {
	Watch           []string
	Ignore          []string
	Debounce        time.Duration
	OnReload        func() error
	OnError         func(error)
	ExcludePatterns []string
}

// HotReloader watches files and triggers reloads
type HotReloader struct {
	config        *HotReloadConfig
	done          chan bool
	debounceTimer *time.Timer
	mu            sync.Mutex
	isRunning     bool
	fileCache     map[string]time.Time
}

// NewHotReloader creates a new hot reloader
func NewHotReloader(config *HotReloadConfig) (*HotReloader, error) {
	if config.Debounce == 0 {
		config.Debounce = 500 * time.Millisecond
	}

	return &HotReloader{
		config:    config,
		done:      make(chan bool),
		fileCache: make(map[string]time.Time),
	}, nil
}

// Start starts the hot reload watcher
func (hr *HotReloader) Start() error {
	hr.mu.Lock()
	if hr.isRunning {
		hr.mu.Unlock()
		return fmt.Errorf("hot reloader already running")
	}
	hr.isRunning = true
	hr.mu.Unlock()

	// Start watching
	go hr.watch()

	return nil
}

// Stop stops the hot reload watcher
func (hr *HotReloader) Stop() error {
	hr.mu.Lock()
	if !hr.isRunning {
		hr.mu.Unlock()
		return fmt.Errorf("hot reloader not running")
	}
	hr.isRunning = false
	hr.mu.Unlock()

	hr.done <- true

	if hr.debounceTimer != nil {
		hr.debounceTimer.Stop()
	}

	return nil
}

func (hr *HotReloader) shouldIgnore(path string) bool {
	base := filepath.Base(path)

	// Check against ignore patterns
	for _, pattern := range hr.config.Ignore {
		if matched, _ := filepath.Match(pattern, base); matched {
			return true
		}
	}

	// Check against exclude patterns
	for _, pattern := range hr.config.ExcludePatterns {
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}
	}

	// Ignore hidden directories
	if len(base) > 0 && base[0] == '.' {
		return true
	}

	// Ignore node_modules
	if base == "node_modules" {
		return true
	}

	return false
}

func (hr *HotReloader) watch() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-hr.done:
			return

		case <-ticker.C:
			// Poll for file changes
			hr.pollFileChanges()
		}
	}
}

func (hr *HotReloader) pollFileChanges() {
	for _, watchPath := range hr.config.Watch {
		hr.checkPath(watchPath)
	}
}

func (hr *HotReloader) checkPath(path string) {
	filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil || hr.shouldIgnore(filePath) {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		if !info.IsDir() {
			modTime := info.ModTime()
			if lastTime, exists := hr.fileCache[filePath]; !exists {
				hr.fileCache[filePath] = modTime
			} else if modTime.After(lastTime) {
				hr.fileCache[filePath] = modTime
				hr.debounceReload()
			}
		}

		return nil
	})
}

func (hr *HotReloader) debounceReload() {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	// Cancel previous timer
	if hr.debounceTimer != nil {
		hr.debounceTimer.Stop()
	}

	// Set new timer
	hr.debounceTimer = time.AfterFunc(hr.config.Debounce, func() {
		hr.reload()
	})
}

func (hr *HotReloader) reload() {
	hr.mu.Lock()
	hr.debounceTimer = nil
	hr.mu.Unlock()

	fmt.Println("\n[HotReload] Files changed, reloading...")

	if hr.config.OnReload != nil {
		if err := hr.config.OnReload(); err != nil {
			if hr.config.OnError != nil {
				hr.config.OnError(fmt.Errorf("reload failed: %w", err))
			}
		} else {
			fmt.Println("[HotReload] Reload successful!")
		}
	}
}

// Watch watches for file changes (blocking call)
func (hr *HotReloader) Watch(onChange func(event FileChangeEvent) error) error {
	if err := hr.Start(); err != nil {
		return err
	}
	defer hr.Stop()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Check for changes
			filepath.Walk(hr.config.Watch[0], func(path string, info os.FileInfo, err error) error {
				if err != nil || hr.shouldIgnore(path) {
					return nil
				}

				if info.IsDir() {
					return nil
				}

				modTime := info.ModTime()
				if lastTime, exists := hr.fileCache[path]; !exists {
					hr.fileCache[path] = modTime
				} else if modTime.After(lastTime) {
					hr.fileCache[path] = modTime
					changeEvent := FileChangeEvent{
						Path:      path,
						EventType: "write",
						Time:      time.Now(),
					}
					if err := onChange(changeEvent); err != nil {
						return err
					}
				}

				return nil
			})

		case <-hr.done:
			return nil
		}
	}
}

// GetFileStatus returns the status of a watched file
func (hr *HotReloader) GetFileStatus(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return !stat.IsDir(), nil
}
