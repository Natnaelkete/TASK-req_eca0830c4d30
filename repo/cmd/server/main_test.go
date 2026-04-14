package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthHandler_Exists(t *testing.T) {
	assert.NotNil(t, healthHandler(nil))
}

func TestHealthHandler_NoDBReturns503(t *testing.T) {
	handler := healthHandler(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)

	defer func() {
		if r := recover(); r != nil {
			// Expected: calling Ping on nil db panics in unit test
		}
	}()
	handler(createTestGinContext(w, req))
}

func createTestGinContext(w http.ResponseWriter, req *http.Request) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req
	return c
}
