package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mindflow/agri-platform/pkg/services"
)

type ConversationHandler struct {
	convSvc *services.ConversationService
}

func NewConversationHandler(svc *services.ConversationService) *ConversationHandler {
	return &ConversationHandler{convSvc: svc}
}

// --- Orders ---

func (h *ConversationHandler) CreateOrder(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var in services.CreateOrderInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	order, err := h.convSvc.CreateOrderWithTitle(c.Request.Context(), userID.(uint), in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order"})
		return
	}
	c.JSON(http.StatusCreated, order)
}

func (h *ConversationHandler) GetOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}
	order, err := h.convSvc.GetOrder(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, services.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get order"})
		return
	}
	c.JSON(http.StatusOK, order)
}

func (h *ConversationHandler) ListOrders(c *gin.Context) {
	userID, _ := c.Get("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.convSvc.ListOrders(c.Request.Context(), services.OrderListParams{
		Page:     page,
		PageSize: pageSize,
		UserID:   userID.(uint),
		Status:   c.Query("status"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list orders"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// --- Conversation Messages ---

func (h *ConversationHandler) PostMessage(c *gin.Context) {
	userID, _ := c.Get("user_id")
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	var in services.PostMessageInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conv, err := h.convSvc.PostMessage(c.Request.Context(), uint(orderID), userID.(uint), in)
	if err != nil {
		if errors.Is(err, services.ErrRateLimitExceeded) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, services.ErrSensitiveWord) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error(), "blocked": true})
			return
		}
		if errors.Is(err, services.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to post message"})
		return
	}

	c.JSON(http.StatusCreated, conv)
}

func (h *ConversationHandler) ListMessages(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	msgs, total, err := h.convSvc.ListMessages(c.Request.Context(), uint(orderID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": msgs, "total": total})
}

func (h *ConversationHandler) MarkRead(c *gin.Context) {
	userID, _ := c.Get("user_id")
	msgID, err := strconv.ParseUint(c.Param("msg_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message id"})
		return
	}

	if err := h.convSvc.MarkRead(c.Request.Context(), uint(msgID), userID.(uint)); err != nil {
		if errors.Is(err, services.ErrConversationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark read"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "marked as read"})
}

// --- Ticket Transfer ---

func (h *ConversationHandler) TransferTicket(c *gin.Context) {
	userID, _ := c.Get("user_id")
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	var in services.TransferInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conv, err := h.convSvc.TransferTicket(c.Request.Context(), uint(orderID), userID.(uint), in)
	if err != nil {
		if errors.Is(err, services.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to transfer ticket"})
		return
	}

	c.JSON(http.StatusOK, conv)
}

// --- Templates ---

func (h *ConversationHandler) CreateTemplate(c *gin.Context) {
	var in services.CreateTemplateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tmpl, err := h.convSvc.CreateTemplate(c.Request.Context(), in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create template"})
		return
	}
	c.JSON(http.StatusCreated, tmpl)
}

func (h *ConversationHandler) ListTemplates(c *gin.Context) {
	templates, err := h.convSvc.ListTemplates(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list templates"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": templates})
}

func (h *ConversationHandler) SendTemplate(c *gin.Context) {
	userID, _ := c.Get("user_id")
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}
	templateID, err := strconv.ParseUint(c.Param("template_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template id"})
		return
	}

	conv, err := h.convSvc.SendTemplate(c.Request.Context(), uint(orderID), userID.(uint), uint(templateID))
	if err != nil {
		if errors.Is(err, services.ErrTemplateNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
			return
		}
		if errors.Is(err, services.ErrRateLimitExceeded) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send template"})
		return
	}
	c.JSON(http.StatusCreated, conv)
}
