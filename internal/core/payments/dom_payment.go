package payments

import (
	"clinic-vet-api/internal/core/appointments"
	"clinic-vet-api/internal/core/customers"
	"clinic-vet-api/internal/core/employees"
	"clinic-vet-api/internal/shared"
	"clinic-vet-api/internal/shared/log"
	"context"
	"slices"
	"time"

	"go.uber.org/zap"
)

type Payment struct {
	shared.Entity[PaymentID]
	Amount           shared.Money
	Status           PaymentStatus
	Method           PaymentMethod
	Reference        string
	ReceiptNumber    *string
	Reason           *string
	Description      *string
	DueDate          *time.Time
	PaidAt           *time.Time
	RefundedAt       *time.Time
	IsActive         bool
	PaidFromCustomer customers.CustomerID
	PaidToEmployee   employees.EmployeeID
	AppointmentID    *appointments.AppointmentID
	InvoiceID        *string
	RefundAmount     *shared.Money
	FailureReason    *string
	DeletedAt        *time.Time
}

type PaymentReport struct {
	StartDate        time.Time             `json:"start_date"`
	EndDate          time.Time             `json:"end_date"`
	TotalPayments    int                   `json:"total_payments"`
	TotalAmount      float64               `json:"total_amount"`
	TotalCurrency    string                `json:"total_currency"`
	PaidAmount       float64               `json:"paid_amount"`
	PendingAmount    float64               `json:"pending_amount"`
	RefundedAmount   float64               `json:"refunded_amount"`
	OverdueAmount    float64               `json:"overdue_amount"`
	PaymentsByMethod map[PaymentMethod]any `json:"payments_by_method"`
	PaymentsByStatus map[PaymentStatus]any `json:"payments_by_status"`
}

type PaymentID struct {
	shared.IntegerID
}

func NewPaymentID(id uint) PaymentID {
	return PaymentID{shared.NewBaseID(id)}
}

func (p *Payment) MarkAsPaid(ctx context.Context, transactionID string, paidAt time.Time) error {
	operation := "MarkAsPaid"

	if p.Status != PaymentStatusPending {
		return CannotMarkAsPaidError(ctx, p.Status, operation)
	}

	if transactionID == "" {
		return ReferenceRequiredError(ctx, operation)
	}

	if paidAt.After(time.Now()) {
		return PaidDateFutureError(ctx, operation)
	}

	p.Status = PaymentStatusPaid
	p.Reference = transactionID
	p.PaidAt = &paidAt
	p.IncrementVersion()

	return nil
}

func (p *Payment) RequestRefund(ctx context.Context, refundAmount shared.Money, reason string) error {
	operation := "RequestRefund"

	if !p.IsRefundable() {
		return CannotRequestRefundError(ctx, p.Status, operation)
	}

	if p.isRefundPeriodExpired() {
		return RefundPeriodExpiredError(ctx, *p.PaidAt, operation)
	}

	diff, err := refundAmount.Subtract(p.Amount)
	if err != nil {
		return err
	}
	if diff.IsPositive() {
		return RefundAmountExceededError(ctx, refundAmount, p.Amount, operation)
	}

	if reason == "" {
		return FailureReasonEmptyError(ctx, operation)
	}

	p.Status = PaymentStatusRefunded
	p.RefundAmount = &refundAmount
	now := time.Now()
	p.RefundedAt = &now
	p.Reason = &reason
	p.UpdatedAt = now
	p.IncrementVersion()

	return nil
}

func (p *Payment) SoftDelete(ctx context.Context) {
	operation := "SoftDelete"

	log.App.Info("Soft deleting payment",
		zap.String("payment_id", p.ID.String()),
		zap.String("operation", operation))

	p.IsActive = false
	deletedAt := time.Now()
	p.DeletedAt = &deletedAt
	p.IncrementVersion()
}

func (p *Payment) Restore(ctx context.Context) {
	operation := "Restore"

	log.App.Info("Restoring payment",
		zap.String("payment_id", p.ID.String()),
		zap.String("operation", operation))

	p.IsActive = true
	p.DeletedAt = nil
	p.IncrementVersion()
}

func (p *Payment) isRefundPeriodExpired() bool {
	if p.PaidAt == nil {
		return true
	}

	// Refund period: 30 days from payment date
	refundDeadline := p.PaidAt.Add(30 * 24 * time.Hour)
	return time.Now().After(refundDeadline)
}

func (p *Payment) Cancel(ctx context.Context, reason string) error {
	operation := "CancelPayment"

	allowedStatuses := []PaymentStatus{
		PaymentStatusAuthorized,
		PaymentStatusPending,
		PaymentStatusPaid,
		PaymentStatusFailed,
	}

	if !p.canTransitionStatus(allowedStatuses) {
		return CannotCancelError(ctx, p.Status, operation)
	}

	if reason == "" {
		return CancellationReasonRequiredError(ctx, operation)
	}

	p.Status = PaymentStatusCancelled
	p.Reason = &reason
	p.IncrementVersion()
	return nil
}

