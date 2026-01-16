package runtime

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

// App represents the runtime-aware framework application
type App struct {
	name            string
	middleware      []Middleware
	routes          map[string]Route
	dynamicRoutes   []*DynamicRoute
	lifecycle       *Lifecycle
	errorHandler    ErrorHandler
	notFoundHandler NotFoundHandler
	panicHandler    PanicHandler
	mu              sync.RWMutex
}

// DynamicRoute represents a route with dynamic parameters
type DynamicRoute struct {
	Method  string
	Pattern *regexp.Regexp
	Path    string
	Handler Handler
}

// ErrorHandler handles errors during request processing
type ErrorHandler func(ctx *Context, err error) error

// NotFoundHandler handles 404 not found errors
type NotFoundHandler func(ctx *Context) error

// PanicHandler handles panics in handlers
type PanicHandler func(ctx *Context, r interface{}) error

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
		name:          name,
		middleware:    make([]Middleware, 0),
		routes:        make(map[string]Route),
		dynamicRoutes: make([]*DynamicRoute, 0),
		lifecycle: &Lifecycle{
			onStart: make([]func() error, 0),
			onStop:  make([]func() error, 0),
		},
		errorHandler:    DefaultErrorHandler,
		notFoundHandler: DefaultNotFoundHandler,
		panicHandler:    DefaultPanicHandler,
	}
}

// DefaultErrorHandler provides default error handling
func DefaultErrorHandler(ctx *Context, err error) error {
	ctx.Response.Status = 500
	ctx.Response.Body = []byte(fmt.Sprintf("Internal Server Error: %v", err))
	return err
}

// DefaultNotFoundHandler provides default 404 handling
func DefaultNotFoundHandler(ctx *Context) error {
	ctx.Response.Status = 404
	ctx.Response.Body = []byte("Not Found")
	return nil
}

// DefaultPanicHandler provides default panic handling
func DefaultPanicHandler(ctx *Context, r interface{}) error {
	ctx.Response.Status = 500
	ctx.Response.Body = []byte(fmt.Sprintf("Internal Server Error: %v", r))
	return fmt.Errorf("panic: %v", r)
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

// Patch registers a PATCH route
func (a *App) Patch(path string, handler Handler) {
	a.registerRoute("PATCH", path, handler)
}

// Options registers an OPTIONS route
func (a *App) Options(path string, handler Handler) {
	a.registerRoute("OPTIONS", path, handler)
}

// Head registers a HEAD route
func (a *App) Head(path string, handler Handler) {
	a.registerRoute("HEAD", path, handler)
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

// Dynamic registers a dynamic route with parameters (e.g., /users/:id/posts/:postid)
func (a *App) Dynamic(method, path string, handler Handler) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Convert path to regex pattern
	pattern := convertPathToPattern(path)
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return
	}

	a.dynamicRoutes = append(a.dynamicRoutes, &DynamicRoute{
		Method:  method,
		Pattern: regex,
		Path:    path,
		Handler: handler,
	})
}

// convertPathToPattern converts a path like /users/:id to a regex pattern
func convertPathToPattern(path string) string {
	pattern := regexp.QuoteMeta(path)
	pattern = strings.ReplaceAll(pattern, "\\:", ":")
	pattern = regexp.MustCompile(`:([a-zA-Z_][a-zA-Z0-9_]*)`).ReplaceAllString(pattern, `(?P<$1>[^/]+)`)
	return "^" + pattern + "$"
}

// SetErrorHandler sets the error handler
func (a *App) SetErrorHandler(handler ErrorHandler) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.errorHandler = handler
}

// SetNotFoundHandler sets the not found handler
func (a *App) SetNotFoundHandler(handler NotFoundHandler) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.notFoundHandler = handler
}

// SetPanicHandler sets the panic handler
func (a *App) SetPanicHandler(handler PanicHandler) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.panicHandler = handler
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
	// Defer panic recovery
	defer func() {
		if r := recover(); r != nil {
			a.mu.RLock()
			panicHandler := a.panicHandler
			a.mu.RUnlock()
			_ = panicHandler(ctx, r)
		}
	}()

	// Build middleware chain
	var next Next
	next = func() error {
		// Find route
		key := fmt.Sprintf("%s:%s", ctx.Request.Method, ctx.Request.Path)
		a.mu.RLock()
		route, ok := a.routes[key]
		a.mu.RUnlock()

		if ok {
			return route.Handler(ctx)
		}

		// Try dynamic routes
		a.mu.RLock()
		for _, dynRoute := range a.dynamicRoutes {
			if dynRoute.Method == ctx.Request.Method && dynRoute.Pattern.MatchString(ctx.Request.Path) {
				// Extract path parameters
				matches := extractNamedMatches(dynRoute.Pattern, ctx.Request.Path)
				if ctx.Request.Params == nil {
					ctx.Request.Params = make(map[string]string)
				}
				for key, val := range matches {
					ctx.Request.Params[key] = val
				}
				a.mu.RUnlock()
				return dynRoute.Handler(ctx)
			}
		}
		a.mu.RUnlock()

		// Not found
		a.mu.RLock()
		notFoundHandler := a.notFoundHandler
		a.mu.RUnlock()
		return notFoundHandler(ctx)
	}

	// Execute middleware in order
	a.mu.RLock()
	middleware := make([]Middleware, len(a.middleware))
	copy(middleware, a.middleware)
	errorHandler := a.errorHandler
	a.mu.RUnlock()

	for i := len(middleware) - 1; i >= 0; i-- {
		mw := middleware[i]
		prevNext := next
		next = func() error {
			return mw(ctx, prevNext)
		}
	}

	// Execute the middleware chain
	err := next()

	// Handle errors
	if err != nil {
		return errorHandler(ctx, err)
	}

	return nil
}

// extractNamedMatches extracts named groups from a regex match
func extractNamedMatches(re *regexp.Regexp, s string) map[string]string {
	captures := make(map[string]string)
	match := re.FindStringSubmatch(s)
	if match == nil {
		return captures
	}
	for i, name := range re.SubexpNames() {
		if i != 0 && name != "" {
			captures[name] = match[i]
		}
	}
	return captures
}
