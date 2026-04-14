package services

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret-key-for-unit-tests"
const testEncKey = "0123456789abcdef0123456789abcdef"

func TestGenerateAndValidateToken(t *testing.T) {
	svc := NewAuthService(nil, testSecret, testEncKey)

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
	svc := NewAuthService(nil, testSecret, testEncKey)

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
	svc := NewAuthService(nil, testSecret, testEncKey)

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
	svc := NewAuthService(nil, testSecret, testEncKey)
	_, err := svc.ValidateToken("not-a-jwt")
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestNewAuthService(t *testing.T) {
	svc := NewAuthService(nil, "secret", testEncKey)
	assert.NotNil(t, svc)
	assert.Equal(t, []byte("secret"), svc.jwtSecret)
}

func TestValidatePasswordComplexity(t *testing.T) {
	tests := []struct {
		password string
		wantErr  bool
	}{
		{"abcdefgh", true},       // no digits
		{"12345678", true},       // no letters
		{"abc12345", false},      // valid
		{"Password1", false},     // valid
		{"ab1", false},           // short but has both (length validated by binding)
		{"!!!!!!!!!", true},      // no letters or digits
	}
	for _, tt := range tests {
		err := validatePasswordComplexity(tt.password)
		if tt.wantErr {
			assert.Error(t, err, "password=%s", tt.password)
		} else {
			assert.NoError(t, err, "password=%s", tt.password)
		}
	}
}
