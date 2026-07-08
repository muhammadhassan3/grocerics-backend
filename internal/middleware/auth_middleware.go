package middleware

import (
	"strings"

	"grocerics-backend/internal/auth"
	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/errs"
	"grocerics-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates the Bearer JWT and stashes a typed
// auth.UserContext on the gin context for downstream handlers.
func AuthMiddleware(jwt *auth.JWTService, users *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.Error(errs.Unauthorized("AUTH_TOKEN_MISSING", "Authorization header missing or invalid"))
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := jwt.Validate(tokenStr)
		if err != nil {
			c.Error(errs.Unauthorized("AUTH_TOKEN_INVALID", "Invalid token").WithCause(err))
			c.Abort()
			return
		}

		role := domain.Role(claims.Role)

		if role != domain.RoleAdmin {
			u, lookupErr := users.FindByID(claims.UserID)
			if lookupErr != nil {
				c.Error(errs.Internal("AUTH_USER_LOOKUP_FAILED", lookupErr))
				c.Abort()
				return
			}
			if u == nil {
				c.Error(errs.Unauthorized("AUTH_USER_GONE", "user no longer exists"))
				c.Abort()
				return
			}
		}

		auth.SetUser(c, auth.UserContext{
			ID:   claims.UserID,
			Name: claims.Name,
			Role: role,
		})
		c.Next()
	}
}
