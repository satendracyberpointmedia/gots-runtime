package runtime

import (
	"fmt"
	"sync"
)

// ServiceNode represents a node in the service graph
type ServiceNode struct {
	ID          string
	ModuleID    string
	ServiceType string
	Dependencies []string
	Dependents  []string
	mu          sync.RWMutex
}

// ServiceGraph represents the service dependency graph
type ServiceGraph struct {
	nodes map[string]*ServiceNode
	mu    sync.RWMutex
}

// NewServiceGraph creates a new service graph
func NewServiceGraph() *ServiceGraph {
	return &ServiceGraph{
		nodes: make(map[string]*ServiceNode),
	}
}

// AddNode adds a node to the graph
func (sg *ServiceGraph) AddNode(node *ServiceNode) {
	sg.mu.Lock()
	defer sg.mu.Unlock()
	sg.nodes[node.ID] = node
}

// GetNode gets a node by ID
func (sg *ServiceGraph) GetNode(id string) (*ServiceNode, bool) {
	sg.mu.RLock()
	defer sg.mu.RUnlock()
	node, ok := sg.nodes[id]
	return node, ok
}

// AddDependency adds a dependency between nodes
func (sg *ServiceGraph) AddDependency(fromID, toID string) error {
	sg.mu.Lock()
	defer sg.mu.Unlock()
	
	from, ok := sg.nodes[fromID]
	if !ok {
		return fmt.Errorf("node not found: %s", fromID)
	}
	
	to, ok := sg.nodes[toID]
	if !ok {
		return fmt.Errorf("node not found: %s", toID)
	}
	
	// Add dependency
	from.mu.Lock()
	from.Dependencies = append(from.Dependencies, toID)
	from.mu.Unlock()
	
	// Add dependent
	to.mu.Lock()
	to.Dependents = append(to.Dependents, fromID)
	to.mu.Unlock()
	
	return nil
}

// GetDependencies gets all dependencies for a node
func (sg *ServiceGraph) GetDependencies(nodeID string) ([]string, error) {
	sg.mu.RLock()
	defer sg.mu.RUnlock()
	
	node, ok := sg.nodes[nodeID]
	if !ok {
		return nil, fmt.Errorf("node not found: %s", nodeID)
	}
	
	node.mu.RLock()
	defer node.mu.RUnlock()
	
	result := make([]string, len(node.Dependencies))
	copy(result, node.Dependencies)
	return result, nil
}

// GetDependents gets all dependents for a node
func (sg *ServiceGraph) GetDependents(nodeID string) ([]string, error) {
	sg.mu.RLock()
	defer sg.mu.RUnlock()
	
	node, ok := sg.nodes[nodeID]
	if !ok {
		return nil, fmt.Errorf("node not found: %s", nodeID)
	}
	
	node.mu.RLock()
	defer node.mu.RUnlock()
	
	result := make([]string, len(node.Dependents))
	copy(result, node.Dependents)
	return result, nil
}

// TopologicalSort performs topological sort of the graph
func (sg *ServiceGraph) TopologicalSort() ([]string, error) {
	sg.mu.RLock()
	defer sg.mu.RUnlock()
	
	// Kahn's algorithm
	inDegree := make(map[string]int)
	for id := range sg.nodes {
		inDegree[id] = 0
	}
	
	// Calculate in-degrees
	for _, node := range sg.nodes {
		node.mu.RLock()
		for range node.Dependencies {
			inDegree[node.ID]++
		}
		node.mu.RUnlock()
	}
	
	// Find nodes with no dependencies
	queue := make([]string, 0)
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}
	
	result := make([]string, 0)
	for len(queue) > 0 {
		nodeID := queue[0]
		queue = queue[1:]
		result = append(result, nodeID)
		
		node, ok := sg.nodes[nodeID]
		if !ok {
			continue
		}
		node.mu.RLock()
		for _, dependent := range node.Dependents {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
		node.mu.RUnlock()
	}
	
	if len(result) != len(sg.nodes) {
		return nil, fmt.Errorf("circular dependency detected")
	}
	
	return result, nil
}

// GetAllNodes returns all nodes
func (sg *ServiceGraph) GetAllNodes() map[string]*ServiceNode {
	sg.mu.RLock()
	defer sg.mu.RUnlock()
	
	result := make(map[string]*ServiceNode)
	for k, v := range sg.nodes {
		result[k] = v
	}
	return result
}

