package usecases

import (
	"context"
	"fmt"
	"testing"

	"MydroX/project-v/internal/common/errorscode"
	"MydroX/project-v/internal/config"
	"MydroX/project-v/internal/users/dto"
	"MydroX/project-v/internal/users/mocks"
	"MydroX/project-v/internal/users/models"

	loggerpkg "MydroX/project-v/pkg/logger"
	passwordpkg "MydroX/project-v/pkg/password"
	"MydroX/project-v/pkg/uuid"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var userPrefix = "user"

func createTestUsecase(t *testing.T) (*mocks.MockUsersRepository, UsersUsecases) {
	ctrl := gomock.NewController(t)
	repository := mocks.NewMockUsersRepository(ctrl)

	logger := loggerpkg.New("TEST")
	u := NewUsecases(logger, repository, &config.JWT{})

	return repository, u
}

func Test_Create(t *testing.T) {
	r, u := createTestUsecase(t)

	t.Run("[V1] Create user", func(t *testing.T) {
		ctx := context.Background()

		request := dto.CreateUserRequest{
			Username: "test",
			Email:    "test@test.com",
			Password: "thisisapassword123",
			Role:     "USER",
		}

		r.EXPECT().CreateUser(&ctx, gomock.Any()).DoAndReturn(func(_ *context.Context, user *models.User) error {
			assert.NotEmpty(t, user.UUID)
			assert.Greater(t, len(user.UUID), 36)

			assert.NotEmpty(t, user.Password)
			return nil
		})
		err := u.Create(&ctx, &request)

		assert.NoError(t, err)
	})

	t.Run("[V1] Create user, default role", func(t *testing.T) {
		ctx := context.Background()

		request := dto.CreateUserRequest{
			Username: "test",
			Email:    "test@test.com",
			Password: "thisisapassword123",
		}

		r.EXPECT().CreateUser(&ctx, gomock.Any()).DoAndReturn(func(_ *context.Context, user *models.User) error {
			assert.Equal(t, user.Role, "USER")
			return nil
		})

		err := u.Create(&ctx, &request)
		assert.NoError(t, err)
	})

	t.Run("[V1] Create user, failed to hash password", func(t *testing.T) {
		ctx := context.Background()

		request := dto.CreateUserRequest{
			Username: "test",
			Email:    "test@test.com",
			Password: "WcYVkLZaCHH5AjzVyyhPaZ0Ny1j8Yqxqu0zYHz8YtvKDzQ7cEx8cXG7VTBq55RmLUFubXPhHgaqwGyQn",
			Role:     "USER",
		}

		err := u.Create(&ctx, &request)

		assert.Error(t, err)
		assert.Equal(t, errorscode.CODE_FAILED_TO_HASH_PASSWORD, ctx.Value(errorscode.CtxErrorCodeKey))
	})

	t.Run("[V1] Create user repository error", func(t *testing.T) {
		ctx := context.Background()

		request := dto.CreateUserRequest{
			Username: "test",
			Email:    "test@test.com",
			Password: "thisisapassword123",
			Role:     "USER",
		}

		r.EXPECT().CreateUser(&ctx, gomock.Any()).Return(fmt.Errorf("error"))
		err := u.Create(&ctx, &request)

		assert.Error(t, err)
	})
}

func Test_Get(t *testing.T) {
	r, u := createTestUsecase(t)

	userUUID := uuid.NewWithPrefix(userPrefix)

	t.Run("[V1] Get user", func(t *testing.T) {
		ctx := context.Background()

		user := models.User{
			UUID:     userUUID,
			Username: "test",
			Email:    "test@test.com",
			Role:     "USER",
		}

		r.EXPECT().GetUserByUUID(&ctx, userUUID).Return(&user, nil)
		res, err := u.Get(&ctx, userUUID)

		assert.Equal(t, res.UUID, userUUID)
		assert.NoError(t, err)
	})

	t.Run("[V1] Get User repository error", func(t *testing.T) {
		ctx := context.Background()

		r.EXPECT().GetUserByUUID(&ctx, userUUID).Return(nil, fmt.Errorf("error"))
		_, err := u.Get(&ctx, userUUID)

		assert.Error(t, err)
	})
}

func Test_Update(t *testing.T) {
	r, u := createTestUsecase(t)

	userUUID := uuid.NewWithPrefix(userPrefix)

	t.Run("[V1] Update user", func(t *testing.T) {
		ctx := context.Background()

		request := dto.UpdateUserRequest{
			UUID:     userUUID,
			Username: "test",
			Email:    "test@test.com",
		}

		user := &models.User{
			UUID:     userUUID,
			Username: "test",
			Email:    "test@test.com",
		}

		r.EXPECT().UpdateUser(&ctx, user).Return(&models.User{}, nil)

		err := u.Update(&ctx, &request)

		assert.NoError(t, err)
	})

	t.Run("[V1] Update user repository error", func(t *testing.T) {
		ctx := context.Background()

		request := dto.UpdateUserRequest{
			UUID:     userUUID,
			Username: "test",
			Email:    "test@test.com",
		}

		user := &models.User{
			UUID:     userUUID,
			Username: "test",
			Email:    "test@test.com",
		}

		r.EXPECT().UpdateUser(&ctx, user).Return(nil, fmt.Errorf("error"))

		err := u.Update(&ctx, &request)

		assert.Error(t, err)
	})
}

func Test_UpdatePassword(t *testing.T) {
	r, u := createTestUsecase(t)

	userUUID := uuid.NewWithPrefix(userPrefix)

	t.Run("[V1] Update password", func(t *testing.T) {
		ctx := context.Background()
		password := "passwordtest123!?"

		r.EXPECT().UpdatePassword(&ctx, userUUID, gomock.Any()).DoAndReturn(func(_ *context.Context, _ string, p string) error {
			assert.NotEmpty(t, p)
			assert.NotEqual(t, p, password)
			return nil
		})

		err := u.UpdatePassword(&ctx, userUUID, password)

		assert.NoError(t, err)
	})

	t.Run("[V1] Update password, failed to hash password", func(t *testing.T) {
		ctx := context.Background()

		password := "WcYVkLZaCHH5AjzVyyhPaZ0Ny1j8Yqxqu0zYHz8YtvKDzQ7cEx8cXG7VTBq55RmLUFubXPhHgaqwGyQn"

		err := u.UpdatePassword(&ctx, userUUID, password)

		assert.Error(t, err)
		assert.Equal(t, errorscode.CODE_FAILED_TO_HASH_PASSWORD, ctx.Value(errorscode.CtxErrorCodeKey))
	})

	t.Run("[V1] Update Repository error", func(t *testing.T) {
		ctx := context.Background()
		password := "passwordtest123!?"

		r.EXPECT().UpdatePassword(&ctx, userUUID, gomock.Any()).Return(fmt.Errorf("error"))

		err := u.UpdatePassword(&ctx, userUUID, password)

		assert.Error(t, err)
	})
}

func Test_UpdateEmail(t *testing.T) {
	r, u := createTestUsecase(t)

	userUUID := uuid.NewWithPrefix(userPrefix)
	email := "jeon.soyeon@cube.kr"

	t.Run("[V1] Update email", func(t *testing.T) {
		ctx := context.Background()

		r.EXPECT().UpdateEmail(&ctx, userUUID, email).Return(nil)

		err := u.UpdateEmail(&ctx, userUUID, email)

		assert.NoError(t, err)
	})

	t.Run("[V1] Update email repository error", func(t *testing.T) {
		ctx := context.Background()

		r.EXPECT().UpdateEmail(&ctx, userUUID, email).Return(fmt.Errorf("error"))

		err := u.UpdateEmail(&ctx, userUUID, email)

		assert.Error(t, err)
	})
}

func Test_Delete(t *testing.T) {
	r, u := createTestUsecase(t)

	userUUID := uuid.NewWithPrefix(userPrefix)

	t.Run("[V1] Delete user", func(t *testing.T) {
		ctx := context.Background()

		r.EXPECT().DeleteUser(&ctx, userUUID).Return(nil)

		err := u.Delete(&ctx, userUUID)

		assert.NoError(t, err)
	})

	t.Run("[V1] Repository error", func(t *testing.T) {
		ctx := context.Background()

		r.EXPECT().DeleteUser(&ctx, userUUID).Return(fmt.Errorf("error"))

		err := u.Delete(&ctx, userUUID)

		assert.Error(t, err)
	})
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

		token, err := u.Login(&ctx, "", req.Email, req.Password)

		assert.NotEmpty(t, token)
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

		token, err := u.Login(&ctx, req.Username, "", req.Password)

		assert.NotEmpty(t, token)
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

		_, err := u.Login(&ctx, req.Username, req.Email, req.Password)
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

		_, err := u.Login(&ctx, "", req.Email, req.Password)

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

		_, err := u.Login(&ctx, req.Username, "", req.Password)

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

		_, err = u.Login(&ctx, "", req.Email, req.Password)

		assert.Error(t, err)
	})
}

func Test_GetAllUsers(t *testing.T) {
	r, u := createTestUsecase(t)
	userUUID := uuid.NewWithPrefix(userPrefix)

	t.Run("[V1] Get all users", func(t *testing.T) {
		ctx := context.Background()

		users := []*models.User{
			{
				UUID:     userUUID,
				Username: "test",
				Email:    "test@test.com",
				Role:     "USER",
			},
		}

		r.EXPECT().GetAllUsers(&ctx).Return(users, nil)

		res, err := u.GetAllUsers(&ctx)
		assert.NoError(t, err)
		assert.Len(t, res.Users, 1)
		assert.Equal(t, res.Users[0].UUID, userUUID)
	})

	t.Run("[V1] Get all users, repository error", func(t *testing.T) {
		ctx := context.Background()

		r.EXPECT().GetAllUsers(&ctx).Return(nil, fmt.Errorf("error"))

		_, err := u.GetAllUsers(&ctx)

		assert.Error(t, err)
	})
}
