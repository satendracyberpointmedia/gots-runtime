package runtime

import (
	"fmt"
	"sync"
	"time"
)

// Domain represents a domain in DDD
type Domain struct {
	Name       string
	Modules    []string
	Boundaries []DomainBoundary
	mu         sync.RWMutex
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
		Name:       name,
		Modules:    modules,
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

// DDDValidator validates DDD compliance
type DDDValidator struct {
	enforcer      *DDDEnforcer
	violations    []BoundaryViolation
	maxViolations int
	mu            sync.RWMutex
}

// BoundaryViolation represents a DDD boundary violation
type BoundaryViolation struct {
	FromModule string
	ToModule   string
	FromDomain string
	ToDomain   string
	Timestamp  time.Time
	Severity   string // "warning", "critical"
}

// NewDDDValidator creates a new DDD validator
func NewDDDValidator(enforcer *DDDEnforcer) *DDDValidator {
	return &DDDValidator{
		enforcer:      enforcer,
		violations:    make([]BoundaryViolation, 0),
		maxViolations: 100,
	}
}

// ValidateModule validates a module
func (dv *DDDValidator) ValidateModule(moduleID string) error {
	return dv.enforcer.CheckAccess(moduleID, moduleID)
}

// ValidateDependency validates a dependency
func (dv *DDDValidator) ValidateDependency(fromModule, toModule string) error {
	err := dv.enforcer.CheckAccess(fromModule, toModule)

	if err != nil {
		dv.mu.Lock()
		defer dv.mu.Unlock()

		fromDomain := dv.enforcer.findDomainForModule(fromModule)
		toDomain := dv.enforcer.findDomainForModule(toModule)

		violation := BoundaryViolation{
			FromModule: fromModule,
			ToModule:   toModule,
			FromDomain: fromDomain,
			ToDomain:   toDomain,
			Timestamp:  time.Now(),
			Severity:   "critical",
		}

		dv.violations = append(dv.violations, violation)
		if len(dv.violations) > dv.maxViolations {
			dv.violations = dv.violations[1:]
		}
	}

	return err
}

// ValidateAllDependencies validates all dependencies in the graph
func (dv *DDDValidator) ValidateAllDependencies() []BoundaryViolation {
	violations := make([]BoundaryViolation, 0)

	nodes := dv.enforcer.graph.GetAllNodes()
	for nodeID, node := range nodes {
		node.mu.RLock()
		dependencies := node.Dependencies
		node.mu.RUnlock()

		for _, depID := range dependencies {
			if err := dv.ValidateDependency(nodeID, depID); err != nil {
				fromDomain := dv.enforcer.findDomainForModule(nodeID)
				toDomain := dv.enforcer.findDomainForModule(depID)

				violations = append(violations, BoundaryViolation{
					FromModule: nodeID,
					ToModule:   depID,
					FromDomain: fromDomain,
					ToDomain:   toDomain,
					Timestamp:  time.Now(),
					Severity:   "critical",
				})
			}
		}
	}

	return violations
}

// GetViolations returns all recorded violations
func (dv *DDDValidator) GetViolations() []BoundaryViolation {
	dv.mu.RLock()
	defer dv.mu.RUnlock()

	result := make([]BoundaryViolation, len(dv.violations))
	copy(result, dv.violations)
	return result
}

// ClearViolations clears all recorded violations
func (dv *DDDValidator) ClearViolations() {
	dv.mu.Lock()
	defer dv.mu.Unlock()
	dv.violations = make([]BoundaryViolation, 0)
}

// GetViolationStats returns statistics about violations
func (dv *DDDValidator) GetViolationStats() map[string]interface{} {
	dv.mu.RLock()
	defer dv.mu.RUnlock()

	criticalCount := 0
	warningCount := 0

	for _, v := range dv.violations {
		if v.Severity == "critical" {
			criticalCount++
		} else {
			warningCount++
		}
	}

	return map[string]interface{}{
		"total_violations":    len(dv.violations),
		"critical_violations": criticalCount,
		"warning_violations":  warningCount,
		"max_violations":      dv.maxViolations,
	}
}

// DomainGraph represents the graph of domains and their relationships
type DomainGraph struct {
	domains map[string]*Domain
	edges   map[string][]string // fromDomain -> [toDomains]
	mu      sync.RWMutex
}

// NewDomainGraph creates a new domain graph
func NewDomainGraph() *DomainGraph {
	return &DomainGraph{
		domains: make(map[string]*Domain),
		edges:   make(map[string][]string),
	}
}

// AddDomain adds a domain to the graph
func (dg *DomainGraph) AddDomain(domain *Domain) {
	dg.mu.Lock()
	defer dg.mu.Unlock()
	dg.domains[domain.Name] = domain
	if _, ok := dg.edges[domain.Name]; !ok {
		dg.edges[domain.Name] = make([]string, 0)
	}
}

// AddEdge adds an edge between domains
func (dg *DomainGraph) AddEdge(fromDomain, toDomain string) {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	if edges, ok := dg.edges[fromDomain]; ok {
		dg.edges[fromDomain] = append(edges, toDomain)
	}
}

// GetDomainDependencies gets the dependencies of a domain
func (dg *DomainGraph) GetDomainDependencies(domainName string) []string {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	if edges, ok := dg.edges[domainName]; ok {
		result := make([]string, len(edges))
		copy(result, edges)
		return result
	}
	return []string{}
}

// DetectCycles detects cyclic dependencies between domains
func (dg *DomainGraph) DetectCycles() [][]string {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	cycles := make([][]string, 0)
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for domain := range dg.domains {
		if !visited[domain] {
			if cycle := dg.detectCycleDFS(domain, visited, recStack, []string{}); cycle != nil {
				cycles = append(cycles, cycle)
			}
		}
	}

	return cycles
}

// detectCycleDFS uses DFS to detect cycles
func (dg *DomainGraph) detectCycleDFS(node string, visited, recStack map[string]bool, path []string) []string {
	visited[node] = true
	recStack[node] = true
	path = append(path, node)

	for _, neighbor := range dg.edges[node] {
		if !visited[neighbor] {
			if cycle := dg.detectCycleDFS(neighbor, visited, recStack, path); cycle != nil {
				return cycle
			}
		} else if recStack[neighbor] {
			// Found a cycle
			cycleStart := -1
			for i, n := range path {
				if n == neighbor {
					cycleStart = i
					break
				}
			}
			if cycleStart != -1 {
				return append(path[cycleStart:], neighbor)
			}
		}
	}

	recStack[node] = false
	return nil
}
