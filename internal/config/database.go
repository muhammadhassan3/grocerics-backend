package config

import (
	"time"

	"grocerics-backend/internal/auth"
	"grocerics-backend/internal/domain"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ConnectDB opens a GORM connection, tunes the pool, and returns the
// handle. No globals — caller owns the *gorm.DB and must pass it to
// repositories.
func ConnectDB(cfg DBConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{
		NowFunc: func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetConnMaxLifetime(5 * time.Minute)
	}
	return db, nil
}

// SeedAdmin upserts an admin from SeedConfig if ENV allows. Called from cmd/main.go after DB connection. --- IGNORE ---
func SeedAdmin(db *gorm.DB, seed SeedConfig, env string) {
	if env == "production" {
		return
	}
	if seed.AdminEmail == "" || seed.AdminPassword == "" {
		return
	}

	var existing domain.Admin
	if err := db.Where("email = ? AND deleted_at IS NULL", seed.AdminEmail).First(&existing).Error; err == nil {
		return // already present
	}

	admin := domain.Admin{
		Name:         "Admin",
		Email:        seed.AdminEmail,
		PasswordHash: auth.HashPassword(seed.AdminPassword),
		Status:       domain.UserStatusActive,
	}
	if err := db.Create(&admin).Error; err != nil {
		zap.S().Warnw("seed admin failed", "error", err)
	}
}
