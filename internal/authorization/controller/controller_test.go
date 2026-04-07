package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"MydroX/anicetus/internal/authorization/dto"
	"MydroX/anicetus/internal/authorization/mocks"
	"MydroX/anicetus/pkg/errs"
	"MydroX/anicetus/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type testServer struct {
	router      *gin.Engine
	mockUsecase *mocks.MockAuthorizationUsecases
}

func newTestServer(t *testing.T) testServer {
	ctrl := gomock.NewController(t)
	mockUC := mocks.NewMockAuthorizationUsecases(ctrl)
	log, _ := logger.New("TEST")

	c := New(log, mockUC)

	router := gin.New()
	v1 := router.Group("/api/v1")
	Router(v1, c)

	return testServer{router: router, mockUsecase: mockUC}
}

// --- CreateRole ---

func TestCreateRole(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		s := newTestServer(t)

		s.mockUsecase.EXPECT().
			CreateRole(gomock.Any(), gomock.Any()).
			Return(nil)

		body, _ := json.Marshal(dto.CreateRoleRequest{Name: "admin", Description: "Administrator role"})
		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("BindFailure", func(t *testing.T) {
		s := newTestServer(t)

		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewBufferString("invalid"))
		httpReq.Header.Set("Content-Type", "application/json")

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ValidationFailure", func(t *testing.T) {
		s := newTestServer(t)

		body, _ := json.Marshal(dto.CreateRoleRequest{Name: ""})
		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		s := newTestServer(t)

		s.mockUsecase.EXPECT().
			CreateRole(gomock.Any(), gomock.Any()).
			Return(errs.New(errs.ErrorUniqueViolation, "role already exists", nil))

		body, _ := json.Marshal(dto.CreateRoleRequest{Name: "admin", Description: "Administrator role"})
		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusConflict, w.Code)
	})
}

// --- GetAllRoles ---