func (p *Payment) Pay(ctx context.Context, transactionID string) error {
	operation := "PayPayment"

	allowedStatuses := []PaymentStatus{
		PaymentStatusAuthorized,
		PaymentStatusFailed,
		PaymentStatusPending,
	}

	if !p.canTransitionStatus(allowedStatuses) {
		return CannotProcessError(ctx, p.Status, operation)
	}

	if transactionID == "" {
		return ReferenceRequiredError(ctx, operation)
	}

	now := time.Now()
	p.Status = PaymentStatusPaid
	p.Reference = transactionID
	p.PaidAt = &now
	p.IncrementVersion()
	p.UpdatedAt = now

	return nil
}

func (p *Payment) Refund(ctx context.Context) error {
	operation := "RefundPayment"

	allowedStatuses := []PaymentStatus{PaymentStatusPaid}

	if !p.canTransitionStatus(allowedStatuses) {
		return CannotRefundError(ctx, p.Status, operation)
	}

	if p.isRefundPeriodExpired() {
		return RefundPeriodExpiredError(ctx, *p.PaidAt, operation)
	}

	now := time.Now()
	p.Status = PaymentStatusRefunded
	p.RefundedAt = &now
	p.IncrementVersion()
	p.UpdatedAt = now

	return nil
}

func (p *Payment) MarkAsOverdue(ctx context.Context) error {
	operation := "MarkAsOverdue"

	allowedStatuses := []PaymentStatus{
		PaymentStatusFailed,
		PaymentStatusPending,
	}

	if !p.canTransitionStatus(allowedStatuses) {
		return CannotMarkOverdueError(ctx, p.Status, operation)
	}

	if p.DueDate != nil && time.Now().Before(*p.DueDate) {
		return DueDateNotReachedError(ctx, *p.DueDate, operation)
	}

	p.Status = PaymentStatusOverdue
	p.IncrementVersion()
	return nil
}

func (p *Payment) MarkAsFailed(ctx context.Context) error {
	operation := "MarkAsFailed"

	allowedStatuses := []PaymentStatus{PaymentStatusPending}

	if !p.canTransitionStatus(allowedStatuses) {
		return CannotMarkFailedError(ctx, p.Status, operation)
	}

	p.Status = PaymentStatusFailed
	p.IncrementVersion()
	return nil
}

func (p *Payment) Update(ctx context.Context, amount *shared.Money, paymentMethod *PaymentMethod, description *string, dueDate *time.Time) error {
	operation := "UpdatePayment"

	if amount != nil {
		if amount.Amount().IsZero() || amount.Amount().IsNegative() {
			return AmountInvalidError(ctx, *amount, operation)
		}
		p.Amount = *amount
	}

	if paymentMethod != nil {
		if !paymentMethod.IsValid() {
			return MethodInvalidError(ctx, *paymentMethod, operation)
		}
		p.Method = *paymentMethod
	}

	if description != nil {
		if len(*description) > 500 {
			return DescriptionTooLongError(ctx, len(*description), operation)
		}
		p.Description = description
	}

	if dueDate != nil {
		if dueDate.Before(time.Now()) {
			return DueDatePastError(ctx, operation)
		}
		p.DueDate = dueDate
	}

	p.IncrementVersion()
	return nil
}

// Métodos auxiliares (se mantienen igual pero con mejor documentación)
func (p *Payment) CanRetry() bool {
	return p.Status == PaymentStatusFailed || p.Status == PaymentStatusPending
}

func (p *Payment) AmountDue() shared.Money {
	if p.IsPaid() || p.IsRefunded() || p.IsCancelled() {
		return shared.NewMoney(p.Amount.Amount(), p.Amount.Currency())
	}
	return p.Amount
}

func (p *Payment) canTransitionStatus(allowedStatuses []PaymentStatus) bool {
	return slices.Contains(allowedStatuses, p.Status)
}

// RequiresSettlement indicates that an authorized local ledger entry awaits staff recording.
func (p *Payment) RequiresSettlement() bool { return p.Status == PaymentStatusAuthorized }

func (p *Payment) CanBeRefunded() bool {
	return p.Status == PaymentStatusPaid &&
		p.PaidAt != nil &&
		!p.isRefundPeriodExpired()
}

func (p *Payment) DaysOverdue() int {
	if !p.IsOverdue() || p.DueDate == nil {
		return 0
	}
	return int(time.Since(*p.DueDate).Hours() / 24)
}

func (p *Payment) IsPaid() bool {
	return p.Status == PaymentStatusPaid
}

func (p *Payment) IsPending() bool {
	return p.Status == PaymentStatusPending
}

func (p *Payment) IsFailed() bool {
	return p.Status == PaymentStatusFailed
}

func (p *Payment) IsRefunded() bool {
	return p.Status == PaymentStatusRefunded
}

func (p *Payment) IsCancelled() bool {
	return p.Status == PaymentStatusCancelled
}

func (p *Payment) IsOverdue() bool {
	return p.Status == PaymentStatusOverdue
}

func (p *Payment) IsDisputed() bool {
	return p.Status == PaymentStatusDisputed
}

func (p *Payment) IsRefundable() bool {
	return p.Status == PaymentStatusPaid &&
		p.PaidAt != nil &&
		!p.isRefundPeriodExpired()
}
