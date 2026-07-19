// Package auth contains authentication helpers and request user context utilities.
package auth

import (
	"grocerics-backend/internal/domain"

	"github.com/gin-gonic/gin"
)

// A struct that allows us to track roles inside the gin context, so that we can do role-based access control in handlers without re-querying the database for every request.
type UserContext struct {
	ID        string
	Name      string
	Role      domain.Role
	Kind      Kind
	CompanyID *string
}

const userContextKey = "auth.user"

func SetUser(c *gin.Context, u UserContext) {
	c.Set(userContextKey, u)
}

func UserFrom(c *gin.Context) (UserContext, bool) {
	v, ok := c.Get(userContextKey)
	if !ok {
		return UserContext{}, false
	}
	u, ok := v.(UserContext)
	return u, ok
}

func MustUser(c *gin.Context) UserContext {
	u, ok := UserFrom(c)
	if !ok {
		panic("auth.MustUser: no UserContext on gin.Context — AuthMiddleware missing from chain")
	}
	return u
}
