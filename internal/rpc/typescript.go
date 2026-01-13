package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/dop251/goja"
)

// TypeScriptRPCServer wraps RPC server for TypeScript
type TypeScriptRPCServer struct {
	server  *RPCServer
	engine  *goja.Runtime
	ctx     context.Context
	mu      sync.RWMutex
}

// NewTypeScriptRPCServer creates a new TypeScript-wrapped RPC server
func NewTypeScriptRPCServer(engine *goja.Runtime, ctx context.Context) *TypeScriptRPCServer {
	return &TypeScriptRPCServer{
		server: NewRPCServer(ctx),
		engine: engine,
		ctx:    ctx,
	}
}

// ToJSObject converts the RPC server to a JavaScript object
func (tsr *TypeScriptRPCServer) ToJSObject() *goja.Object {
	obj := tsr.engine.NewObject()
	
	// Register method
	obj.Set("register", func(method string, handler goja.Value) {
		handlerFunc, ok := goja.AssertFunction(handler)
		if !ok {
			panic(tsr.engine.ToValue("handler must be a function"))
		}
		
		tsr.server.RegisterHandler(method, func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			// Parse params
			var paramsData interface{}
			if len(params) > 0 {
				if err := json.Unmarshal(params, &paramsData); err != nil {
					return nil, fmt.Errorf("failed to parse params: %w", err)
				}
			}
			
			// Call TypeScript handler
			result, err := handlerFunc(nil, tsr.engine.ToValue(paramsData))
			if err != nil {
				return nil, fmt.Errorf("handler error: %w", err)
			}
			
			// Export the result
			return result.Export(), nil
		})
	})
	
	// Unregister method
	obj.Set("unregister", func(method string) {
		// Note: The Go RPC server doesn't have unregister, so we'll register a nil handler
		// In a full implementation, we'd add unregister to the Go server
		tsr.mu.Lock()
		defer tsr.mu.Unlock()
		// For now, we'll just log that unregister was called
	})
	
	// Listen method
	obj.Set("listen", func(address string, callback goja.Value) {
		err := tsr.server.Listen(address)
		
		if callback != nil {
			if callable, ok := goja.AssertFunction(callback); ok {
				if err != nil {
					_, _ = callable(nil, tsr.engine.ToValue(err.Error()))
				} else {
					_, _ = callable(nil, goja.Undefined())
				}
			}
		}
	})
	
	// Close method
	obj.Set("close", func(callback goja.Value) {
		err := tsr.server.Stop()
		
		if callback != nil {
			if callable, ok := goja.AssertFunction(callback); ok {
				if err != nil {
					_, _ = callable(nil, tsr.engine.ToValue(err.Error()))
				} else {
					_, _ = callable(nil, goja.Undefined())
				}
			}
		}
	})
	
	return obj
}

// TypeScriptRPCClient wraps RPC client for TypeScript
type TypeScriptRPCClient struct {
	client *RPCClient
	engine *goja.Runtime
	mu     sync.RWMutex
}

// NewTypeScriptRPCClient creates a new TypeScript-wrapped RPC client
func NewTypeScriptRPCClient(engine *goja.Runtime, address string) (*TypeScriptRPCClient, error) {
	client, err := NewRPCClient(address)
	if err != nil {
		return nil, err
	}
	
	return &TypeScriptRPCClient{
		client: client,
		engine: engine,
	}, nil
}

// ToJSObject converts the RPC client to a JavaScript object
func (tsc *TypeScriptRPCClient) ToJSObject() *goja.Object {
	obj := tsc.engine.NewObject()
	
	// Call method
	obj.Set("call", func(method string, params goja.Value) *goja.Promise {
		promise, resolve, reject := tsc.engine.NewPromise()
		
		go func() {
			var paramsData interface{}
			if params != nil && !goja.IsUndefined(params) {
				paramsData = params.Export()
			}
			
			result, err := tsc.client.Call(method, paramsData)
			if err != nil {
				reject(tsc.engine.ToValue(err.Error()))
			} else {
				resolve(tsc.engine.ToValue(result))
			}
		}()
		
		return promise
	})
	
	// Close method
	obj.Set("close", func() *goja.Promise {
		promise, resolve, reject := tsc.engine.NewPromise()
		
		go func() {
			if err := tsc.client.Close(); err != nil {
				reject(tsc.engine.ToValue(err.Error()))
			} else {
				resolve(tsc.engine.ToValue(true))
			}
		}()
		
		return promise
	})
	
	return obj
}

