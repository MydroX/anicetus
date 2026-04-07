package usecases

import (
	"context"
	"errors"
	"time"

	"MydroX/anicetus/internal/authentication/dto"
	"MydroX/anicetus/internal/authentication/models"
	authrepository "MydroX/anicetus/internal/authentication/repository"
	"MydroX/anicetus/internal/config"
	identitymodels "MydroX/anicetus/internal/identity/models"
	identityrepository "MydroX/anicetus/internal/identity/repository"
	servicesusecases "MydroX/anicetus/internal/services/usecases"
	"MydroX/anicetus/pkg/argon2id"
	"MydroX/anicetus/pkg/errs"
	"MydroX/anicetus/pkg/jwt"
	passwordpkg "MydroX/anicetus/pkg/password"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type usecases struct {
	logger             *zap.SugaredLogger
	identityRepository identityrepository.IdentityRepository
	sessionStore       authrepository.SessionStore
	serviceManager     *servicesusecases.ServiceManager
	config             *config.Config
	jwtService         *jwt.Service
}

func New(
	l *zap.SugaredLogger,
	ir identityrepository.IdentityRepository,
	ss authrepository.SessionStore,
	cfg *config.Config,
	jwtService *jwt.Service,
	serviceManager *servicesusecases.ServiceManager,
) AuthenticationUsecases {
	return &usecases{
		logger:             l,
		identityRepository: ir,
		sessionStore:       ss,
		serviceManager:     serviceManager,
		config:             cfg,
		jwtService:         jwtService,
	}
}

func (u *usecases) Login(ctx context.Context, req *dto.LoginRequest) (accessToken, refreshToken string, err error) {
	switch {
	case req.Email != "":
		user, err := u.identityRepository.GetUserByEmail(ctx, req.Email)
		if err != nil {
			return "", "", errs.New(errs.ErrorInvalidCredentials, errs.MessageInvalidCredentials, nil)
		}

		return u.authenticateUser(ctx, user, &req.Session, req.Password)

	case req.Username != "":
		user, err := u.identityRepository.GetUserByUsername(ctx, req.Username)
		if err != nil {
			return "", "", errs.New(errs.ErrorInvalidCredentials, errs.MessageInvalidCredentials, nil)
		}

		return u.authenticateUser(ctx, user, &req.Session, req.Password)
	}

	return "", "", errs.New(
		errs.ErrorInvalidInput,
		"username or email must be provided",
		errors.New("username or email must be provided"),
	)
}

func (u *usecases) authenticateUser(
	ctx context.Context,
	user *identitymodels.User,
	s *dto.Session,
	reqPwd string,
) (accessToken, refreshToken string, err error) {
	if !passwordpkg.CheckPasswordHash(reqPwd, user.Password) {
		return "", "", errs.New(errs.ErrorInvalidCredentials, errs.MessageInvalidCredentials, nil)
	}

	// Fetch per-user services (JWT audiences)
	audiences, audErr := u.serviceManager.GetUserServices(ctx, user.UUID)
	if audErr != nil {
		u.logger.Warnw("Failed to get user services, using empty", "error", audErr)

		audiences = []string{}
	}

	sessionUUID := uuid.Must(uuid.NewV7()).String()

	accessToken, err = u.jwtService.CreateAccessToken(
		user.UUID,
		nil, // TODO: permissions (will be provided by authorization domain)
		audiences,
	)
	if err != nil {
		return "", "", errs.New(errs.ErrorCreateToken, "failed to create access token", err)
	}

	refreshToken, err = u.jwtService.CreateRefreshToken(
		user.UUID,
		sessionUUID,
		audiences,
	)
	if err != nil {
		return "", "", errs.New(errs.ErrorCreateToken, "failed to create refresh token", err)
	}

	if err := u.saveSession(ctx, user.UUID, sessionUUID, refreshToken, s); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (u *usecases) saveSession(ctx context.Context, userUUID, sessionUUID, refreshToken string, s *dto.Session) error {
	hashParams := argon2id.New(
		argon2id.Iterations(u.config.Session.Hash.Iterations),
		argon2id.Memory(u.config.Session.Hash.Memory),
		argon2id.Parallelism(u.config.Session.Hash.Parallelism),
		argon2id.KeyLength(u.config.Session.Hash.KeyLength),
		argon2id.SaltLength(u.config.Session.Hash.SaltLength),
	)

	refreshTokenHashed, hashErr := argon2id.Hash(refreshToken, hashParams)
	if hashErr != nil {
		return errs.New(errs.ErrorFailedToHashPassword, "failed to hash refresh token", hashErr)
	}

	expirationTimeRefresh := time.Now().Add(time.Second * time.Duration(u.config.JWT.RefreshToken.Expiration))

	session := models.Session{
		UUID:           sessionUUID,
		UserID:         userUUID,
		OS:             s.OS,
		Browser:        s.Browser,
		BrowserVersion: s.BrowserVersion,
		IPv4Address:    s.IPv4Address,
		CreatedAt:      time.Now(),
		LastUsedAt:     time.Now(),
		ExpiresAt:      expirationTimeRefresh,
		RefreshToken:   refreshTokenHashed,
	}

	return u.sessionStore.SaveSession(ctx, &session)
}

func (u *usecases) Logout(ctx context.Context, refreshToken string) error {
	claims, err := u.jwtService.ParseRefreshToken(refreshToken)
	if err != nil {
		return errs.New(errs.ErrorUnauthorized, "invalid refresh token", err)
	}

	return u.sessionStore.DeleteSession(ctx, claims.SessionUUID)
}

func (u *usecases) RefreshToken(ctx context.Context, refreshToken string) (newAccessToken, newRefreshToken string, err error) {
	session, err := u.validateRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", "", err
	}

	// Fetch fresh user data
	user, err := u.identityRepository.GetUserByUUID(ctx, session.UserID)
	if err != nil {
		return "", "", errs.New(errs.ErrorUnauthorized, "user not found", err)
	}

	// Fetch fresh audiences
	audiences, audErr := u.serviceManager.GetUserServices(ctx, user.UUID)
	if audErr != nil {
		u.logger.Warnw("Failed to get user services, using empty", "error", audErr)

		audiences = []string{}
	}

	return u.rotateSession(ctx, user.UUID, session, audiences)
}

func (u *usecases) validateRefreshToken(ctx context.Context, refreshToken string) (*models.Session, error) {
	claims, err := u.jwtService.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, errs.New(errs.ErrorUnauthorized, "invalid refresh token", err)
	}

	session, err := u.sessionStore.GetSessionByUUID(ctx, claims.SessionUUID)
	if err != nil {
		return nil, errs.New(errs.ErrorUnauthorized, "session not found", err)
	}

	if time.Now().After(session.ExpiresAt) {
		_ = u.sessionStore.DeleteSession(ctx, session.UUID)

		return nil, errs.New(errs.ErrorUnauthorized, "session expired", nil)
	}

	match, err := argon2id.Verify(refreshToken, session.RefreshToken)
	if err != nil || !match {
		return nil, errs.New(errs.ErrorUnauthorized, "invalid refresh token", err)
	}

	return session, nil
}

