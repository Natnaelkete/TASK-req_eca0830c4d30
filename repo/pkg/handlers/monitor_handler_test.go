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

func TestNewMonitorHandler(t *testing.T) {
	q := services.NewQueueService(10, 1)
	defer q.Shutdown()
	h := NewMonitorHandler(nil, q)
	assert.NotNil(t, h)
}

func TestMonitorHandler_CheckDevice_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	q := services.NewQueueService(10, 1)
	defer q.Shutdown()
	h := NewMonitorHandler(nil, q)
	r := gin.New()
	r.POST("/v1/monitor/device", h.CheckDevice)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/monitor/device", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMonitorHandler_ThresholdCheck_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	q := services.NewQueueService(10, 1)
	defer q.Shutdown()
	h := NewMonitorHandler(nil, q)
	r := gin.New()
	r.POST("/v1/monitor/threshold", h.ThresholdCheck)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/monitor/threshold", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMonitorHandler_JobStatus_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	q := services.NewQueueService(10, 1)
	defer q.Shutdown()
	h := NewMonitorHandler(nil, q)
	r := gin.New()
	r.GET("/v1/monitor/jobs/:id", h.JobStatus)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/monitor/jobs/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestMonitorHandler_QueueStats(t *testing.T) {
	gin.SetMode(gin.TestMode)
	q := services.NewQueueService(10, 1)
	defer q.Shutdown()
	h := NewMonitorHandler(nil, q)
	r := gin.New()
	r.GET("/v1/monitor/queue/status", h.QueueStats)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/monitor/queue/status", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "capacity")
}

func TestMonitorHandler_ResolveAlert_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	q := services.NewQueueService(10, 1)
	defer q.Shutdown()
	h := NewMonitorHandler(nil, q)
	r := gin.New()
	r.PATCH("/v1/monitor/alerts/:id/resolve", h.ResolveAlert)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/v1/monitor/alerts/abc/resolve", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// GET /v1/monitor/alerts — HTTP route registration and handler invocation check.
func TestMonitorHandler_ListAlerts_HTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	q := services.NewQueueService(10, 1)
	defer q.Shutdown()
	h := NewMonitorHandler(nil, q)
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/v1/monitor/alerts", h.ListAlerts)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/monitor/alerts", nil)
	r.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusNotFound, w.Code)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
