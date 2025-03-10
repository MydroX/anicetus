package usecases

import (
	"MydroX/project-v/internal/config"
	"MydroX/project-v/internal/users/repository"
	"MydroX/project-v/pkg/jwt"
	"MydroX/project-v/pkg/logger"
	passwordpkg "MydroX/project-v/pkg/password"

	"context"
	"fmt"
	"time"
)

type usecases struct {
	logger     *logger.Logger
	repository repository.UsersRepository
	jwtConfig  *config.JWT
}

func NewUsecases(l *logger.Logger, r repository.UsersRepository, jwtConfig *config.JWT) IamUsecasesInterface {
	return &usecases{
		logger:     l,
		repository: r,
		jwtConfig:  jwtConfig,
	}
}

func (u usecases) Login(ctx *context.Context, username, email, password string) (accessToken, refreshToken string, err error) {
	switch {
	case email != "":
		user, err := u.repository.GetUserByEmail(ctx, email)
		if err != nil {
			return "", "", err
		}

		accessToken, refreshToken, err := login(u.jwtConfig, user.UUID, password, user.Password)

		return accessToken, refreshToken, err

	case username != "":
		user, err := u.repository.GetUserByUsername(ctx, username)
		if err != nil {
			return "", "", err
		}

		accessToken, refreshToken, err := login(u.jwtConfig, user.UUID, password, user.Password)

		return accessToken, refreshToken, err
	}
	return "", "", fmt.Errorf("username or email must be provided")
}

func login(jwtConfig *config.JWT, userUUID, password, passwordCrypted string) (accessToken, refreshToken string, err error) {
	if !passwordpkg.CheckPasswordHash(password, passwordCrypted) {
		return "", "", fmt.Errorf("invalid password")
	}

	expirationTimeAccess := time.Now().Add(time.Second * time.Duration(jwtConfig.AccessToken.Expiration))

	accessToken, _ = jwt.CreateAccessToken(
		expirationTimeAccess,
		jwtConfig.AccessToken.Secret,
		userUUID,
	)

	expirationTimeRefresh := time.Now().Add(time.Second * time.Duration(jwtConfig.RefreshToken.Expiration))

	refreshToken, _ = jwt.CreateRefreshToken(
		expirationTimeRefresh,
		jwtConfig.RefreshToken.Secret,
		userUUID,
	)

	return accessToken, refreshToken, nil
}

// nolint
func (u usecases) Logout(ctx *context.Context, token string) error {
	panic("not implemented") // TODO: Implement
}

// nolint
func (u usecases) RefreshToken(ctx *context.Context, token string) (string, error) {
	panic("not implemented") // TODO: Implement
}
