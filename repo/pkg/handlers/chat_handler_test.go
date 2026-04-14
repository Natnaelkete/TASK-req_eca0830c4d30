package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"bytes"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewChatHandler(t *testing.T) {
	h := NewChatHandler(nil)
	assert.NotNil(t, h)
}

func TestChatHandler_Send_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewChatHandler(nil)
	r.POST("/v1/chat", h.Send)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/chat", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestChatHandler_MarkRead_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewChatHandler(nil)
	r.PATCH("/v1/chat/:id/read", h.MarkRead)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/v1/chat/abc/read", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
