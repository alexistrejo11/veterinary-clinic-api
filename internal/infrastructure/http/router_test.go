package http

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"clinic-vet-api/internal/infrastructure/http/handlers"
	"clinic-vet-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func TestRegisterRoutesIncludesPhaseFourGroups(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	router := &APIRouter{
		config: &APIConfig{Router: engine, AuthMiddleware: middleware.NewAuthMiddleware("test")},
		appHandlers: &AppHandlers{
			address:     &handlers.AddressHandler{},
			auth:        &handlers.AuthHandler{},
			user:        &handlers.UserHandler{},
			pets:        &handlers.PetHandler{},
			profile:     &handlers.ProfileHandler{},
			appointment: &handlers.AppointmentHandler{},
			customer:    &handlers.CustomerHandler{},
			employee:    &handlers.EmployeeHandler{},
			payment:     &handlers.PaymentHandler{},
			home:        &handlers.HomeHandler{},
		},
	}
	router.RegisterRoutes()

	want := map[string]bool{
		"GET /health":                              false,
		"GET /api/v2/addresses/":                   false,
		"GET /api/v2/pets/":                        false,
		"GET /api/v2/payments/":                    false,
		"POST /api/v2/me/addresses/":               false,
		"POST /api/v2/me/pets/":                    false,
		"POST /api/v2/me/appointments/:id/cancel":  false,
		"POST /api/v2/payments/:id/process":        false,
		"POST /api/v2/payments/:id/refund":         false,
		"POST /api/v2/payments/:id/cancel":         false,
		"POST /api/v2/pets/:id/restore":            false,
		"PUT /api/v2/addresses/users/:user_id/:id": false,
	}
	for _, route := range engine.Routes() {
		key := route.Method + " " + route.Path
		if _, ok := want[key]; ok {
			want[key] = true
		}
	}
	for route, registered := range want {
		if !registered {
			t.Errorf("missing route %s", route)
		}
	}
}

func TestVeterinarianAppointmentRoutesDoNotExposeCreationOrConfirmation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	router := &APIRouter{
		config: &APIConfig{Router: engine, AuthMiddleware: middleware.NewAuthMiddleware("test")},
		appHandlers: &AppHandlers{
			address: &handlers.AddressHandler{}, auth: &handlers.AuthHandler{}, user: &handlers.UserHandler{}, pets: &handlers.PetHandler{},
			profile: &handlers.ProfileHandler{}, appointment: &handlers.AppointmentHandler{}, customer: &handlers.CustomerHandler{},
			employee: &handlers.EmployeeHandler{}, payment: &handlers.PaymentHandler{}, home: &handlers.HomeHandler{},
		},
	}
	router.RegisterRoutes()

	for _, route := range engine.Routes() {
		if route.Method == "POST" && (route.Path == "/api/v2/employees/appointments/" || route.Path == "/api/v2/employees/appointments/:id/confirm") {
			t.Fatalf("veterinarian-only route must not expose %s", route.Path)
		}
	}
}

func TestSwaggerOperationsAreRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	router := &APIRouter{
		config: &APIConfig{Router: engine, AuthMiddleware: middleware.NewAuthMiddleware("test")},
		appHandlers: &AppHandlers{
			address: &handlers.AddressHandler{}, auth: &handlers.AuthHandler{}, user: &handlers.UserHandler{}, pets: &handlers.PetHandler{},
			profile: &handlers.ProfileHandler{}, appointment: &handlers.AppointmentHandler{}, customer: &handlers.CustomerHandler{},
			employee: &handlers.EmployeeHandler{}, payment: &handlers.PaymentHandler{}, medical: &handlers.MedicalHandler{},
			notification: &handlers.NotificationHandler{}, home: &handlers.HomeHandler{},
		},
	}
	router.RegisterRoutes()

	registered := make(map[string]struct{}, len(engine.Routes()))
	for _, route := range engine.Routes() {
		registered[route.Method+" "+strings.TrimSuffix(route.Path, "/")] = struct{}{}
	}

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate router test file")
	}
	document, err := os.ReadFile(filepath.Join(filepath.Dir(thisFile), "../../../docs/swagger.json"))
	if err != nil {
		t.Fatalf("read Swagger document: %v", err)
	}
	var swagger struct {
		BasePath string                     `json:"basePath"`
		Paths    map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(document, &swagger); err != nil {
		t.Fatalf("parse Swagger document: %v", err)
	}
	for path, operations := range swagger.Paths {
		var methods map[string]json.RawMessage
		if err := json.Unmarshal(operations, &methods); err != nil {
			t.Fatalf("parse Swagger operations for %s: %v", path, err)
		}
		for method := range methods {
			key := strings.ToUpper(method) + " " + swaggerPathToGinPath(swagger.BasePath+path)
			if _, ok := registered[key]; !ok {
				t.Errorf("Swagger documents unregistered operation %s", key)
			}
		}
	}
}

func swaggerPathToGinPath(path string) string {
	segments := strings.Split(strings.TrimSuffix(path, "/"), "/")
	for index, segment := range segments {
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			segments[index] = ":" + strings.TrimSuffix(strings.TrimPrefix(segment, "{"), "}")
		}
	}
	return strings.Join(segments, "/")
}
