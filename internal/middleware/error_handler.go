package middleware

import (
	"net/http"

	"grocerics-backend/internal/dto"
	"grocerics-backend/internal/errs"
	"grocerics-backend/internal/logging"

	"github.com/gin-gonic/gin"
)

// ErrorHandler maps errors emitted by handlers (via c.Error(err)) into
// the project's standard JSON response. *errs.AppError carries its own
// status + code + safe message.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}
		err := c.Errors.Last().Err
		log := logging.FromContext(c.Request.Context())

		if ae, ok := errs.As(err); ok {
			if ae.Status >= 500 {
				log.Errorw("server error", "code", ae.Code, "cause", ae.Cause)
			} else {
				log.Warnw("client error",
					"status", ae.Status, "code", ae.Code, "msg", ae.Message)
			}
			c.JSON(ae.Status, dto.Response{
				Status:  "failed",
				Code:    ae.Code,
				Message: ae.Message,
			})
			return
		}

		log.Errorw("unhandled error", "error", err)
		c.JSON(http.StatusInternalServerError, dto.Response{
			Status:  "failed",
			Message: "internal server error",
		})
	}
}
