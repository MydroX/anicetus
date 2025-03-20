package repository

import (
	"MydroX/anicetus/internal/common/context"
	"MydroX/anicetus/internal/common/errors"
	"MydroX/anicetus/internal/iam/models"
)

//go:generate mockgen -source=repository_interface.go -destination=../mocks/mock_repository.go -package=mocks

type IamRepository interface {
	SaveSession(ctx *context.AppContext, session *models.Session) *errors.Err
}
