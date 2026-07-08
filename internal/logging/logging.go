// Package logging wires up the application's zap logger and provides
// context helpers for propagating a per-request logger instance.
package logging

import (
	"context"

	"go.uber.org/zap"
)

var Log *zap.Logger

// Init builds a zap logger tuned for env (production: JSON, info level;
// otherwise: human-readable console, debug level), installs it as the
// global logger (zap.L()/zap.S()), and returns the sugared handle.
func Init(env string) *zap.SugaredLogger {
	var cfg zap.Config
	if env == "production" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}

	logger, err := cfg.Build()
	if err != nil {
		panic("logging: failed to build zap logger: " + err.Error())
	}

	Log = logger

	zap.ReplaceGlobals(logger)
	return logger.Sugar()
}

type loggerCtxKey struct{}

// WithLogger attaches a per-request logger to ctx.
// Middleware uses it to inject request_id etc. so downstream service-layer code can pull a logger that's already tagged for this request.
func WithLogger(ctx context.Context, l *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, loggerCtxKey{}, l)
}

// FromContext returns the per-request logger, falling back to the
// global sugared logger if none was attached to ctx.
func FromContext(ctx context.Context) *zap.SugaredLogger {
	if l, ok := ctx.Value(loggerCtxKey{}).(*zap.SugaredLogger); ok {
		return l
	}
	return zap.S()
}
