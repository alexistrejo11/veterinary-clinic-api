package payments

import (
	"clinic-vet-api/internal/core/customers"
	"clinic-vet-api/internal/shared/page"
	"time"
)

// PaymentSpecification defines filters and pagination for querying payments.
// All fields are optional; zero values mean “no filter” for that criterion.
type PaymentSpecification struct {
	CustomerID    *customers.CustomerID
	Status        *PaymentStatus
	Method        *PaymentMethod
	FromCreatedAt *time.Time
	ToCreatedAt   *time.Time
	OverdueOnly   bool
	page.Pagination
}
