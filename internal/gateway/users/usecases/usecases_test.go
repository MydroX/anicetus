package usecases

import (
	"MydroX/project-v/internal/gateway/users/dto"
	"MydroX/project-v/internal/gateway/users/mocks"
	"MydroX/project-v/internal/gateway/users/models"
	"MydroX/project-v/pkg/logger"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func createTestUsecase(t *testing.T) (*mocks.MockUsersRepository, UsersUsecases) {
	ctrl := gomock.NewController(t)
	repository := mocks.NewMockUsersRepository(ctrl)

	logger := logger.New("TEST")
	u := NewUsecases(logger, repository)

	return repository, u
}

func Test_Create(t *testing.T) {
	r, u := createTestUsecase(t)

	t.Run("[V1] Create user", func(t *testing.T) {
		request := dto.CreateUserRequest{
			Username: "test",
			Email:    "test@test.com",
			Password: "thisisapassword123",
			Role:     "USER",
		}

		r.EXPECT().CreateUser(gomock.Any()).Return(nil)
		err := u.Create(request)

		assert.NoError(t, err)
	})

	t.Run("[V1] Repository error", func(t *testing.T) {
		request := dto.CreateUserRequest{
			Username: "test",
			Email:    "test@test.com",
			Password: "thisisapassword123",
			Role:     "USER",
		}

		r.EXPECT().CreateUser(gomock.Any()).Return(fmt.Errorf("error"))
		err := u.Create(request)

		assert.Error(t, err)
	})
}

func Test_Get(t *testing.T) {
	r, u := createTestUsecase(t)

	userUUID := uuid.New()

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

	userUUID := uuid.New()

	t.Run("[V1] Update user", func(t *testing.T) {
		userRequest := dto.UpdateUserRequest{
			UUID:     userUUID,
			Username: "test",
			Email:    "test@test.com",
			Role:     "USER",
		}
		userModel := &models.User{
			UUID:     userUUID,
			Username: "test",
			Email:    "test@test.com",
			Role:     "USER",
		}

		r.EXPECT().UpdateUser(userModel).Return(nil)

		err := u.Update(userRequest)

		assert.NoError(t, err)
	})

	t.Run("[V1] Repository error", func(t *testing.T) {
		userRequest := dto.UpdateUserRequest{
			UUID:     userUUID,
			Username: "test",
			Email:    "test@test.com",
			Role:     "USER",
		}
		userModel := &models.User{
			UUID:     userUUID,
			Username: "test",
			Email:    "test@test.com",
			Role:     "USER",
		}

		r.EXPECT().UpdateUser(userModel).Return(fmt.Errorf("error"))

		err := u.Update(userRequest)

		assert.Error(t, err)
	})
}

func Test_UpdatePassword(t *testing.T) {
	r, u := createTestUsecase(t)

	userUUID := uuid.New()
	password := "passwordtest123!?"

	t.Run("[V1] Update password", func(t *testing.T) {
		r.EXPECT().UpdatePassword(userUUID, gomock.Any()).Return(nil)

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

	userUUID := uuid.New()
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

	userUUID := uuid.New()

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
