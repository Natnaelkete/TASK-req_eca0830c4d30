package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"bytes"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewResultHandler(t *testing.T) {
	h := NewResultHandler(nil)
	assert.NotNil(t, h)
}

func TestResultHandler_Create_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewResultHandler(nil)
	r.POST("/v1/results", h.Create)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/results", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestResultHandler_Get_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewResultHandler(nil)
	r.GET("/v1/results/:id", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("role", "researcher")
		h.Get(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/results/abc", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestResultHandler_Delete_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewResultHandler(nil)
	r.DELETE("/v1/results/:id", h.Delete)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/v1/results/abc", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// GET /v1/results — HTTP route registration check.
func TestResultHandler_List_HTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())
	h := NewResultHandler(nil)
	r.GET("/v1/results", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("role", "researcher")
		h.List(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/results", nil)
	r.ServeHTTP(w, req)
	assert.NotEqual(t, http.StatusNotFound, w.Code)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// PUT /v1/results/:id — invalid id path.
func TestResultHandler_Update_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewResultHandler(nil)
	r.PUT("/v1/results/:id", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("role", "researcher")
		h.Update(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/v1/results/abc", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// PATCH /v1/results/:id/transition — invalid id + missing status.
func TestResultHandler_Transition_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewResultHandler(nil)
	r.PATCH("/v1/results/:id/transition", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		h.Transition(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/v1/results/abc/transition", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// POST /v1/results/:id/notes — invalid id path.
func TestResultHandler_AppendNotes_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewResultHandler(nil)
	r.POST("/v1/results/:id/notes", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		h.AppendNotes(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/results/abc/notes", bytes.NewBufferString(`{"notes":"hi"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// POST /v1/results/:id/invalidate — invalid id path.
func TestResultHandler_Invalidate_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewResultHandler(nil)
	r.POST("/v1/results/:id/invalidate", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		h.Invalidate(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/results/abc/invalidate", bytes.NewBufferString(`{"reason":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// POST /v1/results/field-rules — bad JSON.
func TestResultHandler_CreateFieldRule_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewResultHandler(nil)
	r.POST("/v1/results/field-rules", h.CreateFieldRule)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/results/field-rules", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// GET /v1/results/field-rules — HTTP route registration check.
func TestResultHandler_ListFieldRules_HTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())
	h := NewResultHandler(nil)
	r.GET("/v1/results/field-rules", h.ListFieldRules)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/results/field-rules", nil)
	r.ServeHTTP(w, req)
	assert.NotEqual(t, http.StatusNotFound, w.Code)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
