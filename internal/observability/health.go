package observability

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// HealthStatus represents health status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
)

// HealthCheck represents a health check
type HealthCheck struct {
	Name      string
	Status    HealthStatus
	Message   string
	Timestamp time.Time
}

// HealthEndpoint provides health check endpoints
type HealthEndpoint struct {
	checks map[string]*HealthCheck
	mu     sync.RWMutex
}

// NewHealthEndpoint creates a new health endpoint
func NewHealthEndpoint() *HealthEndpoint {
	return &HealthEndpoint{
		checks: make(map[string]*HealthCheck),
	}
}

// RegisterCheck registers a health check
func (he *HealthEndpoint) RegisterCheck(name string, check func() (HealthStatus, string)) {
	he.mu.Lock()
	defer he.mu.Unlock()

	status, message := check()
	he.checks[name] = &HealthCheck{
		Name:      name,
		Status:    status,
		Message:   message,
		Timestamp: time.Now(),
	}
}

// GetHealth returns the overall health status
func (he *HealthEndpoint) GetHealth() HealthStatus {
	he.mu.RLock()
	defer he.mu.RUnlock()

	hasUnhealthy := false
	hasDegraded := false

	for _, check := range he.checks {
		if check.Status == HealthStatusUnhealthy {
			hasUnhealthy = true
		} else if check.Status == HealthStatusDegraded {
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return HealthStatusUnhealthy
	}
	if hasDegraded {
		return HealthStatusDegraded
	}
	return HealthStatusHealthy
}

// Handler returns an HTTP handler for health checks
func (he *HealthEndpoint) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		he.mu.RLock()
		health := he.GetHealth()
		checks := make(map[string]*HealthCheck)
		for k, v := range he.checks {
			checks[k] = v
		}
		he.mu.RUnlock()

		response := map[string]interface{}{
			"status": health,
			"checks": checks,
		}

		w.Header().Set("Content-Type", "application/json")
		if health == HealthStatusUnhealthy {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		json.NewEncoder(w).Encode(response)
	}
}

// ReadinessHandler returns an HTTP handler for readiness checks
func (he *HealthEndpoint) ReadinessHandler() http.HandlerFunc {
	return he.Handler()
}

