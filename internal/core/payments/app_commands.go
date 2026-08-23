package payments

import (
	"clinic-vet-api/internal/core/appointments"
	"clinic-vet-api/internal/core/customers"
	"time"
)

type CancelPaymentCommand struct {
	ID     PaymentID
	Reason string
}

type ProcessPaymentCommand struct {
	ID     PaymentID
	Reason string
}

type RefundPaymentCommand struct {
	ID     PaymentID
	Reason string
}

type DeletePaymentCommand struct {
	ID PaymentID
}

type CreatePaymentCommand struct {
	Amount        float64
	Currency      string
	Status        PaymentStatus
	Method        PaymentMethod
	Reference     string
	Reason        string
	Description   string
	DueDate       time.Time
	CustomerID    customers.CustomerID
	AppointmentID appointments.AppointmentID
	InvoiceID     string
}

type UpdatePaymentCommand struct {
	ID            PaymentID
	Amount        float64
	Currency      string
	Status        PaymentStatus
	Method        PaymentMethod
	Reference     string
	Description   string
	DueDate       time.Time
	CustomerID    customers.CustomerID
	AppointmentID appointments.AppointmentID
	InvoiceID     string
}
