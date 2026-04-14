package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mindflow/agri-platform/pkg/services"
)

type MetricHandler struct {
	metricSvc *services.MetricService
}

func NewMetricHandler(svc *services.MetricService) *MetricHandler {
	return &MetricHandler{metricSvc: svc}
}

func (h *MetricHandler) Create(c *gin.Context) {
	var in services.CreateMetricInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	metric, err := h.metricSvc.Create(c.Request.Context(), in)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, metric)
}

func (h *MetricHandler) BatchCreate(c *gin.Context) {
	var in services.BatchMetricInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	metrics, err := h.metricSvc.BatchCreate(c.Request.Context(), in.Metrics)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"count": len(metrics), "data": metrics})
}

func (h *MetricHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	deviceID, _ := strconv.ParseUint(c.Query("device_id"), 10, 64)

	result, err := h.metricSvc.List(c.Request.Context(), services.MetricListParams{
		Page:       page,
		PageSize:   pageSize,
		DeviceID:   uint(deviceID),
		MetricType: c.Query("metric_type"),
		StartTime:  c.Query("start_time"),
		EndTime:    c.Query("end_time"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list metrics"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *MetricHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid metric id"})
		return
	}

	metric, err := h.metricSvc.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, services.ErrMetricNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "metric not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get metric"})
		return
	}

	c.JSON(http.StatusOK, metric)
}

func (h *MetricHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid metric id"})
		return
	}

	if err := h.metricSvc.Delete(c.Request.Context(), uint(id)); err != nil {
		if errors.Is(err, services.ErrMetricNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "metric not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete metric"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "metric deleted"})
}
