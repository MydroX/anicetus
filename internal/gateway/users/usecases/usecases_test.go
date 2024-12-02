package usecases

import (
	"fmt"
	"testing"

	"MydroX/project-v/internal/gateway/users/config"
	"MydroX/project-v/internal/gateway/users/dto"
	"MydroX/project-v/internal/gateway/users/mocks"
	"MydroX/project-v/internal/gateway/users/models"

	apiError "MydroX/project-v/pkg/errors"
	"MydroX/project-v/pkg/logger"
	passwordPkg "MydroX/project-v/pkg/password"
	"MydroX/project-v/pkg/uuid"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var userPrefix = "user"

func createTestUsecase(t *testing.T) (*mocks.MockUsersRepository, UsersUsecases) {
	ctrl := gomock.NewController(t)
	repository := mocks.NewMockUsersRepository(ctrl)

	logger := logger.New("TEST")
	u := NewUsecases(logger, repository, &config.JWT{})

	return repository, u
}

func Test_Create(t *testing.T) {
	r, u := createTestUsecase(t)

	t.Run("[V1] Create user", func(t *testing.T) {
		request := models.User{
			Username: "test",
			Email:    "test@test.com",
			Password: "thisisapassword123",
			Role:     "USER",
		}

		r.EXPECT().CreateUser(gomock.Any()).DoAndReturn(func(user *models.User) error {
			assert.NotEmpty(t, user.UUID)
			assert.Greater(t, len(user.UUID), 36)

			assert.NotEmpty(t, user.Password)
			return nil
		})
		err := u.Create(&request)

		assert.NoError(t, err)
	})

	t.Run("[V1] Repository error", func(t *testing.T) {
		request := models.User{
			Username: "test",
			Email:    "test@test.com",
			Password: "thisisapassword123",
			Role:     "USER",
		}

		r.EXPECT().CreateUser(gomock.Any()).Return(fmt.Errorf("error"))
		err := u.Create(&request)

		assert.Error(t, err)
	})
}

func Test_Get(t *testing.T) {
	r, u := createTestUsecase(t)

	userUUID := uuid.NewWithPrefix(userPrefix)

	t.Run("[V1] Get user", func(t *testing.T) {
		user := models.User{
			UUID:     userUUID,
			Username: "test",
			Email:    "test@test.com",
			Role:     "USER",
		}

		r.EXPECT().GetUser(userUUID).Return(&user, nil)
		res, err := u.Get(userUUID)

		assert.Equal(t, res.UUID, userUUID)
		assert.NoError(t, err)
	})

	t.Run("[V1] Repository error", func(t *testing.T) {
		r.EXPECT().GetUser(userUUID).Return(nil, fmt.Errorf("error"))
		_, err := u.Get(userUUID)

		assert.Error(t, err)
	})
}

func Test_Update(t *testing.T) {
	r, u := createTestUsecase(t)

	userUUID := uuid.NewWithPrefix(userPrefix)

	t.Run("[V1] Update user", func(t *testing.T) {
		user := &models.User{
			UUID:     userUUID,
			Username: "test",
			Email:    "test@test.com",
			Role:     "USER",
		}

		r.EXPECT().UpdateUser(user).Return(nil)

		err := u.Update(user)

		assert.NoError(t, err)
	})

	t.Run("[V1] Repository error", func(t *testing.T) {
		user := &models.User{
			UUID:     userUUID,
			Username: "test",
			Password: "password",
			Email:    "test@test.com",
			Role:     "USER",
		}

		r.EXPECT().UpdateUser(user).Return(fmt.Errorf("error"))

		err := u.Update(user)

		assert.Error(t, err)
	})
}

func Test_UpdatePassword(t *testing.T) {
	r, u := createTestUsecase(t)

	userUUID := uuid.NewWithPrefix(userPrefix)
	password := "passwordtest123!?"

	t.Run("[V1] Update password", func(t *testing.T) {
		r.EXPECT().UpdatePassword(userUUID, gomock.Any()).DoAndReturn(func(uuid string, p string) error {
			assert.NotEmpty(t, p)
			assert.NotEqual(t, p, password)
			return nil
		})

		err := u.UpdatePassword(userUUID, password)

		assert.NoError(t, err)
	})

	t.Run("[V1] Repository error", func(t *testing.T) {
		r.EXPECT().UpdatePassword(userUUID, gomock.Any()).Return(fmt.Errorf("error"))

		err := u.UpdatePassword(userUUID, password)

		assert.Error(t, err)
	})
}

func Test_UpdateEmail(t *testing.T) {
	r, u := createTestUsecase(t)

	userUUID := uuid.NewWithPrefix(userPrefix)
	email := "jeon.soyeon@cube.kr"

	t.Run("[V1] Update email", func(t *testing.T) {
		r.EXPECT().UpdateEmail(userUUID, email).Return(nil)

		err := u.UpdateEmail(userUUID, email)

		assert.NoError(t, err)
	})

	t.Run("[V1] Repository error", func(t *testing.T) {
		r.EXPECT().UpdateEmail(userUUID, email).Return(fmt.Errorf("error"))

		err := u.UpdateEmail(userUUID, email)

		assert.Error(t, err)
	})
}

func Test_Delete(t *testing.T) {
	r, u := createTestUsecase(t)

	userUUID := uuid.NewWithPrefix(userPrefix)

	t.Run("[V1] Delete user", func(t *testing.T) {
		r.EXPECT().DeleteUser(userUUID).Return(nil)

		err := u.Delete(userUUID)

		assert.NoError(t, err)
	})

	t.Run("[V1] Repository error", func(t *testing.T) {
		r.EXPECT().DeleteUser(userUUID).Return(fmt.Errorf("error"))

		err := u.Delete(userUUID)

		assert.Error(t, err)
	})
}

func Test_Login(t *testing.T) {
	r, u := createTestUsecase(t)

	userUUID := uuid.NewWithPrefix(userPrefix)

	t.Run("[V1] Login user with email", func(t *testing.T) {
		p, err := passwordPkg.Hash("thisisapassword123")
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

		r.EXPECT().GetUserByEmail(req.Email).Return(&userFromDB, nil)

		token, err := u.Login("", req.Email, req.Password)

		assert.NotEmpty(t, token)
		assert.NoError(t, err)
	})

	t.Run("[V1] Login user with username", func(t *testing.T) {
		p, err := passwordPkg.Hash("thisisapassword123")
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

		r.EXPECT().GetUserByUsername(req.Username).Return(&userFromDB, nil)

		token, err := u.Login(req.Username, "", req.Password)

		assert.NotEmpty(t, token)
		assert.NoError(t, err)
	})

	t.Run("[V1] Login user with wrong password", func(t *testing.T) {
		_, err := passwordPkg.Hash("thisisapassword123")
		assert.NoError(t, err)
	})

	t.Run("[V1] Login user without email or username", func(t *testing.T) {
		req := dto.LoginRequest{
			Email:    "",
			Username: "",
			Password: "thisisapassword123",
		}

		_, err := u.Login("", req.Email, req.Password)
		assert.Error(t, err)
	})

	t.Run("[V1] Login user, email not found", func(t *testing.T) {
		req := dto.LoginRequest{
			Email:    "test@test.com",
			Password: "thisisapassword123",
		}

		r.EXPECT().GetUserByEmail(req.Email).Return(nil, apiError.ErrNotFound)

		_, err := u.Login("", req.Email, req.Password)

		assert.Equal(t, apiError.ErrNotFound, err)
	})

	t.Run("[V1] Login user, wrong password", func(t *testing.T) {
		req := dto.LoginRequest{
			Email:    "test@test.com",
			Password: "thisisapassword321",
		}

		p, err := passwordPkg.Hash("thisisapassword123")
		assert.NoError(t, err)

		userFromDB := models.User{
			UUID:     userUUID,
			Email:    "test@test.com",
			Password: p,
		}

		r.EXPECT().GetUserByEmail(req.Email).Return(&userFromDB, nil)

		_, err = u.Login("", req.Email, req.Password)

		assert.Error(t, err)
	})
}
