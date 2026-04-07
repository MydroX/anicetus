package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	authmocks "MydroX/anicetus/internal/authentication/mocks"
	"MydroX/anicetus/internal/authentication/dto"
	"MydroX/anicetus/internal/authentication/models"
	"MydroX/anicetus/internal/config"
	identitymocks "MydroX/anicetus/internal/identity/mocks"
	identitymodels "MydroX/anicetus/internal/identity/models"
	servicesmocks "MydroX/anicetus/internal/services/mocks"
	servicesusecases "MydroX/anicetus/internal/services/usecases"
	"MydroX/anicetus/pkg/argon2id"
	"MydroX/anicetus/pkg/errs"
	"MydroX/anicetus/pkg/jwt"
	"MydroX/anicetus/pkg/logger"
	passwordpkg "MydroX/anicetus/pkg/password"

	valkeymock "github.com/valkey-io/valkey-go/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type testDeps struct {
	identityRepo *identitymocks.MockIdentityRepository
	sessionStore *authmocks.MockSessionStore
	serviceStore *servicesmocks.MockServiceStore
	usecases     AuthenticationUsecases
	jwtService   *jwt.Service
}

func setupTest(t *testing.T) testDeps {
	ctrl := gomock.NewController(t)
	log, _ := logger.New("TEST")

	identityRepo := identitymocks.NewMockIdentityRepository(ctrl)
	sessionStore := authmocks.NewMockSessionStore(ctrl)
	serviceStore := servicesmocks.NewMockServiceStore(ctrl)

	valkeyClient := valkeymock.NewClient(ctrl)
	// Mock valkey Do calls to always return cache miss (error)
	valkeyClient.EXPECT().Do(gomock.Any(), gomock.Any()).Return(valkeymock.ErrorResult(errors.New("cache miss"))).AnyTimes()

	serviceManager := servicesusecases.NewServiceManager(log, serviceStore, valkeyClient)

	jwtConfig := jwt.TokenConfig{
		AccessTokenSecret:    "test-secret-key-long-enough-for-signing-jwt-tokens-securely",
		RefreshTokenSecret:   "test-refresh-secret-long-enough-for-signing-jwt-tok",
		ExpectedIssuer:       "test-issuer",
		ClockSkewSeconds:     60,
		AccessTokenDuration:  3600,
		RefreshTokenDuration: 86400,
	}
	jwtService := jwt.NewJWTService(jwtConfig)

	cfg := &config.Config{
		JWT: config.JWT{
			RefreshToken: config.RefreshToken{Expiration: 86400},
		},
		Session: config.Session{
			Hash: config.Hash{
				Iterations:  1,
				Memory:      1024,
				Parallelism: 1,
				KeyLength:   32,
				SaltLength:  16,
			},
		},
	}

	uc := New(log, identityRepo, sessionStore, cfg, jwtService, serviceManager)

	return testDeps{
		identityRepo: identityRepo,
		sessionStore: sessionStore,
		serviceStore: serviceStore,
		usecases:     uc,
		jwtService:   jwtService,
	}
}

func hashedPassword() string {
	hash, _ := passwordpkg.Hash("password123!")
	return hash
}

func testSession() dto.Session {
	return dto.Session{
		IPv4Address:    "192.168.1.1",
		OS:             "macOS",
		OSVersion:      "14.0",
		Browser:        "Chrome",
		BrowserVersion: "120.0",
	}
}

// --- Login ---

func TestLogin_ByEmail_Success(t *testing.T) {
	deps := setupTest(t)
	ctx := context.Background()

	user := &identitymodels.User{
		UUID:     "user-uuid-123",
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword(),
	}

	deps.identityRepo.EXPECT().GetUserByEmail(gomock.Any(), "test@example.com").Return(user, nil)
	deps.serviceStore.EXPECT().GetUserServices(gomock.Any(), "user-uuid-123").Return([]string{"service-a"}, nil)
	deps.sessionStore.EXPECT().SaveSession(gomock.Any(), gomock.Any()).Return(nil)

	session := testSession()
	req := &dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123!",
		Session:  session,
	}

	accessToken, refreshToken, err := deps.usecases.Login(ctx, req)
	require.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
}

