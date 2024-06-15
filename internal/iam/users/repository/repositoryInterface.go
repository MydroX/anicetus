package repository

import (
	"MydroX/project-v/internal/iam/users/models"

	"github.com/google/uuid"
)

//go:generate mockgen -destination=../mocks/mock_repository.go -imports=models=MydroX/project-v/internal/users/models -package=mocks MydroX/project-v/internal/iam/users/repository UsersRepository

// Repository is the interface to all the implemented db queries
type UsersRepository interface {
	CreateUser(*models.User) error
	GetUser(uuid.UUID) (*models.User, error)
	UpdateUser(*models.User) error
	UpdatePassword(uuid.UUID, string) error
	UpdateEmail(uuid.UUID, string) error
	DeleteUser(uuid.UUID) error
}
