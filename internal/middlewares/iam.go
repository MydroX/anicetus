// nolint
package middlewares

import (
	"MydroX/project-v/pkg/jwt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// func Authenticate() gin.HandlerFunc {
// }

func CheckAccessToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("access_token")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Auth: Bad request"})
			c.Abort()
			return
		}

		_, errJWT := jwt.ParseToken(tokenString)
		if errJWT != jwt.JWTNoError {
			if errJWT == jwt.JWTExpiredToken {
				if isRefreshTokenValid() {
					// TODO
					// Generate new access token
					// Set new access token to cookie
					// Continue
				} else {
					// TODO : logout user

					c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
					c.Abort()
					return
				}
			}
		}

	}
}

func CheckRefreshToken() {}

func isRefreshTokenValid() bool {
	// TODO: Implement this
	return true
}
