package framework

import (
	"fmt"
	"sync"

	"github.com/dop251/goja"
	"gots-runtime/framework/runtime"
	"gots-runtime/internal/api"
	"gots-runtime/internal/eventloop"
)

// TypeScriptApp wraps the Go App for TypeScript
type TypeScriptApp struct {
	app      *runtime.App
	engine   *goja.Runtime
	eventLoop *eventloop.Loop
	httpAPI  *api.HTTP
	server   *api.Server
	mu       sync.RWMutex
}

// NewTypeScriptApp creates a new TypeScript-wrapped app
func NewTypeScriptApp(engine *goja.Runtime, eventLoop *eventloop.Loop, name string) *TypeScriptApp {
	app := runtime.NewApp(name)
	httpAPI := api.NewHTTP(eventLoop)
	
	return &TypeScriptApp{
		app:      app,
		engine:   engine,
		eventLoop: eventLoop,
		httpAPI:  httpAPI,
	}
}

// ToJSObject converts the app to a JavaScript object
func (tsa *TypeScriptApp) ToJSObject() *goja.Object {
	obj := tsa.engine.NewObject()
	
	// Use method - add middleware
	obj.Set("use", func(middleware goja.Value) {
		mwFunc, ok := goja.AssertFunction(middleware)
		if !ok {
			panic(tsa.engine.ToValue("middleware must be a function"))
		}
		
		tsa.app.Use(func(ctx *runtime.Context, next runtime.Next) error {
			// Create TypeScript context
			tsCtx := tsa.createContextObject(ctx)
			
			// Call TypeScript middleware
			nextFunc := tsa.engine.NewObject()
			nextFunc.Set("call", func() *goja.Promise {
				promise, resolve, reject := tsa.engine.NewPromise()
				go func() {
					if err := next(); err != nil {
						reject(tsa.engine.ToValue(err.Error()))
					} else {
						resolve(tsa.engine.ToValue(true))
					}
				}()
				return promise
			})
			
			result, err := mwFunc(nil, tsCtx, nextFunc)
			if err != nil {
				return fmt.Errorf("middleware error: %w", err)
			}
			
			// If middleware returns a promise, wait for it
			// For now, we'll execute synchronously
			_ = result
			
			return nil
		})
	})
	
	// Get method
	obj.Set("get", func(path string, handler goja.Value) {
		handlerFunc, ok := goja.AssertFunction(handler)
		if !ok {
			panic(tsa.engine.ToValue("handler must be a function"))
		}
		
		tsa.app.Get(path, func(ctx *runtime.Context) error {
			tsCtx := tsa.createContextObject(ctx)
			_, err := handlerFunc(nil, tsCtx)
			return err
		})
	})
	
	// Post method
	obj.Set("post", func(path string, handler goja.Value) {
		handlerFunc, ok := goja.AssertFunction(handler)
		if !ok {
			panic(tsa.engine.ToValue("handler must be a function"))
		}
		
		tsa.app.Post(path, func(ctx *runtime.Context) error {
			tsCtx := tsa.createContextObject(ctx)
			_, err := handlerFunc(nil, tsCtx)
			return err
		})
	})
	
	// Put method
	obj.Set("put", func(path string, handler goja.Value) {
		handlerFunc, ok := goja.AssertFunction(handler)
		if !ok {
			panic(tsa.engine.ToValue("handler must be a function"))
		}
		
		tsa.app.Put(path, func(ctx *runtime.Context) error {
			tsCtx := tsa.createContextObject(ctx)
			_, err := handlerFunc(nil, tsCtx)
			return err
		})
	})
	
	// Delete method
	obj.Set("delete", func(path string, handler goja.Value) {
		handlerFunc, ok := goja.AssertFunction(handler)
		if !ok {
			panic(tsa.engine.ToValue("handler must be a function"))
		}
		
		tsa.app.Delete(path, func(ctx *runtime.Context) error {
			tsCtx := tsa.createContextObject(ctx)
			_, err := handlerFunc(nil, tsCtx)
			return err
		})
	})
	
	// OnStart method
	obj.Set("onStart", func(hook goja.Value) {
		hookFunc, ok := goja.AssertFunction(hook)
		if !ok {
			panic(tsa.engine.ToValue("hook must be a function"))
		}
		
		tsa.app.OnStart(func() error {
			_, err := hookFunc(nil)
			if err != nil {
				return fmt.Errorf("startup hook error: %w", err)
			}
			return nil
		})
	})
	
	// OnStop method
	obj.Set("onStop", func(hook goja.Value) {
		hookFunc, ok := goja.AssertFunction(hook)
		if !ok {
			panic(tsa.engine.ToValue("hook must be a function"))
		}
		
		tsa.app.OnStop(func() error {
			_, err := hookFunc(nil)
			if err != nil {
				return fmt.Errorf("shutdown hook error: %w", err)
			}
			return nil
		})
	})
	
	// Start method
	obj.Set("start", func() *goja.Promise {
		promise, resolve, reject := tsa.engine.NewPromise()
		go func() {
			if err := tsa.app.Start(); err != nil {
				reject(tsa.engine.ToValue(err.Error()))
			} else {
				resolve(tsa.engine.ToValue(true))
			}
		}()
		return promise
	})
	
	// Stop method
	obj.Set("stop", func() *goja.Promise {
		promise, resolve, reject := tsa.engine.NewPromise()
		go func() {
			if err := tsa.app.Stop(); err != nil {
				reject(tsa.engine.ToValue(err.Error()))
			} else {
				resolve(tsa.engine.ToValue(true))
			}
		}()
		return promise
	})
	
	// Listen method
	obj.Set("listen", func(port int, callback goja.Value) {
		tsa.mu.Lock()
		if tsa.server == nil {
			addr := fmt.Sprintf(":%d", port)
			tsa.server = tsa.httpAPI.NewServer(addr)
			
			// Register app handler
			tsa.server.Handle("/", func(req *api.Request) (*api.Response, error) {
				// Convert API request to framework request
				fwReq := &runtime.Request{
					Method:  req.Method,
					Path:    req.URL,
					Headers: req.Headers,
					Body:    req.Body,
					Query:   req.Query,
					Params:  req.Params,
				}
				
				fwResp := &runtime.Response{
					Status:  200,
					Headers: make(map[string]string),
					Body:    []byte{},
				}
				
				fwCtx := &runtime.Context{
					Request:  fwReq,
					Response: fwResp,
					App:      tsa.app,
					Data:     make(map[string]interface{}),
				}
				
				if err := tsa.app.Handle(fwCtx); err != nil {
					return nil, err
				}
				
				return &api.Response{
					Status:  fwResp.Status,
					Headers: fwResp.Headers,
					Body:    fwResp.Body,
				}, nil
			})
		}
		tsa.mu.Unlock()
		
		tsa.server.ListenAndServe(func(err error) {
			if callback != nil {
				if callable, ok := goja.AssertFunction(callback); ok {
					if err != nil {
						_, _ = callable(nil, tsa.engine.ToValue(err.Error()))
					} else {
						_, _ = callable(nil, goja.Undefined())
					}
				}
			}
		})
	})
	
	return obj
}

