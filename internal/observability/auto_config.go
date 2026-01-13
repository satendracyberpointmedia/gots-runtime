package observability

import (
	"net/http"
	"sync"
)

// AutoConfig provides zero-config observability setup
type AutoConfig struct {
	logger         *Logger
	metrics        *MetricsCollector
	tracer         *Tracer
	healthEndpoint *HealthEndpoint
	httpServer     *http.Server
	mu             sync.RWMutex
	enabled        bool
}

// NewAutoConfig creates a new auto-config
func NewAutoConfig() *AutoConfig {
	return &AutoConfig{
		logger:         NewLogger(LogLevelInfo),
		metrics:        NewMetricsCollector(),
		tracer:         NewTracer(),
		healthEndpoint: NewHealthEndpoint(),
		enabled:        true,
	}
}

// Setup automatically sets up observability
func (ac *AutoConfig) Setup() error {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	
	if !ac.enabled {
		return nil
	}
	
	// Setup default health checks
	ac.setupDefaultHealthChecks()
	
	// Setup metrics collection
	ac.setupMetrics()
	
	// Setup logging
	ac.setupLogging()
	
	return nil
}

// setupDefaultHealthChecks sets up default health checks
func (ac *AutoConfig) setupDefaultHealthChecks() {
	ac.healthEndpoint.RegisterCheck("runtime", func() (HealthStatus, string) {
		return HealthStatusHealthy, "runtime is operational"
	})
}

// setupMetrics sets up metrics collection
func (ac *AutoConfig) setupMetrics() {
	// Register default metrics
	ac.metrics.Increment("runtime.started", nil)
}

// setupLogging sets up logging
func (ac *AutoConfig) setupLogging() {
	ac.logger.Info("Observability auto-configured")
}

// StartHealthServer starts the health server
func (ac *AutoConfig) StartHealthServer(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", ac.healthEndpoint.Handler())
	mux.HandleFunc("/ready", ac.healthEndpoint.ReadinessHandler())
	mux.HandleFunc("/metrics", ac.metricsHandler())
	
	ac.mu.Lock()
	ac.httpServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	ac.mu.Unlock()
	
	go func() {
		if err := ac.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			ac.logger.Error("Health server error: %v", err)
		}
	}()
	
	ac.logger.Info("Health server started on %s", addr)
	return nil
}

// metricsHandler returns a handler for metrics endpoint
func (ac *AutoConfig) metricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = ac.metrics.GetAll()
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		
		// Simple JSON output (in production, use proper format like Prometheus)
		_, _ = w.Write([]byte("{\"metrics\": \"enabled\"}\n"))
	}
}

// GetLogger returns the logger
func (ac *AutoConfig) GetLogger() *Logger {
	return ac.logger
}

// GetMetrics returns the metrics collector
func (ac *AutoConfig) GetMetrics() *MetricsCollector {
	return ac.metrics
}

// GetTracer returns the tracer
func (ac *AutoConfig) GetTracer() *Tracer {
	return ac.tracer
}

// GetHealthEndpoint returns the health endpoint
func (ac *AutoConfig) GetHealthEndpoint() *HealthEndpoint {
	return ac.healthEndpoint
}

// Enable enables auto-config
func (ac *AutoConfig) Enable() {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.enabled = true
}

// Disable disables auto-config
func (ac *AutoConfig) Disable() {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.enabled = false
}

// Stop stops the health server
func (ac *AutoConfig) Stop() error {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	
	if ac.httpServer != nil {
		return ac.httpServer.Close()
	}
	return nil
}