func TestLogin_ByUsername_Success(t *testing.T) {
	deps := setupTest(t)
	ctx := context.Background()

	user := &identitymodels.User{
		UUID:     "user-uuid-123",
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword(),
	}

	deps.identityRepo.EXPECT().GetUserByUsername(gomock.Any(), "testuser").Return(user, nil)
	deps.serviceStore.EXPECT().GetUserServices(gomock.Any(), "user-uuid-123").Return([]string{}, nil)
	deps.sessionStore.EXPECT().SaveSession(gomock.Any(), gomock.Any()).Return(nil)

	session := testSession()
	req := &dto.LoginRequest{
		Username: "testuser",
		Password: "password123!",
		Session:  session,
	}

	accessToken, refreshToken, err := deps.usecases.Login(ctx, req)
	require.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
}

func TestLogin_UserNotFound(t *testing.T) {
	deps := setupTest(t)
	ctx := context.Background()

	deps.identityRepo.EXPECT().GetUserByEmail(gomock.Any(), "nonexistent@example.com").
		Return(nil, errs.New(errs.ErrorNotFound, "not found", nil))

	req := &dto.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123!",
		Session:  testSession(),
	}

	_, _, err := deps.usecases.Login(ctx, req)
	require.Error(t, err)

	var appErr *errs.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, errs.ErrorInvalidCredentials, appErr.Code)
}

func TestLogin_WrongPassword(t *testing.T) {
	deps := setupTest(t)
	ctx := context.Background()

	user := &identitymodels.User{
		UUID:     "user-uuid-123",
		Password: hashedPassword(),
	}

	deps.identityRepo.EXPECT().GetUserByEmail(gomock.Any(), "test@example.com").Return(user, nil)

	req := &dto.LoginRequest{
		Email:    "test@example.com",
		Password: "wrong-password",
		Session:  testSession(),
	}

	_, _, err := deps.usecases.Login(ctx, req)
	require.Error(t, err)

	var appErr *errs.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, errs.ErrorInvalidCredentials, appErr.Code)
}

func TestLogin_NeitherEmailNorUsername(t *testing.T) {
	deps := setupTest(t)
	ctx := context.Background()

	req := &dto.LoginRequest{
		Password: "password123!",
		Session:  testSession(),
	}

	_, _, err := deps.usecases.Login(ctx, req)
	require.Error(t, err)

	var appErr *errs.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, errs.ErrorInvalidInput, appErr.Code)
}

func TestLogin_SessionSaveFailure(t *testing.T) {
	deps := setupTest(t)
	ctx := context.Background()

	user := &identitymodels.User{
		UUID:     "user-uuid-123",
		Password: hashedPassword(),
	}

	deps.identityRepo.EXPECT().GetUserByEmail(gomock.Any(), "test@example.com").Return(user, nil)
	deps.serviceStore.EXPECT().GetUserServices(gomock.Any(), "user-uuid-123").Return([]string{}, nil)
	deps.sessionStore.EXPECT().SaveSession(gomock.Any(), gomock.Any()).
		Return(errs.New(errs.ErrorInternal, "db error", nil))

	req := &dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123!",
		Session:  testSession(),
	}

	_, _, err := deps.usecases.Login(ctx, req)
	require.Error(t, err)
}

// --- Logout ---

func TestLogout_Success(t *testing.T) {
	deps := setupTest(t)
	ctx := context.Background()

	// Create a real refresh token to parse
	refreshToken, _ := deps.jwtService.CreateRefreshToken("user-uuid-123", "session-uuid-456", []string{})

	deps.sessionStore.EXPECT().DeleteSession(gomock.Any(), "session-uuid-456").Return(nil)

	err := deps.usecases.Logout(ctx, refreshToken)
	assert.NoError(t, err)
}

func TestLogout_InvalidToken(t *testing.T) {
	deps := setupTest(t)
	ctx := context.Background()

	err := deps.usecases.Logout(ctx, "invalid-token")
	require.Error(t, err)

	var appErr *errs.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, errs.ErrorUnauthorized, appErr.Code)
}

