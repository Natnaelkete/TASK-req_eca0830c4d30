package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all application configuration.
type Config struct {
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	JWTSecret     string
	ServerPort    string
	EncryptionKey string // 32-byte hex key for AES-256 field-level encryption
	MigrationsDir string // path to SQL migration files applied at startup
	AppEnv        string // "development" | "production"; gates insecure defaults
}

// defaultJWTSecret is the development-grade fallback for JWT_SECRET. In
// production this value must be overridden or config load fails.
const defaultJWTSecret = "change-me-in-production"

// defaultEncryptionKey is the development-grade fallback for ENCRYPTION_KEY.
const defaultEncryptionKey = "0123456789abcdef0123456789abcdef"

// Load reads configuration from environment variables.
// It attempts to load a .env file but does not fail if one is absent.
func Load() (*Config, error) {
	_ = godotenv.Load() // best-effort; env vars may come from Docker

	cfg := &Config{
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnv("DB_PORT", "3306"),
		DBUser:        getEnv("DB_USER", "root"),
		DBPassword:    getEnv("DB_PASSWORD", "pass"),
		DBName:        getEnv("DB_NAME", "agri"),
		JWTSecret:     getEnv("JWT_SECRET", defaultJWTSecret),
		ServerPort:    getEnv("SERVER_PORT", "8080"),
		EncryptionKey: getEnv("ENCRYPTION_KEY", defaultEncryptionKey),
		MigrationsDir: getEnv("MIGRATIONS_DIR", "migrations"),
		AppEnv:        strings.ToLower(getEnv("APP_ENV", "development")),
	}

	if cfg.DBHost == "" || cfg.DBUser == "" || cfg.DBName == "" {
		return nil, fmt.Errorf("required database configuration is missing")
	}

	// In any non-development environment, refuse to boot with insecure
	// defaults — this prevents the sample secret from leaking into staging
	// or production deployments.
	if cfg.AppEnv != "development" && cfg.AppEnv != "dev" && cfg.AppEnv != "test" {
		if cfg.JWTSecret == "" || cfg.JWTSecret == defaultJWTSecret {
			return nil, fmt.Errorf("JWT_SECRET must be set to a non-default value when APP_ENV=%q", cfg.AppEnv)
		}
		if cfg.EncryptionKey == "" || cfg.EncryptionKey == defaultEncryptionKey {
			return nil, fmt.Errorf("ENCRYPTION_KEY must be set to a non-default value when APP_ENV=%q", cfg.AppEnv)
		}
	}

	return cfg, nil
}

// DSN returns the MySQL data-source name for GORM / sql.Open.
func (c *Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
