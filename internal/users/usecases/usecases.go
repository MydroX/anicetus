package usecases

import (
	"MydroX/anicetus/internal/common/context"
	"MydroX/anicetus/internal/common/errorsutil"
	"MydroX/anicetus/internal/config"
	"MydroX/anicetus/internal/users/dto"
	"MydroX/anicetus/internal/users/models"
	"MydroX/anicetus/internal/users/repository"
	passwordpkg "MydroX/anicetus/pkg/password"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type usecases struct {
	logger     *zap.SugaredLogger
	repository repository.UsersRepository
	config     *config.Config
}

func New(l *zap.SugaredLogger, r repository.UsersRepository, cfg *config.Config) UsersUsecases {
	return &usecases{
		logger:     l,
		repository: r,
		config:     cfg,
	}
}

func (u *usecases) Create(ctx *context.AppContext, req *dto.CreateUserRequest) error {
	passwordHashed, err := passwordpkg.Hash(req.Password)
	if err != nil {
		return &errorsutil.AppError{Code: errorsutil.ErrorInternal, Err: err}
	}

	user := models.User{
		UUID:     uuid.Must(uuid.NewV7()).String(),
		Username: req.Username,
		Email:    req.Email,
		Password: passwordHashed,
		Role:     []string{"USER"},
	}

	err = u.repository.CreateUser(ctx, &user)

	return err
}

func (u *usecases) Get(ctx *context.AppContext, uuid string) (*dto.GetUserResponse, error) {
	user, err := u.repository.GetUserByUUID(ctx, uuid)
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

func (u *usecases) Update(ctx *context.AppContext, userParams *dto.UpdateUserRequest) error {
	user := models.User{
		UUID:     userParams.UUID,
		Username: userParams.Username,
		Email:    userParams.Email,
		Role:     userParams.Role,
	}

	_, err := u.repository.UpdateUser(ctx, &user)

	return err
}

func (u *usecases) UpdatePassword(ctx *context.AppContext, uuid, newPassword string) error {
	newPasswordCrypted, err := passwordpkg.Hash(newPassword)
	if err != nil {
		return &errorsutil.AppError{Code: errorsutil.ErrorInternal, Err: err}
	}

	apiErr := u.repository.UpdatePassword(ctx, uuid, newPasswordCrypted)

	return apiErr
}

func (u *usecases) UpdateEmail(ctx *context.AppContext, uuid, email string) error {
	err := u.repository.UpdateEmail(ctx, uuid, email)

	return err
}

func (u *usecases) Delete(ctx *context.AppContext, uuid string) error {
	err := u.repository.DeleteUser(ctx, uuid)

	return err
}

func (u *usecases) GetAllUsers(ctx *context.AppContext) (*dto.GetAllUsersResponse, error) {
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
			Role:     user.Role,
		})
	}

	return &res, err
}
