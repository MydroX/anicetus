package iam

import (
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

	"MydroX/project-v/internal/common/errorscode"
	"MydroX/project-v/internal/common/response"
	"MydroX/project-v/internal/config"
	"MydroX/project-v/internal/iam/dto"
	"MydroX/project-v/internal/iam/mocks"
	loggerpkg "MydroX/project-v/pkg/logger"
)

const v1 = "/api/v1"

type testServer struct {
	router      *gin.Engine
	mockUsecase *mocks.MockIamUsecasesInterface
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
	usecasesMock := mocks.NewMockIamUsecasesInterface(ctrl)

	c := New(logger, usecasesMock, &config.Config{})
	router := testRouter(logger, c)

	return testServer{
		router:      router,
		mockUsecase: usecasesMock,
	}
}

func Test_Login(t *testing.T) {
	s := newServerTest(t)

	t.Run("[V1] Login with success", func(t *testing.T) {
		ctx := context.Background()

		input := dto.LoginRequest{
			Username: "test",
			Password: "thisisatest123",
		}

		userJSON, _ := json.Marshal(input)

		req, _ := http.NewRequest("POST", v1+"/login", strings.NewReader(string(userJSON)))

		s.mockUsecase.EXPECT().Login(&ctx, input.Username, input.Email, input.Password).Return("access", "refresh", nil)

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

		userJSON, _ := json.Marshal(input)

		req, _ := http.NewRequest("POST", v1+"/login", strings.NewReader(string(userJSON)))

		s.mockUsecase.EXPECT().Login(&ctx, input.Username, input.Email, input.Password).DoAndReturn(
			func(ctx *context.Context, _ string, _ string, _ string) (string, string, error) {
				*ctx = context.WithValue(*ctx, errorscode.CtxErrorCodeKey, errorscode.CODE_ENTITY_NOT_FOUND)
				return "", "", fmt.Errorf("user not found")
			})

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		resp := response.ErrorResponse{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, errorscode.CODE_ENTITY_NOT_FOUND, resp.Code)
	})

	t.Run("[V1] Login - Usecase error", func(t *testing.T) {
		ctx := context.Background()

		input := dto.LoginRequest{
			Username: "test",
			Password: "thisisatest123",
		}

		userJSON, _ := json.Marshal(input)

		req, _ := http.NewRequest("POST", v1+"/login", strings.NewReader(string(userJSON)))

		s.mockUsecase.EXPECT().Login(&ctx, input.Username, input.Email, input.Password).Return("", "", fmt.Errorf("test error"))

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
