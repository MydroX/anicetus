package controller

import (
	"MydroX/anicetus/internal/common/errorsutil"
	"MydroX/anicetus/internal/common/uuidutil"
	"MydroX/anicetus/internal/config"
	"MydroX/anicetus/internal/users/dto"
	"MydroX/anicetus/internal/users/mocks"
	loggerpkg "MydroX/anicetus/pkg/logger"
	uuidpkg "MydroX/anicetus/pkg/uuid"

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
)

const (
	v1    = "/api/v1"
	users = "/users"

	create = "/register"
)

type testServer struct {
	router      *gin.Engine
	mockUsecase *mocks.MockUsersUsecases
}

func testRouter(logger *loggerpkg.Logger, controller ControllerInterface) *gin.Engine {
	router := gin.Default()

	err := router.SetTrustedProxies(nil)
	if err != nil {
		logger.Zap.Fatal("error setting trusted proxies", zap.Error(err))
	}

	api := router.Group("api")
	v1 := api.Group("/v1")

	Router(v1, controller)

	return router
}

func newServerTest(t *testing.T) testServer {
	logger := loggerpkg.New("TEST")

	ctrl := gomock.NewController(t)
	usecasesMock := mocks.NewMockUsersUsecases(ctrl)

	c := New(logger, usecasesMock, &config.Config{})
	router := testRouter(logger, c)

	return testServer{
		router:      router,
		mockUsecase: usecasesMock,
	}
}

