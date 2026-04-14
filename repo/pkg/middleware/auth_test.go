package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mindflow/agri-platform/pkg/services"
	"github.com/stretchr/testify/assert"
)

const testSecret = "test-secret-key"

func setupTestRouter(authSvc *services.AuthService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthMiddleware(authSvc))
	r.GET("/protected", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		role, _ := c.Get("role")
		c.JSON(http.StatusOK, gin.H{"user_id": userID, "role": role})
	})
	return r
}

func makeToken(secret string, userID uint, role string, expired bool) string {
	now := time.Now()
	exp := now.Add(24 * time.Hour)
	if expired {
		now = now.Add(-2 * time.Hour)
		exp = now.Add(1 * time.Hour)
	}
	claims := services.Claims{
		UserID:   userID,
		Username: "testuser",
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
			Subject:   fmt.Sprintf("%d", userID),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(secret))
	return signed
}

func TestAuthMiddleware_NoHeader(t *testing.T) {
	authSvc := services.NewAuthService(nil, testSecret, "0123456789abcdef0123456789abcdef")
	r := setupTestRouter(authSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "missing authorization header")
}

func TestAuthMiddleware_BadFormat(t *testing.T) {
	authSvc := services.NewAuthService(nil, testSecret, "0123456789abcdef0123456789abcdef")
	r := setupTestRouter(authSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Token abc123")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid authorization format")
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	authSvc := services.NewAuthService(nil, testSecret, "0123456789abcdef0123456789abcdef")
	r := setupTestRouter(authSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	authSvc := services.NewAuthService(nil, testSecret, "0123456789abcdef0123456789abcdef")
	r := setupTestRouter(authSvc)

	token := makeToken(testSecret, 1, "admin", true)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	authSvc := services.NewAuthService(nil, testSecret, "0123456789abcdef0123456789abcdef")
	r := setupTestRouter(authSvc)

	token := makeToken(testSecret, 42, "researcher", false)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "42")
	assert.Contains(t, w.Body.String(), "researcher")
}

func TestRoleGuard_Allowed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	authSvc := services.NewAuthService(nil, testSecret, "0123456789abcdef0123456789abcdef")
	r.Use(AuthMiddleware(authSvc), RoleGuard("admin"))
	r.GET("/admin-only", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	token := makeToken(testSecret, 1, "admin", false)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin-only", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRoleGuard_Denied(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	authSvc := services.NewAuthService(nil, testSecret, "0123456789abcdef0123456789abcdef")
	r.Use(AuthMiddleware(authSvc), RoleGuard("admin"))
	r.GET("/admin-only", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	token := makeToken(testSecret, 2, "viewer", false)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin-only", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "insufficient permissions")
}

func TestRoleGuard_NoRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// Skip auth middleware — so no role is set
	r.Use(RoleGuard("admin"))
	r.GET("/admin-only", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin-only", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
