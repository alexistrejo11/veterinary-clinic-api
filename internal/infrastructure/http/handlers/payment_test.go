package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"clinic-vet-api/internal/core/payments"
	"clinic-vet-api/internal/infrastructure/http/handlers/mappers"
	"clinic-vet-api/internal/shared/page"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type paymentQueryStub struct {
	payments.QueryService
	specification payments.PaymentSpecification
}

func (s *paymentQueryStub) GetPaymentsBySpecification(_ context.Context, specification payments.PaymentSpecification) (page.Page[payments.Payment], error) {
	s.specification = specification
	return page.NewPage([]payments.Payment{}, 0, page.PaginationRequest{Page: 2, PageSize: 20}), nil
}

func TestGetPaymentsBySpecificationBindsQueryParameters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	query := &paymentQueryStub{}
	handler := NewPaymentHandler(query, nil, validator.New(), mappers.NewPaymentMapper(), nil)

	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v2/payments?page=2&page_size=20&customer_id=5&status=authorized", nil)

	handler.GetPaymentsBySpecification(ctx)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if query.specification.CustomerID == nil || query.specification.CustomerID.Value() != 5 {
		t.Fatalf("customer filter = %#v, want customer 5", query.specification.CustomerID)
	}
	if query.specification.Status == nil || *query.specification.Status != payments.PaymentStatusAuthorized {
		t.Fatalf("status filter = %#v, want authorized", query.specification.Status)
	}
	if query.specification.Pagination.Number != 2 || query.specification.Pagination.Size != 20 {
		t.Fatalf("pagination = %#v, want page 2 size 20", query.specification.Pagination)
	}
}
