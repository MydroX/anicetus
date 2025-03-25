package usecases

import (
	"errors"
	"time"

	"MydroX/anicetus/internal/common/context"
	"MydroX/anicetus/internal/common/errorsutil"
	"MydroX/anicetus/internal/common/jwt"
	"MydroX/anicetus/internal/config"
	"MydroX/anicetus/internal/iam/dto"
	"MydroX/anicetus/internal/iam/models"
	iamrepository "MydroX/anicetus/internal/iam/repository"
	usersmodels "MydroX/anicetus/internal/users/models"
	usersrepository "MydroX/anicetus/internal/users/repository"
	"MydroX/anicetus/pkg/argon2id"
	"MydroX/anicetus/pkg/logger"
	passwordpkg "MydroX/anicetus/pkg/password"
	"MydroX/anicetus/pkg/uuid"
)

const (
	sessionPrefix = "session"
)

type usecases struct {
	logger          *logger.Logger
	usersRepository usersrepository.UsersRepository
	iamRepository   iamrepository.IamRepository
	sessionConfig   *config.Session
}

func New(l *logger.Logger, ur usersrepository.UsersRepository, iamr iamrepository.IamRepository, sessionConfig *config.Session) IamUsecasesInterface {
	return &usecases{
		logger:          l,
		usersRepository: ur,
		iamRepository:   iamr,
		sessionConfig:   sessionConfig,
	}
}

func (u *usecases) Login(ctx *context.AppContext, req *dto.LoginRequest) (accessToken, refreshToken string, err error) {
	switch {
	case req.Email != "":
		user, err := u.usersRepository.GetUserByEmail(ctx, req.Email)
		if err != nil {
			return "", "", err
		}

		accessToken, refreshToken, err := login(ctx, u, user, &req.Session, req.Password)

		return accessToken, refreshToken, err

	case req.Username != "":
		user, err := u.usersRepository.GetUserByUsername(ctx, req.Username)
		if err != nil {
			return "", "", err
		}

		accessToken, refreshToken, err := login(ctx, u, user, &req.Session, req.Password)

		return accessToken, refreshToken, err
	}
	return "", "", errorsutil.New(
		errorsutil.ERROR_INVALID_INPUT,
		"username or email must be provided",
		errors.New("username or email must be provided"),
	)
}

func login(
	ctx *context.AppContext,
	u *usecases, user *usersmodels.User,
	s *dto.Session,
	password string,
) (accessToken, refreshToken string, err error) {
	if !passwordpkg.CheckPasswordHash(password, user.Password) {
		return "", "", errorsutil.New(
			errorsutil.ERROR_INVALID_INPUT,
			"invalid password",
			errors.New("invalid password"),
		)
	}

	expirationTimeAccess := time.Now().Add(time.Second * time.Duration(u.sessionConfig.AccessToken.Expiration))

	accessToken, _ = jwt.CreateAccessToken(
		expirationTimeAccess,
		u.sessionConfig.AccessToken.Secret,
		user.UUID,
	)

	expirationTimeRefresh := time.Now().Add(time.Second * time.Duration(u.sessionConfig.RefreshToken.Expiration))

	refreshToken, _ = jwt.CreateRefreshToken(
		expirationTimeRefresh,
		u.sessionConfig.RefreshToken.Secret,
		user.UUID,
	)

	hashParams := argon2id.New(
		argon2id.Iterations(uint32(u.sessionConfig.HashConfig.Iterations)),
		argon2id.Memory(uint32(u.sessionConfig.HashConfig.Memory)),
		argon2id.Parallelism(uint8(u.sessionConfig.HashConfig.Parallelism)),
		argon2id.KeyLength(uint32(u.sessionConfig.HashConfig.KeyLength)),
		argon2id.SaltLength(uint32(u.sessionConfig.HashConfig.SaltLength)),
	)

	refreshTokenHashed, hashErr := argon2id.Hash(refreshToken, hashParams)
	if hashErr != nil {
		return "", "", errorsutil.New(
			errorsutil.ERROR_FAILED_TO_HASH_PASSWORD,
			"failed to hash refresh token",
			hashErr,
		)
	}

	session := models.Session{
		UUID:           uuid.NewWithPrefix(sessionPrefix),
		UserId:         user.UUID,
		OS:             s.OS,
		Browser:        s.Browser,
		BrowserVersion: s.BrowserVersion,
		IPv4Address:    s.IPv4Address,
		CreatedAt:      time.Now(),
		LastUsedAt:     time.Now(),
		ExpiresAt:      expirationTimeRefresh,
		RefreshToken:   refreshTokenHashed,
	}

	err = u.iamRepository.SaveSession(ctx, &session)
	if err != nil {
		// TODO
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// nolint
func (u *usecases) Logout(ctx *context.AppContext, token string) error {
	panic("not implemented") // TODO: Implement
}

// nolint
func (u *usecases) RefreshToken(ctx *context.AppContext, token string) (string, error) {
	panic("not implemented") // TODO: Implement
}
