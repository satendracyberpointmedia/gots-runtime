package rpc

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
)

// RPCClient provides RPC client functionality
type RPCClient struct {
	conn   net.Conn
	encoder *json.Encoder
	decoder *json.Decoder
	mu     sync.Mutex
	idGen  uint64
}

// NewRPCClient creates a new RPC client
func NewRPCClient(address string) (*RPCClient, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	
	return &RPCClient{
		conn:    conn,
		encoder: json.NewEncoder(conn),
		decoder: json.NewDecoder(conn),
	}, nil
}

// Call makes an RPC call
func (rc *RPCClient) Call(method string, params interface{}) (interface{}, error) {
	rc.mu.Lock()
	id := fmt.Sprintf("req-%d", rc.idGen)
	rc.idGen++
	rc.mu.Unlock()
	
	req := &RPCRequest{
		ID:     id,
		Method: method,
	}
	
	if params != nil {
		paramsData, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		}
		req.Params = paramsData
	}
	
	if err := rc.encoder.Encode(req); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	
	var response RPCResponse
	if err := rc.decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to receive response: %w", err)
	}
	
	if response.Error != nil {
		return nil, fmt.Errorf("RPC error: %s", response.Error.Message)
	}
	
	return response.Result, nil
}

// Close closes the client connection
func (rc *RPCClient) Close() error {
	return rc.conn.Close()
}