func TestLogout_DeleteSessionError(t *testing.T) {
	deps := setupTest(t)
	ctx := context.Background()

	refreshToken, _ := deps.jwtService.CreateRefreshToken("user-uuid-123", "session-uuid-456", []string{})

	deps.sessionStore.EXPECT().DeleteSession(gomock.Any(), "session-uuid-456").
		Return(errs.New(errs.ErrorNotFound, "session not found", nil))

	err := deps.usecases.Logout(ctx, refreshToken)
	require.Error(t, err)
}

// --- RefreshToken ---

func TestRefreshToken_Success(t *testing.T) {
	deps := setupTest(t)
	ctx := context.Background()

	// Create a real refresh token
	refreshToken, _ := deps.jwtService.CreateRefreshToken("user-uuid-123", "session-uuid-456", []string{})

	// Hash the token the same way saveSession does
	hashParams := argon2id.New(
		argon2id.Iterations(1),
		argon2id.Memory(64*1024),
		argon2id.Parallelism(1),
		argon2id.KeyLength(32),
		argon2id.SaltLength(16),
	)
	hashedToken, _ := argon2id.Hash(refreshToken, hashParams)

	session := &models.Session{
		UUID:         "session-uuid-456",
		UserID:       "user-uuid-123",
		RefreshToken: hashedToken,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		OS:           "macOS",
		OSVersion:    "14.0",
		Browser:      "Chrome",
		BrowserVersion: "120.0",
		IPv4Address:  "192.168.1.1",
	}

	user := &identitymodels.User{
		UUID:     "user-uuid-123",
		Username: "testuser",
		Email:    "test@example.com",
	}

	deps.sessionStore.EXPECT().GetSessionByUUID(gomock.Any(), "session-uuid-456").Return(session, nil)
	deps.identityRepo.EXPECT().GetUserByUUID(gomock.Any(), "user-uuid-123").Return(user, nil)
	deps.serviceStore.EXPECT().GetUserServices(gomock.Any(), "user-uuid-123").Return([]string{}, nil)
	deps.sessionStore.EXPECT().DeleteSession(gomock.Any(), "session-uuid-456").Return(nil)
	deps.sessionStore.EXPECT().SaveSession(gomock.Any(), gomock.Any()).Return(nil)

	newAccess, newRefresh, err := deps.usecases.RefreshToken(ctx, refreshToken)
	require.NoError(t, err)
	assert.NotEmpty(t, newAccess)
	assert.NotEmpty(t, newRefresh)
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	deps := setupTest(t)
	ctx := context.Background()

	_, _, err := deps.usecases.RefreshToken(ctx, "invalid-token")
	require.Error(t, err)

	var appErr *errs.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, errs.ErrorUnauthorized, appErr.Code)
}

func TestRefreshToken_SessionNotFound(t *testing.T) {
	deps := setupTest(t)
	ctx := context.Background()

	refreshToken, _ := deps.jwtService.CreateRefreshToken("user-uuid-123", "session-uuid-456", []string{})

	deps.sessionStore.EXPECT().GetSessionByUUID(gomock.Any(), "session-uuid-456").
		Return(nil, errs.New(errs.ErrorNotFound, "not found", nil))

	_, _, err := deps.usecases.RefreshToken(ctx, refreshToken)
	require.Error(t, err)

	var appErr *errs.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, errs.ErrorUnauthorized, appErr.Code)
}

func TestRefreshToken_SessionExpired(t *testing.T) {
	deps := setupTest(t)
	ctx := context.Background()

	refreshToken, _ := deps.jwtService.CreateRefreshToken("user-uuid-123", "session-uuid-456", []string{})

	session := &models.Session{
		UUID:      "session-uuid-456",
		UserID:    "user-uuid-123",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
	}

	deps.sessionStore.EXPECT().GetSessionByUUID(gomock.Any(), "session-uuid-456").Return(session, nil)
	deps.sessionStore.EXPECT().DeleteSession(gomock.Any(), "session-uuid-456").Return(nil)

	_, _, err := deps.usecases.RefreshToken(ctx, refreshToken)
	require.Error(t, err)

	var appErr *errs.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, errs.ErrorUnauthorized, appErr.Code)
}

