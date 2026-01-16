package runtime

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// LoggerMiddleware provides logging middleware
func LoggerMiddleware(ctx *Context, next Next) error {
	start := time.Now()

	err := next()

	duration := time.Since(start)
	fmt.Printf("[%s] %s %s - %d - %v\n",
		time.Now().Format(time.RFC3339),
		ctx.Request.Method,
		ctx.Request.Path,
		ctx.Response.Status,
		duration)

	return err
}

// CORSMiddleware provides CORS middleware
func CORSMiddleware(ctx *Context, next Next) error {
	if ctx.Response.Headers == nil {
		ctx.Response.Headers = make(map[string]string)
	}

	ctx.Response.Headers["Access-Control-Allow-Origin"] = "*"
	ctx.Response.Headers["Access-Control-Allow-Methods"] = "GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD"
	ctx.Response.Headers["Access-Control-Allow-Headers"] = "Content-Type, Authorization"

	if ctx.Request.Method == "OPTIONS" {
		ctx.Response.Status = 200
		return nil
	}

	return next()
}

// CORSMiddlewareWithOrigins provides CORS middleware with custom origins
func CORSMiddlewareWithOrigins(allowedOrigins []string) Middleware {
	return func(ctx *Context, next Next) error {
		if ctx.Response.Headers == nil {
			ctx.Response.Headers = make(map[string]string)
		}

		origin := ctx.Request.Headers["Origin"]
		allowed := false
		for _, o := range allowedOrigins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}

		if allowed {
			ctx.Response.Headers["Access-Control-Allow-Origin"] = origin
			ctx.Response.Headers["Access-Control-Allow-Methods"] = "GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD"
			ctx.Response.Headers["Access-Control-Allow-Headers"] = "Content-Type, Authorization"
		}

		if ctx.Request.Method == "OPTIONS" {
			ctx.Response.Status = 200
			return nil
		}

		return next()
	}
}

// RecoveryMiddleware provides panic recovery middleware
func RecoveryMiddleware(ctx *Context, next Next) error {
	defer func() {
		if r := recover(); r != nil {
			ctx.Response.Status = 500
			ctx.Response.Body = []byte(fmt.Sprintf("Internal Server Error: %v", r))
		}
	}()

	return next()
}

// TimeoutMiddleware provides timeout middleware
func TimeoutMiddleware(timeout time.Duration) Middleware {
	return func(ctx *Context, next Next) error {
		done := make(chan error, 1)

		go func() {
			done <- next()
		}()

		select {
		case err := <-done:
			return err
		case <-time.After(timeout):
			ctx.Response.Status = 504
			ctx.Response.Body = []byte("Request Timeout")
			return fmt.Errorf("request timeout")
		}
	}
}

// AuthMiddleware provides authentication middleware
func AuthMiddleware(validateToken func(string) bool) Middleware {
	return func(ctx *Context, next Next) error {
		token := ctx.Request.Headers["Authorization"]
		if token == "" {
			ctx.Response.Status = 401
			ctx.Response.Body = []byte("Unauthorized")
			return fmt.Errorf("missing authorization token")
		}

		if !validateToken(token) {
			ctx.Response.Status = 401
			ctx.Response.Body = []byte("Unauthorized")
			return fmt.Errorf("invalid authorization token")
		}

		return next()
	}
}

// BearerTokenMiddleware extracts and validates bearer tokens
func BearerTokenMiddleware(validateToken func(string) bool) Middleware {
	return func(ctx *Context, next Next) error {
		authHeader := ctx.Request.Headers["Authorization"]
		if authHeader == "" {
			ctx.Response.Status = 401
			ctx.Response.Body = []byte("Unauthorized: missing Authorization header")
			return fmt.Errorf("missing authorization header")
		}

		// Extract bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			ctx.Response.Status = 401
			ctx.Response.Body = []byte("Unauthorized: invalid Authorization header format")
			return fmt.Errorf("invalid authorization header format")
		}

		token := parts[1]
		if !validateToken(token) {
			ctx.Response.Status = 401
			ctx.Response.Body = []byte("Unauthorized: invalid token")
			return fmt.Errorf("invalid token")
		}

		// Store token in context
		if ctx.Data == nil {
			ctx.Data = make(map[string]interface{})
		}
		ctx.Data["token"] = token

		return next()
	}
}

// ContentTypeMiddleware enforces content type on requests
func ContentTypeMiddleware(requiredType string) Middleware {
	return func(ctx *Context, next Next) error {
		if ctx.Request.Method == "GET" || ctx.Request.Method == "DELETE" || ctx.Request.Method == "HEAD" {
			return next()
		}

		contentType := ctx.Request.Headers["Content-Type"]
		if !strings.Contains(contentType, requiredType) {
			ctx.Response.Status = 415
			ctx.Response.Body = []byte(fmt.Sprintf("Unsupported Media Type: expected %s", requiredType))
			return fmt.Errorf("unsupported media type")
		}

		return next()
	}
}

// RequestIDMiddleware adds a request ID to the context
func RequestIDMiddleware(idGenerator func() string) Middleware {
	return func(ctx *Context, next Next) error {
		if ctx.Data == nil {
			ctx.Data = make(map[string]interface{})
		}

		requestID := ctx.Request.Headers["X-Request-ID"]
		if requestID == "" {
			requestID = idGenerator()
		}

		ctx.Data["requestId"] = requestID

		if ctx.Response.Headers == nil {
			ctx.Response.Headers = make(map[string]string)
		}
		ctx.Response.Headers["X-Request-ID"] = requestID

		return next()
	}
}

// RateLimitMiddleware provides basic rate limiting
func RateLimitMiddleware(maxRequests int, windowSize time.Duration) Middleware {
	var mu sync.Mutex
	requests := make([]time.Time, 0)

	return func(ctx *Context, next Next) error {
		mu.Lock()
		now := time.Now()

		// Clean old requests
		validRequests := make([]time.Time, 0)
		for _, req := range requests {
			if now.Sub(req) < windowSize {
				validRequests = append(validRequests, req)
			}
		}

		if len(validRequests) >= maxRequests {
			mu.Unlock()
			ctx.Response.Status = 429
			ctx.Response.Body = []byte("Too Many Requests")
			return fmt.Errorf("rate limit exceeded")
		}

		requests = append(validRequests, now)
		mu.Unlock()

		return next()
	}
}

// ContextDataMiddleware adds data to the request context
func ContextDataMiddleware(data map[string]interface{}) Middleware {
	return func(ctx *Context, next Next) error {
		if ctx.Data == nil {
			ctx.Data = make(map[string]interface{})
		}

		for key, val := range data {
			ctx.Data[key] = val
		}

		return next()
	}
}

// ResponseHeaderMiddleware adds custom headers to responses
func ResponseHeaderMiddleware(headers map[string]string) Middleware {
	return func(ctx *Context, next Next) error {
		if ctx.Response.Headers == nil {
			ctx.Response.Headers = make(map[string]string)
		}

		for key, val := range headers {
			ctx.Response.Headers[key] = val
		}

		return next()
	}
}
