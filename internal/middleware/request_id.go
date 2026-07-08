package middleware

import (
	"grocerics-backend/internal/logging"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	RequestIDHeader = "X-Request-Id"
	requestIDKey    = "request_id"
)

// RequestID generates (or accepts via header) a unique ID per request,
// attaches it to gin context, sets it on the response header, and injects
// a per-request zap.SugaredLogger into context that tags every line with the ID.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(RequestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}
		c.Set(requestIDKey, id)
		c.Header(RequestIDHeader, id)

		logger := zap.S().With("request_id", id)
		c.Request = c.Request.WithContext(logging.WithLogger(c.Request.Context(), logger))

		c.Next()
	}
}

// RequestIDFromContext returns the request ID set by RequestID middleware.
func RequestIDFromContext(c *gin.Context) string {
	if v, ok := c.Get(requestIDKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
