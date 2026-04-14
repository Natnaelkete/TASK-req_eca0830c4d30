package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Defaults(t *testing.T) {
	// Unset all relevant vars to test defaults
	for _, k := range []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "JWT_SECRET", "SERVER_PORT"} {
		os.Unsetenv(k)
	}

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "localhost", cfg.DBHost)
	assert.Equal(t, "3306", cfg.DBPort)
	assert.Equal(t, "root", cfg.DBUser)
	assert.Equal(t, "pass", cfg.DBPassword)
	assert.Equal(t, "agri", cfg.DBName)
	assert.Equal(t, "change-me-in-production", cfg.JWTSecret)
	assert.Equal(t, "8080", cfg.ServerPort)
}

func TestLoad_EnvOverrides(t *testing.T) {
	os.Setenv("DB_HOST", "myhost")
	os.Setenv("DB_PORT", "3307")
	os.Setenv("DB_USER", "admin")
	os.Setenv("DB_PASSWORD", "secret")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("JWT_SECRET", "supersecret")
	os.Setenv("SERVER_PORT", "9090")
	defer func() {
		for _, k := range []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "JWT_SECRET", "SERVER_PORT"} {
			os.Unsetenv(k)
		}
	}()

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "myhost", cfg.DBHost)
	assert.Equal(t, "3307", cfg.DBPort)
	assert.Equal(t, "admin", cfg.DBUser)
	assert.Equal(t, "secret", cfg.DBPassword)
	assert.Equal(t, "testdb", cfg.DBName)
	assert.Equal(t, "supersecret", cfg.JWTSecret)
	assert.Equal(t, "9090", cfg.ServerPort)
}

func TestDSN(t *testing.T) {
	cfg := &Config{
		DBHost:     "localhost",
		DBPort:     "3306",
		DBUser:     "root",
		DBPassword: "pass",
		DBName:     "agri",
	}
	expected := "root:pass@tcp(localhost:3306)/agri?charset=utf8mb4&parseTime=True&loc=Local"
	assert.Equal(t, expected, cfg.DSN())
}

func TestLoad_MissingDBHost(t *testing.T) {
	os.Setenv("DB_HOST", "")
	os.Setenv("DB_USER", "root")
	os.Setenv("DB_NAME", "agri")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_NAME")
	}()

	_, err := Load()
	assert.Error(t, err)
}
