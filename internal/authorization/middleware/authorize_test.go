package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	authmiddleware "MydroX/anicetus/internal/authentication/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}


// --- RequirePermission ---

func TestRequirePermission_HasAll(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		authmiddleware.SetPermissions(c, []string{"read", "write", "delete"})
		c.Next()
	})
	r.GET("/test", RequirePermission("read", "write"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequirePermission_MissingOne(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		authmiddleware.SetPermissions(c, []string{"read"})
		c.Next()
	})
	r.GET("/test", RequirePermission("read", "write"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "insufficient permissions", resp["message"])
}

func TestRequirePermission_NoPermissions(t *testing.T) {
	r := gin.New()
	r.GET("/test", RequirePermission("read"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// --- RequireAnyPermission ---

func TestRequireAnyPermission_HasOne(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		authmiddleware.SetPermissions(c, []string{"delete"})
		c.Next()
	})
	r.GET("/test", RequireAnyPermission("read", "delete"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireAnyPermission_HasNone(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		authmiddleware.SetPermissions(c, []string{"other"})
		c.Next()
	})
	r.GET("/test", RequireAnyPermission("read", "write"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireAnyPermission_NoPermissions(t *testing.T) {
	r := gin.New()
	r.GET("/test", RequireAnyPermission("read"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// --- RequireRole (stub) ---

func TestRequireRole_Stub(t *testing.T) {
	r := gin.New()
	r.GET("/test", RequireRole("admin"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Equal(t, http.StatusOK, w.Code)
}
