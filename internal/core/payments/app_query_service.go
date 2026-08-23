package payments

import (
	"clinic-vet-api/internal/core/customers"
	"clinic-vet-api/internal/shared/page"
	"context"
)

type QueryService interface {
	GetPaymentsBySpecification(ctx context.Context, specification PaymentSpecification) (page.Page[Payment], error)
	GetPaymentByReference(ctx context.Context, reference string) (Payment, error)
	GetPaymentByID(ctx context.Context, id PaymentID) (Payment, error)
	GetPaymentsByCustomerID(ctx context.Context, customerID customers.CustomerID) (page.Page[Payment], error)
}

type queryService struct {
	repository PaymentRepository
}

func NewQueryService(repository PaymentRepository) QueryService {
	return &queryService{repository: repository}
}

func (s *queryService) GetPaymentsBySpecification(ctx context.Context, specification PaymentSpecification) (page.Page[Payment], error) {
	return s.repository.GetBySpecification(ctx, specification)
}

func (s *queryService) GetPaymentByReference(ctx context.Context, reference string) (Payment, error) {
	return s.repository.GetByReference(ctx, reference)
}

func (s *queryService) GetPaymentByID(ctx context.Context, id PaymentID) (Payment, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *queryService) GetPaymentsByCustomerID(ctx context.Context, customerID customers.CustomerID) (page.Page[Payment], error) {
	spec := PaymentSpecification{
		CustomerID: &customerID,
		Pagination: page.AllItems().ToPagination(),
	}
	return s.repository.GetBySpecification(ctx, spec)
}
