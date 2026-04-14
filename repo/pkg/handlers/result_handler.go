package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mindflow/agri-platform/pkg/services"
)

type ResultHandler struct {
	resultSvc *services.ResultService
}

func NewResultHandler(svc *services.ResultService) *ResultHandler {
	return &ResultHandler{resultSvc: svc}
}

func (h *ResultHandler) Create(c *gin.Context) {
	var in services.CreateResultInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")
	result, err := h.resultSvc.Create(c.Request.Context(), userID, in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create result"})
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (h *ResultHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	plotID, _ := strconv.ParseUint(c.Query("plot_id"), 10, 64)
	taskID, _ := strconv.ParseUint(c.Query("task_id"), 10, 64)

	result, err := h.resultSvc.List(c.Request.Context(), services.ResultListParams{
		Page:     page,
		PageSize: pageSize,
		PlotID:   uint(plotID),
		TaskID:   uint(taskID),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list results"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *ResultHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid result id"})
		return
	}

	result, err := h.resultSvc.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, services.ErrResultNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "result not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get result"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *ResultHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid result id"})
		return
	}

	var in services.UpdateResultInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.resultSvc.Update(c.Request.Context(), uint(id), in)
	if err != nil {
		if errors.Is(err, services.ErrResultNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "result not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update result"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *ResultHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid result id"})
		return
	}

	if err := h.resultSvc.Delete(c.Request.Context(), uint(id)); err != nil {
		if errors.Is(err, services.ErrResultNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "result not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete result"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "result deleted"})
}
