package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Service implements the TokenService interface
type Service struct {
	config tokenConfig
}

// NewJWTService creates a new JWT service with the given configuration
func NewJWTService(config tokenConfig) *Service {
	return &Service{
		config: config,
	}
}

// CreateAccessToken creates a new access token
func (s *Service) CreateAccessToken(userUUID string, permissions, audiences []string) (string, error) {
	expT := jwt.NewNumericDate(time.Now().Add(time.Duration(s.config.AccessTokenDuration) * time.Second))

	claims := AccessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: expT,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    s.config.ExpectedIssuer,
			Audience:  audiences,
		},
		UserUUID:    userUUID,
		TokenType:   string(AccessToken),
		Permissions: permissions,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	ss, err := token.SignedString([]byte(s.config.SecretKey))
	if err != nil {
		return "", WrapError(err, "failed to sign access token")
	}

	return ss, nil
}

// CreateRefreshToken creates a refresh token with session info
func (s *Service) CreateRefreshToken(userUUID, sessionUUID string, audiences []string) (string, error) {
	expT := jwt.NewNumericDate(time.Now().Add(time.Duration(s.config.RefreshTokenDuration) * time.Second))

	claims := RefreshTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: expT,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    s.config.ExpectedIssuer,
			Audience:  audiences,
		},
		UserUUID:    userUUID,
		TokenType:   string(RefreshToken),
		SessionUUID: sessionUUID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	ss, err := token.SignedString([]byte(s.config.SecretKey))
	if err != nil {
		return "", WrapError(err, "failed to sign refresh token")
	}

	return ss, nil
}

// ParseAccessToken parses and validates an access token
func (s *Service) ParseAccessToken(tokenString string) (*AccessClaims, error) {
	claims, err := s.parseAndValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if err := s.validateAccessTokenType(claims); err != nil {
		return nil, err
	}

	userUUID, err := s.extractUserUUID(claims)
	if err != nil {
		return nil, err
	}

	permissions := s.extractPermissions(claims)

	return s.buildAccessClaims(claims, userUUID, permissions), nil
}

// parseAndValidateToken handles the common token parsing and validation
func (s *Service) parseAndValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, s.keyFunc)
	if err != nil {
		return nil, handleParseError(err, "access token")
	}

	if !token.Valid {
		return nil, WrapError(ErrInvalidToken, "")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, WrapError(ErrInvalidClaimsType, "")
	}

	return claims, s.validateStandardClaims(claims)
}

// validateAccessTokenType verifies the token type is access token
func (_ *Service) validateAccessTokenType(claims jwt.MapClaims) error {
	tokenType, ok := claims[claimTokenType].(string)
	if !ok {
		return WrapError(ErrMissingTokenType, "")
	}

	if TokenType(tokenType) != AccessToken {
		return WrapError(ErrNotAccessToken, "")
	}

	return nil
}

// extractUserUUID extracts and validates the user UUID from claims
func (_ *Service) extractUserUUID(claims jwt.MapClaims) (string, error) {
	userUUID, ok := claims[claimUserUUID].(string)
	if !ok || userUUID == "" {
		return "", WrapError(ErrMissingUserUUID, "")
	}

	return userUUID, nil
}

// extractPermissions safely extracts permissions from claims
func (_ *Service) extractPermissions(claims jwt.MapClaims) []string {
	var permissions []string

	perms, ok := claims[claimPermissions].([]any)
	if !ok {
		return permissions // Return empty slice if permissions claim is missing or wrong type
	}

	for _, p := range perms {
		if perm, ok := p.(string); ok {
			permissions = append(permissions, perm)
		}
		// Silently skip invalid permission entries rather than failing
	}

	return permissions
}

// buildAccessClaims constructs the final AccessClaims object
func (_ *Service) buildAccessClaims(claims jwt.MapClaims, userUUID string, permissions []string) *AccessClaims {
	return &AccessClaims{
		BaseClaims: BaseClaims{
			UserUUID:  userUUID,
			TokenType: AccessToken,
			Exp:       extractInt64Claim(claims, claimExp),
			IssuedAt:  extractTimeClaim(claims, claimIAT),
			Issuer:    extractStringClaim(claims, claimIss),
			Audience:  extractStringClaim(claims, claimAud),
		},
		Permissions: permissions,
	}
}

