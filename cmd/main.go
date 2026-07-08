package main

import (
	"os"

	"grocerics-backend/internal/app"
	"grocerics-backend/internal/config"

	"go.uber.org/zap"
)

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
