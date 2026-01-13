package runtime

import (
	"context"
	"fmt"
	"sync"

	"gots-runtime/internal/eventloop"
	"gots-runtime/internal/observability"
	"gots-runtime/internal/security"
	"gots-runtime/internal/tsengine"
)

// RuntimeIntegration provides the main integration layer
type RuntimeIntegration struct {
	orchestrator    *Orchestrator
	eventLoop       *eventloop.Loop
	tsEngine        *tsengine.Engine
	permManager     *security.PermissionManager
	sandboxManager  *security.SandboxManager
	healthEndpoint  *observability.HealthEndpoint
	logger          *observability.Logger
	metrics         *observability.MetricsCollector
	tracer          *observability.Tracer
	mu              sync.RWMutex
	initialized     bool
}

// NewRuntimeIntegration creates a new runtime integration
func NewRuntimeIntegration() *RuntimeIntegration {
	ctx := context.Background()
	
	// Create orchestrator
	orch := NewOrchestrator()
	
	// Create event loop
	eventLoop := eventloop.NewLoop(ctx)
	
	// Create TypeScript engine
	tsEngine := tsengine.NewEngine()
	
	// Create security managers
	permManager := security.NewPermissionManager()
	sandboxManager := security.NewSandboxManager()
	
	// Create observability
	logger := observability.NewLogger(observability.LogLevelInfo)
	metrics := observability.NewMetricsCollector()
	tracer := observability.NewTracer()
	healthEndpoint := observability.NewHealthEndpoint()
	
	return &RuntimeIntegration{
		orchestrator:   orch,
		eventLoop:      eventLoop,
		tsEngine:       tsEngine,
		permManager:    permManager,
		sandboxManager: sandboxManager,
		healthEndpoint: healthEndpoint,
		logger:         logger,
		metrics:        metrics,
		tracer:         tracer,
	}
}

// Initialize initializes the runtime integration
func (ri *RuntimeIntegration) Initialize() error {
	ri.mu.Lock()
	defer ri.mu.Unlock()
	
	if ri.initialized {
		return fmt.Errorf("runtime already initialized")
	}
	
	// Start orchestrator
	if err := ri.orchestrator.Start(); err != nil {
		return fmt.Errorf("failed to start orchestrator: %w", err)
	}
	
	// Start event loop
	ri.eventLoop.Start()
	
	// Create and start scheduler
	scheduler := NewAdvancedScheduler(ri.orchestrator.Context(), ri.eventLoop)
	scheduler.Start()
	ri.orchestrator.SetScheduler(scheduler)
	
	// Load and register standard library
	stdlibLoader := tsengine.NewStdlibLoader(ri.tsEngine)
	if err := stdlibLoader.Load(); err != nil {
		return fmt.Errorf("failed to load stdlib: %w", err)
	}
	if err := stdlibLoader.Register(); err != nil {
		return fmt.Errorf("failed to register stdlib: %w", err)
	}
	
	// Register default health checks
	ri.setupHealthChecks()
	
	ri.initialized = true
	ri.logger.Info("Runtime initialized successfully")
	
	return nil
}

// setupHealthChecks sets up default health checks
func (ri *RuntimeIntegration) setupHealthChecks() {
	ri.healthEndpoint.RegisterCheck("orchestrator", func() (observability.HealthStatus, string) {
		state := ri.orchestrator.State()
		if state == StateRunning {
			return observability.HealthStatusHealthy, "orchestrator is running"
		}
		return observability.HealthStatusUnhealthy, "orchestrator is not running"
	})
	
	ri.healthEndpoint.RegisterCheck("eventloop", func() (observability.HealthStatus, string) {
		if ri.eventLoop.IsOverloaded() {
			return observability.HealthStatusDegraded, "event loop is overloaded"
		}
		return observability.HealthStatusHealthy, "event loop is healthy"
	})
}

// GetOrchestrator returns the orchestrator
func (ri *RuntimeIntegration) GetOrchestrator() *Orchestrator {
	return ri.orchestrator
}

// GetEventLoop returns the event loop
func (ri *RuntimeIntegration) GetEventLoop() *eventloop.Loop {
	return ri.eventLoop
}

// GetTSEngine returns the TypeScript engine
func (ri *RuntimeIntegration) GetTSEngine() *tsengine.Engine {
	return ri.tsEngine
}

// GetPermissionManager returns the permission manager
func (ri *RuntimeIntegration) GetPermissionManager() *security.PermissionManager {
	return ri.permManager
}

// GetSandboxManager returns the sandbox manager
func (ri *RuntimeIntegration) GetSandboxManager() *security.SandboxManager {
	return ri.sandboxManager
}

// GetHealthEndpoint returns the health endpoint
func (ri *RuntimeIntegration) GetHealthEndpoint() *observability.HealthEndpoint {
	return ri.healthEndpoint
}

// GetLogger returns the logger
func (ri *RuntimeIntegration) GetLogger() *observability.Logger {
	return ri.logger
}

// GetMetrics returns the metrics collector
func (ri *RuntimeIntegration) GetMetrics() *observability.MetricsCollector {
	return ri.metrics
}

// GetTracer returns the tracer
func (ri *RuntimeIntegration) GetTracer() *observability.Tracer {
	return ri.tracer
}

// RegisterModule registers a module with security policy
func (ri *RuntimeIntegration) RegisterModule(moduleID string, permissions ...security.Permission) error {
	policy := security.NewPolicy(moduleID)
	for _, perm := range permissions {
		policy.Allow(perm)
	}
	
	ri.permManager.RegisterPolicy(moduleID, policy)
	ri.logger.Info("Module registered: %s", moduleID)
	
	return nil
}

// ExecuteModule executes a TypeScript module
func (ri *RuntimeIntegration) ExecuteModule(moduleID, filePath string) error {
	// Register APIs for this module
	bindings := tsengine.NewRuntimeBindings(
		ri.tsEngine,
		ri.eventLoop,
		ri.permManager,
		moduleID,
	)
	
	if err := bindings.RegisterAPIs(); err != nil {
		return fmt.Errorf("failed to register APIs: %w", err)
	}
	
	// Execute the module
	_, err := ri.tsEngine.ExecuteFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to execute module: %w", err)
	}
	
	ri.metrics.Increment("modules.executed", map[string]string{"module": moduleID})
	ri.logger.Info("Module executed: %s", moduleID)
	
	return nil
}

// Shutdown shuts down the runtime
func (ri *RuntimeIntegration) Shutdown() error {
	ri.mu.Lock()
	defer ri.mu.Unlock()
	
	if !ri.initialized {
		return nil
	}
	
	ri.logger.Info("Shutting down runtime...")
	
	// Stop event loop
	ri.eventLoop.Stop()
	
	// Stop orchestrator
	if err := ri.orchestrator.Stop(); err != nil {
		return fmt.Errorf("failed to stop orchestrator: %w", err)
	}
	
	ri.initialized = false
	ri.logger.Info("Runtime shut down successfully")
	
	return nil
}

