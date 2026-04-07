package middleware

import (
	"net/http"
	"strings"

	"MydroX/anicetus/pkg/jwt"
	"github.com/gin-gonic/gin"
)

const bearerParts = 2

func AuthMiddleware(jwtService *jwt.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractToken(c)
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "missing authorization token"})

			return
		}

		claims, err := jwtService.ParseAccessToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "invalid or expired token"})

			return
		}

		SetUserUUID(c, claims.UserUUID)
		SetPermissions(c, claims.Permissions)
		SetAudiences(c, claims.Audience)

		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	// Try Authorization header first
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", bearerParts)
		if len(parts) == bearerParts && strings.EqualFold(parts[0], "Bearer") {
			return parts[1]
		}
	}

	// Fall back to cookie
	token, err := c.Cookie("access_token")
	if err == nil && token != "" {
		return token
	}

	return ""
}
