package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Service implements the TokenService interface
type Service struct {
	config           TokenConfig
	audienceProvider AudienceProvider
}

// NewJWTService creates a new JWT service with the given configuration
func NewJWTService(config TokenConfig, audienceProvider AudienceProvider) *Service {
	return &Service{
		config:           config,
		audienceProvider: audienceProvider,
	}
}

// CreateAccessToken creates a new access token
func (s *Service) CreateAccessToken(userUUID string, permissions []string) (string, error) {
	if s.config.SecretKey == "" {
		return "", WrapError(ErrMissingSecretKey, "")
	}

	if s.config.AccessTokenDuration <= 0 {
		s.config.AccessTokenDuration = DefaultAccessTokenDuration
	}

	expT := jwt.NewNumericDate(time.Now().Add(time.Duration(s.config.AccessTokenDuration) * time.Second))

	audiences, err := s.audienceProvider.GetAllowedAudiences()
	if err != nil {
		return "", WrapError(err, "failed to get allowed audiences")
	}

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
func (s *Service) CreateRefreshToken(userUUID string, sessionUUID string) (string, error) {
	if s.config.SecretKey == "" {
		return "", WrapError(ErrMissingSecretKey, "")
	}

	if s.config.RefreshTokenDuration <= 0 {
		s.config.RefreshTokenDuration = DefaultRefreshTokenDuration
	}

	expT := jwt.NewNumericDate(time.Now().Add(time.Duration(s.config.RefreshTokenDuration) * time.Second))

	audiences, err := s.audienceProvider.GetAllowedAudiences()
	if err != nil {
		return "", WrapError(err, "failed to get allowed audiences")
	}

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
	token, err := jwt.Parse(tokenString, s.keyFunc)
	if err != nil {
		return nil, s.handleParseError(err, "access token")
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

	// Verify this is an access token
	tokenType, ok := claims[claimTokenType].(string)
	if !ok {
		return nil, WrapError(ErrMissingTokenType, "")
	}

	if TokenType(tokenType) != AccessToken {
		return nil, WrapError(ErrNotAccessToken, "")
	}

	userUUID, ok := claims[claimUserUUID].(string)
	if !ok || userUUID == "" {
		return nil, WrapError(ErrMissingUserUUID, "")
	}

	// Extract permissions
	var permissions []string
	if perms, ok := claims[claimPermissions].([]any); ok {
		for _, p := range perms {
			if perm, ok := p.(string); ok {
				permissions = append(permissions, perm)
			}
		}
	}

	// Build claims object
	accessClaims := &AccessClaims{
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

	return accessClaims, nil
}

// ParseRefreshToken parses and validates a refresh token
func (s *Service) ParseRefreshToken(tokenString string) (*RefreshClaims, error) {
	token, err := jwt.Parse(tokenString, s.keyFunc)
	if err != nil {
		return nil, s.handleParseError(err, "refresh token")
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

	// Verify this is a refresh token
	tokenType, ok := claims[claimTokenType].(string)
	if !ok {
		return nil, WrapError(ErrMissingTokenType, "")
	}

	if TokenType(tokenType) != RefreshToken {
		return nil, WrapError(ErrNotRefreshToken, "")
	}

	// Extract required fields with validation
	sessionUUID, ok := claims[claimSessionUUID].(string)
	if !ok || sessionUUID == "" {
		return nil, WrapError(ErrMissingSessionUUID, "")
	}

	userUUID, ok := claims[claimUserUUID].(string)
	if !ok || userUUID == "" {
		return nil, WrapError(ErrMissingUserUUID, "")
	}

	// Build claims object
	refreshClaims := &RefreshClaims{
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

	return refreshClaims, nil
}

// ParseToken parses a token and returns base claims without validating token type
func (s *Service) ParseToken(tokenString string) (*BaseClaims, error) {
	token, err := jwt.Parse(tokenString, s.keyFunc)
	if err != nil {
		return nil, s.handleParseError(err, "token")
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

// validateAudience checks if the provided audience is allowed
func (s *Service) validateAudience(audience string) (bool, error) {
	// If no audience provider is available, use the configured audiences
	if s.audienceProvider == nil {
		for _, allowed := range s.config.ExpectedAudiences {
			if audience == allowed {
				return true, nil
			}
		}
		return false, nil
	}

	// Otherwise, check against dynamic audiences from the provider
	allowed, err := s.audienceProvider.GetAllowedAudiences()
	if err != nil {
		return false, err
	}

	for _, a := range allowed {
		if audience == a {
			return true, nil
		}
	}

	return false, nil
}

// keyFunc validates the signing method and gets the key for verification
func (s *Service) keyFunc(token *jwt.Token) (interface{}, error) {
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
func (s *Service) handleParseError(err error, tokenType string) error {
	if err == nil {
		return nil
	}

	if err == jwt.ErrTokenExpired {
		return WrapError(ErrTokenExpired, "")
	}
	if err == jwt.ErrTokenMalformed {
		return WrapError(ErrInvalidTokenFormat, "")
	}
	return WrapError(err, "failed to parse "+tokenType)
}
