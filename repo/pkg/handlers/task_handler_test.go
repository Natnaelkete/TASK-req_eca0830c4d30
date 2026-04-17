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
	r.GET("/v1/tasks/:id", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("role", "researcher")
		h.Get(c)
	})

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

// POST /v1/tasks/generate — bad JSON path.
func TestTaskHandler_Generate_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewTaskHandler(nil)
	r.POST("/v1/tasks/generate", h.Generate)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/tasks/generate", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// GET /v1/tasks — HTTP route registration check.
func TestTaskHandler_List_HTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())
	h := NewTaskHandler(nil)
	r.GET("/v1/tasks", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("role", "researcher")
		h.List(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/tasks", nil)
	r.ServeHTTP(w, req)
	assert.NotEqual(t, http.StatusNotFound, w.Code)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// PUT /v1/tasks/:id — invalid id.
func TestTaskHandler_Update_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewTaskHandler(nil)
	r.PUT("/v1/tasks/:id", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("role", "researcher")
		h.Update(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/v1/tasks/abc", bytes.NewBufferString(`{"title":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// PATCH /v1/tasks/:id/submit — invalid id.
func TestTaskHandler_Submit_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewTaskHandler(nil)
	r.PATCH("/v1/tasks/:id/submit", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("role", "researcher")
		h.Submit(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/v1/tasks/abc/submit", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// PATCH /v1/tasks/:id/review — invalid id.
func TestTaskHandler_Review_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewTaskHandler(nil)
	r.PATCH("/v1/tasks/:id/review", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("role", "reviewer")
		h.Review(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/v1/tasks/abc/review", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// PATCH /v1/tasks/:id/complete — invalid id.
func TestTaskHandler_Complete_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewTaskHandler(nil)
	r.PATCH("/v1/tasks/:id/complete", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("role", "reviewer")
		h.Complete(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/v1/tasks/abc/complete", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
