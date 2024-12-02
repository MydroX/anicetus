package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func CreateToken(expirationTime time.Time, secretKey string, userID string) (string, error) {
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

func VerifyToken(tokenString string) error {
	token, err := jwt.Parse("token", func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return jwt.ErrSignatureInvalid
	}
	return nil
}
