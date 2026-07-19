package middleware

import (
	"strings"

	"grocerics-backend/internal/auth"
	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/errs"
	"grocerics-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

type AuthDeps struct {
	JWT    *auth.JWTService
	Users  *repository.UserRepository
	Admins *repository.AdminRepository
}

func AuthMiddleware(d *AuthDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.Error(errs.Unauthorized("AUTH_TOKEN_MISSING", "Authorization header missing or invalid"))
			c.Abort()
			return
		}
		claims, err := d.JWT.Validate(strings.TrimPrefix(authHeader, "Bearer "))
		if err != nil {
			c.Error(errs.Unauthorized("AUTH_TOKEN_INVALID", "Invalid token").WithCause(err))
			c.Abort()
			return
		}

		kind := auth.Kind(claims.Kind)
		if kind == "" {
			kind = auth.KindClient
		}

		switch kind {
		case auth.KindAdmin:
			a, lookupErr := d.Admins.FindByID(claims.UserID)
			if lookupErr != nil {
				c.Error(errs.Internal("AUTH_ADMIN_LOOKUP_FAILED", lookupErr))
				c.Abort()
				return
			}
			if a == nil {
				c.Error(errs.Unauthorized("AUTH_ADMIN_GONE", "admin no longer exists"))
				c.Abort()
				return
			}
		default:
			u, lookupErr := d.Users.FindByID(claims.UserID)
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
			Role: domain.Role(claims.Role),
			Kind: kind,
		})
		c.Next()
	}
}
