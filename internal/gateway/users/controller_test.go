package users

import (
	"MydroX/project-v/internal/gateway/users/dto"
	"MydroX/project-v/internal/gateway/users/mocks"
	"MydroX/project-v/pkg/errors"
	"MydroX/project-v/pkg/logger"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	v1 = "/api/v1/users"

	create = "/register"
)

type testServer struct {
	router      *gin.Engine
	mockUsecase *mocks.MockUsersUsecases
}

func testRouter(logger *logger.Logger, _ *gorm.DB, controller Controller) *gin.Engine {
	router := gin.Default()

	err := router.SetTrustedProxies(nil)
	if err != nil {
		logger.Zap.Fatal("error setting trusted proxies", zap.Error(err))
	}

	api := router.Group("api")
	v1 := api.Group("/v1")

	users := v1.Group("/users")
	users.POST("/register", controller.CreateUser)
	users.POST("/auth", controller.AuthenticateUser)
	users.GET("/:uuid", controller.GetUser)

	// TODO: Middleware authentification
	users.PUT("/:uuid", controller.UpdateUser)
	users.PATCH("/:uuid/email", controller.UpdateEmail)
	users.PATCH("/:uuid/password", controller.UpdatePassword)

	users.DELETE("/:uuid", controller.DeleteUser)

	return router
}

func newServerTest(t *testing.T) testServer {
	logger := logger.New("TEST")

	ctrl := gomock.NewController(t)
	usecasesMock := mocks.NewMockUsersUsecases(ctrl)

	c := NewController(logger, usecasesMock)
	router := testRouter(logger, nil, *c)

	return testServer{
		router:      router,
		mockUsecase: usecasesMock,
	}
}

