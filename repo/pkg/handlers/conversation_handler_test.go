package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mindflow/agri-platform/pkg/services"
	"github.com/stretchr/testify/assert"
)

func TestNewConversationHandler(t *testing.T) {
	h := NewConversationHandler(nil)
	assert.NotNil(t, h)
}

func TestConversationHandler_CreateOrder_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := services.NewConversationService(nil)
	h := NewConversationHandler(svc)
	r := gin.New()
	r.POST("/v1/orders", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		h.CreateOrder(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/orders", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConversationHandler_PostMessage_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := services.NewConversationService(nil)
	h := NewConversationHandler(svc)
	r := gin.New()
	r.POST("/v1/orders/:id/messages", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		h.PostMessage(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/orders/1/messages", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConversationHandler_PostMessage_InvalidOrderID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := services.NewConversationService(nil)
	h := NewConversationHandler(svc)
	r := gin.New()
	r.POST("/v1/orders/:id/messages", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		h.PostMessage(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/orders/abc/messages", bytes.NewBufferString(`{"message":"hi"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConversationHandler_GetOrder_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := services.NewConversationService(nil)
	h := NewConversationHandler(svc)
	r := gin.New()
	r.GET("/v1/orders/:id", h.GetOrder)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/orders/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConversationHandler_MarkRead_InvalidMsgID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := services.NewConversationService(nil)
	h := NewConversationHandler(svc)
	r := gin.New()
	r.PATCH("/v1/orders/:id/messages/:msg_id/read", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("role", "researcher")
		h.MarkRead(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/v1/orders/1/messages/abc/read", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConversationHandler_MarkRead_InvalidOrderID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := services.NewConversationService(nil)
	h := NewConversationHandler(svc)
	r := gin.New()
	r.PATCH("/v1/orders/:id/messages/:msg_id/read", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("role", "researcher")
		h.MarkRead(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/v1/orders/abc/messages/1/read", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConversationHandler_TransferTicket_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := services.NewConversationService(nil)
	h := NewConversationHandler(svc)
	r := gin.New()
	r.POST("/v1/orders/:id/transfer", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		h.TransferTicket(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/orders/1/transfer", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConversationHandler_CreateTemplate_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := services.NewConversationService(nil)
	h := NewConversationHandler(svc)
	r := gin.New()
	r.POST("/v1/templates", h.CreateTemplate)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/templates", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// GET /v1/orders — HTTP route registration check.
func TestConversationHandler_ListOrders_HTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := services.NewConversationService(nil)
	h := NewConversationHandler(svc)
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/v1/orders", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("role", "researcher")
		h.ListOrders(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/orders", nil)
	r.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusNotFound, w.Code)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// GET /v1/orders/:id/messages — HTTP route registration check.
func TestConversationHandler_ListMessages_HTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := services.NewConversationService(nil)
	h := NewConversationHandler(svc)
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/v1/orders/:id/messages", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("role", "researcher")
		h.ListMessages(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/orders/1/messages", nil)
	r.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusNotFound, w.Code)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// GET /v1/templates — HTTP route registration check.
func TestConversationHandler_ListTemplates_HTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := services.NewConversationService(nil)
	h := NewConversationHandler(svc)
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/v1/templates", h.ListTemplates)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/templates", nil)
	r.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusNotFound, w.Code)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestConversationHandler_SendTemplate_InvalidIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := services.NewConversationService(nil)
	h := NewConversationHandler(svc)
	r := gin.New()
	r.POST("/v1/orders/:id/templates/:template_id", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		h.SendTemplate(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/orders/abc/templates/1", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/v1/orders/1/templates/abc", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusBadRequest, w2.Code)
}
