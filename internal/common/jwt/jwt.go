package jwt

import (
	"MydroX/anicetus/pkg/argon2id"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenJWT string

type JWTError string

var (
	JWTNoError             JWTError
	JWTExpiredToken        JWTError = "expired token"
	JWTInvalidToken        JWTError = "invalid token"
	JWTSigInvalid          JWTError = "signature invalid"
	JWTNoToken             JWTError = "no token"
	JWTUnexpectedSigMethod JWTError = "unexpected signing method"
)

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func CreateAccessToken(expirationTime time.Time, secretKey, userID string) (string, error) {
	expT := jwt.NewNumericDate(expirationTime)

	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: expT,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	ss, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return ss, nil
}

func CreateRefreshToken(expirationTime time.Time, secretKey, userID string) (string, error) {
	expT := jwt.NewNumericDate(expirationTime)

	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: expT,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	ss, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return ss, nil
}

func ParseToken(tokenString string) (*Claims, JWTError) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, keyFunc)

	if token.Valid {
		return claims, JWTNoError
	}

	switch {
	case errors.Is(err, jwt.ErrTokenMalformed):
		return nil, JWTInvalidToken
	case errors.Is(err, jwt.ErrTokenSignatureInvalid):
		return nil, JWTSigInvalid
	case errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet):
		return nil, JWTExpiredToken
	}

	return nil, JWTInvalidToken
}

func (t TokenJWT) String() string {
	return string(t)
}

// HashArgon2 is a function to hash the token with a argon2 algorithm.
func (t TokenJWT) HashArgon2(params *argon2id.HashParams) (string, error) {
	return argon2id.Hash(t.String(), params)
}

func keyFunc(token *jwt.Token) (any, error) {
	secretKey := os.Getenv("JWT_SECRET")

	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	return secretKey, nil
}
