package usecases

import (
	"MydroX/anicetus/internal/common/context"
	"MydroX/anicetus/internal/common/errorsutil"
	"MydroX/anicetus/internal/common/uuidutil"
	"MydroX/anicetus/internal/config"
	"MydroX/anicetus/internal/iam/dto"
	iammocks "MydroX/anicetus/internal/iam/mocks"
	usersmocks "MydroX/anicetus/internal/users/mocks"
	"MydroX/anicetus/internal/users/models"
	"MydroX/anicetus/pkg/logger"
	"MydroX/anicetus/pkg/uuid"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	passwordpkg "MydroX/anicetus/pkg/password"
)

func createTestUsecase(t *testing.T) (*usersmocks.MockUsersRepository, *iammocks.MockIamRepository, IamUsecasesService) {
	ctrl := gomock.NewController(t)
	usersRepositoryMock := usersmocks.NewMockUsersRepository(ctrl)
	iamRepositoryMock := iammocks.NewMockIamRepository(ctrl)

	l, err := logger.New("TEST")
	if err != nil {
		panic(err)
	}
	u := New(l, usersRepositoryMock, iamRepositoryMock, &config.Session{
		HashConfig: config.HashConfig{
			SaltLength:  16,
			Iterations:  4,
			Memory:      64 * 1024,
			Parallelism: 4,
			KeyLength:   32,
		},
		AccessToken: config.AccessToken{
			Expiration: 3600,
			Secret:     "testsecret",
		},
		RefreshToken: config.RefreshToken{
			Expiration: 7200,
			Secret:     "testsecret",
		},
	})

	return usersRepositoryMock, iamRepositoryMock, u
}

func Test_Login(t *testing.T) {
	usersRepository, iamRepository, u := createTestUsecase(t)

	userUUID := uuid.NewWithPrefix(uuidutil.PREFIX_USER)

	t.Run("[V1] Login user with email", func(t *testing.T) {
		ctx := context.NewAppContextTest()

		p, err := passwordpkg.Hash("thisisapassword123")
		assert.NoError(t, err)

		userFromDB := models.User{
			UUID:     userUUID,
			Email:    "test@test.com",
			Password: p,
		}

		req := dto.LoginRequest{
			Email:    "test@test.com",
			Password: "thisisapassword123",
		}

		usersRepository.EXPECT().GetUserByEmail(gomock.Any(), req.Email).Return(&userFromDB, nil)
		iamRepository.EXPECT().SaveSession(gomock.Any(), gomock.Any()).Return(nil)

		accessToken, refreshToken, apiErr := u.Login(ctx, &req)

		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
		assert.Nil(t, apiErr)
	})

	t.Run("[V1] Login user with username", func(t *testing.T) {
		ctx := context.NewAppContextTest()

		p, err := passwordpkg.Hash("thisisapassword123")
		assert.NoError(t, err)

		userFromDB := models.User{
			UUID:     userUUID,
			Email:    "test@test.com",
			Username: "test",
			Password: p,
		}

		req := dto.LoginRequest{
			Username: "test",
			Password: "thisisapassword123",
		}

		usersRepository.EXPECT().GetUserByUsername(gomock.Any(), req.Username).Return(&userFromDB, nil)
		iamRepository.EXPECT().SaveSession(gomock.Any(), gomock.Any()).Return(nil)

		accessToken, refreshToken, apiErr := u.Login(ctx, &req)

		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
		assert.Nil(t, apiErr)
	})

	t.Run("[V1] Login user with wrong password", func(t *testing.T) {
		_, err := passwordpkg.Hash("thisisapassword123")
		assert.NoError(t, err)
	})

	t.Run("[V1] Login user without email or username", func(t *testing.T) {
		ctx := context.NewAppContextTest()
		req := dto.LoginRequest{
			Email:    "",
			Username: "",
			Password: "thisisapassword123",
		}

		_, _, err := u.Login(ctx, &req)
		assert.Error(t, err)
	})

	t.Run("[V1] Login user, email not found", func(t *testing.T) {
		ctx := context.NewAppContextTest()

		req := dto.LoginRequest{
			Email:    "test@test.com",
			Password: "thisisapassword123",
		}

		usersRepository.EXPECT().GetUserByEmail(gomock.Any(), req.Email).Return(nil, errorsutil.New(errorsutil.ERROR_NOT_FOUND, "user not found", errors.New("user not found")))

		_, _, err := u.Login(ctx, &req)

		assert.Error(t, err)
	})

	t.Run("[V1] Login user, username not found", func(t *testing.T) {
		ctx := context.NewAppContextTest()

		req := dto.LoginRequest{
			Username: "test",
			Password: "thisisapassword123",
		}

		usersRepository.EXPECT().GetUserByUsername(gomock.Any(), req.Username).Return(nil, errorsutil.New(errorsutil.ERROR_NOT_FOUND, "user not found", errors.New("user not found")))
		_, _, err := u.Login(ctx, &req)

		assert.Error(t, err)
	})

	t.Run("[V1] Login user, wrong password", func(t *testing.T) {
		ctx := context.NewAppContextTest()

		req := dto.LoginRequest{
			Email:    "test@test.com",
			Password: "thisisapassword321",
		}

		p, err := passwordpkg.Hash("thisisapassword123")
		assert.NoError(t, err)

		userFromDB := models.User{
			UUID:     userUUID,
			Email:    "test@test.com",
			Password: p,
		}

		usersRepository.EXPECT().GetUserByEmail(gomock.Any(), req.Email).Return(&userFromDB, nil)

		_, _, apiErr := u.Login(ctx, &req)

		assert.Error(t, apiErr)
	})
}
