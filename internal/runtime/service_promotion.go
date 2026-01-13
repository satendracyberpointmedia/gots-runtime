package runtime

import (
	"fmt"
	"sync"
)

// Service represents a promoted service
type Service struct {
	ID          string
	ModuleID    string
	Type        ServiceType
	Endpoints   []Endpoint
	HealthCheck func() bool
	mu          sync.RWMutex
}

// ServiceType represents the type of service
type ServiceType int

const (
	ServiceTypeHTTP ServiceType = iota
	ServiceTypeRPC
	ServiceTypeEvent
)

// Endpoint represents a service endpoint
type Endpoint struct {
	Path   string
	Method string
	Handler string
}

// ServiceRegistry manages service registration
type ServiceRegistry struct {
	services map[string]*Service
	mu       sync.RWMutex
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[string]*Service),
	}
}

// RegisterService registers a service
func (sr *ServiceRegistry) RegisterService(service *Service) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.services[service.ID] = service
}

// GetService gets a service by ID
func (sr *ServiceRegistry) GetService(id string) (*Service, bool) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	service, ok := sr.services[id]
	return service, ok
}

// ListServices lists all services
func (sr *ServiceRegistry) ListServices() []*Service {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	
	services := make([]*Service, 0, len(sr.services))
	for _, service := range sr.services {
		services = append(services, service)
	}
	return services
}

// ServicePromoter promotes a module to a service
type ServicePromoter struct {
	registry *ServiceRegistry
	graph    *ServiceGraph
}

// NewServicePromoter creates a new service promoter
func NewServicePromoter(registry *ServiceRegistry, graph *ServiceGraph) *ServicePromoter {
	return &ServicePromoter{
		registry: registry,
		graph:    graph,
	}
}

// PromoteModule promotes a module to a service
func (sp *ServicePromoter) PromoteModule(moduleID string, serviceType ServiceType, endpoints []Endpoint) (*Service, error) {
	// Check if module exists in graph
	_, ok := sp.graph.GetNode(moduleID)
	if !ok {
		return nil, fmt.Errorf("module not found: %s", moduleID)
	}
	
	service := &Service{
		ID:        fmt.Sprintf("service-%s", moduleID),
		ModuleID:  moduleID,
		Type:      serviceType,
		Endpoints: endpoints,
		HealthCheck: func() bool {
			return true // Default health check
		},
	}
	
	sp.registry.RegisterService(service)
	
	// Add service node to graph
	serviceNode := &ServiceNode{
		ID:          service.ID,
		ModuleID:    moduleID,
		ServiceType: serviceType.String(),
	}
	sp.graph.AddNode(serviceNode)
	
	return service, nil
}

// String returns the string representation of ServiceType
func (st ServiceType) String() string {
	switch st {
	case ServiceTypeHTTP:
		return "http"
	case ServiceTypeRPC:
		return "rpc"
	case ServiceTypeEvent:
		return "event"
	default:
		return "unknown"
	}
}

// ServiceDiscovery provides service discovery
type ServiceDiscovery struct {
	registry *ServiceRegistry
	mu       sync.RWMutex
}

// NewServiceDiscovery creates a new service discovery
func NewServiceDiscovery(registry *ServiceRegistry) *ServiceDiscovery {
	return &ServiceDiscovery{
		registry: registry,
	}
}

// DiscoverServices discovers services by type
func (sd *ServiceDiscovery) DiscoverServices(serviceType ServiceType) []*Service {
	services := sd.registry.ListServices()
	
	result := make([]*Service, 0)
	for _, service := range services {
		service.mu.RLock()
		if service.Type == serviceType {
			result = append(result, service)
		}
		service.mu.RUnlock()
	}
	
	return result
}

// DiscoverServiceByID discovers a service by ID
func (sd *ServiceDiscovery) DiscoverServiceByID(id string) (*Service, error) {
	service, ok := sd.registry.GetService(id)
	if !ok {
		return nil, fmt.Errorf("service not found: %s", id)
	}
	return service, nil
}

