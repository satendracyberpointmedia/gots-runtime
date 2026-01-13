package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
)

// RPCRequest represents an RPC request
type RPCRequest struct {
	ID      string
	Method  string
	Params  json.RawMessage
	Module  string
}

// RPCResponse represents an RPC response
type RPCResponse struct {
	ID     string
	Result interface{}
	Error  *RPCError
}

// RPCError represents an RPC error
type RPCError struct {
	Code    int
	Message string
	Data    interface{}
}

// RPCHandler handles RPC calls
type RPCHandler func(ctx context.Context, params json.RawMessage) (interface{}, error)

// RPCServer provides native RPC functionality
type RPCServer struct {
	handlers map[string]RPCHandler
	listener net.Listener
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewRPCServer creates a new RPC server
func NewRPCServer(ctx context.Context) *RPCServer {
	rpcCtx, cancel := context.WithCancel(ctx)
	return &RPCServer{
		handlers: make(map[string]RPCHandler),
		ctx:      rpcCtx,
		cancel:   cancel,
	}
}

// RegisterHandler registers an RPC handler
func (rs *RPCServer) RegisterHandler(method string, handler RPCHandler) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.handlers[method] = handler
}

// Listen starts listening on an address
func (rs *RPCServer) Listen(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	
	rs.mu.Lock()
	rs.listener = listener
	rs.mu.Unlock()
	
	go rs.accept()
	return nil
}

// accept accepts connections
func (rs *RPCServer) accept() {
	for {
		select {
		case <-rs.ctx.Done():
			return
		default:
			conn, err := rs.listener.Accept()
			if err != nil {
				continue
			}
			
			go rs.handleConnection(conn)
		}
	}
}

// handleConnection handles a connection
func (rs *RPCServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	
	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)
	
	for {
		var req RPCRequest
		if err := decoder.Decode(&req); err != nil {
			return
		}
		
		response := rs.handleRequest(&req)
		if err := encoder.Encode(response); err != nil {
			return
		}
	}
}

// handleRequest handles an RPC request
func (rs *RPCServer) handleRequest(req *RPCRequest) *RPCResponse {
	rs.mu.RLock()
	handler, ok := rs.handlers[req.Method]
	rs.mu.RUnlock()
	
	if !ok {
		return &RPCResponse{
			ID: req.ID,
			Error: &RPCError{
				Code:    -32601,
				Message: "Method not found",
			},
		}
	}
	
	result, err := handler(rs.ctx, req.Params)
	if err != nil {
		return &RPCResponse{
			ID: req.ID,
			Error: &RPCError{
				Code:    -32000,
				Message: err.Error(),
			},
		}
	}
	
	return &RPCResponse{
		ID:     req.ID,
		Result: result,
	}
}

// Stop stops the RPC server
func (rs *RPCServer) Stop() error {
	rs.cancel()
	
	rs.mu.RLock()
	listener := rs.listener
	rs.mu.RUnlock()
	
	if listener != nil {
		return listener.Close()
	}
	return nil
}

