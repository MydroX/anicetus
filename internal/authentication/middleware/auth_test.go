package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"MydroX/anicetus/pkg/jwt"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testConfig = jwt.TokenConfig{
	AccessTokenSecret:    "test-secret-key-long-enough-for-signing-jwt-tokens-securely",
	RefreshTokenSecret:   "test-refresh-secret-long-enough-for-signing-jwt-tok",
	ExpectedIssuer:       "test-issuer",
	ClockSkewSeconds:     60,
	AccessTokenDuration:  3600,
	RefreshTokenDuration: 86400,
}

func createTestAccessToken(userUUID string, permissions, audiences []string) string {
	service := jwt.NewJWTService(testConfig)
	token, _ := service.CreateAccessToken(userUUID, permissions, audiences)
	return token
}

func setupRouter(jwtService *jwt.Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthMiddleware(jwtService))
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"user_uuid":   GetUserUUID(c),
			"permissions": GetPermissions(c),
			"audiences":   GetAudiences(c),
		})
	})
	return r
}

func TestAuthMiddleware_ValidBearerToken(t *testing.T) {
	service := jwt.NewJWTService(testConfig)
	router := setupRouter(service)

	token := createTestAccessToken("user-123", []string{"read", "write"}, []string{"test-issuer"})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "user-123", resp["user_uuid"])
}

func TestAuthMiddleware_ValidCookieToken(t *testing.T) {
	service := jwt.NewJWTService(testConfig)
	router := setupRouter(service)

	token := createTestAccessToken("user-456", nil, []string{"test-issuer"})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	service := jwt.NewJWTService(testConfig)
	router := setupRouter(service)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "missing authorization token", resp["message"])
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	service := jwt.NewJWTService(testConfig)
	router := setupRouter(service)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-string")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "invalid or expired token", resp["message"])
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	service := jwt.NewJWTService(testConfig)
	router := setupRouter(service)

	// Create an expired token manually
	claims := jwt.AccessTokenClaims{
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(-10 * time.Minute)),
			IssuedAt:  jwtlib.NewNumericDate(time.Now().Add(-20 * time.Minute)),
			Issuer:    "test-issuer",
		},
		UserUUID:  "user-123",
		TokenType: string(jwt.AccessToken),
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS512, claims)
	tokenString, _ := token.SignedString([]byte(testConfig.AccessTokenSecret))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_BearerCaseInsensitive(t *testing.T) {
	service := jwt.NewJWTService(testConfig)
	router := setupRouter(service)

	token := createTestAccessToken("user-789", nil, []string{"test-issuer"})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
