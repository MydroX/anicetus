package usecases

import (
	"MydroX/project-v/internal/common/errorscode"
	"MydroX/project-v/internal/config"
	"MydroX/project-v/internal/users/dto"
	"MydroX/project-v/internal/users/models"
	"MydroX/project-v/internal/users/repository"
	"MydroX/project-v/pkg/logger"
	passwordpkg "MydroX/project-v/pkg/password"
	uuidpkg "MydroX/project-v/pkg/uuid"
	"context"
)

var prefix = "user"

type usecases struct {
	logger     *logger.Logger
	repository repository.UsersRepository
	jwtConfig  *config.JWT
}

func NewUsecases(l *logger.Logger, r repository.UsersRepository, jwtConfig *config.JWT) UsersUsecases {
	return &usecases{
		logger:     l,
		repository: r,
		jwtConfig:  jwtConfig,
	}
}

func (u *usecases) Create(ctx *context.Context, req *dto.CreateUserRequest) error {
	passwordHashed, err := passwordpkg.Hash(req.Password)
	if err != nil {
		*ctx = context.WithValue(*ctx, errorscode.CtxErrorCodeKey, errorscode.CODE_FAILED_TO_HASH_PASSWORD)
		return err
	}

	// If the role is not provided, the default role is USER
	// Might not be a good idea to force a non modifiable default value
	if req.Role == "" {
		req.Role = "USER"
	}

	user := models.User{
		UUID:     uuidpkg.NewWithPrefix(prefix),
		Username: req.Username,
		Email:    req.Email,
		Password: passwordHashed,
		Role:     req.Role,
	}

	err = u.repository.CreateUser(ctx, &user)

	return err
}

func (u *usecases) Get(ctx *context.Context, uuid string) (*dto.GetUserResponse, error) {
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

func (u *usecases) Update(ctx *context.Context, userParams *dto.UpdateUserRequest) error {
	user := models.User{
		UUID:     userParams.UUID,
		Username: userParams.Username,
		Email:    userParams.Email,
		Role:     userParams.Role,
	}

	_, err := u.repository.UpdateUser(ctx, &user)

	return err
}

func (u *usecases) UpdatePassword(ctx *context.Context, uuid, newPassword string) error {
	newPasswordCrypted, err := passwordpkg.Hash(newPassword)
	if err != nil {
		*ctx = context.WithValue(*ctx, errorscode.CtxErrorCodeKey, errorscode.CODE_FAILED_TO_HASH_PASSWORD)
		return err
	}

	err = u.repository.UpdatePassword(ctx, uuid, newPasswordCrypted)
	return err
}

func (u *usecases) UpdateEmail(ctx *context.Context, uuid, email string) error {
	err := u.repository.UpdateEmail(ctx, uuid, email)
	return err
}

func (u *usecases) Delete(ctx *context.Context, uuid string) error {
	err := u.repository.DeleteUser(ctx, uuid)
	return err
}

func (u *usecases) GetAllUsers(ctx *context.Context) (*dto.GetAllUsersResponse, error) {
	users, err := u.repository.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}

	res := dto.GetAllUsersResponse{
		Users: make([]*dto.User, 0),
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
