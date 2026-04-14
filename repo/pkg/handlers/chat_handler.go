package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mindflow/agri-platform/pkg/services"
)

type ChatHandler struct {
	chatSvc *services.ChatService
}

func NewChatHandler(svc *services.ChatService) *ChatHandler {
	return &ChatHandler{chatSvc: svc}
}

func (h *ChatHandler) Send(c *gin.Context) {
	var in services.SendMessageInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	senderID := c.GetUint("user_id")
	msg, err := h.chatSvc.Send(c.Request.Context(), senderID, in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send message"})
		return
	}

	c.JSON(http.StatusCreated, msg)
}

func (h *ChatHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	plotID, _ := strconv.ParseUint(c.Query("plot_id"), 10, 64)

	userID := c.GetUint("user_id")

	result, err := h.chatSvc.List(c.Request.Context(), services.MessageListParams{
		Page:     page,
		PageSize: pageSize,
		UserID:   userID,
		PlotID:   uint(plotID),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list messages"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *ChatHandler) MarkRead(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message id"})
		return
	}

	userID := c.GetUint("user_id")
	if err := h.chatSvc.MarkRead(c.Request.Context(), uint(id), userID); err != nil {
		if errors.Is(err, services.ErrMessageNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark message as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "marked as read"})
}
