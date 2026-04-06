package usecases

import (
	"errors"
	"testing"

	"context"
	"MydroX/anicetus/internal/common/errorsutil"
	"MydroX/anicetus/internal/common/jwt"
	"MydroX/anicetus/internal/config"
	"MydroX/anicetus/internal/iam/dto"
	iammocks "MydroX/anicetus/internal/iam/mocks"
	usersmodels "MydroX/anicetus/internal/users/models"
	usersmocks "MydroX/anicetus/internal/users/mocks"
	passwordpkg "MydroX/anicetus/pkg/password"
	"github.com/google/uuid"
	valkeymock "github.com/valkey-io/valkey-go/mock"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func createTestUsecase(t *testing.T) (
	*usersmocks.MockUsersRepository,
	*iammocks.MockIamStore,
	*iammocks.MockAudienceStore,
	IamUsecasesService,
) {
	t.Helper()

	ctrl := gomock.NewController(t)
	usersRepositoryMock := usersmocks.NewMockUsersRepository(ctrl)
	iamRepositoryMock := iammocks.NewMockIamStore(ctrl)
	audienceStoreMock := iammocks.NewMockAudienceStore(ctrl)

	cfg := &config.Config{
		Session: config.Session{
			Hash: config.Hash{
				SaltLength:  16,
				Iterations:  4,
				Memory:      64 * 1024,
				Parallelism: 4,
				KeyLength:   32,
			},
		},
		JWT: config.JWT{
			Issuer:      "test-issuer",
			SkewSeconds: 60,
			AccessToken: config.AccessToken{
				Secret:     "test-access-secret-key-long-enough-for-jwt-signing",
				Expiration: 3600,
			},
			RefreshToken: config.RefreshToken{
				Secret:     "test-refresh-secret-key-long-enough-for-jwt-signin",
				Expiration: 7200,
			},
		},
	}

	tokenConfig := jwt.TokenConfig{
		AccessTokenSecret:    cfg.JWT.AccessToken.Secret,
		RefreshTokenSecret:   cfg.JWT.RefreshToken.Secret,
		ExpectedIssuer:       cfg.JWT.Issuer,
		ExpectedAudiences:    []string{cfg.JWT.Issuer},
		ClockSkewSeconds:     cfg.JWT.SkewSeconds,
		AccessTokenDuration:  cfg.JWT.AccessToken.Expiration,
		RefreshTokenDuration: cfg.JWT.RefreshToken.Expiration,
	}

	jwtService := jwt.NewJWTService(tokenConfig)

	valkeyClient := valkeymock.NewClient(ctrl)
	valkeyClient.EXPECT().Do(gomock.Any(), gomock.Any()).Return(valkeymock.Result(valkeymock.ValkeyNil())).AnyTimes()

	audienceManager := NewAudienceManager(nil, audienceStoreMock, valkeyClient)

	u := New(nil, usersRepositoryMock, iamRepositoryMock, cfg, jwtService, audienceStoreMock, audienceManager)

	return usersRepositoryMock, iamRepositoryMock, audienceStoreMock, u
}

func TestLogin(t *testing.T) {
	usersRepository, iamRepository, audienceStore, u := createTestUsecase(t)

	userUUID := uuid.New().String()

	t.Run("Login user with email", func(t *testing.T) {
		ctx := context.Background()

		p, err := passwordpkg.Hash("thisisapassword123")
		assert.NoError(t, err)

		userFromDB := usersmodels.User{
			UUID:     userUUID,
			Email:    "test@test.com",
			Password: p,
		}

		usersRepository.EXPECT().GetUserByEmail(gomock.Any(), "test@test.com").Return(&userFromDB, nil)
		audienceStore.EXPECT().GetUserAudiences(gomock.Any(), gomock.Any()).Return([]string{"test-audience"}, nil).AnyTimes()
		iamRepository.EXPECT().SaveSession(gomock.Any(), gomock.Any()).Return(nil)

		req := dto.LoginRequest{
			Email:    "test@test.com",
			Password: "thisisapassword123",
			Session: dto.Session{
				OS:             "Mac OS X",
				OSVersion:      "10.15",
				IPv4Address:    "192.168.1.1",
				Browser:        "Mozilla",
				BrowserVersion: "138.0a1",
			},
		}

		accessToken, refreshToken, apiErr := u.Login(ctx, &req)

		assert.NoError(t, apiErr)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
	})

	t.Run("Login user with username", func(t *testing.T) {
		ctx := context.Background()

		p, err := passwordpkg.Hash("thisisapassword123")
		assert.NoError(t, err)

		userFromDB := usersmodels.User{
			UUID:     userUUID,
			Email:    "test@test.com",
			Username: "testuser",
			Password: p,
		}

		usersRepository.EXPECT().GetUserByUsername(gomock.Any(), "testuser").Return(&userFromDB, nil)
		audienceStore.EXPECT().GetUserAudiences(gomock.Any(), gomock.Any()).Return([]string{"test-audience"}, nil).AnyTimes()
		iamRepository.EXPECT().SaveSession(gomock.Any(), gomock.Any()).Return(nil)

		req := dto.LoginRequest{
			Username: "testuser",
			Password: "thisisapassword123",
			Session: dto.Session{
				OS:             "Mac OS X",
				OSVersion:      "10.15",
				IPv4Address:    "192.168.1.1",
				Browser:        "Mozilla",
				BrowserVersion: "138.0a1",
			},
		}

		accessToken, refreshToken, apiErr := u.Login(ctx, &req)

		assert.NoError(t, apiErr)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
	})

	t.Run("Login without email or username", func(t *testing.T) {
		ctx := context.Background()

		req := dto.LoginRequest{
			Password: "thisisapassword123",
			Session: dto.Session{
				OS:             "Mac OS X",
				OSVersion:      "10.15",
				IPv4Address:    "192.168.1.1",
				Browser:        "Mozilla",
				BrowserVersion: "138.0a1",
			},
		}

		_, _, err := u.Login(ctx, &req)
		assert.Error(t, err)
	})

	t.Run("Login email not found", func(t *testing.T) {
		ctx := context.Background()

		usersRepository.EXPECT().GetUserByEmail(gomock.Any(), "notfound@test.com").Return(nil, errorsutil.New(errorsutil.ErrorNotFound, "user not found", errors.New("user not found")))

		req := dto.LoginRequest{
			Email:    "notfound@test.com",
			Password: "thisisapassword123",
			Session: dto.Session{
				OS:             "Mac OS X",
				OSVersion:      "10.15",
				IPv4Address:    "192.168.1.1",
				Browser:        "Mozilla",
				BrowserVersion: "138.0a1",
			},
		}

		_, _, err := u.Login(ctx, &req)
		assert.Error(t, err)
	})

	t.Run("Login wrong password", func(t *testing.T) {
		ctx := context.Background()

		p, err := passwordpkg.Hash("thisisapassword123")
		assert.NoError(t, err)

		userFromDB := usersmodels.User{
			UUID:     userUUID,
			Email:    "test@test.com",
			Password: p,
		}

		usersRepository.EXPECT().GetUserByEmail(gomock.Any(), "test@test.com").Return(&userFromDB, nil)

		req := dto.LoginRequest{
			Email:    "test@test.com",
			Password: "wrongpassword321",
			Session: dto.Session{
				OS:             "Mac OS X",
				OSVersion:      "10.15",
				IPv4Address:    "192.168.1.1",
				Browser:        "Mozilla",
				BrowserVersion: "138.0a1",
			},
		}

		_, _, apiErr := u.Login(ctx, &req)
		assert.Error(t, apiErr)
	})
}
