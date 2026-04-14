package services

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret-key-for-unit-tests"

func TestGenerateAndValidateToken(t *testing.T) {
	svc := NewAuthService(nil, testSecret)

	now := time.Now()
	claims := Claims{
		UserID:   1,
		Username: "admin",
		Role:     "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			Subject:   "1",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(testSecret))
	require.NoError(t, err)

	parsed, err := svc.ValidateToken(signed)
	require.NoError(t, err)
	assert.Equal(t, uint(1), parsed.UserID)
	assert.Equal(t, "admin", parsed.Username)
	assert.Equal(t, "admin", parsed.Role)
}

func TestValidateToken_Expired(t *testing.T) {
	svc := NewAuthService(nil, testSecret)

	past := time.Now().Add(-2 * time.Hour)
	claims := Claims{
		UserID:   1,
		Username: "user",
		Role:     "researcher",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(past),
			ExpiresAt: jwt.NewNumericDate(past.Add(1 * time.Hour)), // expired 1h ago
			Subject:   "1",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(testSecret))
	require.NoError(t, err)

	_, err = svc.ValidateToken(signed)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestValidateToken_WrongSecret(t *testing.T) {
	svc := NewAuthService(nil, testSecret)

	now := time.Now()
	claims := Claims{
		UserID:   1,
		Username: "user",
		Role:     "researcher",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			Subject:   "1",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte("wrong-secret"))
	require.NoError(t, err)

	_, err = svc.ValidateToken(signed)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestValidateToken_InvalidString(t *testing.T) {
	svc := NewAuthService(nil, testSecret)
	_, err := svc.ValidateToken("not-a-jwt")
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestNewAuthService(t *testing.T) {
	svc := NewAuthService(nil, "secret")
	assert.NotNil(t, svc)
	assert.Equal(t, []byte("secret"), svc.jwtSecret)
}
