package handlers

import (
	"testing"

	"clinic-vet-api/internal/infrastructure/http/handlers/dtos"
	"clinic-vet-api/internal/infrastructure/http/handlers/mappers"
)

func TestNewCustomerHandlerInitializesMapper(t *testing.T) {
	handler := NewCustomerHandler(nil, nil)
	if handler.mapper == nil {
		t.Fatal("customer handler mapper is nil")
	}
}

func TestCustomerSearchSpecificationUsesQueryFiltersAndDefaults(t *testing.T) {
	active := true
	spec := mappers.NewCustomerMapper().RequestToSpecification(dtos.CustomerSearchRequest{
		Search:   "Ada",
		IsActive: &active,
	})
	if spec.SearchTerm == nil || *spec.SearchTerm != "Ada" {
		t.Fatalf("unexpected search term: %#v", spec.SearchTerm)
	}
	if spec.IsActive == nil || !*spec.IsActive {
		t.Fatalf("unexpected active filter: %#v", spec.IsActive)
	}
	if spec.Pagination.Number != 1 || spec.Pagination.Size != 10 {
		t.Fatalf("unexpected pagination: %#v", spec.Pagination)
	}
}
