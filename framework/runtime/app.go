package runtime

import (
	"fmt"
	"sync"
)

// App represents the runtime-aware framework application
type App struct {
	name        string
	middleware  []Middleware
	routes      map[string]Route
	lifecycle   *Lifecycle
	mu          sync.RWMutex
}

// Middleware is a middleware function
type Middleware func(ctx *Context, next Next) error

// Next is the next middleware/handler in the chain
type Next func() error

// Context represents request context
type Context struct {
	Request  *Request
	Response *Response
	App      *App
	Data     map[string]interface{}
	mu       sync.RWMutex
}

// Request represents an HTTP request
type Request struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    []byte
	Query   map[string]string
	Params  map[string]string
}

// Response represents an HTTP response
type Response struct {
	Status  int
	Headers map[string]string
	Body    []byte
}

// Route represents a route
type Route struct {
	Method  string
	Path    string
	Handler Handler
}

// Handler is a request handler
type Handler func(ctx *Context) error

// Lifecycle manages application lifecycle
type Lifecycle struct {
	onStart []func() error
	onStop  []func() error
	mu      sync.RWMutex
}

// NewApp creates a new application
func NewApp(name string) *App {
	return &App{
		name:      name,
		middleware: make([]Middleware, 0),
		routes:    make(map[string]Route),
		lifecycle: &Lifecycle{
			onStart: make([]func() error, 0),
			onStop:  make([]func() error, 0),
		},
	}
}

// Use adds middleware
func (a *App) Use(middleware Middleware) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.middleware = append(a.middleware, middleware)
}

// Get registers a GET route
func (a *App) Get(path string, handler Handler) {
	a.registerRoute("GET", path, handler)
}

// Post registers a POST route
func (a *App) Post(path string, handler Handler) {
	a.registerRoute("POST", path, handler)
}

// Put registers a PUT route
func (a *App) Put(path string, handler Handler) {
	a.registerRoute("PUT", path, handler)
}

// Delete registers a DELETE route
func (a *App) Delete(path string, handler Handler) {
	a.registerRoute("DELETE", path, handler)
}

// registerRoute registers a route
func (a *App) registerRoute(method, path string, handler Handler) {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	key := fmt.Sprintf("%s:%s", method, path)
	a.routes[key] = Route{
		Method:  method,
		Path:    path,
		Handler: handler,
	}
}

// OnStart registers a startup hook
func (a *App) OnStart(hook func() error) {
	a.lifecycle.mu.Lock()
	defer a.lifecycle.mu.Unlock()
	a.lifecycle.onStart = append(a.lifecycle.onStart, hook)
}

// OnStop registers a shutdown hook
func (a *App) OnStop(hook func() error) {
	a.lifecycle.mu.Lock()
	defer a.lifecycle.mu.Unlock()
	a.lifecycle.onStop = append(a.lifecycle.onStop, hook)
}

// Start starts the application
func (a *App) Start() error {
	a.lifecycle.mu.RLock()
	hooks := a.lifecycle.onStart
	a.lifecycle.mu.RUnlock()
	
	for _, hook := range hooks {
		if err := hook(); err != nil {
			return fmt.Errorf("startup hook failed: %w", err)
		}
	}
	
	return nil
}

// Stop stops the application
func (a *App) Stop() error {
	a.lifecycle.mu.RLock()
	hooks := a.lifecycle.onStop
	a.lifecycle.mu.RUnlock()
	
	for _, hook := range hooks {
		if err := hook(); err != nil {
			return fmt.Errorf("shutdown hook failed: %w", err)
		}
	}
	
	return nil
}

// Handle handles a request
func (a *App) Handle(ctx *Context) error {
	// Build middleware chain
	var next Next
	next = func() error {
		// Find route
		key := fmt.Sprintf("%s:%s", ctx.Request.Method, ctx.Request.Path)
		a.mu.RLock()
		route, ok := a.routes[key]
		a.mu.RUnlock()
		
		if !ok {
			ctx.Response.Status = 404
			ctx.Response.Body = []byte("Not Found")
			return nil
		}
		
		return route.Handler(ctx)
	}
	
	// Execute middleware in reverse order
	a.mu.RLock()
	middleware := make([]Middleware, len(a.middleware))
	copy(middleware, a.middleware)
	a.mu.RUnlock()
	
	for i := len(middleware) - 1; i >= 0; i-- {
		mw := middleware[i]
		prevNext := next
		next = func() error {
			return mw(ctx, prevNext)
		}
	}
	
	return next()
}