func TestRefreshToken_HashMismatch(t *testing.T) {
	deps := setupTest(t)
	ctx := context.Background()

	refreshToken, _ := deps.jwtService.CreateRefreshToken("user-uuid-123", "session-uuid-456", []string{})

	// Use a different hash (not matching the token)
	hashParams := argon2id.New(
		argon2id.Iterations(1),
		argon2id.Memory(64*1024),
		argon2id.Parallelism(1),
		argon2id.KeyLength(32),
		argon2id.SaltLength(16),
	)
	differentHash, _ := argon2id.Hash("different-token", hashParams)

	session := &models.Session{
		UUID:         "session-uuid-456",
		UserID:       "user-uuid-123",
		RefreshToken: differentHash,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	deps.sessionStore.EXPECT().GetSessionByUUID(gomock.Any(), "session-uuid-456").Return(session, nil)

	_, _, err := deps.usecases.RefreshToken(ctx, refreshToken)
	require.Error(t, err)

	var appErr *errs.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, errs.ErrorUnauthorized, appErr.Code)
}

func TestRefreshToken_UserNotFound(t *testing.T) {
	deps := setupTest(t)
	ctx := context.Background()

	refreshToken, _ := deps.jwtService.CreateRefreshToken("user-uuid-123", "session-uuid-456", []string{})

	hashParams := argon2id.New(
		argon2id.Iterations(1),
		argon2id.Memory(64*1024),
		argon2id.Parallelism(1),
		argon2id.KeyLength(32),
		argon2id.SaltLength(16),
	)
	hashedToken, _ := argon2id.Hash(refreshToken, hashParams)

	session := &models.Session{
		UUID:         "session-uuid-456",
		UserID:       "user-uuid-123",
		RefreshToken: hashedToken,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	deps.sessionStore.EXPECT().GetSessionByUUID(gomock.Any(), "session-uuid-456").Return(session, nil)
	deps.identityRepo.EXPECT().GetUserByUUID(gomock.Any(), "user-uuid-123").
		Return(nil, errs.New(errs.ErrorNotFound, "user not found", nil))

	_, _, err := deps.usecases.RefreshToken(ctx, refreshToken)
	require.Error(t, err)

	var appErr *errs.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, errs.ErrorUnauthorized, appErr.Code)
}

func TestRefreshToken_DeleteOldSessionError(t *testing.T) {
	deps := setupTest(t)
	ctx := context.Background()

	refreshToken, _ := deps.jwtService.CreateRefreshToken("user-uuid-123", "session-uuid-456", []string{})

	hashParams := argon2id.New(
		argon2id.Iterations(1),
		argon2id.Memory(64*1024),
		argon2id.Parallelism(1),
		argon2id.KeyLength(32),
		argon2id.SaltLength(16),
	)
	hashedToken, _ := argon2id.Hash(refreshToken, hashParams)

	session := &models.Session{
		UUID:         "session-uuid-456",
		UserID:       "user-uuid-123",
		RefreshToken: hashedToken,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	user := &identitymodels.User{UUID: "user-uuid-123"}

	deps.sessionStore.EXPECT().GetSessionByUUID(gomock.Any(), "session-uuid-456").Return(session, nil)
	deps.identityRepo.EXPECT().GetUserByUUID(gomock.Any(), "user-uuid-123").Return(user, nil)
	deps.serviceStore.EXPECT().GetUserServices(gomock.Any(), "user-uuid-123").Return([]string{}, nil)
	deps.sessionStore.EXPECT().DeleteSession(gomock.Any(), "session-uuid-456").
		Return(errs.New(errs.ErrorInternal, "db error", nil))

	_, _, err := deps.usecases.RefreshToken(ctx, refreshToken)
	require.Error(t, err)

	var appErr *errs.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, errs.ErrorInternal, appErr.Code)
}
