package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mindflow/agri-platform/pkg/services"
)

// AnalysisHandler exposes trend, funnel, and retention analysis endpoints.
type AnalysisHandler struct {
	analysisSvc *services.AnalysisService
}

func NewAnalysisHandler(svc *services.AnalysisService) *AnalysisHandler {
	return &AnalysisHandler{analysisSvc: svc}
}

// Trends handles POST /v1/analysis/trends — time-bucketed trend analysis with drill-down.
func (h *AnalysisHandler) Trends(c *gin.Context) {
	var p services.TrendAnalysisParams
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.analysisSvc.TrendAnalysis(c.Request.Context(), p)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Funnels handles POST /v1/analysis/funnels — sequential stage funnel analysis.
func (h *AnalysisHandler) Funnels(c *gin.Context) {
	var p services.FunnelAnalysisParams
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.analysisSvc.FunnelAnalysis(c.Request.Context(), p)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Retention handles POST /v1/analysis/retention — cohort-based retention analysis.
func (h *AnalysisHandler) Retention(c *gin.Context) {
	var p services.RetentionAnalysisParams
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.analysisSvc.RetentionAnalysis(c.Request.Context(), p)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
