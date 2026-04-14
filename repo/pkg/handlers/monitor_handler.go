package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mindflow/agri-platform/pkg/services"
)

type MonitorHandler struct {
	monitorSvc *services.MonitorService
	queueSvc   *services.QueueService
}

func NewMonitorHandler(monitorSvc *services.MonitorService, queueSvc *services.QueueService) *MonitorHandler {
	return &MonitorHandler{monitorSvc: monitorSvc, queueSvc: queueSvc}
}

// CheckDevice handles POST /v1/monitor/device — enqueues a device health check.
func (h *MonitorHandler) CheckDevice(c *gin.Context) {
	var in services.MonitorDeviceInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	job, err := h.monitorSvc.SubmitMonitorDevice(in.DeviceID)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"job_id": job.ID, "status": job.Status})
}

// ThresholdCheck handles POST /v1/monitor/threshold — enqueues a threshold check.
func (h *MonitorHandler) ThresholdCheck(c *gin.Context) {
	var in services.ThresholdCheckInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if in.Level == "" {
		in.Level = "warning"
	}

	job, err := h.monitorSvc.SubmitThresholdCheck(in)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"job_id": job.ID, "status": job.Status})
}

// JobStatus handles GET /v1/monitor/jobs/:id — returns a job's current state.
func (h *MonitorHandler) JobStatus(c *gin.Context) {
	id := c.Param("id")
	job, ok := h.queueSvc.GetJob(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}
	c.JSON(http.StatusOK, job)
}

// QueueStats handles GET /v1/monitor/queue/status — returns queue metrics.
func (h *MonitorHandler) QueueStats(c *gin.Context) {
	c.JSON(http.StatusOK, h.queueSvc.Stats())
}

// ListAlerts handles GET /v1/monitor/alerts — returns paginated alerts.
func (h *MonitorHandler) ListAlerts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	deviceID, _ := strconv.ParseUint(c.Query("device_id"), 10, 64)
	level := c.Query("level")

	params := services.AlertListParams{
		Page:     page,
		PageSize: pageSize,
		DeviceID: uint(deviceID),
		Level:    level,
	}

	if r := c.Query("resolved"); r != "" {
		resolved := r == "true"
		params.Resolved = &resolved
	}

	result, err := h.monitorSvc.ListAlerts(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list alerts"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ResolveAlert handles PATCH /v1/monitor/alerts/:id/resolve.
func (h *MonitorHandler) ResolveAlert(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid alert id"})
		return
	}

	if err := h.monitorSvc.ResolveAlert(c.Request.Context(), uint(id)); err != nil {
		if errors.Is(err, services.ErrAlertNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "alert not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve alert"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "alert resolved"})
}
