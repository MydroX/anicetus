package hmac

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

func HashWithSalt(input, saltKey string) (string, error) {
	key, err := base64.StdEncoding.DecodeString(saltKey)
	if err != nil {
		return "", err
	}

	h := hmac.New(sha256.New, key)
	_, err = h.Write([]byte(input))
	if err != nil {
		return "", err
	}
	hash := h.Sum(nil)

	return base64.StdEncoding.EncodeToString(hash), nil
}
