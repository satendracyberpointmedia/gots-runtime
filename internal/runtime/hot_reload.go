package runtime

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// HotReloader provides hot reload functionality with state preservation
type HotReloader struct {
	watchPaths   []string
	state        map[string]interface{}
	reloadFunc   func(string) error
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	mu           sync.RWMutex
	lastModified map[string]time.Time
}

// NewHotReloader creates a new hot reloader
func NewHotReloader(ctx context.Context) *HotReloader {
	reloadCtx, cancel := context.WithCancel(ctx)
	return &HotReloader{
		watchPaths:   make([]string, 0),
		state:        make(map[string]interface{}),
		ctx:          reloadCtx,
		cancel:       cancel,
		lastModified: make(map[string]time.Time),
	}
}

// Watch adds a path to watch
func (hr *HotReloader) Watch(path string) {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	
	hr.watchPaths = append(hr.watchPaths, path)
}

// SetReloadFunc sets the function to call on reload
func (hr *HotReloader) SetReloadFunc(fn func(string) error) {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	hr.reloadFunc = fn
}

// SaveState saves state for preservation
func (hr *HotReloader) SaveState(key string, value interface{}) {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	hr.state[key] = value
}

// GetState gets preserved state
func (hr *HotReloader) GetState(key string) (interface{}, bool) {
	hr.mu.RLock()
	defer hr.mu.RUnlock()
	value, ok := hr.state[key]
	return value, ok
}

// Start starts watching for file changes
func (hr *HotReloader) Start() {
	hr.wg.Add(1)
	go hr.watch()
}

// Stop stops watching
func (hr *HotReloader) Stop() {
	hr.cancel()
	hr.wg.Wait()
}

// watch watches for file changes
func (hr *HotReloader) watch() {
	defer hr.wg.Done()
	
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			hr.checkFiles()
		case <-hr.ctx.Done():
			return
		}
	}
}

// checkFiles checks for file changes
func (hr *HotReloader) checkFiles() {
	hr.mu.RLock()
	paths := make([]string, len(hr.watchPaths))
	copy(paths, hr.watchPaths)
	reloadFunc := hr.reloadFunc
	hr.mu.RUnlock()
	
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		
		modified := info.ModTime()
		
		hr.mu.Lock()
		lastMod, exists := hr.lastModified[path]
		if !exists || modified.After(lastMod) {
			hr.lastModified[path] = modified
			hr.mu.Unlock()
			
			if exists && reloadFunc != nil {
				// File changed, trigger reload
				if err := reloadFunc(path); err != nil {
					fmt.Printf("Hot reload failed for %s: %v\n", path, err)
				} else {
					fmt.Printf("Hot reloaded: %s\n", path)
				}
			}
		} else {
			hr.mu.Unlock()
		}
	}
}

// ReloadFile manually triggers a reload for a file
func (hr *HotReloader) ReloadFile(path string) error {
	hr.mu.RLock()
	reloadFunc := hr.reloadFunc
	hr.mu.RUnlock()
	
	if reloadFunc == nil {
		return fmt.Errorf("no reload function set")
	}
	
	return reloadFunc(path)
}

// WatchDirectory watches a directory recursively
func (hr *HotReloader) WatchDirectory(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && filepath.Ext(path) == ".ts" {
			hr.Watch(path)
		}
		
		return nil
	})
}