// createContextObject creates a TypeScript context object from Go context
func (tsa *TypeScriptApp) createContextObject(ctx *runtime.Context) *goja.Object {
	ctxObj := tsa.engine.NewObject()
	
	// Request object
	reqObj := tsa.engine.NewObject()
	reqObj.Set("method", ctx.Request.Method)
	reqObj.Set("path", ctx.Request.Path)
	reqObj.Set("headers", tsa.engine.ToValue(ctx.Request.Headers))
	reqObj.Set("body", tsa.engine.ToValue(string(ctx.Request.Body)))
	reqObj.Set("query", tsa.engine.ToValue(ctx.Request.Query))
	reqObj.Set("params", tsa.engine.ToValue(ctx.Request.Params))
	ctxObj.Set("request", reqObj)
	
	// Response object
	respObj := tsa.engine.NewObject()
	respObj.Set("status", ctx.Response.Status)
	respObj.Set("headers", tsa.engine.ToValue(ctx.Response.Headers))
	respObj.Set("body", tsa.engine.ToValue(string(ctx.Response.Body)))
	ctxObj.Set("response", respObj)
	
	// Data object
	ctxObj.Set("data", tsa.engine.ToValue(ctx.Data))
	
	// Set method
	ctxObj.Set("set", func(key string, value goja.Value) {
		// Data map is already thread-safe through the app's mutex
		ctx.Data[key] = value.Export()
	})
	
	// Get method
	ctxObj.Set("get", func(key string) goja.Value {
		value, ok := ctx.Data[key]
		if !ok {
			return goja.Undefined()
		}
		return tsa.engine.ToValue(value)
	})
	
	return ctxObj
}

