// Package config loads runtime configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	JWTSecret string
	JWTTTL    time.Duration

	CORSAllowedOrigins string

	// R2* are deliberately optional — left empty, main.go skips wiring the
	// R2 client and PatientDocument upload/download degrades to a clear
	// "not configured" error instead of failing app startup. This keeps
	// environments without a real Cloudflare account (CI, a fresh clone)
	// working for everything except document upload.
	R2AccountID       string
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2BucketName      string
}

// Load reads a .env file if one is present (local dev running outside
// Docker) and falls back to whatever is already in the process environment
// otherwise — inside the app's own container, docker-compose injects these
// vars directly, so a missing .env there is expected, not an error.
func Load() (*Config, error) {
	_ = godotenv.Load()

	jwtTTL, err := time.ParseDuration(getEnv("JWT_TTL", "24h"))
	if err != nil {
		return nil, fmt.Errorf("config: invalid JWT_TTL: %w", err)
	}

	cfg := &Config{
		AppPort:    getEnv("APP_PORT", "8080"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     os.Getenv("POSTGRES_USER"),
		DBPassword: os.Getenv("POSTGRES_PASSWORD"),
		DBName:     os.Getenv("POSTGRES_DB"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		JWTSecret:  os.Getenv("JWT_SECRET"),
		JWTTTL:     jwtTTL,

		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173"),

		R2AccountID:       os.Getenv("R2_ACCOUNT_ID"),
		R2AccessKeyID:     os.Getenv("R2_ACCESS_KEY_ID"),
		R2SecretAccessKey: os.Getenv("R2_SECRET_ACCESS_KEY"),
		R2BucketName:      os.Getenv("R2_BUCKET_NAME"),
	}

	if cfg.DBUser == "" || cfg.DBPassword == "" || cfg.DBName == "" {
		return nil, fmt.Errorf("config: POSTGRES_USER, POSTGRES_PASSWORD and POSTGRES_DB must be set")
	}
	if len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("config: JWT_SECRET must be set and at least 32 characters long")
	}

	return cfg, nil
}

// IsR2Configured reports whether all four R2 values are present. Partial
// configuration (e.g. an account ID but no bucket) is treated the same as
// none at all — main.go shouldn't try to guess which combination is a typo.
func (c *Config) IsR2Configured() bool {
	return c.R2AccountID != "" && c.R2AccessKeyID != "" && c.R2SecretAccessKey != "" && c.R2BucketName != ""
}

// DSN builds the Postgres connection string GORM expects.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
