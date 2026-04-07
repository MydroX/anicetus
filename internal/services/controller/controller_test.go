package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"MydroX/anicetus/internal/config"
	"MydroX/anicetus/internal/services/dto"
	"MydroX/anicetus/internal/services/mocks"
	"MydroX/anicetus/pkg/errs"
	"MydroX/anicetus/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type testServer struct {
	router      *gin.Engine
	mockUsecase *mocks.MockServicesUsecases
}

func newTestServer(t *testing.T) testServer {
	ctrl := gomock.NewController(t)
	mockUC := mocks.NewMockServicesUsecases(ctrl)
	log, _ := logger.New("TEST")
	cfg := &config.Config{}

	c := New(log, mockUC, cfg)

	router := gin.New()
	v1 := router.Group("/api/v1")
	Router(v1, c)

	return testServer{router: router, mockUsecase: mockUC}
}

func jsonReq(method, path string, body any) *http.Request {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// --- RegisterService ---

func TestRegisterService_Success(t *testing.T) {
	s := newTestServer(t)
	s.mockUsecase.EXPECT().RegisterService(gomock.Any(), gomock.Any()).Return(nil)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, jsonReq(http.MethodPost, "/api/v1/services", dto.RegisterServiceRequest{
		Audience: "my-svc", ServiceName: "My Service",
	}))

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestRegisterService_BindFailure(t *testing.T) {
	s := newTestServer(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/services", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegisterService_ValidationFailure(t *testing.T) {
	s := newTestServer(t)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, jsonReq(http.MethodPost, "/api/v1/services", dto.RegisterServiceRequest{
		// Missing required fields
	}))

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegisterService_UsecaseError(t *testing.T) {
	s := newTestServer(t)
	s.mockUsecase.EXPECT().RegisterService(gomock.Any(), gomock.Any()).
		Return(errs.New(errs.ErrorUniqueViolation, "exists", nil))

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, jsonReq(http.MethodPost, "/api/v1/services", dto.RegisterServiceRequest{
		Audience: "dup", ServiceName: "Dup",
	}))

	assert.Equal(t, http.StatusConflict, w.Code)
}

// --- RevokeService ---

func TestRevokeService_Success(t *testing.T) {
	s := newTestServer(t)
	s.mockUsecase.EXPECT().RevokeService(gomock.Any(), "my-svc").Return(nil)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/services/my-svc", nil))

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRevokeService_UsecaseError(t *testing.T) {
	s := newTestServer(t)
	s.mockUsecase.EXPECT().RevokeService(gomock.Any(), "unknown").
		Return(errs.New(errs.ErrorNotFound, "not found", nil))

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/services/unknown", nil))

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- GetAllServices ---

func TestGetAllServices_Success(t *testing.T) {
	s := newTestServer(t)
	s.mockUsecase.EXPECT().GetAllServices(gomock.Any()).Return([]string{"svc-a", "svc-b"}, nil)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/services", nil))

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.ServiceListResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, []string{"svc-a", "svc-b"}, resp.Services)
}

func TestGetAllServices_UsecaseError(t *testing.T) {
	s := newTestServer(t)
	s.mockUsecase.EXPECT().GetAllServices(gomock.Any()).
		Return(nil, errs.New(errs.ErrorInternal, "err", nil))

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/services", nil))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- GetUserServices ---

func TestGetUserServices_Success(t *testing.T) {
	s := newTestServer(t)
	userUUID := uuid.New().String()
	s.mockUsecase.EXPECT().GetUserServices(gomock.Any(), userUUID).Return([]string{"svc-1"}, nil)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/services/users/"+userUUID, nil))

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetUserServices_InvalidUUID(t *testing.T) {
	s := newTestServer(t)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/services/users/invalid-uuid", nil))

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetUserServices_UsecaseError(t *testing.T) {
	s := newTestServer(t)
	userUUID := uuid.New().String()
	s.mockUsecase.EXPECT().GetUserServices(gomock.Any(), userUUID).
		Return(nil, errs.New(errs.ErrorInternal, "err", nil))

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/services/users/"+userUUID, nil))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- AssignServiceToUser ---

func TestAssignServiceToUser_Success(t *testing.T) {
	s := newTestServer(t)
	userUUID := uuid.New().String()
	s.mockUsecase.EXPECT().AssignServiceToUser(gomock.Any(), userUUID, gomock.Any()).Return(nil)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, jsonReq(http.MethodPost, "/api/v1/services/users/"+userUUID, dto.AssignServiceRequest{Audience: "svc"}))

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAssignServiceToUser_InvalidUUID(t *testing.T) {
	s := newTestServer(t)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, jsonReq(http.MethodPost, "/api/v1/services/users/invalid-uuid", dto.AssignServiceRequest{Audience: "svc"}))

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAssignServiceToUser_BindFailure(t *testing.T) {
	s := newTestServer(t)
	userUUID := uuid.New().String()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/services/users/"+userUUID, bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAssignServiceToUser_ValidationFailure(t *testing.T) {
	s := newTestServer(t)
	userUUID := uuid.New().String()

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, jsonReq(http.MethodPost, "/api/v1/services/users/"+userUUID, dto.AssignServiceRequest{}))

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAssignServiceToUser_UsecaseError(t *testing.T) {
	s := newTestServer(t)
	userUUID := uuid.New().String()
	s.mockUsecase.EXPECT().AssignServiceToUser(gomock.Any(), userUUID, gomock.Any()).
		Return(errs.New(errs.ErrorNotFound, "not found", nil))

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, jsonReq(http.MethodPost, "/api/v1/services/users/"+userUUID, dto.AssignServiceRequest{Audience: "svc"}))

	assert.Equal(t, http.StatusNotFound, w.Code)
}
