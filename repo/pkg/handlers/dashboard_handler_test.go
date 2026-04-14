package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mindflow/agri-platform/pkg/services"
	"github.com/stretchr/testify/assert"
)

func TestNewDashboardHandler(t *testing.T) {
	h := NewDashboardHandler(nil)
	assert.NotNil(t, h)
}

func TestDashboardHandler_Create_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := services.NewDashboardService(nil)
	h := NewDashboardHandler(svc)
	r := gin.New()
	r.POST("/v1/dashboards", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		h.Create(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/dashboards", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDashboardHandler_Get_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := services.NewDashboardService(nil)
	h := NewDashboardHandler(svc)
	r := gin.New()
	r.GET("/v1/dashboards/:id", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		h.Get(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/dashboards/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDashboardHandler_Update_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := services.NewDashboardService(nil)
	h := NewDashboardHandler(svc)
	r := gin.New()
	r.PUT("/v1/dashboards/:id", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		h.Update(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/v1/dashboards/abc", bytes.NewBufferString(`{"name":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDashboardHandler_Delete_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := services.NewDashboardService(nil)
	h := NewDashboardHandler(svc)
	r := gin.New()
	r.DELETE("/v1/dashboards/:id", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		h.Delete(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/v1/dashboards/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDashboardHandler_Update_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := services.NewDashboardService(nil)
	h := NewDashboardHandler(svc)
	r := gin.New()
	r.PUT("/v1/dashboards/:id", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		h.Update(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/v1/dashboards/1", bytes.NewBufferString(`not-json`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
