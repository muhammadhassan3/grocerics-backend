package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders applies baseline response hardening.
// Cache-Control: no-store on /auth/* prevents intermediaries (proxies,
// browsers in legacy modes) from caching JWT-bearing responses.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		if strings.HasPrefix(c.Request.URL.Path, "/auth/") {
			c.Header("Cache-Control", "no-store")
		}
		c.Next()
	}
}
