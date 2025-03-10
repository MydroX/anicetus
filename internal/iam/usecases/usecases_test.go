package usecases

import (
	"MydroX/project-v/internal/common/errorscode"
	"MydroX/project-v/internal/config"
	"MydroX/project-v/internal/iam/dto"
	"MydroX/project-v/internal/users/mocks"
	"MydroX/project-v/internal/users/models"
	"MydroX/project-v/pkg/uuid"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	loggerpkg "MydroX/project-v/pkg/logger"
	passwordpkg "MydroX/project-v/pkg/password"
)

var userPrefix = "user"

func createTestUsecase(t *testing.T) (*mocks.MockUsersRepository, IamUsecasesInterface) {
	ctrl := gomock.NewController(t)
	repositoryMock := mocks.NewMockUsersRepository(ctrl)

	logger := loggerpkg.New("TEST")
	u := NewUsecases(logger, repositoryMock, &config.JWT{})

	return repositoryMock, u
}

func Test_Login(t *testing.T) {
	r, u := createTestUsecase(t)

	userUUID := uuid.NewWithPrefix(userPrefix)

	t.Run("[V1] Login user with email", func(t *testing.T) {
		ctx := context.Background()

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

		r.EXPECT().GetUserByEmail(&ctx, req.Email).Return(&userFromDB, nil)

		accessToken, refreshToken, err := u.Login(&ctx, "", req.Email, req.Password)

		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
		assert.NoError(t, err)
	})

	t.Run("[V1] Login user with username", func(t *testing.T) {
		ctx := context.Background()

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

		r.EXPECT().GetUserByUsername(&ctx, req.Username).Return(&userFromDB, nil)

		accessToken, refreshToken, err := u.Login(&ctx, req.Username, "", req.Password)

		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
		assert.NoError(t, err)
	})

	t.Run("[V1] Login user with wrong password", func(t *testing.T) {
		_, err := passwordpkg.Hash("thisisapassword123")
		assert.NoError(t, err)
	})

	t.Run("[V1] Login user without email or username", func(t *testing.T) {
		ctx := context.Background()
		req := dto.LoginRequest{
			Email:    "",
			Username: "",
			Password: "thisisapassword123",
		}

		_, _, err := u.Login(&ctx, req.Username, req.Email, req.Password)
		assert.Error(t, err)
	})

	t.Run("[V1] Login user, email not found", func(t *testing.T) {
		ctx := context.Background()

		req := dto.LoginRequest{
			Email:    "test@test.com",
			Password: "thisisapassword123",
		}

		r.EXPECT().GetUserByEmail(&ctx, req.Email).DoAndReturn(
			func(ctx *context.Context, _ string) (*models.User, error) {
				*ctx = context.WithValue(*ctx, errorscode.CtxErrorCodeKey, errorscode.CODE_ENTITY_NOT_FOUND)
				return nil, fmt.Errorf("user not found")
			})

		_, _, err := u.Login(&ctx, "", req.Email, req.Password)

		errorCode := ctx.Value(errorscode.CtxErrorCodeKey)

		assert.Equal(t, errorscode.CODE_ENTITY_NOT_FOUND, errorCode)
		assert.Error(t, err)
	})

	t.Run("[V1] Login user, username not found", func(t *testing.T) {
		ctx := context.Background()

		req := dto.LoginRequest{
			Username: "test",
			Password: "thisisapassword123",
		}

		r.EXPECT().GetUserByUsername(&ctx, req.Username).DoAndReturn(
			func(ctx *context.Context, _ string) (*models.User, error) {
				*ctx = context.WithValue(*ctx, errorscode.CtxErrorCodeKey, errorscode.CODE_ENTITY_NOT_FOUND)
				return nil, fmt.Errorf("user not found")
			})

		_, _, err := u.Login(&ctx, req.Username, "", req.Password)

		errorCode := ctx.Value(errorscode.CtxErrorCodeKey)

		assert.Equal(t, errorscode.CODE_ENTITY_NOT_FOUND, errorCode)
		assert.Error(t, err)
	})

	t.Run("[V1] Login user, wrong password", func(t *testing.T) {
		ctx := context.Background()

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

		r.EXPECT().GetUserByEmail(&ctx, req.Email).Return(&userFromDB, nil)

		_, _, err = u.Login(&ctx, "", req.Email, req.Password)

		assert.Error(t, err)
	})
}
