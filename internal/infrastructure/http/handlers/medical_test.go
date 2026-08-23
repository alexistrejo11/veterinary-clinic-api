package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"clinic-vet-api/internal/core/medical"
	"clinic-vet-api/internal/middleware"
	"clinic-vet-api/internal/shared/page"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type medicalSessionQueryStub struct {
	medical.MedicalSessionService
	specification medical.MedicalSessionSpecification
}

func (s *medicalSessionQueryStub) GetSessionsBySpecification(_ context.Context, specification medical.MedicalSessionSpecification) (page.Page[medical.MedicalSession], error) {
	s.specification = specification
	return page.NewPage([]medical.MedicalSession{}, 0, page.PaginationRequest{Page: 1, PageSize: 10}), nil
}

func TestAssignedEmployeeIDsScopesVeterinariansButNotAdmins(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &MedicalHandler{
		employeeIDResolver: func(_ context.Context, userID uint) (uint, error) {
			return userID + 100, nil
		},
	}

	for _, testCase := range []struct {
		name string
		role string
		want []uint
	}{
		{name: "veterinarian", role: "veterinarian", want: []uint{107}},
		{name: "admin", role: "admin", want: nil},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Request = httptest.NewRequest("GET", "/", nil)
			ctx.Set("user", &middleware.UserContext{UserID: 7, Role: testCase.role})

			got, err := handler.assignedEmployeeIDs(ctx)
			if err != nil {
				t.Fatalf("assigned employee IDs: %v", err)
			}
			if len(got) != len(testCase.want) {
				t.Fatalf("employee IDs = %v, want %v", got, testCase.want)
			}
			for index := range got {
				if got[index] != testCase.want[index] {
					t.Fatalf("employee IDs = %v, want %v", got, testCase.want)
				}
			}
		})
	}
}

func TestGetSessionsBySpecificationScopesVeterinarianToAssignedEmployee(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &medicalSessionQueryStub{}
	handler := &MedicalHandler{
		sessionSvc: service,
		validator:  validator.New(),
		employeeIDResolver: func(_ context.Context, userID uint) (uint, error) {
			return userID + 100, nil
		},
	}

	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v2/medical/sessions?page=1&page_size=10", nil)
	ctx.Set("user", &middleware.UserContext{UserID: 7, Role: "veterinarian"})

	handler.GetSessionsBySpecification(ctx)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if len(service.specification.EmployeeIDs) != 1 || service.specification.EmployeeIDs[0] != 107 {
		t.Fatalf("employee scope = %v, want [107]", service.specification.EmployeeIDs)
	}
}
