package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"MydroX/anicetus/internal/authentication/dto"
	"MydroX/anicetus/internal/authentication/mocks"
	"MydroX/anicetus/internal/config"
	"MydroX/anicetus/pkg/errs"
	"MydroX/anicetus/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type testServer struct {
	router      *gin.Engine
	mockUsecase *mocks.MockAuthenticationUsecases
}

func newTestServer(t *testing.T) testServer {
	ctrl := gomock.NewController(t)
	mockUC := mocks.NewMockAuthenticationUsecases(ctrl)
	log, _ := logger.New("TEST")

	cfg := &config.Config{
		App: config.App{Domain: "localhost"},
		JWT: config.JWT{
			AccessToken:  config.AccessToken{Expiration: 3600},
			RefreshToken: config.RefreshToken{Expiration: 86400},
		},
	}

	c := New(log, mockUC, cfg)

	router := gin.New()
	v1 := router.Group("/api/v1")
	PublicRouter(v1, c)
	AuthenticatedRouter(v1, c)

	return testServer{router: router, mockUsecase: mockUC}
}

func validLoginRequest() dto.LoginRequest {
	return dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123!",
		Session: dto.Session{
			IPv4Address:    "192.168.1.1",
			OS:             "macOS",
			OSVersion:      "14.0",
			Browser:        "Chrome",
			BrowserVersion: "120.0",
		},
	}
}

// --- Login ---

func TestLogin_Success(t *testing.T) {
	s := newTestServer(t)
	req := validLoginRequest()

	s.mockUsecase.EXPECT().
		Login(gomock.Any(), gomock.Any()).
		Return("access-token", "refresh-token", nil)

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	s.router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.LoginResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "access-token", resp.AccessToken)
	assert.Equal(t, "refresh-token", resp.RefreshToken)
}

func TestLogin_BindFailure(t *testing.T) {
	s := newTestServer(t)

	w := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewBufferString("invalid"))
	httpReq.Header.Set("Content-Type", "application/json")

	s.router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogin_ValidationFailure(t *testing.T) {
	s := newTestServer(t)

	// Missing password
	req := dto.LoginRequest{Email: "test@example.com"}
	body, _ := json.Marshal(req)

	w := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	s.router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogin_NeitherUsernameNorEmail(t *testing.T) {
	s := newTestServer(t)

	req := dto.LoginRequest{
		Password: "password123!",
		Session: dto.Session{
			IPv4Address:    "192.168.1.1",
			OS:             "macOS",
			OSVersion:      "14.0",
			Browser:        "Chrome",
			BrowserVersion: "120.0",
		},
	}
	body, _ := json.Marshal(req)

	w := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	s.router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogin_UsecaseError(t *testing.T) {
	s := newTestServer(t)
	req := validLoginRequest()

	s.mockUsecase.EXPECT().
		Login(gomock.Any(), gomock.Any()).
		Return("", "", errs.New(errs.ErrorInvalidCredentials, errs.MessageInvalidCredentials, nil))

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	s.router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// --- Logout ---

func TestLogout_Success(t *testing.T) {
	s := newTestServer(t)

	s.mockUsecase.EXPECT().
		Logout(gomock.Any(), "refresh-token-value").
		Return(nil)

	w := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/logout", nil)
	httpReq.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh-token-value"})

	s.router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogout_MissingCookie(t *testing.T) {
	s := newTestServer(t)

	w := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/logout", nil)

	s.router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogout_EmptyCookie(t *testing.T) {
	s := newTestServer(t)

	w := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/logout", nil)
	httpReq.AddCookie(&http.Cookie{Name: "refresh_token", Value: ""})

	s.router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogout_UsecaseError(t *testing.T) {
	s := newTestServer(t)

	s.mockUsecase.EXPECT().
		Logout(gomock.Any(), "refresh-token-value").
		Return(errs.New(errs.ErrorUnauthorized, "invalid refresh token", nil))

	w := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/logout", nil)
	httpReq.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh-token-value"})

	s.router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// --- RefreshToken ---

func TestRefreshToken_Success(t *testing.T) {
	s := newTestServer(t)

	s.mockUsecase.EXPECT().
		RefreshToken(gomock.Any(), "old-refresh-token").
		Return("new-access-token", "new-refresh-token", nil)

	w := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/refresh", nil)
	httpReq.AddCookie(&http.Cookie{Name: "refresh_token", Value: "old-refresh-token"})

	s.router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.RefreshTokenResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "new-access-token", resp.AccessToken)
	assert.Equal(t, "new-refresh-token", resp.RefreshToken)
}

func TestRefreshToken_MissingCookie(t *testing.T) {
	s := newTestServer(t)

	w := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/refresh", nil)

	s.router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRefreshToken_UsecaseError(t *testing.T) {
	s := newTestServer(t)

	s.mockUsecase.EXPECT().
		RefreshToken(gomock.Any(), "old-refresh-token").
		Return("", "", errs.New(errs.ErrorUnauthorized, "session expired", nil))

	w := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/refresh", nil)
	httpReq.AddCookie(&http.Cookie{Name: "refresh_token", Value: "old-refresh-token"})

	s.router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
