package usecases

import (
	"fmt"
	"testing"

	"MydroX/anicetus/internal/common/context"
	"MydroX/anicetus/internal/common/errorsutil"
	"MydroX/anicetus/internal/config"
	"MydroX/anicetus/internal/users/dto"
	"MydroX/anicetus/internal/users/mocks"
	"MydroX/anicetus/internal/users/models"

	loggerpkg "MydroX/anicetus/pkg/logger"
	"MydroX/anicetus/pkg/uuid"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var userPrefix = "user"

func createTestUsecase(t *testing.T) (*mocks.MockUsersRepository, UsersUsecases) {
	ctrl := gomock.NewController(t)
	repository := mocks.NewMockUsersRepository(ctrl)

	logger := loggerpkg.New("TEST")
	u := New(logger, repository, &config.Session{})

	return repository, u
}

func Test_Create(t *testing.T) {
	r, u := createTestUsecase(t)
	ctx := context.NewAppContextTest()

	t.Run("[V1] Create user", func(t *testing.T) {
		request := dto.CreateUserRequest{
			Username: "test",
			Email:    "test@test.com",
			Password: "thisisapassword123",
			Role:     "USER",
		}

		r.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil)
		err := u.Create(ctx, &request)

		assert.Nil(t, err)
	})

	t.Run("[V1] Create user, default role", func(t *testing.T) {
		request := dto.CreateUserRequest{
			Username: "test",
			Email:    "test@test.com",
			Password: "thisisapassword123",
		}

		r.EXPECT().CreateUser(gomock.Any(), gomock.Any()).DoAndReturn(func(_ *context.AppContext, user *models.User) error {
			assert.Equal(t, user.Role, "USER")
			return nil
		})

		err := u.Create(ctx, &request)
		assert.Nil(t, err)
	})

	t.Run("[V1] Create user, failed to hash password", func(t *testing.T) {
		ctx := context.NewAppContextTest()

		request := dto.CreateUserRequest{
			Username: "test",
			Email:    "test@test.com",
			Password: "WcYVkLZaCHH5AjzVyyhPaZ0Ny1j8Yqxqu0zYHz8YtvKDzQ7cEx8cXG7VTBq55RmLUFubXPhHgaqwGyQn",
			Role:     "USER",
		}

		err := u.Create(ctx, &request)

		assert.NotNil(t, err)
	})

	t.Run("[V1] Create user repository error", func(t *testing.T) {
		ctx := context.NewAppContextTest()

		request := dto.CreateUserRequest{
			Username: "test",
			Email:    "test@test.com",
			Password: "thisisapassword123",
			Role:     "USER",
		}

		r.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(errorsutil.New(errorsutil.ERROR_INTERNAL, "test error", fmt.Errorf("test error")))
		err := u.Create(ctx, &request)

		assert.NotNil(t, err)
	})
}

func Test_Get(t *testing.T) {
	r, u := createTestUsecase(t)

	userUUID := uuid.NewWithPrefix(userPrefix)

	t.Run("[V1] Get user", func(t *testing.T) {
		ctx := context.NewAppContextTest()

		user := models.User{
			UUID:     userUUID,
			Username: "test",
			Email:    "test@test.com",
			Role:     "USER",
		}

		r.EXPECT().GetUserByUUID(gomock.Any(), userUUID).Return(&user, nil)
		res, err := u.Get(ctx, userUUID)

		assert.Equal(t, res.UUID, userUUID)
		assert.Nil(t, err)
	})

	t.Run("[V1] Get User repository error", func(t *testing.T) {
		ctx := context.NewAppContextTest()

		r.EXPECT().GetUserByUUID(gomock.Any(), userUUID).Return(nil, errorsutil.New(errorsutil.ERROR_INTERNAL, "test error", fmt.Errorf("test error")))
		_, err := u.Get(ctx, userUUID)

		assert.Error(t, err)
	})
}

func Test_Update(t *testing.T) {
	r, u := createTestUsecase(t)
	ctx := context.NewAppContextTest()

	userUUID := uuid.NewWithPrefix(userPrefix)

	t.Run("[V1] Update user", func(t *testing.T) {
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

		r.EXPECT().UpdateUser(gomock.Any(), user).Return(&models.User{}, nil)

		err := u.Update(ctx, &request)

		assert.Nil(t, err)
	})

	t.Run("[V1] Update user repository error", func(t *testing.T) {
		ctx := context.NewAppContextTest()

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

		r.EXPECT().UpdateUser(gomock.Any(), user).Return(nil, errorsutil.New(errorsutil.ERROR_INTERNAL, "test error", fmt.Errorf("test error")))

		err := u.Update(ctx, &request)

		assert.Error(t, err)
	})
}

