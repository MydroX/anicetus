package repository

import (
	"MydroX/project-v/internal/gateway/users/models"
)

//go:generate mockgen -destination=../mocks/mock_repository.go -imports=models=MydroX/project-v/internal/users/models -package=mocks MydroX/project-v/internal/gateway/users/repository UsersRepository

// Repository is the interface to all the implemented db queries
type UsersRepository interface {
	CreateUser(*models.User) error
	GetUser(string) (*models.User, error)
	UpdateUser(*models.User) error
	UpdatePassword(string, string) error
	UpdateEmail(string, string) error
	DeleteUser(string) error
	GetUserByEmail(string) (*models.User, error)
	GetUserByUsername(string) (*models.User, error)
}
