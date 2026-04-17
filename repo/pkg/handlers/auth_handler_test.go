package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRegisterHandler_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewAuthHandler(nil)
	r.POST("/v1/auth/register", h.Register)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/auth/register", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLoginHandler_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewAuthHandler(nil)
	r.POST("/v1/auth/login", h.Login)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/auth/login", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegisterHandler_ValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewAuthHandler(nil)
	r.POST("/v1/auth/register", h.Register)

	tests := []struct {
		name string
		body map[string]string
	}{
		{"missing username", map[string]string{"email": "a@b.com", "password": "pass1234"}},
		{"short username", map[string]string{"username": "ab", "email": "a@b.com", "password": "pass1234"}},
		{"bad email", map[string]string{"username": "user1", "email": "not-email", "password": "pass1234"}},
		{"short password", map[string]string{"username": "user1", "email": "a@b.com", "password": "pass1"}},
		{"invalid role", map[string]string{"username": "user1", "email": "a@b.com", "password": "pass1234", "role": "superadmin"}},
		// Privilege-escalation regression guards: binding must reject any
		// attempt to self-assign an elevated role via public registration.
		{"role admin rejected", map[string]string{"username": "user1", "email": "a@b.com", "password": "pass1234", "role": "admin"}},
		{"role reviewer rejected", map[string]string{"username": "user1", "email": "a@b.com", "password": "pass1234", "role": "reviewer"}},
		{"role customer_service rejected", map[string]string{"username": "user1", "email": "a@b.com", "password": "pass1234", "role": "customer_service"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/v1/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestNewAuthHandler(t *testing.T) {
	h := NewAuthHandler(nil)
	assert.NotNil(t, h)
}

// GET /v1/auth/me — HTTP route registration check with nil service. We only
// verify the route is wired, the handler executes and reaches the service
// boundary (surfaced as a 500 via Recovery when the nil service panics).
func TestMeHandler_HTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())
	h := NewAuthHandler(nil)
	r.GET("/v1/auth/me", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		h.Me(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/auth/me", nil)
	r.ServeHTTP(w, req)
	assert.NotEqual(t, http.StatusNotFound, w.Code)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
