package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestTraceID_GeneratesNew(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	handler := TraceID()
	handler(c)

	traceID := GetTraceID(c)
	assert.NotEmpty(t, traceID)
	assert.Equal(t, traceID, w.Header().Get(TraceIDKey))
}

func TestTraceID_ReusesExisting(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set(TraceIDKey, "existing-trace-id")

	handler := TraceID()
	handler(c)

	assert.Equal(t, "existing-trace-id", GetTraceID(c))
	assert.Equal(t, "existing-trace-id", w.Header().Get(TraceIDKey))
}

func TestGetTraceID_Exists(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(TraceIDKey, "my-trace")

	assert.Equal(t, "my-trace", GetTraceID(c))
}

func TestGetTraceID_NotSet(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	assert.Empty(t, GetTraceID(c))
}

func TestGetTraceID_WrongType(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(TraceIDKey, 12345)

	assert.Empty(t, GetTraceID(c))
}
