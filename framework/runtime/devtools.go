package runtime

import (
	"fmt"
	"strings"
	"time"
)

// DevServerConfig configures the development server
type DevServerConfig struct {
	Host           string
	Port           int
	HotReload      bool
	CORS           bool
	VerboseLogging bool
	RequestTimeout time.Duration
	BodySizeLimit  int64
	AllowedOrigins []string
	AutoAPI        bool // Auto-generate API docs
	MockData       bool // Enable mock data endpoints
}

// DevTools provides development utilities
type DevTools struct {
	config *DevServerConfig
	app    *App
}

// NewDevTools creates development tools
func NewDevTools(config *DevServerConfig) *DevTools {
	return &DevTools{
		config: config,
	}
}

// VerboseLoggerMiddleware provides detailed request/response logging
func VerboseLoggerMiddleware(ctx *Context, next Next) error {
	start := time.Now()

	fmt.Printf("\n[REQUEST] %s %s\n", ctx.Request.Method, ctx.Request.Path)

	if len(ctx.Request.Headers) > 0 {
		fmt.Println("Headers:")
		for k, v := range ctx.Request.Headers {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}

	if len(ctx.Request.Query) > 0 {
		fmt.Println("Query:")
		for k, v := range ctx.Request.Query {
			fmt.Printf("  %s=%s\n", k, v)
		}
	}

	if len(ctx.Request.Body) > 0 {
		fmt.Printf("Body: %s\n", string(ctx.Request.Body))
	}

	err := next()

	duration := time.Since(start)

	fmt.Printf("\n[RESPONSE] Status: %d (%vms)\n", ctx.Response.Status, duration.Milliseconds())
	if len(ctx.Response.Headers) > 0 {
		fmt.Println("Headers:")
		for k, v := range ctx.Response.Headers {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}

	if len(ctx.Response.Body) > 0 && len(ctx.Response.Body) < 500 {
		fmt.Printf("Body: %s\n", string(ctx.Response.Body))
	}

	return err
}

// RequestIDMiddlewareWithHeaders adds unique request ID
func RequestIDMiddlewareWithHeaders(ctx *Context, next Next) error {
	if ctx.Request.Headers == nil {
		ctx.Request.Headers = make(map[string]string)
	}

	// Generate request ID if not present
	if _, ok := ctx.Request.Headers["X-Request-ID"]; !ok {
		requestID := fmt.Sprintf("%d", time.Now().UnixNano())
		ctx.Request.Headers["X-Request-ID"] = requestID
		if ctx.Response.Headers == nil {
			ctx.Response.Headers = make(map[string]string)
		}
		ctx.Response.Headers["X-Request-ID"] = requestID
	}

	return next()
}

// StackTraceMiddleware logs stack traces on panic
func StackTraceMiddleware(ctx *Context, next Next) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[PANIC] %v\n", r)
			ctx.Response.Status = 500
		}
	}()

	return next()
}

// MockDataMiddleware provides mock data for development
func MockDataMiddleware(mockEndpoints map[string]interface{}) Middleware {
	return func(ctx *Context, next Next) error {
		// Check if this is a mock endpoint
		path := ctx.Request.Path

		if mockData, ok := mockEndpoints[path]; ok {
			ctx.Response.Status = 200
			if ctx.Response.Headers == nil {
				ctx.Response.Headers = make(map[string]string)
			}
			ctx.Response.Headers["Content-Type"] = "application/json"
			ctx.Response.Headers["X-Mock"] = "true"

			// In production, this would JSON-encode the mock data
			ctx.Response.Body = []byte(fmt.Sprintf("%v", mockData))
			return nil
		}

		return next()
	}
}

// APIDocMiddleware generates API documentation
func APIDocMiddleware(app *App) Middleware {
	return func(ctx *Context, next Next) error {
		if ctx.Request.Path == "/api/docs" {
			ctx.Response.Status = 200
			if ctx.Response.Headers == nil {
				ctx.Response.Headers = make(map[string]string)
			}
			ctx.Response.Headers["Content-Type"] = "text/html"

			docs := generateAPIDocs(app)
			ctx.Response.Body = []byte(docs)
			return nil
		}

		return next()
	}
}

func generateAPIDocs(app *App) string {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>GoTS API Documentation</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .route { margin: 20px 0; padding: 10px; background: #f5f5f5; border-left: 4px solid #007bff; }
        .method { display: inline-block; padding: 5px 10px; margin-right: 10px; }
        .get { background: #28a745; color: white; }
        .post { background: #007bff; color: white; }
        .put { background: #ffc107; color: black; }
        .delete { background: #dc3545; color: white; }
    </style>
</head>
<body>
    <h1>API Documentation</h1>
    <div id="routes">`

	app.mu.RLock()
	defer app.mu.RUnlock()

	for _, route := range app.routes {
		methodClass := strings.ToLower(route.Method)
		html += fmt.Sprintf(`
    <div class="route">
        <span class="method %s">%s</span> %s
    </div>`, methodClass, route.Method, route.Path)
	}

	html += `
    </div>
</body>
</html>`

	return html
}

// HealthCheckMiddleware provides health check endpoint
func HealthCheckMiddleware(app *App) Middleware {
	return func(ctx *Context, next Next) error {
		if ctx.Request.Path == "/health" {
			ctx.Response.Status = 200
			if ctx.Response.Headers == nil {
				ctx.Response.Headers = make(map[string]string)
			}
			ctx.Response.Headers["Content-Type"] = "application/json"
			ctx.Response.Body = []byte(`{"status": "healthy"}`)
			return nil
		}

		if ctx.Request.Path == "/ready" {
			ctx.Response.Status = 200
			if ctx.Response.Headers == nil {
				ctx.Response.Headers = make(map[string]string)
			}
			ctx.Response.Headers["Content-Type"] = "application/json"
			ctx.Response.Body = []byte(`{"ready": true}`)
			return nil
		}

		return next()
	}
}

// MetricsMiddleware collects request metrics
type MetricsData struct {
	TotalRequests  int64
	TotalErrors    int64
	AverageLatency int64
	RequestCounts  map[string]int64
	ErrorCounts    map[string]int64
}

var globalMetrics = &MetricsData{
	RequestCounts: make(map[string]int64),
	ErrorCounts:   make(map[string]int64),
}

// MetricsMiddleware collects metrics
func MetricsMiddleware(ctx *Context, next Next) error {
	start := time.Now()

	globalMetrics.TotalRequests++
	key := fmt.Sprintf("%s %s", ctx.Request.Method, ctx.Request.Path)
	globalMetrics.RequestCounts[key]++

	err := next()

	duration := time.Since(start)
	globalMetrics.AverageLatency = int64(duration / time.Millisecond)

	if err != nil {
		globalMetrics.TotalErrors++
		globalMetrics.ErrorCounts[key]++
	}

	return err
}

// GetMetrics returns current metrics
func GetMetrics() *MetricsData {
	return globalMetrics
}