func (u *usecases) rotateSession(
	ctx context.Context,
	userUUID string,
	oldSession *models.Session,
	audiences []string,
) (accessToken, refreshToken string, err error) {
	newAccessToken, err := u.jwtService.CreateAccessToken(userUUID, nil, audiences)
	if err != nil {
		return "", "", errs.New(errs.ErrorCreateToken, "failed to create access token", err)
	}

	newSessionUUID := uuid.Must(uuid.NewV7()).String()

	newRefreshToken, err := u.jwtService.CreateRefreshToken(userUUID, newSessionUUID, audiences)
	if err != nil {
		return "", "", errs.New(errs.ErrorCreateToken, "failed to create refresh token", err)
	}

	if err := u.sessionStore.DeleteSession(ctx, oldSession.UUID); err != nil {
		return "", "", errs.New(errs.ErrorInternal, "failed to delete old session", err)
	}

	sessionDTO := &dto.Session{
		OS:             oldSession.OS,
		OSVersion:      oldSession.OSVersion,
		Browser:        oldSession.Browser,
		BrowserVersion: oldSession.BrowserVersion,
		IPv4Address:    oldSession.IPv4Address,
	}

	if err := u.saveSession(ctx, userUUID, newSessionUUID, newRefreshToken, sessionDTO); err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}