func Test_Create(t *testing.T) {
	s := newServerTest(t)

	t.Run("[V1] Create with success", func(t *testing.T) {
		input := dto.CreateUserRequest{
			Username: "test",
			Email:    "test@test.com",
			Password: "thisisatestpassword1234!@#$",
			Role:     "USER",
		}
		userJSON, _ := json.Marshal(input)

		req, _ := http.NewRequest("POST", v1+create, strings.NewReader(string(userJSON)))

		s.mockUsecase.EXPECT().Create(gomock.Any()).Return(nil)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("[V1] Failed to bind JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", v1+create, strings.NewReader(""))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Failed to validate JSON", func(t *testing.T) {
		input := dto.CreateUserRequest{
			Username: "test",
			Email:    "",
			Password: "thisisatestpassword1234!@#$",
			Role:     "USER",
		}
		userJSON, _ := json.Marshal(input)

		req, _ := http.NewRequest("POST", v1+create, strings.NewReader(string(userJSON)))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Usecase error", func(t *testing.T) {
		input := dto.CreateUserRequest{
			Username: "test",
			Email:    "test@test.com",
			Password: "thisisatestpassword1234!@#$",
			Role:     "USER",
		}
		userJSON, _ := json.Marshal(input)

		req, _ := http.NewRequest("POST", v1+create, strings.NewReader(string(userJSON)))

		s.mockUsecase.EXPECT().Create(gomock.Any()).Return(fmt.Errorf("test error"))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func Test_Get(t *testing.T) {
	s := newServerTest(t)

	uuid := uuid.New()
	user := dto.GetUserResponse{
		UUID:     uuid,
		Username: "testusername",
		Email:    "test@test.com",
		Role:     "USER",
	}

	t.Run("[V1] Get with success", func(t *testing.T) {
		req, _ := http.NewRequest("GET", v1+"/"+uuid.String(), nil)

		s.mockUsecase.EXPECT().Get(uuid).Return(&user, nil)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("[V1] Failed to validate UUID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", v1+"/"+"1", nil)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Failed to find user", func(t *testing.T) {
		req, _ := http.NewRequest("GET", v1+"/"+uuid.String(), nil)

		s.mockUsecase.EXPECT().Get(uuid).Return(nil, errors.ErrNotFound)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("[V1] Usecase error", func(t *testing.T) {
		req, _ := http.NewRequest("GET", v1+"/"+uuid.String(), nil)

		s.mockUsecase.EXPECT().Get(uuid).Return(nil, fmt.Errorf("test error"))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func Test_Update(t *testing.T) {
	s := newServerTest(t)

	uuid := uuid.New()

	t.Run("[V1] Update with success", func(t *testing.T) {
		user := dto.UpdateUserRequest{
			Username: "testusername",
			Email:    "test@test.com",
			Role:     "USER",
			Password: "thisisatestpassword1234!@#$",
		}
		userJSON, _ := json.Marshal(user)

		req, _ := http.NewRequest("PUT", v1+"/"+uuid.String(), strings.NewReader(string(userJSON)))

		s.mockUsecase.EXPECT().Update(user).Return(nil)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("[V1] Failed to bind JSON", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", v1+"/"+uuid.String(), strings.NewReader(""))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Failed to validate JSON", func(t *testing.T) {
		user := dto.UpdateUserRequest{
			Username: "testusername",
			Email:    "",
			Role:     "USER",
			Password: "thisisatestpassword1234!@#$",
		}
		userJSON, _ := json.Marshal(user)

		req, _ := http.NewRequest("PUT", v1+"/"+uuid.String(), strings.NewReader(string(userJSON)))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Usecase error", func(t *testing.T) {
		user := dto.UpdateUserRequest{
			Username: "testusername",
			Email:    "test@test.com",
			Role:     "USER",
			Password: "thisisatestpassword1234!@#$",
		}
		userJSON, _ := json.Marshal(user)

		req, _ := http.NewRequest("PUT", v1+"/"+uuid.String(), strings.NewReader(string(userJSON)))

		s.mockUsecase.EXPECT().Update(gomock.Any()).Return(fmt.Errorf("test error"))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)

	})
}

func Test_UpdateEmail(t *testing.T) {
	s := newServerTest(t)

	uuid := uuid.New()

	t.Run("[V1] Update email with success", func(t *testing.T) {
		user := dto.UpdateEmailRequest{
			Email: "test@test.com",
		}
		userJSON, _ := json.Marshal(user)

		req, _ := http.NewRequest("PATCH", v1+"/"+uuid.String()+"/email", strings.NewReader(string(userJSON)))

		s.mockUsecase.EXPECT().UpdateEmail(uuid, user.Email).Return(nil)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("[V1] Failed to bind JSON", func(t *testing.T) {
		req, _ := http.NewRequest("PATCH", v1+"/"+uuid.String()+"/email", strings.NewReader(""))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Failed to validate JSON", func(t *testing.T) {
		user := dto.UpdateEmailRequest{
			Email: "erthgfderftrfe",
		}
		userJSON, _ := json.Marshal(user)

		req, _ := http.NewRequest("PATCH", v1+"/"+uuid.String()+"/email", strings.NewReader(string(userJSON)))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Failed to validate UUID", func(t *testing.T) {
		req, _ := http.NewRequest("PATCH", v1+"/"+"notanuuid"+"/email", strings.NewReader(""))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Usecase error", func(t *testing.T) {
		user := dto.UpdateEmailRequest{
			Email: "test@test.com",
		}
		userJSON, _ := json.Marshal(user)

		req, _ := http.NewRequest("PATCH", v1+"/"+uuid.String()+"/email", strings.NewReader(string(userJSON)))

		s.mockUsecase.EXPECT().UpdateEmail(uuid, user.Email).Return(fmt.Errorf("test error"))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func Test_UpdatePassword(t *testing.T) {
	s := newServerTest(t)

	uuid := uuid.New()

	t.Run("[V1] Update password with success", func(t *testing.T) {
		user := dto.UpdatePasswordRequest{
			Password: "thisisatestpassword123?",
		}
		userJSON, _ := json.Marshal(user)

		req, _ := http.NewRequest("PATCH", v1+"/"+uuid.String()+"/password", strings.NewReader(string(userJSON)))

		s.mockUsecase.EXPECT().UpdatePassword(uuid, user.Password).Return(nil)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("[V1] Failed to bind JSON", func(t *testing.T) {
		req, _ := http.NewRequest("PATCH", v1+"/"+uuid.String()+"/password", strings.NewReader(""))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Failed to validate JSON", func(t *testing.T) {
		user := dto.UpdatePasswordRequest{
			Password: "a",
		}
		userJSON, _ := json.Marshal(user)

		req, _ := http.NewRequest("PATCH", v1+"/"+uuid.String()+"/password", strings.NewReader(string(userJSON)))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Failed to validate UUID", func(t *testing.T) {
		req, _ := http.NewRequest("PATCH", v1+"/"+"notanuuid"+"/password", strings.NewReader(""))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Usecase error", func(t *testing.T) {
		user := dto.UpdatePasswordRequest{
			Password: "thisisatestpassword123?",
		}
		userJSON, _ := json.Marshal(user)

		req, _ := http.NewRequest("PATCH", v1+"/"+uuid.String()+"/password", strings.NewReader(string(userJSON)))

		s.mockUsecase.EXPECT().UpdatePassword(uuid, user.Password).Return(fmt.Errorf("test error"))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func Test_Delete(t *testing.T) {
	s := newServerTest(t)

	uuid := uuid.New()

	t.Run("[V1] Delete with success", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", v1+"/"+uuid.String(), nil)

		s.mockUsecase.EXPECT().Delete(uuid).Return(nil)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("[V1] Failed to validate UUID", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", v1+"/"+"notanuuid", nil)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Usecase error", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", v1+"/"+uuid.String(), nil)

		s.mockUsecase.EXPECT().Delete(uuid).Return(fmt.Errorf("test error"))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
