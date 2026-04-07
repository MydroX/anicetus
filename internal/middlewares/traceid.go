package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const TraceIDKey = "X-Trace-ID"

// TraceID ensures every request has a trace ID.
// If the client sends one in the header, it's reused. Otherwise a new one is generated.
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(TraceIDKey)
		if traceID == "" {
			traceID = uuid.New().String()
		}

		c.Set(TraceIDKey, traceID)
		c.Header(TraceIDKey, traceID)
		c.Next()
	}
}

// GetTraceID retrieves the trace ID from a gin context.
func GetTraceID(c *gin.Context) string {
	if traceID, exists := c.Get(TraceIDKey); exists {
		if id, ok := traceID.(string); ok {
			return id
		}
	}

	return ""
}
