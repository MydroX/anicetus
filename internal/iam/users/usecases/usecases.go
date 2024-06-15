// Package usecases is the internal logic. Describes an action that the user wants to perform.
// Also interact with repository and determines how the data has to be transmitted to the external layer.
package usecases

import (
	"MydroX/project-v/internal/iam/users/dto"
	"MydroX/project-v/internal/iam/users/models"
	"MydroX/project-v/internal/iam/users/repository"
	"MydroX/project-v/pkg/logger"
	"MydroX/project-v/pkg/password"
	"fmt"

	"github.com/google/uuid"
)

type usecases struct {
	logger     *logger.Logger
	repository repository.UsersRepository
}

// NewUsecases is creating an interface for all the usecases of the service.
func NewUsecases(l *logger.Logger, r repository.UsersRepository) UsersUsecases {
	return &usecases{
		logger:     l,
		repository: r,
	}
}

func (u *usecases) Create(req dto.CreateUserRequest) error {
	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Role:     req.Role,
	}

	user.Password, _ = password.Hash(req.Password)

	userUUID := uuid.New()
	user.UUID = userUUID

	err := u.repository.CreateUser(&user)

	return err
}

func (u *usecases) Get(uuid uuid.UUID) (*dto.GetUserResponse, error) {
	user, err := u.repository.GetUser(uuid)
	if err != nil {
		return nil, err
	}

	res := dto.GetUserResponse{
		UUID:     user.UUID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	}

	return &res, err
}

func (u *usecases) Update(user dto.UpdateUserRequest) error {
	userModel := models.User{
		UUID:     user.UUID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		Password: user.Password,
	}

	err := u.repository.UpdateUser(&userModel)
	return err
}

func (u *usecases) UpdatePassword(uuid uuid.UUID, newPassword string) error {
	newPasswordCrypted, _ := password.Hash(newPassword)

	err := u.repository.UpdatePassword(uuid, newPasswordCrypted)
	return err
}

func (u *usecases) UpdateEmail(uuid uuid.UUID, email string) error {
	err := u.repository.UpdateEmail(uuid, email)
	return err
}

func (u *usecases) Delete(uuid uuid.UUID) error {
	err := u.repository.DeleteUser(uuid)
	return err
}

func (u *usecases) Auth(username, email, password string) (string, error) {
	return "", fmt.Errorf("not implemented")
}
