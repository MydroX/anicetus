package middleware

import "github.com/gin-gonic/gin"

const (
	keyUserUUID    = "user_uuid"
	keyPermissions = "permissions"
	keyAudiences   = "audiences"
)

func SetUserUUID(c *gin.Context, uuid string) {
	c.Set(keyUserUUID, uuid)
}

func GetUserUUID(c *gin.Context) string {
	val, exists := c.Get(keyUserUUID)
	if !exists {
		return ""
	}

	uuid, ok := val.(string)
	if !ok {
		return ""
	}

	return uuid
}

func SetPermissions(c *gin.Context, permissions []string) {
	c.Set(keyPermissions, permissions)
}

func GetPermissions(c *gin.Context) []string {
	val, exists := c.Get(keyPermissions)
	if !exists {
		return nil
	}

	permissions, ok := val.([]string)
	if !ok {
		return nil
	}

	return permissions
}

func SetAudiences(c *gin.Context, audiences []string) {
	c.Set(keyAudiences, audiences)
}

func GetAudiences(c *gin.Context) []string {
	val, exists := c.Get(keyAudiences)
	if !exists {
		return nil
	}

	audiences, ok := val.([]string)
	if !ok {
		return nil
	}

	return audiences
}
