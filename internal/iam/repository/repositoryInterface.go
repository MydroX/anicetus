package repository

import "github.com/google/uuid"

//go:generate mockgen -destination=../mocks/mock_repository.go -package=mocks MydroX/project-v/internal/iam/repository RepositoryInterface
type RepositoryInterface interface {
	CreateUser(uuid uuid.UUID, name, email, password, role string) error
}
