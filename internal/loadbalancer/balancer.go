package loadbalancer

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Backend represents a backend server
type Backend struct {
	URL          string
	Healthy     bool
	Weight       int
	ActiveConns  int
	LastHealthCheck time.Time
	mu           sync.RWMutex
}

// NewBackend creates a new backend
func NewBackend(url string, weight int) *Backend {
	return &Backend{
		URL:          url,
		Healthy:     true,
		Weight:       weight,
		ActiveConns: 0,
		LastHealthCheck: time.Now(),
	}
}

// SetHealthy sets the health status
func (b *Backend) SetHealthy(healthy bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Healthy = healthy
	b.LastHealthCheck = time.Now()
}

// IncrementConn increments active connections
func (b *Backend) IncrementConn() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.ActiveConns++
}

// DecrementConn decrements active connections
func (b *Backend) DecrementConn() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.ActiveConns > 0 {
		b.ActiveConns--
	}
}

// LoadBalancer provides runtime-aware load balancing
type LoadBalancer struct {
	backends      []*Backend
	strategy      Strategy
	healthChecker *HealthChecker
	mu            sync.RWMutex
}

// Strategy represents a load balancing strategy
type Strategy int

const (
	StrategyRoundRobin Strategy = iota
	StrategyLeastConnections
	StrategyWeightedRoundRobin
	StrategyIPHash
)

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer(strategy Strategy) *LoadBalancer {
	return &LoadBalancer{
		backends:      make([]*Backend, 0),
		strategy:      strategy,
		healthChecker: NewHealthChecker(),
	}
}

// AddBackend adds a backend
func (lb *LoadBalancer) AddBackend(backend *Backend) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.backends = append(lb.backends, backend)
	lb.healthChecker.AddBackend(backend)
}

// RemoveBackend removes a backend
func (lb *LoadBalancer) RemoveBackend(url string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	for i, backend := range lb.backends {
		if backend.URL == url {
			lb.backends = append(lb.backends[:i], lb.backends[i+1:]...)
			lb.healthChecker.RemoveBackend(url)
			break
		}
	}
}

// SelectBackend selects a backend based on strategy
func (lb *LoadBalancer) SelectBackend(req *http.Request) (*Backend, error) {
	lb.mu.RLock()
	backends := make([]*Backend, 0)
	for _, backend := range lb.backends {
		backend.mu.RLock()
		if backend.Healthy {
			backends = append(backends, backend)
		}
		backend.mu.RUnlock()
	}
	lb.mu.RUnlock()
	
	if len(backends) == 0 {
		return nil, fmt.Errorf("no healthy backends available")
	}
	
	switch lb.strategy {
	case StrategyRoundRobin:
		return lb.roundRobin(backends)
	case StrategyLeastConnections:
		return lb.leastConnections(backends)
	case StrategyWeightedRoundRobin:
		return lb.weightedRoundRobin(backends)
	case StrategyIPHash:
		return lb.ipHash(backends, req)
	default:
		return lb.roundRobin(backends)
	}
}

// roundRobin selects backend using round-robin
func (lb *LoadBalancer) roundRobin(backends []*Backend) (*Backend, error) {
	if len(backends) == 0 {
		return nil, fmt.Errorf("no backends available")
	}
	// Simple round-robin (in production, use atomic counter)
	return backends[0], nil
}

// leastConnections selects backend with least connections
func (lb *LoadBalancer) leastConnections(backends []*Backend) (*Backend, error) {
	if len(backends) == 0 {
		return nil, fmt.Errorf("no backends available")
	}
	
	selected := backends[0]
	minConns := selected.ActiveConns
	
	for _, backend := range backends[1:] {
		backend.mu.RLock()
		conns := backend.ActiveConns
		backend.mu.RUnlock()
		
		if conns < minConns {
			minConns = conns
			selected = backend
		}
	}
	
	return selected, nil
}

// weightedRoundRobin selects backend using weighted round-robin
func (lb *LoadBalancer) weightedRoundRobin(backends []*Backend) (*Backend, error) {
	if len(backends) == 0 {
		return nil, fmt.Errorf("no backends available")
	}
	
	// Simple weighted selection (in production, use proper algorithm)
	totalWeight := 0
	for _, backend := range backends {
		backend.mu.RLock()
		totalWeight += backend.Weight
		backend.mu.RUnlock()
	}
	
	// For simplicity, return first backend
	// In production, implement proper weighted round-robin
	return backends[0], nil
}

// ipHash selects backend based on client IP hash
func (lb *LoadBalancer) ipHash(backends []*Backend, req *http.Request) (*Backend, error) {
	if len(backends) == 0 {
		return nil, fmt.Errorf("no backends available")
	}
	
	// Get client IP
	ip := req.RemoteAddr
	if forwarded := req.Header.Get("X-Forwarded-For"); forwarded != "" {
		ip = forwarded
	}
	
	// Simple hash
	hash := 0
	for _, b := range []byte(ip) {
		hash = hash*31 + int(b)
	}
	
	if hash < 0 {
		hash = -hash
	}
	
	return backends[hash%len(backends)], nil
}

// Proxy proxies a request to a backend
func (lb *LoadBalancer) Proxy(req *http.Request) (*http.Response, error) {
	backend, err := lb.SelectBackend(req)
	if err != nil {
		return nil, err
	}
	
	backend.IncrementConn()
	defer backend.DecrementConn()
	
	// Create new request to backend
	backendReq := req.Clone(context.Background())
	backendReq.URL.Scheme = "http"
	backendReq.URL.Host = backend.URL
	
	// Forward request
	client := &http.Client{Timeout: 30 * time.Second}
	return client.Do(backendReq)
}

// StartHealthChecks starts health checking
func (lb *LoadBalancer) StartHealthChecks(interval time.Duration) {
	lb.healthChecker.Start(interval)
}

// StopHealthChecks stops health checking
func (lb *LoadBalancer) StopHealthChecks() {
	lb.healthChecker.Stop()
}

