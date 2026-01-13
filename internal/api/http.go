package api

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"gots-runtime/internal/eventloop"
)

// HTTP provides HTTP server functionality
type HTTP struct {
	eventLoop *eventloop.Loop
	server    *http.Server
}

// NewHTTP creates a new HTTP API
func NewHTTP(eventLoop *eventloop.Loop) *HTTP {
	return &HTTP{
		eventLoop: eventLoop,
	}
}

// Request represents an HTTP request
type Request struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    []byte
	Params  map[string]string
	Query   map[string]string
}

// Response represents an HTTP response
type Response struct {
	Status  int
	Headers map[string]string
	Body    []byte
}

// Handler is a function that handles HTTP requests
type Handler func(*Request) (*Response, error)

// Middleware is a function that processes requests/responses
type Middleware func(Handler) Handler

// Server represents an HTTP server
type Server struct {
	http      *HTTP
	mux       *http.ServeMux
	handlers  map[string]Handler
	middleware []Middleware
}

// NewServer creates a new HTTP server
func (h *HTTP) NewServer(addr string) *Server {
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	s := &Server{
		http:     h,
		mux:      mux,
		handlers: make(map[string]Handler),
		middleware: make([]Middleware, 0),
	}

	// Set the server
	h.server = server

	return s
}

// Handle registers a handler for a path
func (s *Server) Handle(path string, handler Handler) {
	s.handlers[path] = handler
	
	// Wrap handler with middleware
	wrappedHandler := handler
	for i := len(s.middleware) - 1; i >= 0; i-- {
		wrappedHandler = s.middleware[i](wrappedHandler)
	}

	s.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		// Convert http.Request to our Request type
		req := s.convertRequest(r)
		
		// Execute handler in event loop
		s.http.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
			resp, err := wrappedHandler(req)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return nil
			}

			// Write response
			for k, v := range resp.Headers {
				w.Header().Set(k, v)
			}
			w.WriteHeader(resp.Status)
			_, _ = w.Write(resp.Body)
			return nil
		}, 0))
	})
}

// Use adds middleware
func (s *Server) Use(middleware Middleware) {
	s.middleware = append(s.middleware, middleware)
}

// ListenAndServe starts the HTTP server
func (s *Server) ListenAndServe(callback func(error)) {
	s.http.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		err := s.http.server.ListenAndServe()
		callback(err)
		return nil
	}, 0))
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context, callback func(error)) {
	s.http.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		err := s.http.server.Shutdown(ctx)
		callback(err)
		return nil
	}, 0))
}

// convertRequest converts http.Request to our Request type
func (s *Server) convertRequest(r *http.Request) *Request {
	// Read body
	body, _ := io.ReadAll(r.Body)
	
	// Parse query parameters
	query := make(map[string]string)
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			query[k] = v[0]
		}
	}

	// Parse headers
	headers := make(map[string]string)
	for k, v := range r.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	return &Request{
		Method:  r.Method,
		URL:     r.URL.Path,
		Headers: headers,
		Body:    body,
		Query:   query,
		Params:  make(map[string]string), // Would be populated by router
	}
}

// Client represents an HTTP client
type Client struct {
	http      *HTTP
	client    *http.Client
	timeout   time.Duration
}

// NewClient creates a new HTTP client
func (h *HTTP) NewClient(timeout time.Duration) *Client {
	return &Client{
		http:    h,
		client:  &http.Client{Timeout: timeout},
		timeout: timeout,
	}
}

// Get performs a GET request
func (c *Client) Get(url string, callback func(*Response, error)) {
	c.http.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		resp, err := c.client.Get(url)
		if err != nil {
			callback(nil, err)
			return nil
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			callback(nil, err)
			return nil
		}

		headers := make(map[string]string)
		for k, v := range resp.Header {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}

		response := &Response{
			Status:  resp.StatusCode,
			Headers: headers,
			Body:    body,
		}

		callback(response, nil)
		return nil
	}, 0))
}

// Post performs a POST request
func (c *Client) Post(url, contentType string, body []byte, callback func(*Response, error)) {
	c.http.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		resp, err := c.client.Post(url, contentType, bytes.NewReader(body))
		if err != nil {
			callback(nil, err)
			return nil
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			callback(nil, err)
			return nil
		}

		headers := make(map[string]string)
		for k, v := range resp.Header {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}

		response := &Response{
			Status:  resp.StatusCode,
			Headers: headers,
			Body:    respBody,
		}

		callback(response, nil)
		return nil
	}, 0))
}