// ParseRefreshToken parses and validates a refresh token
func (s *Service) ParseRefreshToken(tokenString string) (*RefreshClaims, error) {
	claims, err := s.parseAndValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if err := s.validateRefreshTokenType(claims); err != nil {
		return nil, err
	}

	userUUID, err := s.extractUserUUID(claims)
	if err != nil {
		return nil, err
	}

	sessionUUID, err := s.extractSessionUUID(claims)
	if err != nil {
		return nil, err
	}

	return s.buildRefreshClaims(claims, userUUID, sessionUUID), nil
}

// validateRefreshTokenType verifies the token type is refresh token
func (_ *Service) validateRefreshTokenType(claims jwt.MapClaims) error {
	tokenType, ok := claims[claimTokenType].(string)
	if !ok {
		return WrapError(ErrMissingTokenType, "")
	}

	if TokenType(tokenType) != RefreshToken {
		return WrapError(ErrNotRefreshToken, "")
	}

	return nil
}

// extractSessionUUID extracts and validates the session UUID from claims
func (_ *Service) extractSessionUUID(claims jwt.MapClaims) (string, error) {
	sessionUUID, ok := claims[claimSessionUUID].(string)
	if !ok || sessionUUID == "" {
		return "", WrapError(ErrMissingSessionUUID, "")
	}

	return sessionUUID, nil
}

// buildRefreshClaims constructs the final RefreshClaims object
func (_ *Service) buildRefreshClaims(claims jwt.MapClaims, userUUID, sessionUUID string) *RefreshClaims {
	return &RefreshClaims{
		BaseClaims: BaseClaims{
			UserUUID:  userUUID,
			TokenType: RefreshToken,
			Exp:       extractInt64Claim(claims, claimExp),
			IssuedAt:  extractTimeClaim(claims, claimIAT),
			Issuer:    extractStringClaim(claims, claimIss),
			Audience:  extractStringClaim(claims, claimAud),
		},
		SessionUUID: sessionUUID,
	}
}

// ParseToken parses a token and returns base claims without validating token type
func (s *Service) ParseToken(tokenString string) (*BaseClaims, error) {
	token, err := jwt.Parse(tokenString, s.keyFunc)
	if err != nil {
		return nil, handleParseError(err, "token")
	}

	if !token.Valid {
		return nil, WrapError(ErrInvalidToken, "")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, WrapError(ErrInvalidClaimsType, "")
	}

	// Validate standard claims
	if err := s.validateStandardClaims(claims); err != nil {
		return nil, err
	}

	// Extract token type with validation
	tokenType, ok := claims[claimTokenType].(string)
	if !ok {
		return nil, WrapError(ErrMissingTokenType, "")
	}

	userUUID, ok := claims[claimUserUUID].(string)
	if !ok || userUUID == "" {
		return nil, WrapError(ErrMissingUserUUID, "")
	}

	// Create base claims
	baseClaims := &BaseClaims{
		UserUUID:  userUUID,
		TokenType: TokenType(tokenType),
		Exp:       extractInt64Claim(claims, claimExp),
		IssuedAt:  extractTimeClaim(claims, claimIAT),
		Issuer:    extractStringClaim(claims, claimIss),
		Audience:  extractStringClaim(claims, claimAud),
	}

	return baseClaims, nil
}

// keyFunc validates the signing method and gets the key for verification
func (s *Service) keyFunc(token *jwt.Token) (any, error) {
	if s.config.SecretKey == "" {
		return nil, ErrMissingSecretKey
	}

	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, WrapError(
			ErrInvalidSigningAlg,
			"unexpected signing method: "+token.Header["alg"].(string),
		)
	}

	return []byte(s.config.SecretKey), nil
}

// handleParseError handles common parse errors with better context
func handleParseError(err error, tokenType string) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, jwt.ErrTokenExpired) {
		return ErrTokenExpired
	}

	if errors.Is(err, jwt.ErrTokenMalformed) {
		return ErrInvalidTokenFormat
	}

	return WrapError(err, "failed to parse "+tokenType)
}
