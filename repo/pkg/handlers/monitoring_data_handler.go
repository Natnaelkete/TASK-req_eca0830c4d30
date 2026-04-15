package handlers

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mindflow/agri-platform/pkg/services"
)

type MonitoringDataHandler struct {
	monDataSvc *services.MonitoringDataService
	queueSvc   *services.QueueService
}

func NewMonitoringDataHandler(monDataSvc *services.MonitoringDataService, queueSvc *services.QueueService) *MonitoringDataHandler {
	return &MonitoringDataHandler{monDataSvc: monDataSvc, queueSvc: queueSvc}
}

// BatchIngest handles POST /v1/monitoring/ingest — enqueues batch data for async processing.
func (h *MonitoringDataHandler) BatchIngest(c *gin.Context) {
	var in services.BatchMonitoringDataInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	job, err := h.monDataSvc.SubmitBatchIngest(in.Data)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"job_id": job.ID, "status": job.Status, "count": len(in.Data)})
}

// List handles GET /v1/monitoring/data — returns paginated monitoring data.
func (h *MonitoringDataHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	plotID, _ := strconv.ParseUint(c.Query("plot_id"), 10, 64)
	deviceID, _ := strconv.ParseUint(c.Query("device_id"), 10, 64)

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	result, err := h.monDataSvc.List(c.Request.Context(), services.MonitoringDataListParams{
		Page:       page,
		PageSize:   pageSize,
		PlotID:     uint(plotID),
		DeviceID:   uint(deviceID),
		MetricCode: c.Query("metric_code"),
		StartTime:  c.Query("start_time"),
		EndTime:    c.Query("end_time"),
		Tags:       c.Query("tags"),
		UserID:     userID.(uint),
		Role:       role.(string),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list monitoring data"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Get handles GET /v1/monitoring/data/:id — returns a single monitoring record.
func (h *MonitoringDataHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	record, err := h.monDataSvc.GetByID(c.Request.Context(), uint(id), userID.(uint), role.(string))
	if err != nil {
		if errors.Is(err, services.ErrMonitoringDataNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "monitoring data not found"})
			return
		}
		if errors.Is(err, services.ErrMonitoringDataForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to access this monitoring data"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get monitoring data"})
		return
	}

	c.JSON(http.StatusOK, record)
}

// Aggregate handles POST /v1/monitoring/aggregate — multi-dimensional aggregation.
func (h *MonitoringDataHandler) Aggregate(c *gin.Context) {
	var p services.AggregationParams
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	p.UserID = userID.(uint)
	p.Role = role.(string)

	results, err := h.monDataSvc.Aggregate(c.Request.Context(), p)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
}

// RealtimeCurve handles POST /v1/monitoring/curve — real-time time-series data.
func (h *MonitoringDataHandler) RealtimeCurve(c *gin.Context) {
	var p services.CurveParams
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	p.UserID = userID.(uint)
	p.Role = role.(string)

	points, err := h.monDataSvc.RealtimeCurve(c.Request.Context(), p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query realtime curve"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"points": points, "count": len(points)})
}

// Trends handles POST /v1/monitoring/trends — daily/weekly/monthly trends with YoY/MoM.
func (h *MonitoringDataHandler) Trends(c *gin.Context) {
	var p services.TrendParams
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	p.UserID = userID.(uint)
	p.Role = role.(string)

	result, err := h.monDataSvc.TrendStatistics(c.Request.Context(), p)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ExportJSON handles GET /v1/monitoring/export/json — exports data as JSON download.
func (h *MonitoringDataHandler) ExportJSON(c *gin.Context) {
	plotID, _ := strconv.ParseUint(c.Query("plot_id"), 10, 64)
	deviceID, _ := strconv.ParseUint(c.Query("device_id"), 10, 64)
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	data, err := h.monDataSvc.ExportData(c.Request.Context(), services.ExportParams{
		PlotID:     uint(plotID),
		DeviceID:   uint(deviceID),
		MetricCode: c.Query("metric_code"),
		StartTime:  c.Query("start_time"),
		EndTime:    c.Query("end_time"),
		Tags:       c.Query("tags"),
		UserID:     userID.(uint),
		Role:       role.(string),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to export data"})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=monitoring_data.json")
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, data)
}

// ExportCSV handles GET /v1/monitoring/export/csv — exports data as CSV download.
func (h *MonitoringDataHandler) ExportCSV(c *gin.Context) {
	plotID, _ := strconv.ParseUint(c.Query("plot_id"), 10, 64)
	deviceID, _ := strconv.ParseUint(c.Query("device_id"), 10, 64)
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	data, err := h.monDataSvc.ExportData(c.Request.Context(), services.ExportParams{
		PlotID:     uint(plotID),
		DeviceID:   uint(deviceID),
		MetricCode: c.Query("metric_code"),
		StartTime:  c.Query("start_time"),
		EndTime:    c.Query("end_time"),
		Tags:       c.Query("tags"),
		UserID:     userID.(uint),
		Role:       role.(string),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to export data"})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=monitoring_data.csv")
	c.Header("Content-Type", "text/csv")
	c.Status(http.StatusOK)

	writer := csv.NewWriter(c.Writer)
	// Write header
	writer.Write([]string{"id", "source_id", "device_id", "plot_id", "metric_code", "value", "unit", "event_time", "tags"})

	for _, d := range data {
		tagsStr := d.Tags
		if tagsStr == "" {
			tagsStr = "{}"
		}
		writer.Write([]string{
			fmt.Sprintf("%d", d.ID),
			d.SourceID,
			fmt.Sprintf("%d", d.DeviceID),
			fmt.Sprintf("%d", d.PlotID),
			d.MetricCode,
			fmt.Sprintf("%.4f", d.Value),
			d.Unit,
			d.EventTime.Format("2006-01-02T15:04:05Z07:00"),
			tagsStr,
		})
	}

	writer.Flush()
}

// JobStatus handles GET /v1/monitoring/jobs/:id — returns an ingest job's status.
func (h *MonitoringDataHandler) JobStatus(c *gin.Context) {
	id := c.Param("id")
	job, ok := h.queueSvc.GetJob(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	resp := gin.H{
		"id":         job.ID,
		"type":       job.Type,
		"status":     job.Status,
		"created_at": job.CreatedAt,
	}
	if job.DoneAt != nil {
		resp["done_at"] = job.DoneAt
	}
	if job.Result != "" {
		var resultMap map[string]interface{}
		if err := json.Unmarshal([]byte(job.Result), &resultMap); err == nil {
			resp["result"] = resultMap
		} else {
			resp["result"] = job.Result
		}
	}
	if job.Error != "" {
		resp["error"] = job.Error
	}

	c.JSON(http.StatusOK, resp)
}
