package runtime

import (
	"fmt"
	"sync"
)

// Domain represents a domain in DDD
type Domain struct {
	Name        string
	Modules     []string
	Boundaries  []DomainBoundary
	mu          sync.RWMutex
}

// DomainBoundary represents a boundary between domains
type DomainBoundary struct {
	FromDomain string
	ToDomain   string
	Allowed    bool
	Protocol   string // RPC, HTTP, etc.
}

// DDDEnforcer enforces Domain-Driven Design principles
type DDDEnforcer struct {
	domains map[string]*Domain
	graph   *ServiceGraph
	mu      sync.RWMutex
}

// NewDDDEnforcer creates a new DDD enforcer
func NewDDDEnforcer(graph *ServiceGraph) *DDDEnforcer {
	return &DDDEnforcer{
		domains: make(map[string]*Domain),
		graph:   graph,
	}
}

// RegisterDomain registers a domain
func (de *DDDEnforcer) RegisterDomain(name string, modules []string) {
	de.mu.Lock()
	defer de.mu.Unlock()
	
	de.domains[name] = &Domain{
		Name:   name,
		Modules: modules,
		Boundaries: make([]DomainBoundary, 0),
	}
}

// AddBoundary adds a boundary between domains
func (de *DDDEnforcer) AddBoundary(fromDomain, toDomain, protocol string, allowed bool) {
	de.mu.Lock()
	defer de.mu.Unlock()
	
	from, ok := de.domains[fromDomain]
	if !ok {
		return
	}
	
	from.mu.Lock()
	from.Boundaries = append(from.Boundaries, DomainBoundary{
		FromDomain: fromDomain,
		ToDomain:   toDomain,
		Allowed:    allowed,
		Protocol:   protocol,
	})
	from.mu.Unlock()
}

// CheckAccess checks if a module can access another module
func (de *DDDEnforcer) CheckAccess(fromModule, toModule string) error {
	de.mu.RLock()
	defer de.mu.RUnlock()
	
	// Find domains for modules
	fromDomain := de.findDomainForModule(fromModule)
	toDomain := de.findDomainForModule(toModule)
	
	if fromDomain == "" || toDomain == "" {
		return fmt.Errorf("module not in any domain")
	}
	
	// Same domain, allow
	if fromDomain == toDomain {
		return nil
	}
	
	// Check boundary
	domain, ok := de.domains[fromDomain]
	if !ok {
		return fmt.Errorf("domain not found: %s", fromDomain)
	}
	
	domain.mu.RLock()
	defer domain.mu.RUnlock()
	
	for _, boundary := range domain.Boundaries {
		if boundary.ToDomain == toDomain {
			if !boundary.Allowed {
				return fmt.Errorf("access denied: boundary between %s and %s is not allowed", 
					fromDomain, toDomain)
			}
			return nil
		}
	}
	
	// No boundary defined, deny by default
	return fmt.Errorf("access denied: no boundary defined between %s and %s", 
		fromDomain, toDomain)
}

// findDomainForModule finds the domain for a module
func (de *DDDEnforcer) findDomainForModule(moduleID string) string {
	for domainName, domain := range de.domains {
		domain.mu.RLock()
		for _, mod := range domain.Modules {
			if mod == moduleID {
				domain.mu.RUnlock()
				return domainName
			}
		}
		domain.mu.RUnlock()
	}
	return ""
}

// EnforceModuleBoundaries enforces module boundaries
func (de *DDDEnforcer) EnforceModuleBoundaries() error {
	// Get all nodes from service graph
	nodes := de.graph.GetAllNodes()
	
	for nodeID, node := range nodes {
		node.mu.RLock()
		dependencies := node.Dependencies
		node.mu.RUnlock()
		
		for _, depID := range dependencies {
			if err := de.CheckAccess(nodeID, depID); err != nil {
				return fmt.Errorf("boundary violation: %s -> %s: %w", nodeID, depID, err)
			}
		}
	}
	
	return nil
}

// GetDomainForModule gets the domain for a module
func (de *DDDEnforcer) GetDomainForModule(moduleID string) (string, error) {
	de.mu.RLock()
	defer de.mu.RUnlock()
	
	domain := de.findDomainForModule(moduleID)
	if domain == "" {
		return "", fmt.Errorf("module not in any domain: %s", moduleID)
	}
	
	return domain, nil
}

