// Package errs defines application error types and helpers for centralized HTTP error mapping.
package errs

import (
	"errors"
	"fmt"
	"net/http"
)

// Handlers don't map errors themselves: they call c.Error(err); return,
// and the ErrorHandler middleware does the mapping in one place.
type AppError struct {
	Status  int    // HTTP status to return
	Code    string // machine-readable: "INVALID_CREDENTIALS", "USER_NOT_FOUND"
	Message string // human-readable; safe to show to the client
	Cause   error  // wrapped; for logs only
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Cause }

// WithCause attaches a (private) cause to an existing AppError. Returns
// the same pointer for chaining.
func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

// As extracts an *AppError from any wrapped error chain.
func As(err error) (*AppError, bool) {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}

func BadRequest(code, msg string) *AppError {
	return &AppError{Status: http.StatusBadRequest, Code: code, Message: msg}
}

func Unauthorized(code, msg string) *AppError {
	return &AppError{Status: http.StatusUnauthorized, Code: code, Message: msg}
}

func Forbidden(code, msg string) *AppError {
	return &AppError{Status: http.StatusForbidden, Code: code, Message: msg}
}

func NotFound(code, msg string) *AppError {
	return &AppError{Status: http.StatusNotFound, Code: code, Message: msg}
}

func Conflict(code, msg string) *AppError {
	return &AppError{Status: http.StatusConflict, Code: code, Message: msg}
}

// Internal wraps a sensitive cause. The cause is logged; the client only sees a generic "internal server error" message.
func Internal(code string, cause error) *AppError {
	return &AppError{
		Status:  http.StatusInternalServerError,
		Code:    code,
		Message: "internal server error",
		Cause:   cause,
	}
}
