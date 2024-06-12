package iam

import (
	"MydroX/project-v/internal/iam/mocks"
	"MydroX/project-v/pkg/logger"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const (
	v1 = "/api/v1"

	create = "/register"
)

type testServer struct {
	router      *gin.Engine
	mockUsecase *mocks.MockUsecasesInterface
}

func newServerTest(t *testing.T) testServer {
	logger := logger.New("TEST")
	validator := validator.New()

	ctrl := gomock.NewController(t)
	usecasesMock := mocks.NewMockUsecasesInterface(ctrl)
	controller := NewController(logger, validator, usecasesMock)

	return testServer{
		router:      Router(logger, validator, nil, controller),
		mockUsecase: usecasesMock,
	}
}

func Test_Create(t *testing.T) {
	s := newServerTest(t)

	t.Run("Create with success V1", func(t *testing.T) {
		input := CreateUserRequest{
			Username: "test",
			Email:    "test@test.com",
			Password: "thisisatestpassword1234!@#$",
			Role:     "USER",
		}
		userJSON, _ := json.Marshal(input)

		req, _ := http.NewRequest("POST", v1+create, strings.NewReader(string(userJSON)))

		s.mockUsecase.EXPECT().Create(gomock.Any(), input.Username, gomock.Any(), input.Email, input.Role).Return(nil)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Failed to bind JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", v1+create, strings.NewReader(""))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Failed to validate JSON", func(t *testing.T) {
		input := CreateUserRequest{
			Username: "test",
			Email:    "",
			Password: "thisisatestpassword1234!@#$",
			Role:     "USER",
		}
		userJSON, _ := json.Marshal(input)

		req, _ := http.NewRequest("POST", v1+create, strings.NewReader(string(userJSON)))

		s.mockUsecase.EXPECT().Create(gomock.Any(), input.Username, gomock.Any(), input.Email, input.Role).Return(nil).AnyTimes()

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Usecase error V1", func(t *testing.T) {
		input := CreateUserRequest{
			Username: "test",
			Email:    "test@test.com",
			Password: "thisisatestpassword1234!@#$",
			Role:     "USER",
		}
		userJSON, _ := json.Marshal(input)

		req, _ := http.NewRequest("POST", v1+create, strings.NewReader(string(userJSON)))

		s.mockUsecase.EXPECT().Create(gomock.Any(), input.Username, gomock.Any(), input.Email, input.Role).Return(fmt.Errorf("test error"))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
