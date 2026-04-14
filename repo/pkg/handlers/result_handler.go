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
	userID, _ := c.Get("user_id")
	result, err := h.resultSvc.Create(c.Request.Context(), userID.(uint), in)
	if err != nil {
		if errors.Is(err, services.ErrFieldValidationFailed) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
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
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	result, err := h.resultSvc.List(c.Request.Context(), services.ResultListParams{
		Page: page, PageSize: pageSize, PlotID: uint(plotID), TaskID: uint(taskID),
		Type: c.Query("type"), Status: c.Query("status"),
		UserID: userID.(uint), Role: role.(string),
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
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	result, err := h.resultSvc.GetByID(c.Request.Context(), uint(id), userID.(uint), role.(string))
	if err != nil {
		if errors.Is(err, services.ErrResultNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "result not found"})
			return
		}
		if errors.Is(err, services.ErrResultForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to access this result"})
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
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	var in services.UpdateResultInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.resultSvc.Update(c.Request.Context(), uint(id), userID.(uint), role.(string), in)
	if err != nil {
		if errors.Is(err, services.ErrResultNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "result not found"})
			return
		}
		if errors.Is(err, services.ErrResultArchived) || errors.Is(err, services.ErrResultForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	if err := h.resultSvc.Delete(c.Request.Context(), uint(id), userID.(uint), role.(string)); err != nil {
		if errors.Is(err, services.ErrResultNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "result not found"})
			return
		}
		if errors.Is(err, services.ErrResultArchived) {
			c.JSON(http.StatusForbidden, gin.H{"error": "archived results cannot be deleted"})
			return
		}
		if errors.Is(err, services.ErrResultForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to delete this result"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete result"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "result deleted"})
}

// Transition handles PATCH /v1/results/:id/transition — changes status via state machine.
func (h *ResultHandler) Transition(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid result id"})
		return
	}
	userID, _ := c.Get("user_id")
	var in struct {
		Status string `json:"status" binding:"required"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.resultSvc.Transition(c.Request.Context(), uint(id), userID.(uint), in.Status, in.Reason)
	if err != nil {
		if errors.Is(err, services.ErrResultNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "result not found"})
			return
		}
		if errors.Is(err, services.ErrInvalidTransition) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to transition"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// AppendNotes handles POST /v1/results/:id/notes — adds notes to archived results.
func (h *ResultHandler) AppendNotes(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid result id"})
		return
	}
	userID, _ := c.Get("user_id")
	var in struct {
		Notes string `json:"notes" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.resultSvc.AppendNotes(c.Request.Context(), uint(id), userID.(uint), in.Notes)
	if err != nil {
		if errors.Is(err, services.ErrResultNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "result not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// Invalidate handles POST /v1/results/:id/invalidate — invalidates an archived result.
func (h *ResultHandler) Invalidate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid result id"})
		return
	}
	userID, _ := c.Get("user_id")
	var in struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.resultSvc.Invalidate(c.Request.Context(), uint(id), userID.(uint), in.Reason)
	if err != nil {
		if errors.Is(err, services.ErrResultNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "result not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// CreateFieldRule handles POST /v1/results/field-rules.
func (h *ResultHandler) CreateFieldRule(c *gin.Context) {
	var in services.CreateFieldRuleInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rule, err := h.resultSvc.CreateFieldRule(c.Request.Context(), in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create field rule"})
		return
	}
	c.JSON(http.StatusCreated, rule)
}

// ListFieldRules handles GET /v1/results/field-rules.
func (h *ResultHandler) ListFieldRules(c *gin.Context) {
	rules, err := h.resultSvc.ListFieldRules(c.Request.Context(), c.Query("result_type"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list field rules"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rules})
}
