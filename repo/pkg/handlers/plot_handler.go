package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mindflow/agri-platform/pkg/services"
)

type PlotHandler struct {
	plotSvc *services.PlotService
}

func NewPlotHandler(svc *services.PlotService) *PlotHandler {
	return &PlotHandler{plotSvc: svc}
}

func (h *PlotHandler) Create(c *gin.Context) {
	var in services.CreatePlotInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")
	plot, err := h.plotSvc.Create(c.Request.Context(), userID, in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create plot"})
		return
	}

	c.JSON(http.StatusCreated, plot)
}

func (h *PlotHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	userIDFilter, _ := strconv.ParseUint(c.Query("user_id"), 10, 64)

	result, err := h.plotSvc.List(c.Request.Context(), services.PlotListParams{
		Page:     page,
		PageSize: pageSize,
		UserID:   uint(userIDFilter),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list plots"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *PlotHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plot id"})
		return
	}

	plot, err := h.plotSvc.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, services.ErrPlotNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "plot not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get plot"})
		return
	}

	c.JSON(http.StatusOK, plot)
}

func (h *PlotHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plot id"})
		return
	}

	var in services.UpdatePlotInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	plot, err := h.plotSvc.Update(c.Request.Context(), uint(id), in)
	if err != nil {
		if errors.Is(err, services.ErrPlotNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "plot not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update plot"})
		return
	}

	c.JSON(http.StatusOK, plot)
}

func (h *PlotHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plot id"})
		return
	}

	if err := h.plotSvc.Delete(c.Request.Context(), uint(id)); err != nil {
		if errors.Is(err, services.ErrPlotNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "plot not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete plot"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "plot deleted"})
}
