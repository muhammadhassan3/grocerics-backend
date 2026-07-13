package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Env         string   // Environment (e.g., "development", "production")
	Port        string   // Port on which the application will run
	FrontendURL []string // URL of the frontend application

	DB   DBConfig
	JWT  JWTConfig
	Seed SeedConfig

	AWS AWSConfig
}

type DBConfig struct {
	Host     string // Database host
	Port     string // Database port
	User     string // Database user
	Password string // Database password
	Name     string // Database name
	SSLMode  string // SSL mode (e.g., "disable", "require")
}

func (db *DBConfig) GetDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", db.Host, db.User, db.Password, db.Name, db.Port, db.SSLMode)
}

type JWTConfig struct {
	SecretKey string // Secret key for signing JWT tokens
}

type SeedConfig struct {
	AdminEmail    string // Email for the initial admin user
	AdminPassword string // Password for the initial admin user
}

type AWSConfig struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	BucketName      string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Env:  envOr("ENV", "development"),
		Port: envOr("PORT", "8080"),
		DB: DBConfig{
			Host:     envOr("DB_HOST", "localhost"),
			Port:     envOr("DB_PORT", "5432"),
			User:     envOr("DB_USER", "postgres"),
			Password: envOr("DB_PASSWORD", "password"),
			Name:     envOr("DB_NAME", "grocerics"),
			SSLMode:  envOr("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			SecretKey: envOr("JWT_SECRET_KEY", "your-secret-key"),
		},
		Seed: SeedConfig{
			AdminEmail:    envOr("SEED_ADMIN_EMAIL", "admin@example.com"),
			AdminPassword: envOr("SEED_ADMIN_PASSWORD", "adminpassword"),
		},
		FrontendURL: splitCSV(envOr("FRONTEND_ORIGINS", "http://localhost:5500,http://localhost:3000")),
		AWS: AWSConfig{
			AccessKeyID:     envOr("AWS_ACCESS_KEY_ID", ""),
			SecretAccessKey: envOr("AWS_SECRET_ACCESS_KEY", ""),
			Region:          envOr("AWS_REGION", "us-east-1"),
			BucketName:      envOr("AWS_BUCKET_NAME", ""),
		},
	}

	var missing []string
	if cfg.DB.Host == "" {
		missing = append(missing, "DB_HOST")
	}
	if cfg.DB.Port == "" {
		missing = append(missing, "DB_PORT")
	}
	if cfg.DB.User == "" {
		missing = append(missing, "DB_USER (or POSTGRES_USER)")
	}
	if cfg.DB.Password == "" {
		missing = append(missing, "DB_PASSWORD (or POSTGRES_PASSWORD)")
	}
	if cfg.DB.Name == "" {
		missing = append(missing, "DB_NAME (or POSTGRES_DB)")
	}
	if cfg.JWT.SecretKey == "" {
		missing = append(missing, "JWT_SECRET_KEY")
	}

	if cfg.AWS.AccessKeyID == "" {
		missing = append(missing, "AWS_ACCESS_KEY_ID")
	}
	if cfg.AWS.SecretAccessKey == "" {
		missing = append(missing, "AWS_SECRET_ACCESS_KEY")
	}
	if cfg.AWS.BucketName == "" {
		missing = append(missing, "AWS_BUCKET_NAME")
	}
	if cfg.AWS.Region == "" {
		missing = append(missing, "AWS_REGION")
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
