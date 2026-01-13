package observability

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Span represents a tracing span
type Span struct {
	TraceID    string
	SpanID     string
	ParentID   string
	Name       string
	StartTime  time.Time
	EndTime    time.Time
	Duration   time.Duration
	Tags       map[string]string
	Logs       []LogEntry
}

// LogEntry represents a log entry in a span
type LogEntry struct {
	Timestamp time.Time
	Fields    map[string]interface{}
}

// Tracer represents a distributed tracer
type Tracer struct {
	spans map[string]*Span
	mu    sync.RWMutex
}

// NewTracer creates a new tracer
func NewTracer() *Tracer {
	return &Tracer{
		spans: make(map[string]*Span),
	}
}

// StartSpan starts a new span
func (t *Tracer) StartSpan(ctx context.Context, name string) (context.Context, *Span) {
	span := &Span{
		SpanID:    generateSpanID(),
		Name:      name,
		StartTime: time.Now(),
		Tags:      make(map[string]string),
		Logs:      make([]LogEntry, 0),
	}

	// Get trace ID from context or create new
	traceID := getTraceIDFromContext(ctx)
	if traceID == "" {
		traceID = generateTraceID()
	}
	span.TraceID = traceID

	// Get parent span ID from context
	parentID := getParentSpanIDFromContext(ctx)
	span.ParentID = parentID

	t.mu.Lock()
	t.spans[span.SpanID] = span
	t.mu.Unlock()

	// Add span to context
	ctx = context.WithValue(ctx, "spanID", span.SpanID)
	ctx = context.WithValue(ctx, "traceID", span.TraceID)

	return ctx, span
}

// FinishSpan finishes a span
func (t *Tracer) FinishSpan(spanID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if span, ok := t.spans[spanID]; ok {
		span.EndTime = time.Now()
		span.Duration = span.EndTime.Sub(span.StartTime)
	}
}

// AddTag adds a tag to a span
func (t *Tracer) AddTag(spanID string, key, value string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if span, ok := t.spans[spanID]; ok {
		span.Tags[key] = value
	}
}

// AddLog adds a log entry to a span
func (t *Tracer) AddLog(spanID string, fields map[string]interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if span, ok := t.spans[spanID]; ok {
		span.Logs = append(span.Logs, LogEntry{
			Timestamp: time.Now(),
			Fields:    fields,
		})
	}
}

// GetSpan gets a span by ID
func (t *Tracer) GetSpan(spanID string) (*Span, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	span, ok := t.spans[spanID]
	return span, ok
}

// GetSpansByTraceID gets all spans for a trace
func (t *Tracer) GetSpansByTraceID(traceID string) []*Span {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var spans []*Span
	for _, span := range t.spans {
		if span.TraceID == traceID {
			spans = append(spans, span)
		}
	}
	return spans
}

var (
	spanIDCounter  uint64
	spanIDMu       sync.Mutex
	traceIDCounter uint64
	traceIDMu      sync.Mutex
)

func generateSpanID() string {
	spanIDMu.Lock()
	defer spanIDMu.Unlock()
	spanIDCounter++
	return fmt.Sprintf("span-%d", spanIDCounter)
}

func generateTraceID() string {
	traceIDMu.Lock()
	defer traceIDMu.Unlock()
	traceIDCounter++
	return fmt.Sprintf("trace-%d", traceIDCounter)
}

func getTraceIDFromContext(ctx context.Context) string {
	if traceID, ok := ctx.Value("traceID").(string); ok {
		return traceID
	}
	return ""
}

func getParentSpanIDFromContext(ctx context.Context) string {
	if spanID, ok := ctx.Value("spanID").(string); ok {
		return spanID
	}
	return ""
}

