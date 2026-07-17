package main

import (
	"flag"
	"os"

	"grocerics-backend/internal/app"
	"grocerics-backend/internal/config"
	"grocerics-backend/internal/logging"
	"grocerics-backend/internal/migrate"

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
	seed := flag.Bool("seed", false,
		"seed reference data (cities, platforms, categories, subcategories, brands) and exit")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		zap.S().Fatalw("failed to load config", "error", err)
		os.Exit(1)
	}

	if *seed {
		runSeed(cfg)
		return
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

func runSeed(cfg *config.Config) {
	if _, err := logging.Init(cfg.Env); err != nil {
		os.Exit(1)
	}
	db, err := config.ConnectDB(cfg.DB)
	if err != nil {
		zap.S().Fatalw("seed: database connect failed", "error", err)
	}
	if err := migrate.Up(db); err != nil {
		zap.S().Fatalw("seed: migrations failed", "error", err)
	}
	if err := config.SeedReference(db); err != nil {
		zap.S().Fatalw("seed: failed", "error", err)
	}
}
