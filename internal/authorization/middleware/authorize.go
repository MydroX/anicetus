package middleware

import (
	"net/http"

	authmiddleware "MydroX/anicetus/internal/authentication/middleware"
	"github.com/gin-gonic/gin"
)

// RequirePermission checks that the authenticated user has ALL of the specified permissions
func RequirePermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userPermissions := authmiddleware.GetPermissions(c)

		permSet := make(map[string]struct{}, len(userPermissions))
		for _, p := range userPermissions {
			permSet[p] = struct{}{}
		}

		for _, required := range permissions {
			if _, ok := permSet[required]; !ok {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "insufficient permissions"})

				return
			}
		}

		c.Next()
	}
}

// RequireAnyPermission checks that the authenticated user has at least one of the specified permissions
func RequireAnyPermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userPermissions := authmiddleware.GetPermissions(c)

		permSet := make(map[string]struct{}, len(userPermissions))
		for _, p := range userPermissions {
			permSet[p] = struct{}{}
		}

		for _, required := range permissions {
			if _, ok := permSet[required]; ok {
				c.Next()

				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "insufficient permissions"})
	}
}

// RequireRole checks that the authenticated user has ALL of the specified roles.
// Note: This requires roles to be stored in the context (not yet implemented in JWT claims).
// For now, this middleware is a placeholder for future role-based checks.
func RequireRole(_ ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement when roles are added to JWT claims or fetched from DB
		c.Next()
	}
}
