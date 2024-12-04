package users

import (
	"MydroX/project-v/internal/common/errors"
	"MydroX/project-v/internal/common/response"
	"MydroX/project-v/internal/gateway/users/config"
	"MydroX/project-v/internal/gateway/users/dto"
	"MydroX/project-v/internal/gateway/users/mocks"
	"MydroX/project-v/internal/gateway/users/models"
	loggerpkg "MydroX/project-v/pkg/logger"
	uuidpkg "MydroX/project-v/pkg/uuid"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	v1 = "/api/v1/users"

	create = "/register"

	prefix = "user"
)

type testServer struct {
	router      *gin.Engine
	mockUsecase *mocks.MockUsersUsecases
}

func testRouter(logger *loggerpkg.Logger, _ *gorm.DB, controller Controller) *gin.Engine {
	router := gin.Default()

	err := router.SetTrustedProxies(nil)
	if err != nil {
		logger.Zap.Fatal("error setting trusted proxies", zap.Error(err))
	}

	api := router.Group("api")
	v1 := api.Group("/v1")

	users := v1.Group("/users")
	users.POST("/register", controller.CreateUser)
	users.POST("/login", controller.Login)
	users.GET("/:uuid", controller.GetUser)

	// TODO: Middleware authentification
	users.PUT("/:uuid", controller.UpdateUser)
	users.PATCH("/:uuid/email", controller.UpdateEmail)
	users.PATCH("/:uuid/password", controller.UpdatePassword)

	users.DELETE("/:uuid", controller.DeleteUser)

	return router
}

func newServerTest(t *testing.T) testServer {
	logger := loggerpkg.New("TEST")

	ctrl := gomock.NewController(t)
	usecasesMock := mocks.NewMockUsersUsecases(ctrl)

	c := NewController(logger, usecasesMock, &config.Config{})
	router := testRouter(logger, nil, *c)

	return testServer{
		router:      router,
		mockUsecase: usecasesMock,
	}
}

