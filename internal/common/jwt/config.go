package jwt

// Constants for claims
const (
	claimTokenType   = "token_type"
	claimUserUUID    = "user_uuid"
	claimSessionUUID = "session_uuid"
	claimPermissions = "permissions"
	claimExp         = "exp"
	claimIAT         = "iat"
	claimIss         = "iss"
	claimAud         = "aud"
	claimNbf         = "nbf"
)

// Config constants
const (
	// JWTSecretEnvVar is the environment variable name for JWT secret
	JWTSecretEnvVar = "JWT_SECRET"

	// Default clock skew in seconds
	DefaultClockSkewSeconds = 60

	// Time-related constants
	SecondsInMinute = 60
	MinutesInHour   = 60
	HoursInDay      = 24

	// Default token durations (in seconds)
	DefaultAccessTokenDuration  = 15 * SecondsInMinute                             // 15 minutes
	DefaultRefreshTokenDuration = 7 * HoursInDay * SecondsInMinute * MinutesInHour // 7 days
)

// TokenConfig holds configuration for the JWT service
type TokenConfig struct {
	// Secret key used for signing tokens
	SecretKey string

	// ExpectedIssuer is the expected issuer claim value
	ExpectedIssuer string

	// ExpectedAudiences is a list of valid audience values
	ExpectedAudiences []string

	// ClockSkewSeconds is the tolerance for time-based claims validation
	ClockSkewSeconds int

	// AccessTokenDuration is the default duration of access tokens in seconds
	AccessTokenDuration int

	// RefreshTokenDuration is the default duration of refresh tokens in seconds
	RefreshTokenDuration int
}

// NewDefaultConfig creates a new TokenConfig with default values
func NewDefaultConfig(secretKey, issuer string) TokenConfig {
	return TokenConfig{
		SecretKey:            secretKey,
		ExpectedIssuer:       issuer,
		ExpectedAudiences:    []string{issuer},
		ClockSkewSeconds:     DefaultClockSkewSeconds,
		AccessTokenDuration:  DefaultAccessTokenDuration,
		RefreshTokenDuration: DefaultRefreshTokenDuration,
	}
}