func Test_Create(t *testing.T) {
	s := newServerTest(t)

	t.Run("[V1] Create user with success", func(t *testing.T) {
		input := dto.CreateUserRequest{
			Username: "test",
			Email:    "test@test.com",
			Password: "Thisisatestpassword1234!@#$",
		}
		userJSON, _ := json.Marshal(input)

		req, _ := http.NewRequest("POST", v1+create, strings.NewReader(string(userJSON)))

		user := &dto.CreateUserRequest{
			Username: input.Username,
			Email:    input.Email,
			Password: input.Password,
		}

		s.mockUsecase.EXPECT().Create(gomock.Any(), user).Return(nil)

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
			Password: "Thisisatestpassword1234!@#$",
		}

		userJSON, _ := json.Marshal(input)

		req, _ := http.NewRequest("POST", v1+create, strings.NewReader(string(userJSON)))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Create - Usecase error", func(t *testing.T) {
		input := dto.CreateUserRequest{
			Username: "test",
			Email:    "test@test.com",
			Password: "Thisisatestpassword1234!@#$",
		}
		userJSON, _ := json.Marshal(input)

		req, _ := http.NewRequest("POST", v1+create, strings.NewReader(string(userJSON)))

		user := &dto.CreateUserRequest{
			Username: input.Username,
			Email:    input.Email,
			Password: input.Password,
		}

		s.mockUsecase.EXPECT().Create(gomock.Any(), user).Return(&errorsutil.AppError{Code: errorsutil.ERROR_INTERNAL, Err: fmt.Errorf("test error")})

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("[V1] Create - Duplicate entity", func(t *testing.T) {
		// ctx := context.NewAppContextTest()

		input := dto.CreateUserRequest{
			Username: "test@test.com",
			Email:    "test@test.com",
			Password: "Thisisatestpassword1234!@#$",
		}

		userJSON, _ := json.Marshal(input)

		req, _ := http.NewRequest("POST", v1+create, strings.NewReader(string(userJSON)))

		user := &dto.CreateUserRequest{
			Username: input.Username,
			Email:    input.Email,
			Password: input.Password,
		}

		s.mockUsecase.EXPECT().Create(gomock.Any(), user).Return(&errorsutil.AppError{Code: errorsutil.ERROR_DUPLICATE_ENTITY, Err: fmt.Errorf("test error")})

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("[V1] Create - Invalid password", func(t *testing.T) {
		input := dto.CreateUserRequest{
			Username: "test@test.com",
			Email:    "test@test.com",
			Password: "passwordpassword",
		}

		userJSON, _ := json.Marshal(input)

		req, _ := http.NewRequest("POST", v1+create, strings.NewReader(string(userJSON)))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func Test_Get(t *testing.T) {
	s := newServerTest(t)

	uuid := uuidpkg.NewWithPrefix(uuidutil.PREFIX_USER)
	user := dto.GetUserResponse{
		UUID:     uuid,
		Username: "testusername",
		Email:    "test@test.com",
		Role:     "USER",
	}

	t.Run("[V1] Get user with success", func(t *testing.T) {
		req, _ := http.NewRequest("GET", v1+users+"/"+uuid, nil)

		s.mockUsecase.EXPECT().Get(gomock.Any(), uuid).Return(&user, nil)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("[V1] Get - Failed to validate UUID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", v1+users+"/"+"1", nil)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Get - Failed to find user", func(t *testing.T) {
		req, _ := http.NewRequest("GET", v1+users+"/"+uuid, nil)

		s.mockUsecase.EXPECT().Get(gomock.Any(), uuid).Return(&dto.GetUserResponse{}, &errorsutil.AppError{Code: errorsutil.ERROR_NOT_FOUND, Err: fmt.Errorf("test error")})

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("[V1] Get -  Usecase error", func(t *testing.T) {
		req, _ := http.NewRequest("GET", v1+users+"/"+uuid, nil)

		s.mockUsecase.EXPECT().Get(gomock.Any(), uuid).Return(nil, &errorsutil.AppError{Code: errorsutil.ERROR_INTERNAL, Err: fmt.Errorf("test error")})

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func Test_Update(t *testing.T) {
	s := newServerTest(t)

	uuid := uuidpkg.NewWithPrefix(uuidutil.PREFIX_USER)

	t.Run("[V1] Update with success", func(t *testing.T) {
		user := dto.UpdateUserRequest{
			UUID:     uuid,
			Username: "testusername",
			Email:    "test@test.com",
			Role:     "USER",
			Password: "thisisatestpassword1234!@#$",
		}
		userJSON, _ := json.Marshal(user)

		req, _ := http.NewRequest("PUT", v1+users+"/"+uuid, strings.NewReader(string(userJSON)))

		s.mockUsecase.EXPECT().Update(gomock.Any(), &user).Return(nil)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("[V1] Failed to bind JSON", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", v1+users+"/"+uuid, strings.NewReader(""))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Failed to validate JSON", func(t *testing.T) {
		user := dto.UpdateUserRequest{
			UUID:     uuid,
			Username: "testusername",
			Email:    "",
			Role:     "USER",
			Password: "thisisatestpassword1234!@#$",
		}
		userJSON, _ := json.Marshal(user)

		req, _ := http.NewRequest("PUT", v1+users+"/"+uuid, strings.NewReader(string(userJSON)))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Usecase error", func(t *testing.T) {
		user := dto.UpdateUserRequest{
			UUID:     uuid,
			Username: "testusername",
			Email:    "test@test.com",
			Role:     "USER",
			Password: "thisisatestpassword1234!@#$",
		}
		userJSON, _ := json.Marshal(user)

		req, _ := http.NewRequest("PUT", v1+users+"/"+uuid, strings.NewReader(string(userJSON)))

		s.mockUsecase.EXPECT().Update(gomock.Any(), &user).Return(&errorsutil.AppError{Code: errorsutil.ERROR_INTERNAL, Err: fmt.Errorf("test error")})

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func Test_UpdateEmail(t *testing.T) {
	s := newServerTest(t)

	uuid := uuidpkg.NewWithPrefix(uuidutil.PREFIX_USER)

	t.Run("[V1] Update email with success", func(t *testing.T) {
		user := dto.UpdateEmailRequest{
			Email: "test@test.com",
		}
		userJSON, _ := json.Marshal(user)

		req, _ := http.NewRequest("PATCH", v1+users+"/"+uuid+"/email", strings.NewReader(string(userJSON)))

		s.mockUsecase.EXPECT().UpdateEmail(gomock.Any(), uuid, user.Email).Return(nil)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("[V1] Update - Failed to bind JSON", func(t *testing.T) {
		req, _ := http.NewRequest("PATCH", v1+users+"/"+uuid+"/email", strings.NewReader(""))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Update - Failed to validate JSON", func(t *testing.T) {
		user := dto.UpdateEmailRequest{
			Email: "erthgfderftrfe",
		}
		userJSON, _ := json.Marshal(user)

		req, _ := http.NewRequest("PATCH", v1+users+"/"+uuid+"/email", strings.NewReader(string(userJSON)))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Update - Failed to validate UUID", func(t *testing.T) {
		req, _ := http.NewRequest("PATCH", v1+users+"/"+"notanuuid"+"/email", strings.NewReader(""))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Update - Usecase error", func(t *testing.T) {
		user := dto.UpdateEmailRequest{
			Email: "test@test.com",
		}
		userJSON, _ := json.Marshal(user)

		req, _ := http.NewRequest("PATCH", v1+users+"/"+uuid+"/email", strings.NewReader(string(userJSON)))

		s.mockUsecase.EXPECT().UpdateEmail(gomock.Any(), uuid, user.Email).Return(&errorsutil.AppError{Code: errorsutil.ERROR_INTERNAL, Err: fmt.Errorf("test error")})

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func Test_UpdatePassword(t *testing.T) {
	s := newServerTest(t)

	uuid := uuidpkg.NewWithPrefix(uuidutil.PREFIX_USER)

	t.Run("[V1] Update password with success", func(t *testing.T) {
		user := dto.UpdatePasswordRequest{
			Password: "Thisisatestpassword123?",
		}
		userJSON, _ := json.Marshal(user)

		req, _ := http.NewRequest("PATCH", v1+users+"/"+uuid+"/password", strings.NewReader(string(userJSON)))

		s.mockUsecase.EXPECT().UpdatePassword(gomock.Any(), uuid, user.Password).Return(nil)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("[V1] Update - Failed to bind JSON", func(t *testing.T) {
		req, _ := http.NewRequest("PATCH", v1+users+"/"+uuid+"/password", strings.NewReader(""))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Update - Failed to validate JSON", func(t *testing.T) {
		user := dto.UpdatePasswordRequest{
			Password: "a",
		}
		userJSON, _ := json.Marshal(user)

		req, _ := http.NewRequest("PATCH", v1+users+"/"+uuid+"/password", strings.NewReader(string(userJSON)))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Update - Failed to validate UUID", func(t *testing.T) {
		req, _ := http.NewRequest("PATCH", v1+users+"/"+"notanuuid"+"/password", strings.NewReader(""))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Update - Usecase error", func(t *testing.T) {
		user := dto.UpdatePasswordRequest{
			Password: "Thisisatestpassword123?",
		}
		userJSON, _ := json.Marshal(user)

		req, _ := http.NewRequest("PATCH", v1+users+"/"+uuid+"/password", strings.NewReader(string(userJSON)))

		s.mockUsecase.EXPECT().UpdatePassword(gomock.Any(), uuid, user.Password).Return(&errorsutil.AppError{Code: errorsutil.ERROR_INTERNAL, Err: fmt.Errorf("test error")})

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("[V1] Update - Invalid password", func(t *testing.T) {
		user := dto.UpdatePasswordRequest{
			Password: "passwordpassword",
		}
		userJSON, _ := json.Marshal(user)

		req, _ := http.NewRequest("PATCH", v1+users+"/"+uuid+"/password", strings.NewReader(string(userJSON)))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func Test_Delete(t *testing.T) {
	s := newServerTest(t)

	uuid := uuidpkg.NewWithPrefix(uuidutil.PREFIX_USER)

	t.Run("[V1] Delete with success", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", v1+users+"/"+uuid, nil)

		s.mockUsecase.EXPECT().Delete(gomock.Any(), uuid).Return(nil)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("[V1] Delete - failed to validate UUID", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", v1+users+"/"+"notanuuid", nil)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Delete - Usecase error", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", v1+users+"/"+uuid, nil)

		s.mockUsecase.EXPECT().Delete(gomock.Any(), uuid).Return(&errorsutil.AppError{Code: errorsutil.ERROR_INTERNAL, Err: fmt.Errorf("test error")})

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func Test_GetAllUsers(t *testing.T) {
	s := newServerTest(t)
	uuid := uuidpkg.NewWithPrefix(uuidutil.PREFIX_USER)

	t.Run("[V1] Get all users with success", func(t *testing.T) {
		req, _ := http.NewRequest("GET", v1+users+"/", nil)

		usecaseResp := dto.GetAllUsersResponse{
			Users: []*dto.User{
				{UUID: uuid, Username: "test", Email: "test@test.com", Role: "USER"},
				{UUID: uuid, Username: "test", Email: "test2@test.com", Role: "USER"},
			},
		}

		s.mockUsecase.EXPECT().GetAllUsers(gomock.Any()).Return(&usecaseResp, nil)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("[V1] Get all users - Usecase error", func(t *testing.T) {
		req, _ := http.NewRequest("GET", v1+users+"/", nil)

		s.mockUsecase.EXPECT().GetAllUsers(gomock.Any()).Return(nil, &errorsutil.AppError{Code: errorsutil.ERROR_INTERNAL, Err: fmt.Errorf("test error")})

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
