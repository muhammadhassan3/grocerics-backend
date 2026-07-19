package middleware

import (
	"grocerics-backend/internal/auth"
	"grocerics-backend/internal/errs"

	"github.com/gin-gonic/gin"
)

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		if auth.MustUser(c).Kind != auth.KindAdmin {
			c.Error(errs.Forbidden("FORBIDDEN", "admin access required"))
			c.Abort()
			return
		}
		c.Next()
	}
}

func ClientOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		if auth.MustUser(c).Kind != auth.KindClient {
			c.Error(errs.Forbidden("FORBIDDEN", "client access required"))
			c.Abort()
			return
		}
		c.Next()
	}
}
