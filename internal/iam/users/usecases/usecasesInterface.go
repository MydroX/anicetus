package usecases

import (
	"MydroX/project-v/internal/iam/users/dto"

	"github.com/google/uuid"
)

//go:generate mockgen -destination=../mocks/mock_usecases.go -package=mocks MydroX/project-v/internal/iam/users/usecases UsersUsecases

// Usecases is the interface to all the implemented usecases
type UsersUsecases interface {
	Create(user dto.CreateUserRequest) error
	Get(uuid uuid.UUID) (*dto.GetUserResponse, error)
	Update(user dto.UpdateUserRequest) error
	UpdatePassword(uuid uuid.UUID, password string) error
	UpdateEmail(uuid uuid.UUID, email string) error
	Delete(uuid uuid.UUID) error
	Auth(username, email, password string) (string, error)
}