func TestGetAllRoles(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		s := newTestServer(t)

		roles := []*dto.RoleResponse{
			{UUID: uuid.New().String(), Name: "admin", Description: "Admin role"},
			{UUID: uuid.New().String(), Name: "user", Description: "User role"},
		}

		s.mockUsecase.EXPECT().
			GetAllRoles(gomock.Any()).
			Return(roles, nil)

		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/roles", nil)

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		s := newTestServer(t)

		s.mockUsecase.EXPECT().
			GetAllRoles(gomock.Any()).
			Return(nil, errs.New(errs.ErrorInternal, "internal error", nil))

		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/roles", nil)

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// --- UpdateRole ---

func TestUpdateRole(t *testing.T) {
	validUUID := uuid.New().String()

	t.Run("Success", func(t *testing.T) {
		s := newTestServer(t)

		s.mockUsecase.EXPECT().
			UpdateRole(gomock.Any(), validUUID, gomock.Any()).
			Return(nil)

		body, _ := json.Marshal(dto.UpdateRoleRequest{Name: "admin-updated", Description: "Updated"})
		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/roles/%s", validUUID), bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("InvalidUUID", func(t *testing.T) {
		s := newTestServer(t)

		body, _ := json.Marshal(dto.UpdateRoleRequest{Name: "admin"})
		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/roles/invalid-uuid", bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("BindFailure", func(t *testing.T) {
		s := newTestServer(t)

		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/roles/%s", validUUID), bytes.NewBufferString("invalid"))
		httpReq.Header.Set("Content-Type", "application/json")

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ValidationFailure", func(t *testing.T) {
		s := newTestServer(t)

		body, _ := json.Marshal(dto.UpdateRoleRequest{Name: ""})
		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/roles/%s", validUUID), bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		s := newTestServer(t)

		s.mockUsecase.EXPECT().
			UpdateRole(gomock.Any(), validUUID, gomock.Any()).
			Return(errs.New(errs.ErrorNotFound, "role not found", nil))

		body, _ := json.Marshal(dto.UpdateRoleRequest{Name: "admin-updated", Description: "Updated"})
		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/roles/%s", validUUID), bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// --- DeleteRole ---

func TestDeleteRole(t *testing.T) {
	validUUID := uuid.New().String()

	t.Run("Success", func(t *testing.T) {
		s := newTestServer(t)

		s.mockUsecase.EXPECT().
			DeleteRole(gomock.Any(), validUUID).
			Return(nil)

		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/roles/%s", validUUID), nil)

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("InvalidUUID", func(t *testing.T) {
		s := newTestServer(t)

		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodDelete, "/api/v1/roles/invalid-uuid", nil)

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		s := newTestServer(t)

		s.mockUsecase.EXPECT().
			DeleteRole(gomock.Any(), validUUID).
			Return(errs.New(errs.ErrorNotFound, "role not found", nil))

		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/roles/%s", validUUID), nil)

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// --- CreatePermission ---

func TestCreatePermission(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		s := newTestServer(t)

		s.mockUsecase.EXPECT().
			CreatePermission(gomock.Any(), gomock.Any()).
			Return(nil)

		body, _ := json.Marshal(dto.CreatePermissionRequest{Name: "users:read", Description: "Read users"})
		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/permissions", bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("BindFailure", func(t *testing.T) {
		s := newTestServer(t)

		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/permissions", bytes.NewBufferString("invalid"))
		httpReq.Header.Set("Content-Type", "application/json")

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ValidationFailure", func(t *testing.T) {
		s := newTestServer(t)

		body, _ := json.Marshal(dto.CreatePermissionRequest{Name: ""})
		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/permissions", bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		s := newTestServer(t)

		s.mockUsecase.EXPECT().
			CreatePermission(gomock.Any(), gomock.Any()).
			Return(errs.New(errs.ErrorUniqueViolation, "permission already exists", nil))

		body, _ := json.Marshal(dto.CreatePermissionRequest{Name: "users:read", Description: "Read users"})
		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/permissions", bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusConflict, w.Code)
	})
}

// --- GetAllPermissions ---

func TestGetAllPermissions(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		s := newTestServer(t)

		permissions := []*dto.PermissionResponse{
			{UUID: uuid.New().String(), Name: "users:read", Description: "Read users"},
			{UUID: uuid.New().String(), Name: "users:write", Description: "Write users"},
		}

		s.mockUsecase.EXPECT().
			GetAllPermissions(gomock.Any()).
			Return(permissions, nil)

		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/permissions", nil)

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		s := newTestServer(t)

		s.mockUsecase.EXPECT().
			GetAllPermissions(gomock.Any()).
			Return(nil, errs.New(errs.ErrorInternal, "internal error", nil))

		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/permissions", nil)

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// --- AssignPermissionToRole ---

func TestAssignPermissionToRole(t *testing.T) {
	validRoleUUID := uuid.New().String()
	validPermUUID := uuid.New().String()

	t.Run("Success", func(t *testing.T) {
		s := newTestServer(t)

		s.mockUsecase.EXPECT().
			AssignPermissionToRole(gomock.Any(), validRoleUUID, validPermUUID).
			Return(nil)

		body, _ := json.Marshal(dto.AssignPermissionToRoleRequest{PermissionUUID: validPermUUID})
		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/roles/%s/permissions", validRoleUUID), bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("InvalidUUID", func(t *testing.T) {
		s := newTestServer(t)

		body, _ := json.Marshal(dto.AssignPermissionToRoleRequest{PermissionUUID: validPermUUID})
		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/roles/invalid-uuid/permissions", bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("BindFailure", func(t *testing.T) {
		s := newTestServer(t)

		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/roles/%s/permissions", validRoleUUID), bytes.NewBufferString("invalid"))
		httpReq.Header.Set("Content-Type", "application/json")

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ValidationFailure", func(t *testing.T) {
		s := newTestServer(t)

		body, _ := json.Marshal(dto.AssignPermissionToRoleRequest{PermissionUUID: ""})
		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/roles/%s/permissions", validRoleUUID), bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		s := newTestServer(t)

		s.mockUsecase.EXPECT().
			AssignPermissionToRole(gomock.Any(), validRoleUUID, validPermUUID).
			Return(errs.New(errs.ErrorNotFound, "role not found", nil))

		body, _ := json.Marshal(dto.AssignPermissionToRoleRequest{PermissionUUID: validPermUUID})
		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/roles/%s/permissions", validRoleUUID), bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// --- RemovePermissionFromRole ---

func TestRemovePermissionFromRole(t *testing.T) {
	validRoleUUID := uuid.New().String()
	validPermUUID := uuid.New().String()

	t.Run("Success", func(t *testing.T) {
		s := newTestServer(t)

		s.mockUsecase.EXPECT().
			RemovePermissionFromRole(gomock.Any(), validRoleUUID, validPermUUID).
			Return(nil)

		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/roles/%s/permissions/%s", validRoleUUID, validPermUUID), nil)

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("InvalidRoleUUID", func(t *testing.T) {
		s := newTestServer(t)

		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/roles/invalid-uuid/permissions/%s", validPermUUID), nil)

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("InvalidPermUUID", func(t *testing.T) {
		s := newTestServer(t)

		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/roles/%s/permissions/invalid-uuid", validRoleUUID), nil)

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		s := newTestServer(t)

		s.mockUsecase.EXPECT().
			RemovePermissionFromRole(gomock.Any(), validRoleUUID, validPermUUID).
			Return(errs.New(errs.ErrorNotFound, "assignment not found", nil))

		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/roles/%s/permissions/%s", validRoleUUID, validPermUUID), nil)

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// --- AssignRoleToUser ---

func TestAssignRoleToUser(t *testing.T) {
	validUserUUID := uuid.New().String()
	validRoleUUID := uuid.New().String()

	t.Run("Success", func(t *testing.T) {
		s := newTestServer(t)

		s.mockUsecase.EXPECT().
			AssignRoleToUser(gomock.Any(), validUserUUID, validRoleUUID).
			Return(nil)

		body, _ := json.Marshal(dto.AssignRoleToUserRequest{RoleUUID: validRoleUUID})
		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/users/%s/roles", validUserUUID), bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("InvalidUUID", func(t *testing.T) {
		s := newTestServer(t)

		body, _ := json.Marshal(dto.AssignRoleToUserRequest{RoleUUID: validRoleUUID})
		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/users/invalid-uuid/roles", bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("BindFailure", func(t *testing.T) {
		s := newTestServer(t)

		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/users/%s/roles", validUserUUID), bytes.NewBufferString("invalid"))
		httpReq.Header.Set("Content-Type", "application/json")

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ValidationFailure", func(t *testing.T) {
		s := newTestServer(t)

		body, _ := json.Marshal(dto.AssignRoleToUserRequest{RoleUUID: ""})
		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/users/%s/roles", validUserUUID), bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		s := newTestServer(t)

		s.mockUsecase.EXPECT().
			AssignRoleToUser(gomock.Any(), validUserUUID, validRoleUUID).
			Return(errs.New(errs.ErrorNotFound, "user not found", nil))

		body, _ := json.Marshal(dto.AssignRoleToUserRequest{RoleUUID: validRoleUUID})
		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/users/%s/roles", validUserUUID), bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// --- RemoveRoleFromUser ---

func TestRemoveRoleFromUser(t *testing.T) {
	validUserUUID := uuid.New().String()
	validRoleUUID := uuid.New().String()

	t.Run("Success", func(t *testing.T) {
		s := newTestServer(t)

		s.mockUsecase.EXPECT().
			RemoveRoleFromUser(gomock.Any(), validUserUUID, validRoleUUID).
			Return(nil)

		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/users/%s/roles/%s", validUserUUID, validRoleUUID), nil)

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("InvalidUserUUID", func(t *testing.T) {
		s := newTestServer(t)

		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/users/invalid-uuid/roles/%s", validRoleUUID), nil)

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("InvalidRoleUUID", func(t *testing.T) {
		s := newTestServer(t)

		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/users/%s/roles/invalid-uuid", validUserUUID), nil)

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		s := newTestServer(t)

		s.mockUsecase.EXPECT().
			RemoveRoleFromUser(gomock.Any(), validUserUUID, validRoleUUID).
			Return(errs.New(errs.ErrorNotFound, "assignment not found", nil))

		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/users/%s/roles/%s", validUserUUID, validRoleUUID), nil)

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// --- GetUserRoles ---

func TestGetUserRoles(t *testing.T) {
	validUUID := uuid.New().String()

	t.Run("Success", func(t *testing.T) {
		s := newTestServer(t)

		roles := []*dto.RoleResponse{
			{UUID: uuid.New().String(), Name: "admin", Description: "Admin role"},
		}

		s.mockUsecase.EXPECT().
			GetUserRoles(gomock.Any(), validUUID).
			Return(roles, nil)

		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/users/%s/roles", validUUID), nil)

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("InvalidUUID", func(t *testing.T) {
		s := newTestServer(t)

		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/users/invalid-uuid/roles", nil)

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		s := newTestServer(t)

		s.mockUsecase.EXPECT().
			GetUserRoles(gomock.Any(), validUUID).
			Return(nil, errs.New(errs.ErrorNotFound, "user not found", nil))

		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/users/%s/roles", validUUID), nil)

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// --- GetUserPermissions ---

func TestGetUserPermissions(t *testing.T) {
	validUUID := uuid.New().String()

	t.Run("Success", func(t *testing.T) {
		s := newTestServer(t)

		permissions := []string{"users:read", "users:write"}

		s.mockUsecase.EXPECT().
			GetUserPermissions(gomock.Any(), validUUID).
			Return(permissions, nil)

		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/users/%s/permissions", validUUID), nil)

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp dto.UserPermissionsResponse
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, permissions, resp.Permissions)
	})

	t.Run("InvalidUUID", func(t *testing.T) {
		s := newTestServer(t)

		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/users/invalid-uuid/permissions", nil)

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		s := newTestServer(t)

		s.mockUsecase.EXPECT().
			GetUserPermissions(gomock.Any(), validUUID).
			Return(nil, errs.New(errs.ErrorNotFound, "user not found", nil))

		w := httptest.NewRecorder()
		httpReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/users/%s/permissions", validUUID), nil)

		s.router.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
