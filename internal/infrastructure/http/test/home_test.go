package test

import (
	"clinic-vet-api/internal/infrastructure/http/handlers"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	home := handlers.NewHomeHandler(router)
	router.GET("/health", home.HealthCheck)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not JSON valid: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf(`expected status "ok", got %v`, body["status"])
	}
}

func TestPublicHomeEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	home := handlers.NewHomeHandler(router)
	router.GET("/", home.Home)
	router.GET("/home", home.Home)

	for _, path := range []string{"/", "/home"} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected status 200 for %s, got %d", path, rec.Code)
			}
			var body map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("response is not JSON valid: %v", err)
			}
			if body["status"] != "ok" {
				t.Errorf(`expected status "ok", got %v`, body["status"])
			}
		})
	}
}