func Test_Create(t *testing.T) {
	s := newServerTest(t)

	t.Run("[V1] Create user with success", func(t *testing.T) {
		ctx := context.Background()

		input := dto.CreateUserRequest{
			Username: "test",
			Email:    "test@test.com",
			Password: "thisisatestpassword1234!@#$",
			Role:     "USER",
		}
		userJSON, _ := json.Marshal(input)

		req, _ := http.NewRequest("POST", v1+create, strings.NewReader(string(userJSON)))

		user := &models.User{
			Username: input.Username,
			Email:    input.Email,
			Password: input.Password,
			Role:     input.Role,
		}

		s.mockUsecase.EXPECT().Create(&ctx, user).Return(nil)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("[V1] Create - Failed to bind JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", v1+create, strings.NewReader(""))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Create - Failed to validate JSON", func(t *testing.T) {
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

	t.Run("[V1] Create - Invalid username", func(t *testing.T) {
		input := dto.CreateUserRequest{
			Username: "test@@@",
			Email:    "test@test.com",
			Password: "thisisatestpassword1234!@#$",
			Role:     "USER",
		}

		userJSON, _ := json.Marshal(input)

		req, _ := http.NewRequest("POST", v1+create, strings.NewReader(string(userJSON)))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		var resp response.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, errors.CODE_INVALID_USERNAME, resp.Code)
	})

	t.Run("[V1] Create - Usecase error", func(t *testing.T) {
		ctx := context.Background()

		input := dto.CreateUserRequest{
			Username: "test",
			Email:    "test@test.com",
			Password: "thisisatestpassword1234!@#$",
			Role:     "USER",
		}
		userJSON, _ := json.Marshal(input)

		req, _ := http.NewRequest("POST", v1+create, strings.NewReader(string(userJSON)))

		user := &models.User{
			Username: input.Username,
			Email:    input.Email,
			Password: input.Password,
			Role:     input.Role,
		}

		s.mockUsecase.EXPECT().Create(&ctx, user).Return(fmt.Errorf("test error"))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func Test_Get(t *testing.T) {
	s := newServerTest(t)

	uuid := uuidpkg.NewWithPrefix(prefix)
	user := models.User{
		UUID:     uuid,
		Username: "testusername",
		Email:    "test@test.com",
		Role:     "USER",
	}

	t.Run("[V1] Get user with success", func(t *testing.T) {
		ctx := context.Background()

		req, _ := http.NewRequest("GET", v1+"/"+uuid, nil)

		s.mockUsecase.EXPECT().Get(&ctx, uuid).Return(&user, nil)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("[V1] Get - Failed to validate UUID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", v1+"/"+"1", nil)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Get - Failed to find user", func(t *testing.T) {
		ctx := context.Background()

		req, _ := http.NewRequest("GET", v1+"/"+uuid, nil)

		s.mockUsecase.EXPECT().Get(&ctx, uuid).DoAndReturn(
			func(ctx *context.Context, _ string) (*models.User, error) {
				*ctx = context.WithValue(*ctx, errors.CtxErrorCodeKey, errors.CODE_ENTITY_NOT_FOUND)
				return nil, fmt.Errorf("user not found")
			})

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		resp := response.ErrorResponse{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, errors.CODE_ENTITY_NOT_FOUND, resp.Code)
	})

	t.Run("[V1] Get -  Usecase error", func(t *testing.T) {
		ctx := context.Background()

		req, _ := http.NewRequest("GET", v1+"/"+uuid, nil)

		s.mockUsecase.EXPECT().Get(&ctx, uuid).Return(nil, fmt.Errorf("test error"))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func Test_Update(t *testing.T) {
	s := newServerTest(t)

	uuid := uuidpkg.NewWithPrefix(prefix)

	t.Run("[V1] Update with success", func(t *testing.T) {
		ctx := context.Background()

		user := models.User{
			UUID:     uuid,
			Username: "testusername",
			Email:    "test@test.com",
			Role:     "USER",
			Password: "thisisatestpassword1234!@#$",
		}
		userJSON, _ := json.Marshal(user)

		req, _ := http.NewRequest("PUT", v1+"/"+uuid, strings.NewReader(string(userJSON)))

		s.mockUsecase.EXPECT().Update(&ctx, &user).Return(nil)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("[V1] Failed to bind JSON", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", v1+"/"+uuid, strings.NewReader(""))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Failed to validate JSON", func(t *testing.T) {
		user := models.User{
			UUID:     uuid,
			Username: "testusername",
			Email:    "",
			Role:     "USER",
			Password: "thisisatestpassword1234!@#$",
		}
		userJSON, _ := json.Marshal(user)

		req, _ := http.NewRequest("PUT", v1+"/"+uuid, strings.NewReader(string(userJSON)))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Usecase error", func(t *testing.T) {
		ctx := context.Background()

		user := models.User{
			UUID:     uuid,
			Username: "testusername",
			Email:    "test@test.com",
			Role:     "USER",
			Password: "thisisatestpassword1234!@#$",
		}
		userJSON, _ := json.Marshal(user)

		req, _ := http.NewRequest("PUT", v1+"/"+uuid, strings.NewReader(string(userJSON)))

		s.mockUsecase.EXPECT().Update(&ctx, &user).Return(fmt.Errorf("test error"))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func Test_UpdateEmail(t *testing.T) {
	s := newServerTest(t)

	uuid := uuidpkg.NewWithPrefix(prefix)

	t.Run("[V1] Update email with success", func(t *testing.T) {
		ctx := context.Background()

		user := dto.UpdateEmailRequest{
			Email: "test@test.com",
		}
		userJSON, _ := json.Marshal(user)

		req, _ := http.NewRequest("PATCH", v1+"/"+uuid+"/email", strings.NewReader(string(userJSON)))

		s.mockUsecase.EXPECT().UpdateEmail(&ctx, uuid, user.Email).Return(nil)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("[V1] Update - Failed to bind JSON", func(t *testing.T) {
		req, _ := http.NewRequest("PATCH", v1+"/"+uuid+"/email", strings.NewReader(""))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Update - Failed to validate JSON", func(t *testing.T) {
		user := dto.UpdateEmailRequest{
			Email: "erthgfderftrfe",
		}
		userJSON, _ := json.Marshal(user)

		req, _ := http.NewRequest("PATCH", v1+"/"+uuid+"/email", strings.NewReader(string(userJSON)))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Update - Failed to validate UUID", func(t *testing.T) {
		req, _ := http.NewRequest("PATCH", v1+"/"+"notanuuid"+"/email", strings.NewReader(""))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Update - Usecase error", func(t *testing.T) {
		ctx := context.Background()

		user := dto.UpdateEmailRequest{
			Email: "test@test.com",
		}
		userJSON, _ := json.Marshal(user)

		req, _ := http.NewRequest("PATCH", v1+"/"+uuid+"/email", strings.NewReader(string(userJSON)))

		s.mockUsecase.EXPECT().UpdateEmail(&ctx, uuid, user.Email).Return(fmt.Errorf("test error"))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func Test_UpdatePassword(t *testing.T) {
	s := newServerTest(t)

	uuid := uuidpkg.NewWithPrefix(prefix)

	t.Run("[V1] Update password with success", func(t *testing.T) {
		ctx := context.Background()

		user := dto.UpdatePasswordRequest{
			Password: "thisisatestpassword123?",
		}
		userJSON, _ := json.Marshal(user)

		req, _ := http.NewRequest("PATCH", v1+"/"+uuid+"/password", strings.NewReader(string(userJSON)))

		s.mockUsecase.EXPECT().UpdatePassword(&ctx, uuid, user.Password).Return(nil)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("[V1] Update - Failed to bind JSON", func(t *testing.T) {
		req, _ := http.NewRequest("PATCH", v1+"/"+uuid+"/password", strings.NewReader(""))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Update - Failed to validate JSON", func(t *testing.T) {
		user := dto.UpdatePasswordRequest{
			Password: "a",
		}
		userJSON, _ := json.Marshal(user)

		req, _ := http.NewRequest("PATCH", v1+"/"+uuid+"/password", strings.NewReader(string(userJSON)))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Update - Failed to validate UUID", func(t *testing.T) {
		req, _ := http.NewRequest("PATCH", v1+"/"+"notanuuid"+"/password", strings.NewReader(""))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Update - Usecase error", func(t *testing.T) {
		ctx := context.Background()

		user := dto.UpdatePasswordRequest{
			Password: "thisisatestpassword123?",
		}
		userJSON, _ := json.Marshal(user)

		req, _ := http.NewRequest("PATCH", v1+"/"+uuid+"/password", strings.NewReader(string(userJSON)))

		s.mockUsecase.EXPECT().UpdatePassword(&ctx, uuid, user.Password).Return(fmt.Errorf("test error"))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func Test_Delete(t *testing.T) {
	s := newServerTest(t)

	uuid := uuidpkg.NewWithPrefix(prefix)

	t.Run("[V1] Delete with success", func(t *testing.T) {
		ctx := context.Background()

		req, _ := http.NewRequest("DELETE", v1+"/"+uuid, nil)

		s.mockUsecase.EXPECT().Delete(&ctx, uuid).Return(nil)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("[V1] Delete - failed to validate UUID", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", v1+"/"+"notanuuid", nil)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Delete - Usecase error", func(t *testing.T) {
		ctx := context.Background()

		req, _ := http.NewRequest("DELETE", v1+"/"+uuid, nil)

		s.mockUsecase.EXPECT().Delete(&ctx, uuid).Return(fmt.Errorf("test error"))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func Test_Login(t *testing.T) {
	s := newServerTest(t)

	t.Run("[V1] Login with success", func(t *testing.T) {
		ctx := context.Background()

		input := dto.LoginRequest{
			Username: "test",
			Password: "thisisatest123",
		}

		user := models.User{
			Username: input.Username,
			Email:    input.Email,
			Password: input.Password,
		}

		userJSON, _ := json.Marshal(input)

		fmt.Println(v1 + "/login")
		req, _ := http.NewRequest("POST", v1+"/login", strings.NewReader(string(userJSON)))

		s.mockUsecase.EXPECT().Login(&ctx, user.Username, user.Email, user.Password).Return("token", nil)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		var resp dto.LoginResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, resp.Token)
	})

	t.Run("[V1] Login - Failed to bind JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", v1+"/login", strings.NewReader(""))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Login - Failed to validate JSON", func(t *testing.T) {
		input := dto.LoginRequest{
			Username: "tes",
			Password: "thisisatest123",
		}

		userJSON, _ := json.Marshal(input)

		req, _ := http.NewRequest("POST", v1+"/login", strings.NewReader(string(userJSON)))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Login - No username and email in body", func(t *testing.T) {
		input := dto.LoginRequest{
			Password: "thisisatest123",
		}

		userJSON, _ := json.Marshal(input)

		req, _ := http.NewRequest("POST", v1+"/login", strings.NewReader(string(userJSON)))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Login - Failed to find user", func(t *testing.T) {
		ctx := context.Background()

		input := dto.LoginRequest{
			Username: "test",
			Password: "thisisatest123",
		}

		user := models.User{
			Username: input.Username,
			Email:    input.Email,
			Password: input.Password,
		}

		userJSON, _ := json.Marshal(input)

		req, _ := http.NewRequest("POST", v1+"/login", strings.NewReader(string(userJSON)))

		s.mockUsecase.EXPECT().Login(&ctx, user.Username, user.Email, user.Password).DoAndReturn(
			func(ctx *context.Context, _ string, _ string, _ string) (string, error) {
				*ctx = context.WithValue(*ctx, errors.CtxErrorCodeKey, errors.CODE_ENTITY_NOT_FOUND)
				return "", fmt.Errorf("user not found")
			})

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		resp := response.ErrorResponse{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, errors.CODE_ENTITY_NOT_FOUND, resp.Code)
	})

	t.Run("[V1] Login - Usecase error", func(t *testing.T) {
		ctx := context.Background()

		input := dto.LoginRequest{
			Username: "test",
			Password: "thisisatest123",
		}

		user := models.User{
			Username: input.Username,
			Email:    input.Email,
			Password: input.Password,
		}

		userJSON, _ := json.Marshal(input)

		req, _ := http.NewRequest("POST", v1+"/login", strings.NewReader(string(userJSON)))

		s.mockUsecase.EXPECT().Login(&ctx, user.Username, user.Email, user.Password).Return("", fmt.Errorf("test error"))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
