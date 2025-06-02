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

	// Validate issuer if present
	if iss, ok := claims[claimIss].(string); ok && s.config.ExpectedIssuer != "" {
		if iss != s.config.ExpectedIssuer {
			return WrapError(
				ErrInvalidIssuer,
				fmt.Sprintf("expected %s, got %s", s.config.ExpectedIssuer, iss),
			)
		}
	}

	// Validate audience if present
	if aud, ok := claims[claimAud].(string); ok {
		valid, err := s.validateAudience(aud)
		if err != nil {
			return WrapError(err, "failed to validate audience")
		}
		if !valid {
			return WrapError(
				ErrInvalidAudience,
				fmt.Sprintf("audience %s not in allowed list", aud),
			)
		}
	} else if audArray, ok := claims[claimAud].([]interface{}); ok {
		valid := false
		for _, a := range audArray {
			audStr, ok := a.(string)
			if !ok {
				continue
			}

			isValid, err := s.validateAudience(audStr)
			if err != nil {
				return WrapError(err, "failed to validate audience")
			}
			if isValid {
				valid = true
				break
			}
		}
		if !valid {
			return WrapError(ErrInvalidAudience, "no valid audience found in token")
		}
	}

	// Validate expiration time (required)
	if exp, ok := claims[claimExp].(float64); !ok {
		return WrapError(ErrMissingExpiration, "")
	} else {
		expTime := time.Unix(int64(exp), 0)
		if now.After(expTime.Add(clockSkew)) {
			return WrapError(ErrTokenExpired, "")
		}
	}

	// Validate issued at if present
	if iat, ok := claims[claimIAT].(float64); ok {
		issuedAt := time.Unix(int64(iat), 0)
		if now.Before(issuedAt.Add(-clockSkew)) {
			return WrapError(
				ErrMissingIssuedAt,
				"token used before issued",
			)
		}
	}

	// Validate not before if present
	if nbf, ok := claims[claimNbf].(float64); ok {
		notBefore := time.Unix(int64(nbf), 0)
		if now.Before(notBefore.Add(-clockSkew)) {
			return WrapError(ErrTokenNotValidYet, "")
		}
	}

	return nil
}
