package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mindflow/agri-platform/pkg/services"
)

type DeviceHandler struct {
	deviceSvc *services.DeviceService
}

func NewDeviceHandler(svc *services.DeviceService) *DeviceHandler {
	return &DeviceHandler{deviceSvc: svc}
}

func (h *DeviceHandler) Create(c *gin.Context) {
	var in services.CreateDeviceInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	device, err := h.deviceSvc.Create(c.Request.Context(), in)
	if err != nil {
		if errors.Is(err, services.ErrDuplicateSerial) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create device"})
		return
	}

	c.JSON(http.StatusCreated, device)
}

func (h *DeviceHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	plotID, _ := strconv.ParseUint(c.Query("plot_id"), 10, 64)
	status := c.Query("status")
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	result, err := h.deviceSvc.List(c.Request.Context(), services.DeviceListParams{
		Page:     page,
		PageSize: pageSize,
		PlotID:   uint(plotID),
		Status:   status,
		UserID:   userID.(uint),
		Role:     role.(string),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list devices"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *DeviceHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid device id"})
		return
	}

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	device, err := h.deviceSvc.GetByID(c.Request.Context(), uint(id), userID.(uint), role.(string))
	if err != nil {
		if errors.Is(err, services.ErrDeviceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
			return
		}
		if errors.Is(err, services.ErrDeviceForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to access this device"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get device"})
		return
	}

	c.JSON(http.StatusOK, device)
}

func (h *DeviceHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid device id"})
		return
	}

	var in services.UpdateDeviceInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	device, err := h.deviceSvc.Update(c.Request.Context(), uint(id), userID.(uint), role.(string), in)
	if err != nil {
		if errors.Is(err, services.ErrDeviceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update device"})
		return
	}

	c.JSON(http.StatusOK, device)
}

func (h *DeviceHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid device id"})
		return
	}

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	if err := h.deviceSvc.Delete(c.Request.Context(), uint(id), userID.(uint), role.(string)); err != nil {
		if errors.Is(err, services.ErrDeviceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
			return
		}
		if errors.Is(err, services.ErrDeviceForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to access this device"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete device"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "device deleted"})
}
