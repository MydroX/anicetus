package usecases

import (
	"context"

	"MydroX/anicetus/internal/config"
	"MydroX/anicetus/internal/identity/dto"
	"MydroX/anicetus/internal/identity/models"
	"MydroX/anicetus/internal/identity/repository"
	"MydroX/anicetus/pkg/errs"
	passwordpkg "MydroX/anicetus/pkg/password"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type usecases struct {
	logger     *zap.SugaredLogger
	repository repository.IdentityRepository
	config     *config.Config
}

func New(l *zap.SugaredLogger, r repository.IdentityRepository, cfg *config.Config) IdentityUsecases {
	return &usecases{
		logger:     l,
		repository: r,
		config:     cfg,
	}
}

func (u *usecases) Create(ctx context.Context, req *dto.CreateUserRequest) error {
	passwordHashed, err := passwordpkg.Hash(req.Password)
	if err != nil {
		return &errs.AppError{Code: errs.ErrorInternal, Err: err}
	}

	user := models.User{
		UUID:     uuid.Must(uuid.NewV7()).String(),
		Username: req.Username,
		Email:    req.Email,
		Password: passwordHashed,
	}

	err = u.repository.CreateUser(ctx, &user)

	return err
}

func (u *usecases) Get(ctx context.Context, uuid string) (*dto.GetUserResponse, error) {
	user, err := u.repository.GetUserByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	res := dto.GetUserResponse{
		UUID:     user.UUID,
		Username: user.Username,
		Email:    user.Email,
	}

	return &res, err
}

func (u *usecases) Update(ctx context.Context, userParams *dto.UpdateUserRequest) error {
	user := models.User{
		UUID:     userParams.UUID,
		Username: userParams.Username,
		Email:    userParams.Email,
	}

	_, err := u.repository.UpdateUser(ctx, &user)

	return err
}

func (u *usecases) UpdatePassword(ctx context.Context, uuid, newPassword string) error {
	newPasswordCrypted, err := passwordpkg.Hash(newPassword)
	if err != nil {
		return &errs.AppError{Code: errs.ErrorInternal, Err: err}
	}

	apiErr := u.repository.UpdatePassword(ctx, uuid, newPasswordCrypted)

	return apiErr
}

func (u *usecases) UpdateEmail(ctx context.Context, uuid, email string) error {
	err := u.repository.UpdateEmail(ctx, uuid, email)

	return err
}

func (u *usecases) Delete(ctx context.Context, uuid string) error {
	err := u.repository.DeleteUser(ctx, uuid)

	return err
}

func (u *usecases) GetAllUsers(ctx context.Context) (*dto.GetAllUsersResponse, error) {
	users, err := u.repository.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}

	res := dto.GetAllUsersResponse{
		Users: []*dto.User{},
	}

	for _, user := range users {
		res.Users = append(res.Users, &dto.User{
			UUID:     user.UUID,
			Username: user.Username,
			Email:    user.Email,
		})
	}

	return &res, err
}
