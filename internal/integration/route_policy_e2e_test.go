//go:build integration

package integration

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"clinic-vet-api/internal/testkit"
)

var routeParameter = regexp.MustCompile(`:[^/]+`)

// TestRegisteredRoutesRejectUnauthenticatedRequests exercises every protected
// operation through the real router before its handler can process input.
func TestRegisteredRoutesRejectUnauthenticatedRequests(t *testing.T) {
	env := testkit.NewEnvironment(t)
	public := map[string]struct{}{
		"GET /health":                              {},
		"POST /api/v2/auth/register":               {},
		"POST /api/v2/auth/login":                  {},
		"POST /api/v2/auth/activate":               {},
		"POST /api/v2/auth/verify-email":           {},
		"POST /api/v2/auth/refresh":                {},
		"POST /api/v2/auth/2fa/verify":             {},
		"POST /api/v2/auth/reset-password/request": {},
		"POST /api/v2/auth/reset-password/reset":   {},
	}

	for _, route := range env.Router.Routes() {
		key := route.Method + " " + route.Path
		if _, ok := public[key]; ok {
			continue
		}
		t.Run(key, func(t *testing.T) {
			path := routeParameter.ReplaceAllString(route.Path, "999")
			req := httptest.NewRequest(route.Method, path, nil)
			response := httptest.NewRecorder()
			env.Router.ServeHTTP(response, req)
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("unauthenticated %s status = %d, want %d: %s", key, response.Code, http.StatusUnauthorized, response.Body.String())
			}
		})
	}
}
