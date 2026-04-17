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

func TestNewIndicatorHandler(t *testing.T) {
	h := NewIndicatorHandler(nil)
	assert.NotNil(t, h)
}

func TestIndicatorHandler_Create_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewIndicatorHandler(nil)
	r.POST("/v1/indicators", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		h.Create(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/indicators", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestIndicatorHandler_Update_BadID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewIndicatorHandler(nil)
	r.PUT("/v1/indicators/:id", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		h.Update(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/v1/indicators/abc", bytes.NewBufferString(`{"diff_summary":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestIndicatorHandler_Update_RequiresDiffSummary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewIndicatorHandler(nil)
	r.PUT("/v1/indicators/:id", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		h.Update(c)
	})

	body, _ := json.Marshal(map[string]string{"name": "new name"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/v1/indicators/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestIndicatorHandler_Get_BadID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewIndicatorHandler(nil)
	r.GET("/v1/indicators/:id", h.Get)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/indicators/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestIndicatorHandler_Delete_BadID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewIndicatorHandler(nil)
	r.DELETE("/v1/indicators/:id", h.Delete)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/v1/indicators/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestIndicatorHandler_ListVersions_BadID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewIndicatorHandler(nil)
	r.GET("/v1/indicators/:id/versions", h.ListVersions)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/indicators/abc/versions", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// GET /v1/indicators — HTTP route registration check.
func TestIndicatorHandler_List_HTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())
	h := NewIndicatorHandler(nil)
	r.GET("/v1/indicators", h.List)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/indicators", nil)
	r.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusNotFound, w.Code)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestIndicatorHandler_GetVersion_BadVersion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewIndicatorHandler(nil)
	r.GET("/v1/indicators/:id/versions/:version", h.GetVersion)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/indicators/1/versions/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
