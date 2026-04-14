package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"bytes"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewTaskHandler(t *testing.T) {
	h := NewTaskHandler(nil)
	assert.NotNil(t, h)
}

func TestTaskHandler_Create_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewTaskHandler(nil)
	r.POST("/v1/tasks", h.Create)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/tasks", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_Get_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewTaskHandler(nil)
	r.GET("/v1/tasks/:id", h.Get)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/tasks/abc", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_Delete_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewTaskHandler(nil)
	r.DELETE("/v1/tasks/:id", h.Delete)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/v1/tasks/abc", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
