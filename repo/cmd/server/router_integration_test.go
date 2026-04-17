package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mindflow/agri-platform/internal/config"
	"github.com/mindflow/agri-platform/pkg/services"
)

func makeAdminToken(secret string) string {
	now := time.Now()
	claims := services.Claims{
		UserID:   1,
		Username: "admin",
		Role:     "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			Subject:   fmt.Sprintf("%d", 1),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		panic(err)
	}
	return signed
}

func TestSetupRouter_AllRoutesAreReachable(t *testing.T) {
	jwtSecret := "test-secret"
	cfg := &config.Config{
		JWTSecret:     jwtSecret,
		EncryptionKey: "0123456789abcdef0123456789abcdef",
	}
	adminToken := makeAdminToken(jwtSecret)

	queueSvc := services.NewQueueService(32, 0)
	defer queueSvc.Shutdown()

	router := setupRouter(nil, cfg, queueSvc, services.NewTaskService(nil), services.NewCapacityService(nil))

	tests := []struct {
		name         string
		method       string
		path         string
		body         string
		expectStatus int
		protected    bool
	}{
		{name: "health", method: http.MethodGet, path: "/health", expectStatus: http.StatusInternalServerError},
		{name: "register bad json", method: http.MethodPost, path: "/v1/auth/register", body: `{}`, expectStatus: http.StatusBadRequest},
		{name: "login bad json", method: http.MethodPost, path: "/v1/auth/login", body: `{}`, expectStatus: http.StatusBadRequest},
		{name: "me", method: http.MethodGet, path: "/v1/auth/me", protected: true},

		{name: "plots post", method: http.MethodPost, path: "/v1/plots", body: `{}`, protected: true},
		{name: "plots list", method: http.MethodGet, path: "/v1/plots", protected: true},
		{name: "plots get", method: http.MethodGet, path: "/v1/plots/1", protected: true},
		{name: "plots put", method: http.MethodPut, path: "/v1/plots/1", body: `{}`, protected: true},
		{name: "plots delete", method: http.MethodDelete, path: "/v1/plots/1", protected: true},

		{name: "devices post", method: http.MethodPost, path: "/v1/devices", body: `{}`, protected: true},
		{name: "devices list", method: http.MethodGet, path: "/v1/devices", protected: true},
		{name: "devices get", method: http.MethodGet, path: "/v1/devices/1", protected: true},
		{name: "devices put", method: http.MethodPut, path: "/v1/devices/1", body: `{}`, protected: true},
		{name: "devices delete", method: http.MethodDelete, path: "/v1/devices/1", protected: true},

		{name: "metrics post", method: http.MethodPost, path: "/v1/metrics", body: `{}`, protected: true},
		{name: "metrics batch", method: http.MethodPost, path: "/v1/metrics/batch", body: `{}`, protected: true},
		{name: "metrics list", method: http.MethodGet, path: "/v1/metrics", protected: true},
		{name: "metrics get", method: http.MethodGet, path: "/v1/metrics/1", protected: true},
		{name: "metrics delete", method: http.MethodDelete, path: "/v1/metrics/1", protected: true},

		{name: "monitor device", method: http.MethodPost, path: "/v1/monitor/device", body: `{}`, protected: true},
		{name: "monitor threshold", method: http.MethodPost, path: "/v1/monitor/threshold", body: `{}`, protected: true},
		{name: "monitor job", method: http.MethodGet, path: "/v1/monitor/jobs/job-1", protected: true},
		{name: "monitor queue", method: http.MethodGet, path: "/v1/monitor/queue/status", protected: true},
		{name: "monitor alerts", method: http.MethodGet, path: "/v1/monitor/alerts", protected: true},
		{name: "monitor resolve", method: http.MethodPatch, path: "/v1/monitor/alerts/1/resolve", protected: true},

		{name: "monitoring ingest", method: http.MethodPost, path: "/v1/monitoring/ingest", body: `{}`, protected: true},
		{name: "monitoring list", method: http.MethodGet, path: "/v1/monitoring/data", protected: true},
		{name: "monitoring get", method: http.MethodGet, path: "/v1/monitoring/data/1", protected: true},
		{name: "monitoring aggregate", method: http.MethodPost, path: "/v1/monitoring/aggregate", body: `{}`, protected: true},
		{name: "monitoring curve", method: http.MethodPost, path: "/v1/monitoring/curve", body: `{}`, protected: true},
		{name: "monitoring trends", method: http.MethodPost, path: "/v1/monitoring/trends", body: `{}`, protected: true},
		{name: "monitoring export json", method: http.MethodGet, path: "/v1/monitoring/export/json", protected: true},
		{name: "monitoring export csv", method: http.MethodGet, path: "/v1/monitoring/export/csv", protected: true},
		{name: "monitoring job", method: http.MethodGet, path: "/v1/monitoring/jobs/job-1", protected: true},

		{name: "dashboards post", method: http.MethodPost, path: "/v1/dashboards", body: `{}`, protected: true},
		{name: "dashboards list", method: http.MethodGet, path: "/v1/dashboards", protected: true},
		{name: "dashboards get", method: http.MethodGet, path: "/v1/dashboards/1", protected: true},
		{name: "dashboards put", method: http.MethodPut, path: "/v1/dashboards/1", body: `{}`, protected: true},
		{name: "dashboards delete", method: http.MethodDelete, path: "/v1/dashboards/1", protected: true},

		{name: "analysis trends", method: http.MethodPost, path: "/v1/analysis/trends", body: `{}`, protected: true},
		{name: "analysis funnels", method: http.MethodPost, path: "/v1/analysis/funnels", body: `{}`, protected: true},
		{name: "analysis retention", method: http.MethodPost, path: "/v1/analysis/retention", body: `{}`, protected: true},

		{name: "indicators post", method: http.MethodPost, path: "/v1/indicators", body: `{}`, protected: true},
		{name: "indicators list", method: http.MethodGet, path: "/v1/indicators", protected: true},
		{name: "indicators get", method: http.MethodGet, path: "/v1/indicators/1", protected: true},
		{name: "indicators put", method: http.MethodPut, path: "/v1/indicators/1", body: `{}`, protected: true},
		{name: "indicators delete", method: http.MethodDelete, path: "/v1/indicators/1", protected: true},
		{name: "indicators versions", method: http.MethodGet, path: "/v1/indicators/1/versions", protected: true},
		{name: "indicators version", method: http.MethodGet, path: "/v1/indicators/1/versions/1", protected: true},

		{name: "orders post", method: http.MethodPost, path: "/v1/orders", body: `{}`, protected: true},
		{name: "orders list", method: http.MethodGet, path: "/v1/orders", protected: true},
		{name: "orders get", method: http.MethodGet, path: "/v1/orders/1", protected: true},
		{name: "orders post message", method: http.MethodPost, path: "/v1/orders/1/messages", body: `{}`, protected: true},
		{name: "orders list messages", method: http.MethodGet, path: "/v1/orders/1/messages", protected: true},
		{name: "orders mark read", method: http.MethodPatch, path: "/v1/orders/1/messages/1/read", protected: true},
		{name: "orders transfer", method: http.MethodPost, path: "/v1/orders/1/transfer", body: `{}`, protected: true},
		{name: "orders send template", method: http.MethodPost, path: "/v1/orders/1/templates/1", protected: true},

		{name: "templates post", method: http.MethodPost, path: "/v1/templates", body: `{}`, protected: true},
		{name: "templates list", method: http.MethodGet, path: "/v1/templates", protected: true},

		{name: "tasks post", method: http.MethodPost, path: "/v1/tasks", body: `{}`, protected: true},
		{name: "tasks generate", method: http.MethodPost, path: "/v1/tasks/generate", body: `{}`, protected: true},
		{name: "tasks list", method: http.MethodGet, path: "/v1/tasks", protected: true},
		{name: "tasks get", method: http.MethodGet, path: "/v1/tasks/1", protected: true},
		{name: "tasks put", method: http.MethodPut, path: "/v1/tasks/1", body: `{}`, protected: true},
		{name: "tasks delete", method: http.MethodDelete, path: "/v1/tasks/1", protected: true},
		{name: "tasks submit", method: http.MethodPatch, path: "/v1/tasks/1/submit", protected: true},
		{name: "tasks review", method: http.MethodPatch, path: "/v1/tasks/1/review", protected: true},
		{name: "tasks complete", method: http.MethodPatch, path: "/v1/tasks/1/complete", protected: true},

		{name: "chat post", method: http.MethodPost, path: "/v1/chat", body: `{}`, protected: true},
		{name: "chat list", method: http.MethodGet, path: "/v1/chat", protected: true},
		{name: "chat read", method: http.MethodPatch, path: "/v1/chat/1/read", protected: true},

		{name: "results post", method: http.MethodPost, path: "/v1/results", body: `{}`, protected: true},
		{name: "results list", method: http.MethodGet, path: "/v1/results", protected: true},
		{name: "results get", method: http.MethodGet, path: "/v1/results/1", protected: true},
		{name: "results put", method: http.MethodPut, path: "/v1/results/1", body: `{}`, protected: true},
		{name: "results delete", method: http.MethodDelete, path: "/v1/results/1", protected: true},
		{name: "results transition", method: http.MethodPatch, path: "/v1/results/1/transition", protected: true},
		{name: "results notes", method: http.MethodPost, path: "/v1/results/1/notes", body: `{}`, protected: true},
		{name: "results invalidate", method: http.MethodPost, path: "/v1/results/1/invalidate", body: `{}`, protected: true},
		{name: "results field rules create", method: http.MethodPost, path: "/v1/results/field-rules", body: `{}`, protected: true},
		{name: "results field rules list", method: http.MethodGet, path: "/v1/results/field-rules", protected: true},

		{name: "system capacity", method: http.MethodGet, path: "/v1/system/capacity", protected: true},
		{name: "system notifications", method: http.MethodGet, path: "/v1/system/notifications", protected: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var body *bytes.Buffer
			if tc.body != "" {
				body = bytes.NewBufferString(tc.body)
			} else {
				body = bytes.NewBuffer(nil)
			}

			req := httptest.NewRequest(tc.method, tc.path, body)
			if tc.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			if tc.protected {
				req.Header.Set("Authorization", "Bearer "+adminToken)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if tc.protected {
				if w.Code == http.StatusUnauthorized || w.Code == http.StatusForbidden {
					t.Fatalf("%s %s did not reach handler (status=%d), body=%s", tc.method, tc.path, w.Code, w.Body.String())
				}
				return
			}

			if w.Code != tc.expectStatus {
				t.Fatalf("%s %s expected status %d, got %d, body=%s", tc.method, tc.path, tc.expectStatus, w.Code, w.Body.String())
			}
		})
	}
}
