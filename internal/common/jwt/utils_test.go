package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestExtractStringClaim(t *testing.T) {
	claims := jwt.MapClaims{"test": "value"}
	result := extractStringClaim(claims, "test")
	assert.Equal(t, "value", result)

	// Missing claim
	result = extractStringClaim(claims, "missing")
	assert.Equal(t, "", result)

	// Wrong type
	claims["wrong"] = 123
	result = extractStringClaim(claims, "wrong")
	assert.Equal(t, "", result)
}

func TestExtractInt64Claim(t *testing.T) {
	claims := jwt.MapClaims{"test": float64(123)}
	result := extractInt64Claim(claims, "test")
	assert.Equal(t, int64(123), result)

	// Missing claim
	result = extractInt64Claim(claims, "missing")
	assert.Equal(t, int64(0), result)
}

func TestExtractTimeClaim(t *testing.T) {
	timestamp := float64(1609459200) // 2021-01-01 00:00:00 UTC
	claims := jwt.MapClaims{"test": timestamp}
	result := extractTimeClaim(claims, "test")
	expected := time.Unix(1609459200, 0)
	assert.True(t, expected.Equal(result))

	// Missing claim
	result = extractTimeClaim(claims, "missing")
	assert.True(t, time.Time{}.Equal(result))
}

func TestValidateStandardClaims(t *testing.T) {
	now := time.Now()
	service := &Service{
		config: TokenConfig{
			ClockSkewSeconds: 60,
			ExpectedIssuer:   "test-issuer",
		},
	}

	// Valid token
	claims := jwt.MapClaims{
		claimIss: "test-issuer",
		claimExp: float64(now.Add(time.Hour).Unix()),
		claimIAT: float64(now.Add(-time.Minute).Unix()),
	}
	err := service.validateStandardClaims(claims)
	assert.NoError(t, err)

	// Expired token
	claims[claimExp] = float64(now.Add(-time.Hour).Unix())
	err = service.validateStandardClaims(claims)
	assert.Error(t, err)

	// Wrong issuer
	claims[claimExp] = float64(now.Add(time.Hour).Unix())
	claims[claimIss] = "wrong-issuer"
	err = service.validateStandardClaims(claims)
	assert.Error(t, err)

	// Token issued in future (beyond clock skew)
	claims[claimIss] = "test-issuer"
	claims[claimIAT] = float64(now.Add(2 * time.Minute).Unix())
	err = service.validateStandardClaims(claims)
	assert.Error(t, err)

	// Token not valid yet (nbf in future)
	claims[claimIAT] = float64(now.Add(-time.Minute).Unix())
	claims[claimNbf] = float64(now.Add(2 * time.Minute).Unix())
	err = service.validateStandardClaims(claims)
	assert.Error(t, err)
}

func TestValidateExpiration(t *testing.T) {
	now := time.Now()
	clockSkew := 60 * time.Second

	// Valid expiration
	claims := jwt.MapClaims{claimExp: float64(now.Add(time.Hour).Unix())}
	err := validateExpiration(claims, now, clockSkew)
	assert.NoError(t, err)

	// Expired token
	claims[claimExp] = float64(now.Add(-time.Hour).Unix())
	err = validateExpiration(claims, now, clockSkew)
	assert.Error(t, err)

	// Missing expiration
	delete(claims, claimExp)
	err = validateExpiration(claims, now, clockSkew)
	assert.Error(t, err)
}

func TestValidateIssuer(t *testing.T) {
	service := &Service{
		config: TokenConfig{ExpectedIssuer: "test-issuer"},
	}

	// Valid issuer
	claims := jwt.MapClaims{claimIss: "test-issuer"}
	err := service.validateIssuer(claims)
	assert.NoError(t, err)

	// Wrong issuer
	claims[claimIss] = "wrong-issuer"
	err = service.validateIssuer(claims)
	assert.Error(t, err)

	// Missing issuer (allowed)
	delete(claims, claimIss)
	err = service.validateIssuer(claims)
	assert.NoError(t, err)
}

func TestValidateIssuedAt(t *testing.T) {
	now := time.Now()
	clockSkew := 60 * time.Second

	// Valid issued at - past time
	claims := jwt.MapClaims{claimIAT: float64(now.Add(-time.Hour).Unix())}
	err := validateIssuedAt(claims, now, clockSkew)
	assert.NoError(t, err)

	// Valid issued at - within clock skew
	claims[claimIAT] = float64(now.Add(30 * time.Second).Unix())
	err = validateIssuedAt(claims, now, clockSkew)
	assert.NoError(t, err)

	// Invalid - issued too far in future
	claims[claimIAT] = float64(now.Add(2 * time.Minute).Unix())
	err = validateIssuedAt(claims, now, clockSkew)
	assert.Error(t, err)

	// Missing issued at (should be valid)
	delete(claims, claimIAT)
	err = validateIssuedAt(claims, now, clockSkew)
	assert.NoError(t, err)
}

func TestValidateNotBefore(t *testing.T) {
	now := time.Now()
	clockSkew := 60 * time.Second

	// Valid not before - past time
	claims := jwt.MapClaims{claimNbf: float64(now.Add(-time.Hour).Unix())}
	err := validateNotBefore(claims, now, clockSkew)
	assert.NoError(t, err)

	// Valid not before - within clock skew
	claims[claimNbf] = float64(now.Add(30 * time.Second).Unix())
	err = validateNotBefore(claims, now, clockSkew)
	assert.NoError(t, err)

	// Invalid - not before too far in future
	claims[claimNbf] = float64(now.Add(2 * time.Minute).Unix())
	err = validateNotBefore(claims, now, clockSkew)
	assert.Error(t, err)

	// Missing not before (should be valid)
	delete(claims, claimNbf)
	err = validateNotBefore(claims, now, clockSkew)
	assert.NoError(t, err)
}

func TestSecurityEdgeCases(t *testing.T) {
	now := time.Now()
	service := &Service{
		config: TokenConfig{
			ClockSkewSeconds: 60,
			ExpectedIssuer:   "test-issuer",
		},
	}

	// Clock skew boundary - token expired just beyond allowed skew
	claims := jwt.MapClaims{
		claimIss: "test-issuer",
		claimExp: float64(now.Add(-61 * time.Second).Unix()),
		claimIAT: float64(now.Add(-time.Hour).Unix()),
	}
	err := service.validateStandardClaims(claims)
	assert.Error(t, err, "Should reject token expired beyond clock skew")

	// Type confusion attack - non-string issuer
	claims = jwt.MapClaims{
		claimIss: 123, // Not a string
		claimExp: float64(now.Add(time.Hour).Unix()),
		claimIAT: float64(now.Add(-time.Minute).Unix()),
	}
	err = service.validateStandardClaims(claims)
	assert.NoError(t, err, "Should ignore non-string issuer types")
}
