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
