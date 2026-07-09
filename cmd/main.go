package main

import (
	"os"

	"grocerics-backend/internal/app"
	"grocerics-backend/internal/config"

	"go.uber.org/zap"
)

// @title Grocerics API
// @version 1.1
// @description This is the API documentation for the Grocerics backend service.
// @contact.name API Support
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cfg, err := config.Load()
	if err != nil {
		zap.S().Fatalw("failed to load config", "error", err)
		os.Exit(1)
	}

	a, err := app.New(cfg)
	if err != nil {
		zap.S().Fatalw("failed to initialize app", "error", err)
		os.Exit(1)
	}

	if err := a.Run(); err != nil {
		zap.S().Fatalw("failed to run app", "error", err)
		os.Exit(1)
	}
}
