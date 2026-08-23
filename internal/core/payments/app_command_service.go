package payments

import (
	"clinic-vet-api/internal/core/appointments"
	"clinic-vet-api/internal/shared"
	"context"
	"fmt"
	"time"
)

type CommandService interface {
	CancelPayment(ctx context.Context, cmd CancelPaymentCommand) error
	ProcessPayment(ctx context.Context, cmd ProcessPaymentCommand) error
	RefundPayment(ctx context.Context, cmd RefundPaymentCommand) error
	DeletePayment(ctx context.Context, cmd DeletePaymentCommand) error
	CreatePayment(ctx context.Context, cmd CreatePaymentCommand) (Payment, error)
	UpdatePayment(ctx context.Context, cmd UpdatePaymentCommand) error
}

type commandService struct {
	repository   PaymentRepository
	appointments appointments.AppointmentRepository
}

func NewCommandService(repository PaymentRepository, appointmentRepository appointments.AppointmentRepository) CommandService {
	return &commandService{repository: repository, appointments: appointmentRepository}
}

func (s *commandService) CancelPayment(ctx context.Context, cmd CancelPaymentCommand) error {
	if cmd.Reason == "" {
		return fmt.Errorf("cancellation reason is required")
	}
	payment, err := s.repository.GetByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	if payment.IsCancelled() {
		return nil
	}
	_, err = s.repository.CancelAuthorized(ctx, cmd.ID, cmd.Reason)
	if err != nil {
		if current, getErr := s.repository.GetByID(ctx, cmd.ID); getErr == nil && current.IsCancelled() {
			return nil
		}
	}
	return err
}

func (s *commandService) ProcessPayment(ctx context.Context, cmd ProcessPaymentCommand) error {
	payment, err := s.repository.GetByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	if payment.IsPaid() {
		return nil
	}
	receiptNumber := fmt.Sprintf("RCT-%d-%d", payment.ID.Value(), time.Now().UnixNano())
	_, err = s.repository.SettleAuthorized(ctx, cmd.ID, receiptNumber, cmd.Reason)
	if err != nil {
		if current, getErr := s.repository.GetByID(ctx, cmd.ID); getErr == nil && current.IsPaid() {
			return nil
		}
	}
	return err
}

func (s *commandService) RefundPayment(ctx context.Context, cmd RefundPaymentCommand) error {
	if cmd.Reason == "" {
		return fmt.Errorf("refund reason is required")
	}
	payment, err := s.repository.GetByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	if payment.IsRefunded() {
		return nil
	}
	_, err = s.repository.RefundSettled(ctx, cmd.ID, cmd.Reason)
	if err != nil {
		if current, getErr := s.repository.GetByID(ctx, cmd.ID); getErr == nil && current.IsRefunded() {
			return nil
		}
	}
	return err
}

func (s *commandService) DeletePayment(ctx context.Context, cmd DeletePaymentCommand) error {
	// Use soft delete for safety; this keeps a history of payments.
	if err := s.repository.SoftDelete(ctx, cmd.ID); err != nil {
		return err
	}
	return nil
}

func (s *commandService) CreatePayment(ctx context.Context, cmd CreatePaymentCommand) (Payment, error) {
	operation := "CreatePayment"

	if cmd.Reference == "" {
		return Payment{}, fmt.Errorf("payment reference is required")
	}
	if existing, err := s.repository.GetByReference(ctx, cmd.Reference); err == nil {
		return existing, nil
	}

	if cmd.Amount <= 0 {
		money := shared.NewMoney(shared.NewDecimalFromFloat(cmd.Amount), cmd.Currency)
		return Payment{}, AmountInvalidError(ctx, money, operation)
	}

	amountDecimal := shared.NewDecimalFromFloat(cmd.Amount)
	amount := shared.NewMoney(amountDecimal, cmd.Currency)

	if !cmd.Method.IsValid() {
		return Payment{}, MethodInvalidError(ctx, cmd.Method, operation)
	}

	if cmd.CustomerID.IsZero() {
		return Payment{}, CustomerIDRequiredError(ctx, operation)
	}
	if !cmd.AppointmentID.IsZero() {
		if s.appointments == nil {
			return Payment{}, fmt.Errorf("appointment repository is required")
		}
		if _, err := s.appointments.FindByIDAndCustomerID(ctx, cmd.AppointmentID, cmd.CustomerID); err != nil {
			return Payment{}, err
		}
	}

	if !cmd.DueDate.IsZero() && cmd.DueDate.Before(time.Now()) {
		return Payment{}, DueDatePastError(ctx, operation)
	}

	entity := Payment{
		Amount:           amount,
		Status:           PaymentStatusAuthorized,
		Method:           cmd.Method,
		Reference:        cmd.Reference,
		Reason:           nil,
		Description:      nil,
		DueDate:          nil,
		PaidAt:           nil,
		RefundedAt:       nil,
		IsActive:         true,
		PaidFromCustomer: cmd.CustomerID,
		InvoiceID:        nil,
	}

	if cmd.Reason != "" {
		entity.Reason = &cmd.Reason
	}
	if cmd.Description != "" {
		entity.Description = &cmd.Description
	}
	if !cmd.DueDate.IsZero() {
		due := cmd.DueDate
		entity.DueDate = &due
	}
	if cmd.InvoiceID != "" {
		inv := cmd.InvoiceID
		entity.InvoiceID = &inv
	}
	if !cmd.AppointmentID.IsZero() {
		appt := cmd.AppointmentID
		entity.AppointmentID = &appt
	}

	if err := s.repository.Save(ctx, &entity); err != nil {
		return Payment{}, err
	}

	return entity, nil
}

func (s *commandService) UpdatePayment(ctx context.Context, cmd UpdatePaymentCommand) error {
	operation := "UpdatePayment"

	payment, err := s.repository.GetByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	// Authorized ledger entries are editable until they are settled or cancelled.
	if !payment.RequiresSettlement() {
		return AlreadyProcessedError(ctx, payment.Status, operation)
	}

	customerID := payment.PaidFromCustomer
	if cmd.CustomerID.IntegerID != nil && !cmd.CustomerID.IsZero() {
		customerID = cmd.CustomerID
	}
	appointmentID := payment.AppointmentID
	if cmd.AppointmentID.IntegerID != nil && !cmd.AppointmentID.IsZero() {
		appointmentID = &cmd.AppointmentID
	}
	if appointmentID != nil {
		if s.appointments == nil {
			return fmt.Errorf("appointment repository is required")
		}
		if _, err := s.appointments.FindByIDAndCustomerID(ctx, *appointmentID, customerID); err != nil {
			return err
		}
	}

	var amount *shared.Money
	if cmd.Amount > 0 {
		dec := shared.NewDecimalFromFloat(cmd.Amount)
		m := shared.NewMoney(dec, cmd.Currency)
		amount = &m
	}

	var method *PaymentMethod
	if cmd.Method != "" {
		m := cmd.Method
		method = &m
	}

	var description *string
	if cmd.Description != "" {
		desc := cmd.Description
		description = &desc
	}

	var dueDate *time.Time
	if !cmd.DueDate.IsZero() {
		d := cmd.DueDate
		dueDate = &d
	}

	if err := payment.Update(ctx, amount, method, description, dueDate); err != nil {
		return err
	}

	// Update relations if provided.
	payment.PaidFromCustomer = customerID
	payment.AppointmentID = appointmentID
	if cmd.InvoiceID != "" {
		inv := cmd.InvoiceID
		payment.InvoiceID = &inv
	}

	return s.repository.Save(ctx, &payment)
}
