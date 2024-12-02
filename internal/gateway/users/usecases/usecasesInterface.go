package usecases

import (
	"MydroX/project-v/internal/gateway/users/models"
)

//go:generate mockgen -destination=../mocks/mock_usecases.go -package=mocks MydroX/project-v/internal/gateway/users/usecases UsersUsecases

// Usecases is the interface to all the implemented usecases
type UsersUsecases interface {
	Create(user *models.User) error
	Get(uuid string) (*models.User, error)
	Update(user *models.User) error
	UpdatePassword(uuid string, password string) error
	UpdateEmail(uuid string, email string) error
	Delete(uuid string) error
	Login(username, email, password string) (string, error)
}
