package middleware

import (
	"grocerics-backend/internal/auth"
	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/errs"

	"github.com/gin-gonic/gin"
)

func RequireRole(allowed ...domain.Role) gin.HandlerFunc {
	set := make(map[domain.Role]struct{}, len(allowed))
	for _, r := range allowed {
		set[r] = struct{}{}
	}
	return func(c *gin.Context) {
		u := auth.MustUser(c)
		if _, ok := set[u.Role]; !ok {
			c.Error(errs.Forbidden("FORBIDDEN", "insufficient role"))
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequireSameCompany(extract func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		u := auth.MustUser(c)
		if u.Role == domain.RoleAdmin {
			c.Next()
			return
		}
		target := extract(c)
		if u.CompanyID == nil || *u.CompanyID != target {
			c.Error(errs.Forbidden("FORBIDDEN", "company mismatch"))
			c.Abort()
			return
		}
		c.Next()
	}
}
