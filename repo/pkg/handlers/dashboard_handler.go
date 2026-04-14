package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mindflow/agri-platform/pkg/services"
)

type DashboardHandler struct {
	dashSvc *services.DashboardService
}

func NewDashboardHandler(svc *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashSvc: svc}
}

// Create handles POST /v1/dashboards — saves a new dashboard config for the authenticated user.
func (h *DashboardHandler) Create(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var in services.CreateDashboardInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cfg, err := h.dashSvc.Create(c.Request.Context(), userID.(uint), in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create dashboard"})
		return
	}

	c.JSON(http.StatusCreated, cfg)
}

// List handles GET /v1/dashboards — lists dashboard configs for the authenticated user.
func (h *DashboardHandler) List(c *gin.Context) {
	userID, _ := c.Get("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.dashSvc.List(c.Request.Context(), services.DashboardListParams{
		Page:     page,
		PageSize: pageSize,
		UserID:   userID.(uint),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list dashboards"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Get handles GET /v1/dashboards/:id — loads a saved dashboard config.
func (h *DashboardHandler) Get(c *gin.Context) {
	userID, _ := c.Get("user_id")
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid dashboard id"})
		return
	}

	cfg, err := h.dashSvc.GetByID(c.Request.Context(), uint(id), userID.(uint))
	if err != nil {
		if errors.Is(err, services.ErrDashboardNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "dashboard not found"})
			return
		}
		if errors.Is(err, services.ErrDashboardForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to access this dashboard"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get dashboard"})
		return
	}

	c.JSON(http.StatusOK, cfg)
}

// Update handles PUT /v1/dashboards/:id — updates a dashboard config.
func (h *DashboardHandler) Update(c *gin.Context) {
	userID, _ := c.Get("user_id")
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid dashboard id"})
		return
	}

	var in services.UpdateDashboardInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cfg, err := h.dashSvc.Update(c.Request.Context(), uint(id), userID.(uint), in)
	if err != nil {
		if errors.Is(err, services.ErrDashboardNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "dashboard not found"})
			return
		}
		if errors.Is(err, services.ErrDashboardForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to update this dashboard"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update dashboard"})
		return
	}

	c.JSON(http.StatusOK, cfg)
}

// Delete handles DELETE /v1/dashboards/:id — removes a dashboard config.
func (h *DashboardHandler) Delete(c *gin.Context) {
	userID, _ := c.Get("user_id")
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid dashboard id"})
		return
	}

	if err := h.dashSvc.Delete(c.Request.Context(), uint(id), userID.(uint)); err != nil {
		if errors.Is(err, services.ErrDashboardNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "dashboard not found"})
			return
		}
		if errors.Is(err, services.ErrDashboardForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to delete this dashboard"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete dashboard"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "dashboard deleted"})
}
