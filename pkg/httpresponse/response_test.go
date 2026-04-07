package httpresponse

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"MydroX/anicetus/internal/middlewares"
	"MydroX/anicetus/pkg/errs"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func testContext(traceID string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middlewares.TraceIDKey, traceID)

	return c, w
}

func testLogger() *zap.SugaredLogger {
	l, _ := zap.NewDevelopment()
	return l.Sugar()
}

func TestError_AppError(t *testing.T) {
	c, w := testContext("trace-123")

	appErr := &errs.AppError{Code: errs.ErrorNotFound, Message: "user not found", Err: errors.New("db")}
	Error(testLogger(), c, appErr)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp errorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "user not found", resp.Message)
	assert.Equal(t, errs.ErrorNotFound, resp.Code)
	assert.Equal(t, "trace-123", resp.TraceID)
}

func TestError_NonAppError(t *testing.T) {
	c, w := testContext("trace-456")

	Error(testLogger(), c, errors.New("unexpected"))

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp errorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "internal server error", resp.Message)
	assert.Equal(t, errs.ErrorUnknownError, resp.Code)
	assert.Equal(t, "trace-456", resp.TraceID)
}

func TestError_ZeroCodeAppError(t *testing.T) {
	c, w := testContext("trace-789")

	appErr := &errs.AppError{Code: 0, Message: "unknown", Err: errors.New("oops")}
	Error(testLogger(), c, appErr)

	var resp errorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, errs.ErrorUnknownError, resp.Code)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestBadRequest(t *testing.T) {
	c, w := testContext("trace-bad")

	BadRequest(testLogger(), c, errs.ErrorInvalidInput, "invalid email")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp errorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "invalid email", resp.Message)
	assert.Equal(t, errs.ErrorInvalidInput, resp.Code)
	assert.Equal(t, "trace-bad", resp.TraceID)
}

func TestError_NoTraceID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Error(testLogger(), c, errors.New("no trace"))

	var resp errorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Empty(t, resp.TraceID)
}
