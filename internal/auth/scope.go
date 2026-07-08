package auth

import (
	"grocerics-backend/internal/domain"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Scope struct {
	UserID    string
	Role      domain.Role
	CompanyID *string
}

func ScopeFromContext(c *gin.Context) (Scope, bool) {
	u, ok := UserFrom(c)
	if !ok {
		return Scope{}, false
	}
	return Scope{UserID: u.ID, Role: u.Role, CompanyID: u.CompanyID}, true
}

func MustScope(c *gin.Context) Scope {
	s, ok := ScopeFromContext(c)
	if !ok {
		panic("auth.MustScope: no UserContext — AuthMiddleware missing from chain")
	}
	return s
}

// WhereClause returns the SQL fragment + args needed to scope a query to the current user's context, basically permissions.
func (s Scope) WhereClause(column string) (string, []any, bool) {
	if s.Role == domain.RoleAdmin {
		return "", nil, false
	}
	if s.CompanyID == nil {
		return "1 = 0", nil, true
	}
	return column + " = ?", []any{*s.CompanyID}, true
}

// Apply is the legacy *gorm.DB adapter. Callers using the GORM v2
// generics API (gorm.G[T]) should use WhereClause directly.
func (s Scope) Apply(db *gorm.DB, column string) *gorm.DB {
	clause, args, applied := s.WhereClause(column)
	if !applied {
		return db
	}
	return db.Where(clause, args...)
}