func Test_UpdatePassword(t *testing.T) {
	r, u := createTestUsecase(t)

	userUUID := uuid.NewWithPrefix(userPrefix)

	t.Run("[V1] Update password", func(t *testing.T) {
		ctx := context.NewAppContextTest()
		password := "passwordtest123!?"

		r.EXPECT().UpdatePassword(gomock.Any(), userUUID, gomock.Any()).DoAndReturn(func(_ *context.AppContext, _ string, p string) error {
			assert.NotEmpty(t, p)
			assert.NotEqual(t, p, password)
			return nil
		})

		err := u.UpdatePassword(ctx, userUUID, password)

		assert.Nil(t, err)
	})

	t.Run("[V1] Update password, failed to hash password", func(t *testing.T) {
		ctx := context.NewAppContextTest()

		password := "WcYVkLZaCHH5AjzVyyhPaZ0Ny1j8Yqxqu0zYHz8YtvKDzQ7cEx8cXG7VTBq55RmLUFubXPhHgaqwGyQn"

		err := u.UpdatePassword(ctx, userUUID, password)

		assert.Error(t, err)
	})

	t.Run("[V1] Update Repository error", func(t *testing.T) {
		ctx := context.NewAppContextTest()
		password := "passwordtest123!?"

		r.EXPECT().UpdatePassword(gomock.Any(), userUUID, gomock.Any()).Return(errorsutil.New(errorsutil.ERROR_INTERNAL, "test error", fmt.Errorf("test error")))

		err := u.UpdatePassword(ctx, userUUID, password)

		assert.Error(t, err)
	})
}

func Test_UpdateEmail(t *testing.T) {
	r, u := createTestUsecase(t)

	userUUID := uuid.NewWithPrefix(userPrefix)
	email := "jeon.soyeon@cube.kr"

	t.Run("[V1] Update email", func(t *testing.T) {
		ctx := context.NewAppContextTest()

		r.EXPECT().UpdateEmail(gomock.Any(), userUUID, email).Return(nil)

		err := u.UpdateEmail(ctx, userUUID, email)

		assert.Nil(t, err)
	})

	t.Run("[V1] Update email repository error", func(t *testing.T) {
		ctx := context.NewAppContextTest()

		r.EXPECT().UpdateEmail(gomock.Any(), userUUID, email).Return(errorsutil.New(errorsutil.ERROR_INTERNAL, "test error", fmt.Errorf("test error")))

		err := u.UpdateEmail(ctx, userUUID, email)

		assert.Error(t, err)
	})
}

func Test_Delete(t *testing.T) {
	r, u := createTestUsecase(t)

	userUUID := uuid.NewWithPrefix(userPrefix)

	t.Run("[V1] Delete user", func(t *testing.T) {
		ctx := context.NewAppContextTest()

		r.EXPECT().DeleteUser(gomock.Any(), userUUID).Return(nil)

		err := u.Delete(ctx, userUUID)

		assert.Nil(t, err)
	})

	t.Run("[V1] Repository error", func(t *testing.T) {
		ctx := context.NewAppContextTest()

		r.EXPECT().DeleteUser(gomock.Any(), userUUID).Return(errorsutil.New(errorsutil.ERROR_INTERNAL, "test error", fmt.Errorf("test error")))

		err := u.Delete(ctx, userUUID)

		assert.Error(t, err)
	})
}

func Test_GetAllUsers(t *testing.T) {
	r, u := createTestUsecase(t)
	userUUID := uuid.NewWithPrefix(userPrefix)

	t.Run("[V1] Get all users", func(t *testing.T) {
		ctx := context.NewAppContextTest()

		users := []*models.User{
			{
				UUID:     userUUID,
				Username: "test",
				Email:    "test@test.com",
				Role:     "USER",
			},
		}

		r.EXPECT().GetAllUsers(gomock.Any()).Return(users, nil)

		res, err := u.GetAllUsers(ctx)
		assert.Nil(t, err)
		assert.Len(t, res.Users, 1)
		assert.Equal(t, res.Users[0].UUID, userUUID)
	})

	t.Run("[V1] Get all users, repository error", func(t *testing.T) {
		ctx := context.NewAppContextTest()

		r.EXPECT().GetAllUsers(gomock.Any()).Return(nil, errorsutil.New(errorsutil.ERROR_INTERNAL, "test error", fmt.Errorf("test error")))

		_, err := u.GetAllUsers(ctx)

		assert.Error(t, err)
	})
}
