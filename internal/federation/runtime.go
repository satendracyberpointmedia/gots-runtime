package federation

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

// RuntimeNode represents a node in the federation
type RuntimeNode struct {
	ID       string
	Address  string
	Healthy  bool
	LastSeen time.Time
	mu       sync.RWMutex
}

// NewRuntimeNode creates a new runtime node
func NewRuntimeNode(id, address string) *RuntimeNode {
	return &RuntimeNode{
		ID:       id,
		Address:  address,
		Healthy:  true,
		LastSeen: time.Now(),
	}
}

// SetHealthy sets the health status
func (rn *RuntimeNode) SetHealthy(healthy bool) {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	rn.Healthy = healthy
	rn.LastSeen = time.Now()
}

// FederationMessage represents a federation message
type FederationMessage struct {
	Type      string
	From      string
	To        string
	Payload   json.RawMessage
	Timestamp time.Time
}

// Federation provides multi-runtime federation
type Federation struct {
	localID  string
	nodes    map[string]*RuntimeNode
	listener net.Listener
	handlers map[string]MessageHandler
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

// MessageHandler handles federation messages
type MessageHandler func(ctx context.Context, msg *FederationMessage) (*FederationMessage, error)

// NewFederation creates a new federation
func NewFederation(localID string, ctx context.Context) *Federation {
	fedCtx, cancel := context.WithCancel(ctx)
	return &Federation{
		localID:  localID,
		nodes:    make(map[string]*RuntimeNode),
		handlers: make(map[string]MessageHandler),
		ctx:      fedCtx,
		cancel:   cancel,
	}
}

// RegisterNode registers a node
func (f *Federation) RegisterNode(node *RuntimeNode) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nodes[node.ID] = node
}

// UnregisterNode unregisters a node
func (f *Federation) UnregisterNode(nodeID string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.nodes, nodeID)
}

// RegisterHandler registers a message handler
func (f *Federation) RegisterHandler(msgType string, handler MessageHandler) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.handlers[msgType] = handler
}

// Send sends a message to a node
func (f *Federation) Send(nodeID string, msgType string, payload interface{}) error {
	f.mu.RLock()
	node, ok := f.nodes[nodeID]
	f.mu.RUnlock()

	if !ok {
		return fmt.Errorf("node not found: %s", nodeID)
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	msg := &FederationMessage{
		Type:      msgType,
		From:      f.localID,
		To:        nodeID,
		Payload:   payloadJSON,
		Timestamp: time.Now(),
	}

	return f.sendMessage(node.Address, msg)
}

// Broadcast broadcasts a message to all nodes
func (f *Federation) Broadcast(msgType string, payload interface{}) error {
	f.mu.RLock()
	nodes := make([]*RuntimeNode, 0, len(f.nodes))
	for _, node := range f.nodes {
		if node.Healthy {
			nodes = append(nodes, node)
		}
	}
	f.mu.RUnlock()

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	for _, node := range nodes {
		msg := &FederationMessage{
			Type:      msgType,
			From:      f.localID,
			To:        node.ID,
			Payload:   payloadJSON,
			Timestamp: time.Now(),
		}

		_ = f.sendMessage(node.Address, msg)
	}

	return nil
}

// sendMessage sends a message to an address
func (f *Federation) sendMessage(address string, msg *FederationMessage) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	return encoder.Encode(msg)
}

// Listen starts listening for federation messages
func (f *Federation) Listen(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	f.mu.Lock()
	f.listener = listener
	f.mu.Unlock()

	go f.accept()
	return nil
}

// accept accepts connections
func (f *Federation) accept() {
	for {
		select {
		case <-f.ctx.Done():
			return
		default:
			f.mu.RLock()
			listener := f.listener
			f.mu.RUnlock()

			if listener == nil {
				return
			}

			conn, err := listener.Accept()
			if err != nil {
				continue
			}

			go f.handleConnection(conn)
		}
	}
}

// handleConnection handles a connection
func (f *Federation) handleConnection(conn net.Conn) {
	defer conn.Close()

	decoder := json.NewDecoder(conn)

	var msg FederationMessage
	if err := decoder.Decode(&msg); err != nil {
		return
	}

	// Handle message
	f.mu.RLock()
	handler, ok := f.handlers[msg.Type]
	f.mu.RUnlock()

	if !ok {
		return
	}

	response, err := handler(f.ctx, &msg)
	if err != nil {
		return
	}

	if response != nil {
		encoder := json.NewEncoder(conn)
		_ = encoder.Encode(response)
	}
}

// Stop stops the federation
func (f *Federation) Stop() error {
	f.cancel()

	f.mu.RLock()
	listener := f.listener
	f.mu.RUnlock()

	if listener != nil {
		return listener.Close()
	}
	return nil
}

// GetNodeStats returns statistics for a specific node
func (f *Federation) GetNodeStats(nodeID string) (*NodeStats, bool) {
	f.mu.RLock()
	node, ok := f.nodes[nodeID]
	f.mu.RUnlock()

	if !ok {
		return nil, false
	}

	node.mu.RLock()
	defer node.mu.RUnlock()

	return &NodeStats{
		ID:       node.ID,
		Address:  node.Address,
		Healthy:  node.Healthy,
		LastSeen: node.LastSeen,
	}, true
}

// DiscoverNodes discovers new nodes (for dynamic federation)
func (f *Federation) DiscoverNodes(discoverAddress string) error {
	// In a real implementation, this would contact a discovery service
	// For now, this is a placeholder
	return nil
}

// NodeStats represents statistics about a node
type NodeStats struct {
	ID       string
	Address  string
	Healthy  bool
	LastSeen time.Time
}
