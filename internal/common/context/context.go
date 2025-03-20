package context

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ContextKey string

// TraceIDKey is the key used to store the trace ID in contexts
const (
	TraceIDKey   ContextKey = "X-Trace-ID"
	ErrorCodeKey ContextKey = "X-Error-Code"
)

// AppContext is a custom context that wraps both gin.Context and context.Context
type AppContext struct {
	gc  *gin.Context
	ctx context.Context
}

// NewAppContext creates a new AppContext
func NewAppContext(gc *gin.Context) *AppContext {
	ac := &AppContext{
		gc:  gc,
		ctx: gc.Request.Context(),
	}

	// Ensure trace ID exists when context is created
	ac.EnsureTraceID()

	return ac
}

func NewAppContextTest() *AppContext {
	return &AppContext{
		gc:  &gin.Context{},
		ctx: context.Background(),
	}
}

// EnsureTraceID makes sure a trace ID exists in the context
// If one doesn't exist, it creates a new one and adds it to both contexts
func (a *AppContext) EnsureTraceID() string {
	// Check if trace ID already exists in either context
	if traceID, exists := a.Get(string(TraceIDKey)); exists {
		if strTraceID, ok := traceID.(string); ok {
			return strTraceID
		}
		return ""
	}

	// Check if it came from an HTTP header
	if traceID := a.gc.GetHeader(string(TraceIDKey)); traceID != "" {
		a.Set(TraceIDKey, traceID)
		return traceID
	}

	// Generate a new trace ID
	traceID := uuid.New().String()
	a.Set(TraceIDKey, traceID)
	return traceID
}

// GetTraceID gets the current trace ID or creates one if it doesn't exist
func (a *AppContext) GetTraceID() string {
	return a.EnsureTraceID()
}

// Set stores a value in both contexts
func (a *AppContext) Set(key ContextKey, value any) {
	a.gc.Set(string(key), value)
	a.ctx = context.WithValue(a.ctx, key, value)
}

// Get retrieves a value, checking both contexts
func (a *AppContext) Get(key string) (any, bool) {
	// Try gin.Context first
	if value, exists := a.gc.Get(key); exists {
		return value, true
	}

	// Then try standard context
	if value := a.ctx.Value(key); value != nil {
		return value, true
	}

	return nil, false
}

func (a *AppContext) Value(key any) any {
	// Check string keys in gin.Context
	if keyStr, ok := key.(string); ok {
		if val, exists := a.gc.Get(keyStr); exists {
			return val
		}
	}

	// Fall back to standard context
	return a.ctx.Value(key)
}

// GinContext returns the underlying gin.Context
func (a *AppContext) GinContext() *gin.Context {
	return a.gc
}

// StdContext returns the underlying context.Context
func (a *AppContext) StdContext() context.Context {
	return a.ctx
}
