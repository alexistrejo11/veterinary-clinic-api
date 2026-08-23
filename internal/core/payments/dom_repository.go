package payments

import (
	"clinic-vet-api/internal/core/customers"
	"clinic-vet-api/internal/shared"
	"clinic-vet-api/internal/shared/page"
	"context"
	"time"
)

type PaymentRepository interface {
	GetByID(ctx context.Context, id PaymentID) (Payment, error)
	GetByReference(ctx context.Context, reference string) (Payment, error)
	GetByIDAndCustomerID(ctx context.Context, id PaymentID, customerID customers.CustomerID) (Payment, error)

	GetBySpecification(ctx context.Context, spec PaymentSpecification) (page.Page[Payment], error)

	/* Spec should include this methods
	FindByCustomerID(ctx context.Context, customerID customers.CustomerID, pageInput page.Pagination) (page.Page[Payment], error)
	FindByStatus(ctx context.Context, status PaymentStatus, pageInput page.Pagination) (page.Page[Payment], error)
	FindOverdue(ctx context.Context, pageInput page.Pagination) (page.Page[Payment], error)
	FindByDateRange(ctx context.Context, startDate, endDate time.Time, pageInput page.Pagination) (page.Page[Payment], error)

	FindRecentByCustomerID(ctx context.Context, customerID customers.CustomerID, limit int) ([]Payment, error)
	FindPendingPayments(ctx context.Context, pageInput page.Pagination) (page.Page[Payment], error)
	FindSuccessfulPayments(ctx context.Context, pageInput page.Pagination) (page.Page[Payment], error)
	*/
	ExistsByID(ctx context.Context, id PaymentID) (bool, error)
	ExistsByReference(ctx context.Context, reference string) (bool, error)
	ExistsPendingByCustomerID(ctx context.Context, customerID customers.CustomerID) (bool, error)

	Save(ctx context.Context, paymentEntity *Payment) error
	SettleAuthorized(ctx context.Context, id PaymentID, receiptNumber, reason string) (Payment, error)
	CancelAuthorized(ctx context.Context, id PaymentID, reason string) (Payment, error)
	RefundSettled(ctx context.Context, id PaymentID, reason string) (Payment, error)
	SoftDelete(ctx context.Context, id PaymentID) error
	HardDelete(ctx context.Context, id PaymentID) error

	CountByStatus(ctx context.Context, status PaymentStatus) (int64, error)
	CountByCustomerID(ctx context.Context, customerID customers.CustomerID) (int64, error)
	CountOverdue(ctx context.Context) (int64, error)
	TotalRevenueByDateRange(ctx context.Context, startDate, endDate time.Time) (float64, error)
}

type PaymentSummary struct {
	Count  int          `json:"count"`
	Amount shared.Money `json:"amount"`
}
