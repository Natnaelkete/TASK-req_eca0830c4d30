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

func TestNewMonitoringDataHandler(t *testing.T) {
	q := services.NewQueueService(10, 1)
	defer q.Shutdown()
	h := NewMonitoringDataHandler(nil, q)
	assert.NotNil(t, h)
}

func TestMonitoringDataHandler_BatchIngest_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	q := services.NewQueueService(10, 1)
	defer q.Shutdown()
	h := NewMonitoringDataHandler(nil, q)
	r := gin.New()
	r.POST("/v1/monitoring/ingest", h.BatchIngest)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/monitoring/ingest", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMonitoringDataHandler_BatchIngest_EmptyData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	q := services.NewQueueService(10, 1)
	defer q.Shutdown()
	h := NewMonitoringDataHandler(nil, q)
	r := gin.New()
	r.POST("/v1/monitoring/ingest", h.BatchIngest)

	w := httptest.NewRecorder()
	body := `{"data":[]}`
	req, _ := http.NewRequest("POST", "/v1/monitoring/ingest", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMonitoringDataHandler_Aggregate_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	q := services.NewQueueService(10, 1)
	defer q.Shutdown()
	h := NewMonitoringDataHandler(nil, q)
	r := gin.New()
	r.POST("/v1/monitoring/aggregate", h.Aggregate)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/monitoring/aggregate", bytes.NewBufferString(`not-json`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMonitoringDataHandler_RealtimeCurve_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	q := services.NewQueueService(10, 1)
	defer q.Shutdown()
	h := NewMonitoringDataHandler(nil, q)
	r := gin.New()
	r.POST("/v1/monitoring/curve", h.RealtimeCurve)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/monitoring/curve", bytes.NewBufferString(`not-json`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMonitoringDataHandler_Trends_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	q := services.NewQueueService(10, 1)
	defer q.Shutdown()
	h := NewMonitoringDataHandler(nil, q)
	r := gin.New()
	r.POST("/v1/monitoring/trends", h.Trends)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/monitoring/trends", bytes.NewBufferString(`not-json`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMonitoringDataHandler_Get_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	q := services.NewQueueService(10, 1)
	defer q.Shutdown()
	h := NewMonitoringDataHandler(nil, q)
	r := gin.New()
	r.GET("/v1/monitoring/data/:id", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("role", "researcher")
		h.Get(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/monitoring/data/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMonitoringDataHandler_JobStatus_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	q := services.NewQueueService(10, 1)
	defer q.Shutdown()
	h := NewMonitoringDataHandler(nil, q)
	r := gin.New()
	r.GET("/v1/monitoring/jobs/:id", h.JobStatus)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/monitoring/jobs/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestMonitoringDataHandler_BatchIngest_Accepted(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// Use a large buffer so jobs queue but don't get processed by the worker
	// (we only test the HTTP accept path, not the async worker)
	q := services.NewQueueService(100, 0)
	svc := services.NewMonitoringDataService(nil, q)
	h := NewMonitoringDataHandler(svc, q)
	r := gin.New()
	r.POST("/v1/monitoring/ingest", h.BatchIngest)

	body := `{"data":[{"source_id":"s1","device_id":1,"plot_id":1,"metric_code":"temp","value":20,"event_time":"2026-01-15T10:00:00Z"}]}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/monitoring/ingest", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)
	assert.Contains(t, w.Body.String(), "job_id")
	assert.Contains(t, w.Body.String(), "count")
}
