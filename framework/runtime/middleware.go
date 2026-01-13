package runtime

import (
	"fmt"
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
	ctx.Response.Headers["Access-Control-Allow-Methods"] = "GET, POST, PUT, DELETE, OPTIONS"
	ctx.Response.Headers["Access-Control-Allow-Headers"] = "Content-Type, Authorization"
	
	if ctx.Request.Method == "OPTIONS" {
		ctx.Response.Status = 200
		return nil
	}
	
	return next()
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

