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

// GET /v1/chat — HTTP route registration check.
func TestChatHandler_List_HTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())
	h := NewChatHandler(nil)
	r.GET("/v1/chat", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		h.List(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/chat", nil)
	r.ServeHTTP(w, req)
	assert.NotEqual(t, http.StatusNotFound, w.Code)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
