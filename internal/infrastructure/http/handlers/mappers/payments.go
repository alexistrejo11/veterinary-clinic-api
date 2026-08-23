package mappers

import (
	"clinic-vet-api/internal/core/appointments"
	"clinic-vet-api/internal/core/customers"
	"clinic-vet-api/internal/core/payments"
	"clinic-vet-api/internal/infrastructure/http/handlers/dtos"
	"clinic-vet-api/internal/shared"
	"clinic-vet-api/internal/shared/page"
	"time"
)

type PaymentMapper struct{}

func NewPaymentMapper() *PaymentMapper {
	return &PaymentMapper{}
}

// RequestToCreateCommand maps a HTTP create request into a domain create command.
func (m *PaymentMapper) RequestToCreateCommand(r dtos.PaymentCreateRequest) (payments.CreatePaymentCommand, error) {
	amountDecimal := shared.NewDecimalFromFloat(r.Amount)
	money := shared.NewMoney(amountDecimal, r.Currency)

	method, err := payments.ParsePaymentMethod(r.Method)
	if err != nil {
		return payments.CreatePaymentCommand{}, err
	}

	var description *string
	if r.Description != nil {
		description = r.Description
	}

	var invoiceID string
	if r.InvoiceID != nil {
		invoiceID = *r.InvoiceID
	}

	customerID := customers.NewCustomerID(r.CustomerID)

	var apptID appointments.AppointmentID
	if r.AppointmentID != nil {
		apptID = appointments.NewAppointmentID(*r.AppointmentID)
	}

	cmd := payments.CreatePaymentCommand{
		Amount:        money.Amount().Float64(),
		Currency:      money.Currency(),
		Status:        payments.PaymentStatusAuthorized,
		Method:        method,
		Reference:     r.Reference,
		Description:   "",
		CustomerID:    customerID,
		AppointmentID: apptID,
		InvoiceID:     invoiceID,
	}

	if description != nil {
		cmd.Description = *description
	}
	if r.Reason != nil {
		cmd.Reason = *r.Reason
	}
	if r.DueDate != nil {
		cmd.DueDate = *r.DueDate
	}

	return cmd, nil
}

// RequestToUpdateCommand maps a HTTP update request plus route ID into a domain update command.
func (m *PaymentMapper) RequestToUpdateCommand(id uint, r dtos.PaymentUpdateRequest, current payments.Payment) (payments.UpdatePaymentCommand, error) {
	paymentID := payments.NewPaymentID(id)

	cmd := payments.UpdatePaymentCommand{
		ID:          paymentID,
		Amount:      current.Amount.Amount().Float64(),
		Currency:    current.Amount.Currency(),
		Status:      current.Status,
		Method:      current.Method,
		Reference:   current.Reference,
		Description: "",
		DueDate:     time.Time{},
		CustomerID:  current.PaidFromCustomer,
		AppointmentID: func() appointments.AppointmentID {
			if current.AppointmentID != nil {
				return *current.AppointmentID
			}
			return appointments.AppointmentID{}
		}(),
		InvoiceID: func() string {
			if current.InvoiceID != nil {
				return *current.InvoiceID
			}
			return ""
		}(),
	}

	if r.Amount != nil {
		cmd.Amount = *r.Amount
	}
	if r.Currency != nil {
		cmd.Currency = *r.Currency
	}
	if r.Method != nil {
		method, err := payments.ParsePaymentMethod(*r.Method)
		if err != nil {
			return payments.UpdatePaymentCommand{}, err
		}
		cmd.Method = method
	}
	if r.Description != nil {
		cmd.Description = *r.Description
	}
	if r.DueDate != nil {
		cmd.DueDate = *r.DueDate
	}
	if r.InvoiceID != nil {
		cmd.InvoiceID = *r.InvoiceID
	}

	return cmd, nil
}

// RequestToSpecification converts a search request into a PaymentSpecification.
func (m *PaymentMapper) RequestToSpecification(r dtos.PaymentSearchRequest) (payments.PaymentSpecification, error) {
	spec := payments.PaymentSpecification{
		OverdueOnly: r.OverdueOnly,
		Pagination:  r.PaginationRequest.WithDefaults().ToPagination(),
	}

	if r.CustomerID != nil {
		customerID := customers.NewCustomerID(*r.CustomerID)
		spec.CustomerID = &customerID
	}

	if r.Status != nil {
		status, err := payments.ParsePaymentStatus(*r.Status)
		if err != nil {
			return payments.PaymentSpecification{}, err
		}
		spec.Status = &status
	}

	if r.Method != nil {
		method, err := payments.ParsePaymentMethod(*r.Method)
		if err != nil {
			return payments.PaymentSpecification{}, err
		}
		spec.Method = &method
	}

	if r.FromCreated != nil {
		spec.FromCreatedAt = r.FromCreated
	}
	if r.ToCreated != nil {
		spec.ToCreatedAt = r.ToCreated
	}

	return spec, nil
}

// ToPaymentResponse maps a Payment aggregate into a HTTP response DTO.
func (m *PaymentMapper) ToPaymentResponse(p payments.Payment) dtos.PaymentResponse {
	amount := p.Amount.Amount().Float64()

	var refundAmount *float64
	if p.RefundAmount != nil {
		val := p.RefundAmount.Amount().Float64()
		refundAmount = &val
	}

	customerID := uint(p.PaidFromCustomer.Int32())

	return dtos.PaymentResponse{
		ID: p.ID.Value(),

		Amount:   amount,
		Currency: p.Amount.Currency(),

		Status:        p.Status.String(),
		StatusDisplay: p.Status.DisplayName(),

		Method:        p.Method.String(),
		MethodDisplay: p.Method.DisplayName(),

		Reference:     p.Reference,
		ReceiptNumber: p.ReceiptNumber,
		Reason:        p.Reason,
		Description:   p.Description,
		DueDate:       p.DueDate,
		PaidAt:        p.PaidAt,
		RefundedAt:    p.RefundedAt,
		IsActive:      p.IsActive,

		CustomerID:    customerID,
		InvoiceID:     p.InvoiceID,
		FailureReason: p.FailureReason,

		RefundAmount: refundAmount,

		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
		DeletedAt: p.DeletedAt,
	}
}

// ToPaginatedResponse maps a paginated domain response into a paginated DTO response.
func (m *PaymentMapper) ToPaginatedResponse(p page.Page[payments.Payment]) page.Page[dtos.PaymentResponse] {
	return page.MapItems(p, m.ToPaymentResponse)
}
