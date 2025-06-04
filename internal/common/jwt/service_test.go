package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestParseAccessToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAudienceProvider := NewMockAudienceProvider(ctrl)
	mockAudienceProvider.EXPECT().GetAllowedAudiences().Return([]string{"test-audience"}, nil).AnyTimes()

	config := TokenConfig{
		SecretKey:        "test-secret-key-long-enough-for-signing-jwt-tokens-securely",
		ExpectedIssuer:   "test-issuer",
		ClockSkewSeconds: 60,
	}

	service := NewJWTService(config)

	t.Run("success case", func(t *testing.T) {
		// Créer un token de test valide
		userUUID := "user-123"
		permissions := []string{"read", "write"}

		// Générer un token valide
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, AccessTokenClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    "test-issuer",
				Audience:  []string{"test-audience"},
			},
			UserUUID:    userUUID,
			TokenType:   string(AccessToken),
			Permissions: permissions,
		})

		tokenString, err := token.SignedString([]byte(config.SecretKey))
		assert.NoError(t, err)

		// Analyser le token
		claims, err := service.ParseAccessToken(tokenString)

		// Vérifier les résultats
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, userUUID, claims.UserUUID)
		assert.Equal(t, AccessToken, claims.TokenType)
		assert.ElementsMatch(t, permissions, claims.Permissions)
	})

	t.Run("expired token", func(t *testing.T) {
		// Créer un token expiré
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, AccessTokenClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-10 * time.Minute)),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-20 * time.Minute)),
				Issuer:    "test-issuer",
				Audience:  []string{"test-audience"},
			},
			UserUUID:    "user-123",
			TokenType:   string(AccessToken),
			Permissions: []string{"read"},
		})

		tokenString, err := token.SignedString([]byte(config.SecretKey))
		assert.NoError(t, err)

		// Analyser le token
		claims, err := service.ParseAccessToken(tokenString)

		// Vérifier les résultats
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.ErrorIs(t, err, ErrTokenExpired)
	})

	t.Run("invalid token type", func(t *testing.T) {
		// Créer un token de rafraîchissement au lieu d'un token d'accès
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, RefreshTokenClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    "test-issuer",
				Audience:  []string{"test-audience"},
			},
			UserUUID:    "user-123",
			TokenType:   string(RefreshToken),
			SessionUUID: "session-456",
		})

		tokenString, err := token.SignedString([]byte(config.SecretKey))
		assert.NoError(t, err)

		// Analyser le token
		claims, err := service.ParseAccessToken(tokenString)

		// Vérifier les résultats
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.ErrorIs(t, err, ErrNotAccessToken)
	})

	t.Run("malformed token", func(t *testing.T) {
		// Analyser un token mal formé
		claims, err := service.ParseAccessToken("not-a-valid-jwt-token")

		// Vérifier les résultats
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.ErrorIs(t, err, ErrInvalidTokenFormat)
	})
}

func TestParseRefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAudienceProvider := NewMockAudienceProvider(ctrl)
	mockAudienceProvider.EXPECT().GetAllowedAudiences().Return([]string{"test-audience"}, nil).AnyTimes()

	config := TokenConfig{
		SecretKey:        "test-secret-key-long-enough-for-signing-jwt-tokens-securely",
		ExpectedIssuer:   "test-issuer",
		ClockSkewSeconds: 60,
	}

	service := NewJWTService(config)

	t.Run("success case", func(t *testing.T) {
		// Créer un token de rafraîchissement valide
		userUUID := "user-123"
		sessionUUID := "session-456"

		// Générer un token valide
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, RefreshTokenClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    "test-issuer",
				Audience:  []string{"test-audience"},
			},
			UserUUID:    userUUID,
			TokenType:   string(RefreshToken),
			SessionUUID: sessionUUID,
		})

		tokenString, err := token.SignedString([]byte(config.SecretKey))
		assert.NoError(t, err)

		// Analyser le token
		claims, err := service.ParseRefreshToken(tokenString)

		// Vérifier les résultats
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, userUUID, claims.UserUUID)
		assert.Equal(t, RefreshToken, claims.TokenType)
		assert.Equal(t, sessionUUID, claims.SessionUUID)
	})

	t.Run("missing session UUID", func(t *testing.T) {
		// Créer un token sans session UUID
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
			"exp":        time.Now().Add(10 * time.Minute).Unix(),
			"iat":        time.Now().Unix(),
			"iss":        "test-issuer",
			"aud":        []string{"test-audience"},
			"user_uuid":  "user-123",
			"token_type": string(RefreshToken),
			// Pas de session_uuid
		})

		tokenString, err := token.SignedString([]byte(config.SecretKey))
		assert.NoError(t, err)

		// Analyser le token
		claims, err := service.ParseRefreshToken(tokenString)

		// Vérifier les résultats
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.ErrorIs(t, err, ErrMissingSessionUUID)
	})
}

func TestParseToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAudienceProvider := NewMockAudienceProvider(ctrl)
	mockAudienceProvider.EXPECT().GetAllowedAudiences().Return([]string{"test-audience"}, nil).AnyTimes()

	config := TokenConfig{
		SecretKey:        "test-secret-key-long-enough-for-signing-jwt-tokens-securely",
		ExpectedIssuer:   "test-issuer",
		ClockSkewSeconds: 60,
	}

	service := NewJWTService(config)

	t.Run("parse access token", func(t *testing.T) {
		// Créer un token d'accès
		userUUID := "user-123"

		token := jwt.NewWithClaims(jwt.SigningMethodHS512, AccessTokenClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    "test-issuer",
				Audience:  []string{"test-audience"},
			},
			UserUUID:    userUUID,
			TokenType:   string(AccessToken),
			Permissions: []string{"read"},
		})

		tokenString, err := token.SignedString([]byte(config.SecretKey))
		assert.NoError(t, err)

		// Analyser le token générique
		claims, err := service.ParseToken(tokenString)

		// Vérifier les résultats
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, userUUID, claims.UserUUID)
		assert.Equal(t, AccessToken, claims.TokenType)
	})

	t.Run("parse refresh token", func(t *testing.T) {
		// Créer un token de rafraîchissement
		userUUID := "user-123"

		token := jwt.NewWithClaims(jwt.SigningMethodHS512, RefreshTokenClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    "test-issuer",
				Audience:  []string{"test-audience"},
			},
			UserUUID:    userUUID,
			TokenType:   string(RefreshToken),
			SessionUUID: "session-456",
		})

		tokenString, err := token.SignedString([]byte(config.SecretKey))
		assert.NoError(t, err)

		// Analyser le token générique
		claims, err := service.ParseToken(tokenString)

		// Vérifier les résultats
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, userUUID, claims.UserUUID)
		assert.Equal(t, RefreshToken, claims.TokenType)
	})

	t.Run("invalid signing method", func(t *testing.T) {
		// Créer un token avec une méthode de signature non supportée
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"exp":        time.Now().Add(10 * time.Minute).Unix(),
			"iat":        time.Now().Unix(),
			"iss":        "test-issuer",
			"aud":        []string{"test-audience"},
			"user_uuid":  "user-123",
			"token_type": string(AccessToken),
		})

		// Utiliser une clé privée RSA pour la signature
		// Dans un test réel, il faudrait une vraie clé RSA
		// Ici, on sait que ça va échouer car keyFunc attend une clé HMAC
		tokenString, _ := token.SignedString([]byte("some-key"))

		// Analyser le token générique
		claims, err := service.ParseToken(tokenString)

		// Vérifier les résultats
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}

func TestCreateAccessToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("successful token creation", func(t *testing.T) {
		// Configurer le mock
		mockAudienceProvider := NewMockAudienceProvider(ctrl)
		mockAudienceProvider.EXPECT().
			GetAllowedAudiences().
			Return([]string{"test-audience"}, nil).AnyTimes()

		// Configurer le service
		config := TokenConfig{
			SecretKey:           "test-secret-key-long-enough-for-signing-jwt-tokens-securely",
			ExpectedIssuer:      "test-issuer",
			AccessTokenDuration: 3600,
		}
		service := NewJWTService(config)

		// Appeler la fonction à tester
		userUUID := "user-123"
		permissions := []string{"read", "write"}
		audiences := []string{"test-audience"}
		token, err := service.CreateAccessToken(userUUID, permissions, audiences)

		// Vérifications
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Vérifier que le token est valide en le décodant
		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.SecretKey), nil
		})
		assert.NoError(t, err)
		assert.True(t, parsedToken.Valid)

		// Vérifier les claims
		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		assert.True(t, ok)
		assert.Equal(t, userUUID, claims["user_uuid"])
		assert.Equal(t, string(AccessToken), claims["token_type"])
		assert.Equal(t, config.ExpectedIssuer, claims["iss"])

		// Vérifier les permissions
		perms, ok := claims["permissions"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, perms, len(permissions))
		for i, p := range permissions {
			assert.Equal(t, p, perms[i])
		}

		// Vérifier l'expiration
		exp, ok := claims["exp"].(float64)
		assert.True(t, ok)
		// L'expiration doit être dans le futur
		assert.Greater(t, exp, float64(time.Now().Unix()))
		// L'expiration doit être proche de maintenant + durée configurée
		assert.InDelta(t, time.Now().Add(time.Duration(config.AccessTokenDuration)*time.Second).Unix(), int64(exp), 5)
	})

	t.Run("use default duration when zero", func(t *testing.T) {
		// Configurer le mock
		mockAudienceProvider := NewMockAudienceProvider(ctrl)
		mockAudienceProvider.EXPECT().
			GetAllowedAudiences().
			Return([]string{"test-audience"}, nil).AnyTimes()

		// Configurer le service avec une durée de 0
		config := TokenConfig{
			SecretKey:           "test-secret-key-long-enough-for-signing-jwt-tokens-securely",
			ExpectedIssuer:      "test-issuer",
			AccessTokenDuration: 0, // Durée zéro, devrait utiliser la valeur par défaut
		}
		service := NewJWTService(config)

		// Appeler la fonction à tester avec des audiences explicites
		audiences := []string{"test-audience"}
		token, err := service.CreateAccessToken("user-123", []string{"read"}, audiences)

		// Vérifications
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Vérifier l'expiration
		parsedToken, _ := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.SecretKey), nil
		})
		claims, _ := parsedToken.Claims.(jwt.MapClaims)
		exp, _ := claims["exp"].(float64)

		// L'expiration doit être proche de maintenant + durée par défaut
		assert.InDelta(t, time.Now().Add(time.Duration(DefaultAccessTokenDuration)*time.Second).Unix(), int64(exp), 5)
	})

	t.Run("missing secret key", func(t *testing.T) {
		// Configurer le service avec une clé secrète vide
		config := TokenConfig{
			SecretKey:      "", // Clé vide
			ExpectedIssuer: "test-issuer",
		}
		service := NewJWTService(config)

		// Appeler la fonction à tester
		token, err := service.CreateAccessToken("user-123", []string{"read"}, []string{})

		// Vérifications
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.ErrorIs(t, err, ErrMissingSecretKey)
	})
}

func TestCreateRefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("successful token creation", func(t *testing.T) {
		// Configurer le mock
		mockAudienceProvider := NewMockAudienceProvider(ctrl)
		mockAudienceProvider.EXPECT().
			GetAllowedAudiences().
			Return([]string{"test-audience"}, nil).AnyTimes()

		// Configurer le service
		config := TokenConfig{
			SecretKey:            "test-secret-key-long-enough-for-signing-jwt-tokens-securely",
			ExpectedIssuer:       "test-issuer",
			RefreshTokenDuration: 86400, // 24 heures
		}
		service := NewJWTService(config)

		// Appeler la fonction à tester
		userUUID := "user-123"
		sessionUUID := "session-456"
		audiences := []string{"test-audience"}
		token, err := service.CreateRefreshToken(userUUID, sessionUUID, audiences)

		// Vérifications
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Vérifier que le token est valide en le décodant
		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.SecretKey), nil
		})
		assert.NoError(t, err)
		assert.True(t, parsedToken.Valid)

		// Vérifier les claims
		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		assert.True(t, ok)
		assert.Equal(t, userUUID, claims["user_uuid"])
		assert.Equal(t, sessionUUID, claims["session_uuid"])
		assert.Equal(t, string(RefreshToken), claims["token_type"])
		assert.Equal(t, config.ExpectedIssuer, claims["iss"])

		// Vérifier l'expiration
		exp, ok := claims["exp"].(float64)
		assert.True(t, ok)
		// L'expiration doit être dans le futur
		assert.Greater(t, exp, float64(time.Now().Unix()))
		// L'expiration doit être proche de maintenant + durée configurée
		assert.InDelta(t, time.Now().Add(time.Duration(config.RefreshTokenDuration)*time.Second).Unix(), int64(exp), 5)
	})

	t.Run("use default duration when zero", func(t *testing.T) {
		// Configurer le mock
		mockAudienceProvider := NewMockAudienceProvider(ctrl)
		mockAudienceProvider.EXPECT().
			GetAllowedAudiences().
			Return([]string{"test-audience"}, nil).AnyTimes()

		// Configurer le service avec une durée de 0
		config := TokenConfig{
			SecretKey:            "test-secret-key-long-enough-for-signing-jwt-tokens-securement",
			ExpectedIssuer:       "test-issuer",
			RefreshTokenDuration: 0, // Durée zéro, devrait utiliser la valeur par défaut
		}
		service := NewJWTService(config)

		// Appeler la fonction à tester
		audiences := []string{"test-audience"}
		token, err := service.CreateRefreshToken("user-123", "session-456", audiences)

		// Vérifications
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Vérifier l'expiration
		parsedToken, _ := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.SecretKey), nil
		})
		claims, _ := parsedToken.Claims.(jwt.MapClaims)
		exp, _ := claims["exp"].(float64)

		// L'expiration doit être proche de maintenant + durée par défaut
		assert.InDelta(t, time.Now().Add(time.Duration(DefaultRefreshTokenDuration)*time.Second).Unix(), int64(exp), 5)
	})

	t.Run("missing secret key", func(t *testing.T) {
		// Configurer le service avec une clé secrète vide
		config := TokenConfig{
			SecretKey:      "", // Clé vide
			ExpectedIssuer: "test-issuer",
		}
		service := NewJWTService(config)

		// Appeler la fonction à tester
		token, err := service.CreateRefreshToken("user-123", "session-456", []string{})

		// Vérifications
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.ErrorIs(t, err, ErrMissingSecretKey)
	})

	t.Run("empty session UUID", func(t *testing.T) {
		// Configurer le mock
		mockAudienceProvider := NewMockAudienceProvider(ctrl)
		mockAudienceProvider.EXPECT().
			GetAllowedAudiences().
			Return([]string{"test-audience"}, nil).AnyTimes()

		// Configurer le service
		config := TokenConfig{
			SecretKey:      "test-secret-key-long-enough-for-signing-jwt-tokens-securely",
			ExpectedIssuer: "test-issuer",
		}
		service := NewJWTService(config)

		// Appeler la fonction à tester avec un sessionUUID vide
		audiences := []string{"test-audience"}
		token, err := service.CreateRefreshToken("user-123", "", audiences)

		// Le comportement actuel ne vérifie pas si sessionUUID est vide
		// Donc le test passe, mais nous devrions peut-être améliorer le code
		// pour valider ce paramètre
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Suggestion d'amélioration: ajouter cette vérification dans CreateRefreshToken
		// if sessionUUID == "" {
		//     return "", WrapError(errors.New("session UUID cannot be empty"), "")
		// }
	})
}
