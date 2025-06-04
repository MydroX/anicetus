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
	passwordpkg "MydroX/anicetus/pkg/password"
	"MydroX/anicetus/pkg/uuid"
	"go.uber.org/zap"
)

const (
	sessionPrefix = "session"
)

type usecases struct {
	logger          *zap.SugaredLogger
	usersRepository usersrepository.UsersRepository
	iamRepository   iamrepository.IamStore
	config          *config.Config
	jwtService      *jwt.Service
}

func New(
	l *zap.SugaredLogger,
	ur usersrepository.UsersRepository,
	iamr iamrepository.IamStore,
	cfg *config.Config,
	jwtService *jwt.Service,
) IamUsecasesService {
	return &usecases{
		logger:          l,
		usersRepository: ur,
		iamRepository:   iamr,
		config:          cfg,
		jwtService:      jwtService,
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
		errorsutil.ErrorInvalidInput,
		"username or email must be provided",
		errors.New("username or email must be provided"),
	)
}

func login(
	ctx *context.AppContext,
	u *usecases,
	user *usersmodels.User,
	s *dto.Session,
	reqPwd string,
) (accessToken, refreshToken string, err error) {
	if !passwordpkg.CheckPasswordHash(reqPwd, user.Password) {
		return "", "", errorsutil.New(
			errorsutil.ErrorInvalidInput,
			"invalid password",
			errors.New("invalid password"),
		)
	}

	accessToken, err = u.jwtService.CreateAccessToken(
		user.UUID,
		nil, // TEMP
		[]string{},
	)
	if err != nil {
		return "", "", errorsutil.New(
			errorsutil.ErrorCreateToken,
			"failed to create access token",
			err,
		)
	}

	refreshToken, err = u.jwtService.CreateRefreshToken(
		user.UUID,
		uuid.NewWithPrefix(sessionPrefix),
		[]string{},
	)
	if err != nil {
		return "", "", errorsutil.New(
			errorsutil.ErrorCreateToken,
			"failed to create refresh token",
			err,
		)
	}

	hashParams := argon2id.New(
		argon2id.Iterations(uint32(u.config.Session.Hash.Iterations)),
		argon2id.Memory(uint32(u.config.Session.Hash.Memory)),
		argon2id.Parallelism(uint8(u.config.Session.Hash.Parallelism)),
		argon2id.KeyLength(uint32(u.config.Session.Hash.KeyLength)),
		argon2id.SaltLength(uint32(u.config.Session.Hash.SaltLength)),
	)

	refreshTokenHashed, hashErr := argon2id.Hash(refreshToken, hashParams)
	if hashErr != nil {
		return "", "", errorsutil.New(
			errorsutil.ErrorFailedToHashPassword,
			"failed to hash refresh token",
			hashErr,
		)
	}

	expirationTimeRefresh := time.Now().Add(time.Second * time.Duration(u.config.JWT.RefreshToken.Expiration))

	session := models.Session{
		UUID:           uuid.NewWithPrefix(sessionPrefix),
		UserID:         user.UUID,
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
