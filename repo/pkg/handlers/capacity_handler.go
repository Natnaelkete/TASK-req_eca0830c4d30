package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mindflow/agri-platform/pkg/services"
)

type CapacityHandler struct {
	capacitySvc *services.CapacityService
}

func NewCapacityHandler(svc *services.CapacityService) *CapacityHandler {
	return &CapacityHandler{capacitySvc: svc}
}

// CheckDisk handles GET /v1/system/capacity — returns current disk usage.
func (h *CapacityHandler) CheckDisk(c *gin.Context) {
	usage, err := h.capacitySvc.CheckDiskUsage(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check capacity"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"disk_usage_percent": usage, "threshold": services.DiskUsageThreshold})
}

// ListNotifications handles GET /v1/system/notifications — returns system notifications.
func (h *CapacityHandler) ListNotifications(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	notifications, total, err := h.capacitySvc.ListNotifications(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list notifications"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": notifications, "total": total})
}
