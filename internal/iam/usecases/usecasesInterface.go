package usecases

import "context"

//go:generate mockgen -destination=../mocks/mock_usecases.go -package=mocks MydroX/project-v/internal/iam/usecases UsecasesInterface

type UsecasesInterface interface {
	Create(ctx context.Context, username, password, email, role string) error
	Get(ctx *context.Context)
	Update(ctx *context.Context)
	Delete(ctx *context.Context)
	Auth(ctx *context.Context)
}
