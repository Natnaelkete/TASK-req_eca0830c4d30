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
		{"missing username", map[string]string{"email": "a@b.com", "password": "123456"}},
		{"short username", map[string]string{"username": "ab", "email": "a@b.com", "password": "123456"}},
		{"bad email", map[string]string{"username": "user1", "email": "not-email", "password": "123456"}},
		{"short password", map[string]string{"username": "user1", "email": "a@b.com", "password": "12345"}},
		{"invalid role", map[string]string{"username": "user1", "email": "a@b.com", "password": "123456", "role": "superadmin"}},
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
