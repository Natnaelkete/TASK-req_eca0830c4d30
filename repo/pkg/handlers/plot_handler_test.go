package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewPlotHandler(t *testing.T) {
	h := NewPlotHandler(nil)
	assert.NotNil(t, h)
}

func TestPlotHandler_Create_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewPlotHandler(nil)
	r.POST("/v1/plots", h.Create)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/plots", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPlotHandler_Get_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewPlotHandler(nil)
	r.GET("/v1/plots/:id", h.Get)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/plots/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid plot id")
}

func TestPlotHandler_Update_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewPlotHandler(nil)
	r.PUT("/v1/plots/:id", h.Update)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/v1/plots/abc", bytes.NewBufferString(`{"name":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPlotHandler_Delete_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewPlotHandler(nil)
	r.DELETE("/v1/plots/:id", h.Delete)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/v1/plots/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// GET /v1/plots — verifies route is wired and handler is invoked via HTTP.
func TestPlotHandler_List_HTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())
	h := NewPlotHandler(nil)
	r.GET("/v1/plots", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("role", "researcher")
		h.List(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/plots?page=1&page_size=10", nil)
	r.ServeHTTP(w, req)

	// Route is registered (not 404) and the handler executed; the nil-DB
	// path surfaces as a 500 captured by Recovery.
	assert.NotEqual(t, http.StatusNotFound, w.Code)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
