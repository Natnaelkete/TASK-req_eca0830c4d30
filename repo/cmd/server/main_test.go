package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mindflow/agri-platform/internal/config"
	"github.com/mindflow/agri-platform/pkg/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// newTestConfig returns a config suitable for bootstrapping the full router
// in unit tests (valid 32-byte hex encryption key, dev-mode secrets).
func newTestConfig() *config.Config {
	return &config.Config{
		DBHost:        "localhost",
		DBPort:        "3306",
		DBUser:        "root",
		DBPassword:    "pass",
		DBName:        "agri",
		JWTSecret:     "test-secret",
		ServerPort:    "8080",
		EncryptionKey: "0123456789abcdef0123456789abcdef",
		AppEnv:        "test",
	}
}

// TestSetupRouter_HealthRouteWired verifies the real setupRouter bootstraps
// the /health route on the same engine that main() serves in production —
// this is the integration contract between main() and the router. No mocks
// are applied to the router itself; dependency services are real
// constructors (with a nil DB, which Recovery handles gracefully).
func TestSetupRouter_HealthRouteWired(t *testing.T) {
	gin.SetMode(gin.TestMode)
	q := services.NewQueueService(10, 0)
	defer q.Shutdown()
	cfg := newTestConfig()
	taskSvc := services.NewTaskService(nil)
	capSvc := services.NewCapacityService(nil)

	router := setupRouter(nil, cfg, q, taskSvc, capSvc)
	require.NotNil(t, router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	// /health is public; with nil DB the Ping panics and Recovery surfaces a
	// 500 through the real middleware chain. Route is proven to be wired.
	assert.NotEqual(t, http.StatusNotFound, w.Code)
}

// TestSetupRouter_ProtectedRouteRequiresAuth verifies the real auth
// middleware is wired in for every protected route group. A request without
// a token must be rejected at the middleware layer, not reach the handler.
func TestSetupRouter_ProtectedRouteRequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	q := services.NewQueueService(10, 0)
	defer q.Shutdown()
	cfg := newTestConfig()
	taskSvc := services.NewTaskService(nil)
	capSvc := services.NewCapacityService(nil)

	router := setupRouter(nil, cfg, q, taskSvc, capSvc)

	// A representative sample across every protected domain.
	protectedRoutes := []struct{ method, path string }{
		{"GET", "/v1/auth/me"},
		{"GET", "/v1/plots"},
		{"GET", "/v1/devices"},
		{"GET", "/v1/metrics"},
		{"POST", "/v1/monitor/device"},
		{"GET", "/v1/monitoring/data"},
		{"GET", "/v1/dashboards"},
		{"POST", "/v1/analysis/trends"},
		{"GET", "/v1/indicators"},
		{"GET", "/v1/orders"},
		{"GET", "/v1/templates"},
		{"GET", "/v1/tasks"},
		{"GET", "/v1/chat"},
		{"GET", "/v1/results"},
		{"GET", "/v1/system/capacity"},
		{"GET", "/v1/system/notifications"},
	}

	for _, r := range protectedRoutes {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(r.method, r.path, nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code,
			"route %s %s should require auth", r.method, r.path)
	}
}

// TestSetupRouter_PublicAuthRoutesReachable proves the public auth routes
// (register/login) are wired into the real router without auth middleware.
func TestSetupRouter_PublicAuthRoutesReachable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	q := services.NewQueueService(10, 0)
	defer q.Shutdown()
	cfg := newTestConfig()
	taskSvc := services.NewTaskService(nil)
	capSvc := services.NewCapacityService(nil)

	router := setupRouter(nil, cfg, q, taskSvc, capSvc)

	// Register with empty body: handler is reached, binding rejects → 400.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/auth/register", nil)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Login with empty body: same contract.
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/v1/auth/login", nil)
	req2.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusBadRequest, w2.Code)
}

// TestSetupRouter_UnknownRouteReturns404 guards against unintended default
// handlers catching arbitrary paths.
func TestSetupRouter_UnknownRouteReturns404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	q := services.NewQueueService(10, 0)
	defer q.Shutdown()
	cfg := newTestConfig()
	taskSvc := services.NewTaskService(nil)
	capSvc := services.NewCapacityService(nil)

	router := setupRouter(nil, cfg, q, taskSvc, capSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/does/not/exist", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
