package usecases

import (
	"errors"
	"time"

	"context"
	"MydroX/anicetus/pkg/errs"
	"MydroX/anicetus/pkg/jwt"
	"MydroX/anicetus/internal/config"
	"MydroX/anicetus/internal/iam/dto"
	"MydroX/anicetus/internal/iam/models"
	iamrepository "MydroX/anicetus/internal/iam/repository"
	usersmodels "MydroX/anicetus/internal/users/models"
	usersrepository "MydroX/anicetus/internal/users/repository"
	"MydroX/anicetus/pkg/argon2id"
	passwordpkg "MydroX/anicetus/pkg/password"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type usecases struct {
	logger          *zap.SugaredLogger
	usersRepository usersrepository.UsersRepository
	iamRepository   iamrepository.IamStore
	audienceStore   iamrepository.AudienceStore
	audienceManager *AudienceManager
	config          *config.Config
	jwtService      *jwt.Service
}

func New(
	l *zap.SugaredLogger,
	ur usersrepository.UsersRepository,
	iamr iamrepository.IamStore,
	cfg *config.Config,
	jwtService *jwt.Service,
	audienceStore iamrepository.AudienceStore,
	audienceManager *AudienceManager,
) IamUsecasesService {
	return &usecases{
		logger:          l,
		usersRepository: ur,
		iamRepository:   iamr,
		audienceStore:   audienceStore,
		audienceManager: audienceManager,
		config:          cfg,
		jwtService:      jwtService,
	}
}

func (u *usecases) Login(ctx context.Context, req *dto.LoginRequest) (accessToken, refreshToken string, err error) {
	switch {
	case req.Email != "":
		user, err := u.usersRepository.GetUserByEmail(ctx, req.Email)
		if err != nil {
			return "", "", errs.New(errs.ErrorInvalidCredentials, errs.MessageInvalidCredentials, nil)
		}

		return login(ctx, u, user, &req.Session, req.Password)

	case req.Username != "":
		user, err := u.usersRepository.GetUserByUsername(ctx, req.Username)
		if err != nil {
			return "", "", errs.New(errs.ErrorInvalidCredentials, errs.MessageInvalidCredentials, nil)
		}

		return login(ctx, u, user, &req.Session, req.Password)
	}

	return "", "", errs.New(
		errs.ErrorInvalidInput,
		"username or email must be provided",
		errors.New("username or email must be provided"),
	)
}

func login(
	ctx context.Context,
	u *usecases,
	user *usersmodels.User,
	s *dto.Session,
	reqPwd string,
) (accessToken, refreshToken string, err error) {
	if !passwordpkg.CheckPasswordHash(reqPwd, user.Password) {
		return "", "", errs.New(errs.ErrorInvalidCredentials, errs.MessageInvalidCredentials, nil)
	}

	// Fetch per-user audiences
	audiences, audErr := u.audienceManager.GetUserAudiences(ctx, user.UUID)
	if audErr != nil {
		u.logger.Warnw("Failed to get user audiences, using empty", "error", audErr)

		audiences = []string{}
	}

	accessToken, err = u.jwtService.CreateAccessToken(
		user.UUID,
		nil, // TODO: permissions
		audiences,
	)
	if err != nil {
		return "", "", errs.New(
			errs.ErrorCreateToken,
			"failed to create access token",
			err,
		)
	}

	refreshToken, err = u.jwtService.CreateRefreshToken(
		user.UUID,
		uuid.Must(uuid.NewV7()).String(),
		audiences,
	)
	if err != nil {
		return "", "", errs.New(
			errs.ErrorCreateToken,
			"failed to create refresh token",
			err,
		)
	}

	if err := u.saveSession(ctx, user.UUID, refreshToken, s); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (u *usecases) saveSession(ctx context.Context, userUUID, refreshToken string, s *dto.Session) error {
	hashParams := argon2id.New(
		argon2id.Iterations(u.config.Session.Hash.Iterations),
		argon2id.Memory(u.config.Session.Hash.Memory),
		argon2id.Parallelism(u.config.Session.Hash.Parallelism),
		argon2id.KeyLength(u.config.Session.Hash.KeyLength),
		argon2id.SaltLength(u.config.Session.Hash.SaltLength),
	)

	refreshTokenHashed, hashErr := argon2id.Hash(refreshToken, hashParams)
	if hashErr != nil {
		return errs.New(
			errs.ErrorFailedToHashPassword,
			"failed to hash refresh token",
			hashErr,
		)
	}

	expirationTimeRefresh := time.Now().Add(time.Second * time.Duration(u.config.JWT.RefreshToken.Expiration))

	session := models.Session{
		UUID:           uuid.Must(uuid.NewV7()).String(),
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

	return u.iamRepository.SaveSession(ctx, &session)
}

// nolint
func (u *usecases) Logout(ctx context.Context, token string) error {
	panic("not implemented") // TODO: Implement
}

// nolint
func (u *usecases) RefreshToken(ctx context.Context, token string) (string, error) {
	panic("not implemented") // TODO: Implement
}

func (u *usecases) RegisterAudience(ctx context.Context, req *dto.RegisterAudienceRequest) error {
	metadata := map[string]any{
		"service_name": req.ServiceName,
		"description":  req.Description,
		"permissions":  req.Permissions,
	}

	err := u.audienceStore.RegisterAudience(ctx, req.Audience, metadata)
	if err != nil {
		return err
	}

	u.audienceManager.InvalidateAllAudiencesCache(ctx)

	return nil
}

func (u *usecases) RevokeAudience(ctx context.Context, audience string) error {
	err := u.audienceStore.RevokeAudience(ctx, audience)
	if err != nil {
		return err
	}

	u.audienceManager.InvalidateAllAudiencesCache(ctx)

	return nil
}

func (u *usecases) GetAllAudiences(ctx context.Context) ([]string, error) {
	return u.audienceManager.GetAllowedAudiences(ctx)
}

func (u *usecases) GetUserAudiences(ctx context.Context, userUUID string) ([]string, error) {
	return u.audienceManager.GetUserAudiences(ctx, userUUID)
}

func (u *usecases) AssignAudienceToUser(ctx context.Context, userUUID string, req *dto.AssignAudienceRequest) error {
	err := u.audienceStore.AssignAudienceToUser(ctx, userUUID, req.Audience)
	if err != nil {
		return err
	}

	u.audienceManager.InvalidateUserAudiencesCache(ctx, userUUID)

	return nil
}
