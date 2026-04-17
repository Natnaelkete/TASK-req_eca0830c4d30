package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewDeviceHandler(t *testing.T) {
	h := NewDeviceHandler(nil)
	assert.NotNil(t, h)
}

func TestDeviceHandler_Create_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewDeviceHandler(nil)
	r.POST("/v1/devices", h.Create)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/devices", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeviceHandler_Get_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewDeviceHandler(nil)
	r.GET("/v1/devices/:id", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("role", "researcher")
		h.Get(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/devices/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid device id")
}

func TestDeviceHandler_Update_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewDeviceHandler(nil)
	r.PUT("/v1/devices/:id", h.Update)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/v1/devices/abc", bytes.NewBufferString(`{"name":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeviceHandler_Delete_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewDeviceHandler(nil)
	r.DELETE("/v1/devices/:id", h.Delete)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/v1/devices/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// GET /v1/devices — HTTP route registration and handler invocation check.
func TestDeviceHandler_List_HTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())
	h := NewDeviceHandler(nil)
	r.GET("/v1/devices", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("role", "researcher")
		h.List(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/devices", nil)
	r.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusNotFound, w.Code)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
