package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Helper functions for extracting claims
func extractStringClaim(claims jwt.MapClaims, key string) string {
	if val, ok := claims[key].(string); ok {
		return val
	}

	return ""
}

func extractInt64Claim(claims jwt.MapClaims, key string) int64 {
	if val, ok := claims[key].(float64); ok {
		return int64(val)
	}

	return 0
}

//nolint:unparam // key is parameterized for consistency with other extract functions
func extractStringSliceClaim(
	claims jwt.MapClaims,
	key string,
) []string {
	switch v := claims[key].(type) {
	case []any:
		var result []string

		for _, item := range v {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}

		return result
	case string:
		return []string{v}
	default:
		return nil
	}
}

func extractTimeClaim(claims jwt.MapClaims, key string) time.Time {
	if val, ok := claims[key].(float64); ok {
		return time.Unix(int64(val), 0)
	}

	return time.Time{}
}

// validateStandardClaims validates standard JWT claims
func (s *Service) validateStandardClaims(claims jwt.MapClaims) error {
	now := time.Now()
	clockSkew := time.Duration(s.config.ClockSkewSeconds) * time.Second

	// Validate each claim type
	if err := s.validateIssuer(claims); err != nil {
		return err
	}

	if err := s.validateAudience(claims); err != nil {
		return err
	}

	if err := validateExpiration(claims, now, clockSkew); err != nil {
		return err
	}

	if err := validateIssuedAt(claims, now, clockSkew); err != nil {
		return err
	}

	return validateNotBefore(claims, now, clockSkew)
}

// validateIssuer validates the issuer claim if present
func (s *Service) validateIssuer(claims jwt.MapClaims) error {
	if iss, ok := claims[claimIss].(string); ok && s.config.ExpectedIssuer != "" {
		if iss != s.config.ExpectedIssuer {
			return WrapError(
				ErrInvalidIssuer,
				fmt.Sprintf("expected %s, got %s", s.config.ExpectedIssuer, iss),
			)
		}
	}

	return nil
}

// validateAudience validates the audience claim against expected audiences
func (s *Service) validateAudience(claims jwt.MapClaims) error {
	if len(s.config.ExpectedAudiences) == 0 {
		return nil
	}

	tokenAudiences := extractStringSliceClaim(claims, claimAud)
	if len(tokenAudiences) == 0 {
		return WrapError(ErrInvalidAudience, "token has no audience claim")
	}

	for _, tokenAud := range tokenAudiences {
		for _, expectedAud := range s.config.ExpectedAudiences {
			if tokenAud == expectedAud {
				return nil
			}
		}
	}

	return WrapError(ErrInvalidAudience, "token audience does not match any expected audience")
}

// validateExpiration validates the expiration claim
func validateExpiration(claims jwt.MapClaims, now time.Time, clockSkew time.Duration) error {
	exp, ok := claims[claimExp].(float64)
	if !ok {
		return WrapError(ErrMissingExpiration, "")
	}

	expTime := time.Unix(int64(exp), 0)
	if now.After(expTime.Add(clockSkew)) {
		return WrapError(ErrTokenExpired, "")
	}

	return nil
}

// validateIssuedAt validates the issued at claim if present
func validateIssuedAt(claims jwt.MapClaims, now time.Time, clockSkew time.Duration) error {
	if iat, ok := claims[claimIAT].(float64); ok {
		issuedAt := time.Unix(int64(iat), 0)
		if now.Before(issuedAt.Add(-clockSkew)) {
			return WrapError(
				ErrMissingIssuedAt,
				"token used before issued",
			)
		}
	}

	return nil
}

// validateNotBefore validates the not before claim if present
func validateNotBefore(claims jwt.MapClaims, now time.Time, clockSkew time.Duration) error {
	if nbf, ok := claims[claimNbf].(float64); ok {
		notBefore := time.Unix(int64(nbf), 0)
		if now.Before(notBefore.Add(-clockSkew)) {
			return WrapError(ErrTokenNotValidYet, "")
		}
	}

	return nil
}
