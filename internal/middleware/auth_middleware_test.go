package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sharedlog "clinic-vet-api/internal/shared/log"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestRequireAnyRoleCanonicalRoles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previousLogger := sharedlog.App
	sharedlog.App = zap.NewNop()
	t.Cleanup(func() {
		sharedlog.App = previousLogger
	})

	tests := []struct {
		name       string
		role       string
		allowed    []string
		wantStatus int
	}{
		{name: "admin can access admin route", role: "admin", allowed: []string{"admin"}, wantStatus: http.StatusNoContent},
		{name: "receptionist can access reception route", role: "receptionist", allowed: []string{"admin", "receptionist"}, wantStatus: http.StatusNoContent},
		{name: "veterinarian can access veterinarian route", role: "veterinarian", allowed: []string{"admin", "veterinarian"}, wantStatus: http.StatusNoContent},
		{name: "customer cannot access staff route", role: "customer", allowed: []string{"admin", "receptionist"}, wantStatus: http.StatusForbidden},
		{name: "legacy manager role is denied", role: "manager", allowed: []string{"admin", "receptionist"}, wantStatus: http.StatusForbidden},
		{name: "legacy employee role is denied", role: "employee", allowed: []string{"admin", "veterinarian"}, wantStatus: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("userRole", tt.role)
			})
			router.Use(NewAuthMiddleware("test-secret").RequireAnyRole(tt.allowed...))
			router.GET("/", func(c *gin.Context) {
				c.Status(http.StatusNoContent)
			})

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			router.ServeHTTP(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, recorder.Code)
			}
		})
	}
}

func TestAuditLogDoesNotRecordPayloadsOrHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previousAudit, previousApp := sharedlog.Audit, sharedlog.App
	core, observed := observer.New(zap.InfoLevel)
	sharedlog.Audit = zap.New(core)
	sharedlog.App = zap.NewNop()
	t.Cleanup(func() {
		sharedlog.Audit = previousAudit
		sharedlog.App = previousApp
	})

	const secret = "refresh-token-and-password"
	router := gin.New()
	router.Use(AuditLog())
	router.POST("/auth/logout", func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil || string(body) != secret {
			t.Fatalf("request body was not preserved: %v", err)
		}
		c.String(http.StatusOK, secret)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/auth/logout?token="+secret, bytes.NewBufferString(secret))
	request.Header.Set("Authorization", "Bearer "+secret)
	router.ServeHTTP(recorder, request)

	if recorder.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected request ID response header")
	}
	entries := observed.All()
	if len(entries) != 1 {
		t.Fatalf("expected one audit entry, got %d", len(entries))
	}
	if strings.Contains(fmt.Sprint(entries[0].ContextMap()), secret) || strings.Contains(entries[0].Message, secret) {
		t.Fatal("audit log contains secret data")
	}
	if _, exists := entries[0].ContextMap()["headers"]; exists {
		t.Fatal("audit log contains request headers")
	}
}
