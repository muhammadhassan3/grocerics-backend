package middleware

import (
	"grocerics-backend/internal/auth"
	"grocerics-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func ActivityTracker(repo *repository.AnalyticsRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		if u, ok := auth.UserFrom(c); ok && u.ID != "" {
			uid := u.ID
			go func() { _ = repo.MarkActive(uid) }()
		}
		c.Next()
	}
}
