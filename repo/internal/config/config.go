package config

import (
	"fmt"
	"os"

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
}

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
		JWTSecret:     getEnv("JWT_SECRET", "change-me-in-production"),
		ServerPort:    getEnv("SERVER_PORT", "8080"),
		EncryptionKey: getEnv("ENCRYPTION_KEY", "0123456789abcdef0123456789abcdef"), // 32-byte default for dev
	}

	if cfg.DBHost == "" || cfg.DBUser == "" || cfg.DBName == "" {
		return nil, fmt.Errorf("required database configuration is missing")
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
