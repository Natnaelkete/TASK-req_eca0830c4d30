package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mindflow/agri-platform/pkg/services"
)

// IndicatorHandler exposes HTTP endpoints for indicator version management.
type IndicatorHandler struct {
	indicatorSvc *services.IndicatorService
}

func NewIndicatorHandler(svc *services.IndicatorService) *IndicatorHandler {
	return &IndicatorHandler{indicatorSvc: svc}
}

// Create handles POST /v1/indicators — creates a new indicator definition.
func (h *IndicatorHandler) Create(c *gin.Context) {
	var in services.CreateIndicatorInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := c.Get("user_id")
	indicator, err := h.indicatorSvc.Create(c.Request.Context(), userID.(uint), in)
	if err != nil {
		if errors.Is(err, services.ErrIndicatorExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create indicator"})
		return
	}
	c.JSON(http.StatusCreated, indicator)
}

// List handles GET /v1/indicators — lists indicator definitions.
func (h *IndicatorHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.indicatorSvc.List(c.Request.Context(), services.IndicatorListParams{
		Page:     page,
		PageSize: pageSize,
		Category: c.Query("category"),
		Status:   c.Query("status"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list indicators"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// Get handles GET /v1/indicators/:id — returns an indicator definition.
func (h *IndicatorHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid indicator id"})
		return
	}
	indicator, err := h.indicatorSvc.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, services.ErrIndicatorNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "indicator not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get indicator"})
		return
	}
	c.JSON(http.StatusOK, indicator)
}

// Update handles PUT /v1/indicators/:id — updates an indicator and records a new version.
func (h *IndicatorHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid indicator id"})
		return
	}
	userID, _ := c.Get("user_id")
	var in services.UpdateIndicatorInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	indicator, err := h.indicatorSvc.Update(c.Request.Context(), uint(id), userID.(uint), in)
	if err != nil {
		if errors.Is(err, services.ErrIndicatorNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "indicator not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, indicator)
}

// Delete handles DELETE /v1/indicators/:id — deprecates an indicator.
func (h *IndicatorHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid indicator id"})
		return
	}
	if err := h.indicatorSvc.Delete(c.Request.Context(), uint(id)); err != nil {
		if errors.Is(err, services.ErrIndicatorNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "indicator not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deprecate indicator"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "indicator deprecated"})
}

// ListVersions handles GET /v1/indicators/:id/versions — lists all versions of an indicator.
func (h *IndicatorHandler) ListVersions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid indicator id"})
		return
	}
	versions, err := h.indicatorSvc.ListVersions(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list versions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": versions})
}

// GetVersion handles GET /v1/indicators/:id/versions/:version — returns a specific version.
func (h *IndicatorHandler) GetVersion(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid indicator id"})
		return
	}
	ver, err := strconv.Atoi(c.Param("version"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid version number"})
		return
	}
	version, err := h.indicatorSvc.GetVersion(c.Request.Context(), uint(id), ver)
	if err != nil {
		if errors.Is(err, services.ErrIndicatorNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "version not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get version"})
		return
	}
	c.JSON(http.StatusOK, version)
}
