package iam

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"MydroX/anicetus/pkg/errs"
	"MydroX/anicetus/internal/config"
	"MydroX/anicetus/internal/iam/dto"
	"MydroX/anicetus/internal/iam/mocks"
	"MydroX/anicetus/pkg/logger"
)

const v1 = "/api/v1"

type testServer struct {
	logger      *zap.SugaredLogger
	router      *gin.Engine
	mockUsecase *mocks.MockIamUsecasesService
}

func testRouter(controller ControllerInterface) *gin.Engine {
	router := gin.Default()

	err := router.SetTrustedProxies(nil)
	if err != nil {
		panic(err)
	}

	api := router.Group("api")
	v1 := api.Group("/v1")

	Router(v1, controller)

	return router
}

func newServerTest(t *testing.T) testServer {
	l, err := logger.New("TEST")
	if err != nil {
		panic(err)
	}

	ctrl := gomock.NewController(t)
	usecasesMock := mocks.NewMockIamUsecasesService(ctrl)

	c := New(l, usecasesMock, &config.Config{})
	router := testRouter(c)

	return testServer{
		logger:      l,
		router:      router,
		mockUsecase: usecasesMock,
	}
}

// Helper function to make login requests
func makeLoginRequest(router *gin.Engine, body string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("POST", v1+"/login", strings.NewReader(body))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// Helper function to create a valid login request
func createValidLoginRequest() dto.LoginRequest {
	return dto.LoginRequest{
		Username: "test",
		Password: "thisisatest123",
		Session: dto.Session{
			IPv4Address:    "0.0.0.0",
			OS:             "linux",
			OSVersion:      "1.0",
			Browser:        "chrome",
			BrowserVersion: "1.0",
		},
	}
}

func Test_Login(t *testing.T) {
	s := newServerTest(t)

	t.Run("[V1] Login with success", func(t *testing.T) {
		input := createValidLoginRequest()
		userJSON, _ := json.Marshal(input)

		s.mockUsecase.EXPECT().Login(gomock.Any(), &input).Return("access", "refresh", nil)

		w := makeLoginRequest(s.router, string(userJSON))

		var resp dto.LoginResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, resp.AccessToken)
		assert.NotEmpty(t, resp.RefreshToken)
	})

	t.Run("[V1] Login - Failed to bind JSON", func(t *testing.T) {
		w := makeLoginRequest(s.router, "")
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Login - Failed to validate JSON - short username", func(t *testing.T) {
		input := dto.LoginRequest{
			Username: "tes", // Too short
			Password: "thisisatest123",
		}
		userJSON, _ := json.Marshal(input)

		w := makeLoginRequest(s.router, string(userJSON))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Login - No username in body", func(t *testing.T) {
		input := dto.LoginRequest{
			Password: "thisisatest123",
		}
		userJSON, _ := json.Marshal(input)

		w := makeLoginRequest(s.router, string(userJSON))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("[V1] Login - Failed to find user", func(t *testing.T) {
		input := createValidLoginRequest()
		userJSON, _ := json.Marshal(input)

		userNotFoundErr := errs.New(errs.ErrorNotFound, "user not found", nil)
		s.mockUsecase.EXPECT().Login(gomock.Any(), &input).Return("", "", userNotFoundErr)

		w := makeLoginRequest(s.router, string(userJSON))

		var resp errs.AppError
		_ = json.Unmarshal(w.Body.Bytes(), &resp)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, errs.ErrorNotFound, resp.Code)
	})

	t.Run("[V1] Login - Usecase error", func(t *testing.T) {
		input := createValidLoginRequest()
		userJSON, _ := json.Marshal(input)

		internalErr := errs.New(errs.ErrorInternal, "internal error", nil)
		s.mockUsecase.EXPECT().Login(gomock.Any(), &input).Return("", "", internalErr)

		w := makeLoginRequest(s.router, string(userJSON))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("[V1] Login - No email and username provided - Error", func(t *testing.T) {
		input := dto.LoginRequest{
			Password: "thisisatest123",
			Session: dto.Session{
				IPv4Address:    "0.0.0.0",
				OS:             "linux",
				OSVersion:      "1.0",
				Browser:        "chrome",
				BrowserVersion: "1.0",
			},
		}
		userJSON, _ := json.Marshal(input)

		w := makeLoginRequest(s.router, string(userJSON))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
