package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewMetricHandler(t *testing.T) {
	h := NewMetricHandler(nil)
	assert.NotNil(t, h)
}

func TestMetricHandler_Create_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewMetricHandler(nil)
	r.POST("/v1/metrics", h.Create)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/metrics", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMetricHandler_BatchCreate_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewMetricHandler(nil)
	r.POST("/v1/metrics/batch", h.BatchCreate)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/metrics/batch", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMetricHandler_Get_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewMetricHandler(nil)
	r.GET("/v1/metrics/:id", h.Get)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/metrics/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid metric id")
}

func TestMetricHandler_Delete_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewMetricHandler(nil)
	r.DELETE("/v1/metrics/:id", h.Delete)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/v1/metrics/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// GET /v1/metrics — HTTP route registration and handler invocation check.
func TestMetricHandler_List_HTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())
	h := NewMetricHandler(nil)
	r.GET("/v1/metrics", h.List)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/metrics", nil)
	r.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusNotFound, w.Code)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
