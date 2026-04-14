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

func TestNewAnalysisHandler(t *testing.T) {
	h := NewAnalysisHandler(nil)
	assert.NotNil(t, h)
}

func TestAnalysisHandler_Trends_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := services.NewAnalysisService(nil)
	h := NewAnalysisHandler(svc)
	r := gin.New()
	r.POST("/v1/analysis/trends", h.Trends)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/analysis/trends", bytes.NewBufferString(`not-json`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalysisHandler_Trends_MissingRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := services.NewAnalysisService(nil)
	h := NewAnalysisHandler(svc)
	r := gin.New()
	r.POST("/v1/analysis/trends", h.Trends)

	// Missing metric_code, start_time, end_time
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/analysis/trends", bytes.NewBufferString(`{"interval":"daily"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalysisHandler_Funnels_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := services.NewAnalysisService(nil)
	h := NewAnalysisHandler(svc)
	r := gin.New()
	r.POST("/v1/analysis/funnels", h.Funnels)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/analysis/funnels", bytes.NewBufferString(`not-json`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalysisHandler_Funnels_MissingStages(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := services.NewAnalysisService(nil)
	h := NewAnalysisHandler(svc)
	r := gin.New()
	r.POST("/v1/analysis/funnels", h.Funnels)

	// stages requires min=2
	w := httptest.NewRecorder()
	body := `{"stages":["only_one"],"start_time":"2026-01-01T00:00:00Z","end_time":"2026-12-31T23:59:59Z"}`
	req, _ := http.NewRequest("POST", "/v1/analysis/funnels", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalysisHandler_Retention_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := services.NewAnalysisService(nil)
	h := NewAnalysisHandler(svc)
	r := gin.New()
	r.POST("/v1/analysis/retention", h.Retention)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/analysis/retention", bytes.NewBufferString(`not-json`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalysisHandler_Retention_MissingRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := services.NewAnalysisService(nil)
	h := NewAnalysisHandler(svc)
	r := gin.New()
	r.POST("/v1/analysis/retention", h.Retention)

	// Missing metric_code, start_time, end_time
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/analysis/retention", bytes.NewBufferString(`{"cohort_interval":"weekly"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
