package loadbalancer

import (
	"context"
	"net/http"
	"sync"
	"time"
)

// HealthChecker checks backend health
type HealthChecker struct {
	backends map[string]*Backend
	stop     chan struct{}
	wg       sync.WaitGroup
	mu       sync.RWMutex
}

// NewHealthChecker creates a new health checker
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		backends: make(map[string]*Backend),
		stop:     make(chan struct{}),
	}
}

// AddBackend adds a backend to health checking
func (hc *HealthChecker) AddBackend(backend *Backend) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.backends[backend.URL] = backend
}

// RemoveBackend removes a backend from health checking
func (hc *HealthChecker) RemoveBackend(url string) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	delete(hc.backends, url)
}

// Start starts health checking
func (hc *HealthChecker) Start(interval time.Duration) {
	hc.wg.Add(1)
	go hc.check(interval)
}

// Stop stops health checking
func (hc *HealthChecker) Stop() {
	close(hc.stop)
	hc.wg.Wait()
}

// check periodically checks backend health
func (hc *HealthChecker) check(interval time.Duration) {
	defer hc.wg.Done()
	
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			hc.checkAll()
		case <-hc.stop:
			return
		}
	}
}

// checkAll checks all backends
func (hc *HealthChecker) checkAll() {
	hc.mu.RLock()
	backends := make([]*Backend, 0, len(hc.backends))
	for _, backend := range hc.backends {
		backends = append(backends, backend)
	}
	hc.mu.RUnlock()
	
	for _, backend := range backends {
		go hc.checkBackend(backend)
	}
}

// checkBackend checks a single backend
func (hc *HealthChecker) checkBackend(backend *Backend) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, "GET", backend.URL+"/health", nil)
	if err != nil {
		backend.SetHealthy(false)
		return
	}
	
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		backend.SetHealthy(false)
		return
	}
	defer resp.Body.Close()
	
	healthy := resp.StatusCode == http.StatusOK
	backend.SetHealthy(healthy)
}

