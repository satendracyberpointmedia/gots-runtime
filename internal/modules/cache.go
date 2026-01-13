package modules

import (
	"sync"
)

// Cache stores loaded modules
type Cache struct {
	modules map[string]*CachedModule
	mu      sync.RWMutex
}

// CachedModule represents a cached module
type CachedModule struct {
	Path      string
	Exports   map[string]interface{}
	Timestamp int64
}

// NewCache creates a new module cache
func NewCache() *Cache {
	return &Cache{
		modules: make(map[string]*CachedModule),
	}
}

// Get retrieves a module from cache
func (c *Cache) Get(path string) (*CachedModule, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	module, ok := c.modules[path]
	return module, ok
}

// Set stores a module in cache
func (c *Cache) Set(path string, module *CachedModule) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.modules[path] = module
}

// Clear removes all modules from cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.modules = make(map[string]*CachedModule)
}

// Remove removes a specific module from cache
func (c *Cache) Remove(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.modules, path)
}

