package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func testGinContext() *gin.Context {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c
}

func TestSetGetUserUUID(t *testing.T) {
	c := testGinContext()
	SetUserUUID(c, "user-123")
	assert.Equal(t, "user-123", GetUserUUID(c))
}

func TestGetUserUUID_NotSet(t *testing.T) {
	c := testGinContext()
	assert.Empty(t, GetUserUUID(c))
}

func TestGetUserUUID_WrongType(t *testing.T) {
	c := testGinContext()
	c.Set(keyUserUUID, 12345)
	assert.Empty(t, GetUserUUID(c))
}

func TestSetGetPermissions(t *testing.T) {
	c := testGinContext()
	perms := []string{"read", "write"}
	SetPermissions(c, perms)
	assert.Equal(t, perms, GetPermissions(c))
}

func TestGetPermissions_NotSet(t *testing.T) {
	c := testGinContext()
	assert.Nil(t, GetPermissions(c))
}

func TestGetPermissions_WrongType(t *testing.T) {
	c := testGinContext()
	c.Set(keyPermissions, "not-a-slice")
	assert.Nil(t, GetPermissions(c))
}

func TestSetGetAudiences(t *testing.T) {
	c := testGinContext()
	auds := []string{"service-a", "service-b"}
	SetAudiences(c, auds)
	assert.Equal(t, auds, GetAudiences(c))
}

func TestGetAudiences_NotSet(t *testing.T) {
	c := testGinContext()
	assert.Nil(t, GetAudiences(c))
}

func TestGetAudiences_WrongType(t *testing.T) {
	c := testGinContext()
	c.Set(keyAudiences, 42)
	assert.Nil(t, GetAudiences(c))
}
