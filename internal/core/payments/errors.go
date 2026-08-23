package payments

import (
	"clinic-vet-api/internal/shared"
	"clinic-vet-api/internal/shared/errors"
	"context"
	"fmt"
	"time"
)

type PaymentErrorCode string

const (
	PaymentAmountInvalid         PaymentErrorCode = "PAYMENT_AMOUNT_INVALID"
	PaymentMethodInvalid         PaymentErrorCode = "PAYMENT_METHOD_INVALID"
	PaymentStatusInvalid         PaymentErrorCode = "PAYMENT_STATUS_INVALID"
	PaymentReferenceEmpty        PaymentErrorCode = "PAYMENT_REFERENCE_EMPTY"
	PaymentDescriptionTooLong    PaymentErrorCode = "PAYMENT_DESCRIPTION_TOO_LONG"
	PaymentDueDatePast           PaymentErrorCode = "PAYMENT_DUE_DATE_PAST"
	PaymentPaidDateFuture        PaymentErrorCode = "PAYMENT_PAID_DATE_FUTURE"
	PaymentRefundDateFuture      PaymentErrorCode = "PAYMENT_REFUND_DATE_FUTURE"
	PaymentCustomerIDRequired    PaymentErrorCode = "PAYMENT_CUSTOMER_ID_REQUIRED"
	PaymentAppointmentIDRequired PaymentErrorCode = "PAYMENT_APPOINTMENT_ID_REQUIRED"
	PaymentInvoiceIDEmpty        PaymentErrorCode = "PAYMENT_INVOICE_ID_EMPTY"
	PaymentRefundAmountNegative  PaymentErrorCode = "PAYMENT_REFUND_AMOUNT_NEGATIVE"
	PaymentFailureReasonEmpty    PaymentErrorCode = "PAYMENT_FAILURE_REASON_EMPTY"
	PaymentCannotProcess         PaymentErrorCode = "PAYMENT_CANNOT_PROCESS"
	PaymentCannotRefund          PaymentErrorCode = "PAYMENT_CANNOT_REFUND"
	PaymentCannotCancel          PaymentErrorCode = "PAYMENT_CANNOT_CANCEL"
	PaymentAlreadyProcessed      PaymentErrorCode = "PAYMENT_ALREADY_PROCESSED"
)

func paymentValidationError(ctx context.Context, code PaymentErrorCode, field, message, operation string) error {
	return errors.ValidationError(ctx, "payment", field,
		fmt.Sprintf("Payment %s: %s", field, message), operation)
}

func paymentBusinessError(ctx context.Context, code PaymentErrorCode, rule, operation string) error {
	return errors.BusinessRuleError(ctx, rule, "payment", "", operation)
}

func AmountInvalidError(ctx context.Context, amount shared.Money, operation string) error {
	return paymentValidationError(ctx, PaymentAmountInvalid, "amount",
		fmt.Sprintf("amount must be positive, got: %s", amount.Currency()), operation)
}

func MethodInvalidError(ctx context.Context, method PaymentMethod, operation string) error {
	return paymentValidationError(ctx, PaymentMethodInvalid, "method",
		fmt.Sprintf("invalid payment method: %s", method.String()), operation)
}

func StatusInvalidError(ctx context.Context, status PaymentStatus, operation string) error {
	return paymentValidationError(ctx, PaymentStatusInvalid, "status",
		fmt.Sprintf("invalid payment status: %s", status.String()), operation)
}

func ReferenceEmptyError(ctx context.Context, operation string) error {
	return paymentValidationError(ctx, PaymentReferenceEmpty, "reference",
		"payment reference cannot be empty", operation)
}

func DescriptionTooLongError(ctx context.Context, length int, operation string) error {
	return paymentValidationError(ctx, PaymentDescriptionTooLong, "description",
		fmt.Sprintf("description too long (%d characters, max 500)", length), operation)
}

func DueDatePastError(ctx context.Context, operation string) error {
	return paymentValidationError(ctx, PaymentDueDatePast, "due_date",
		"due date cannot be in the past", operation)
}

func PaidDateFutureError(ctx context.Context, operation string) error {
	return paymentValidationError(ctx, PaymentPaidDateFuture, "paid_at",
		"paid date cannot be in the future", operation)
}

func RefundDateFutureError(ctx context.Context, operation string) error {
	return paymentValidationError(ctx, PaymentRefundDateFuture, "refunded_at",
		"refunded date cannot be in the future", operation)
}

func CustomerIDRequiredError(ctx context.Context, operation string) error {
	return paymentValidationError(ctx, PaymentCustomerIDRequired, "customer_id",
		"customer ID is required", operation)
}

func AppointmentIDRequiredError(ctx context.Context, operation string) error {
	return paymentValidationError(ctx, PaymentAppointmentIDRequired, "appointment_id",
		"appointment ID is required", operation)
}

func InvoiceIDEmptyError(ctx context.Context, operation string) error {
	return paymentValidationError(ctx, PaymentInvoiceIDEmpty, "invoice_id",
		"invoice ID cannot be empty", operation)
}

func RefundAmountNegativeError(ctx context.Context, amount shared.Money, operation string) error {
	return paymentValidationError(ctx, PaymentRefundAmountNegative, "refund_amount",
		fmt.Sprintf("refund amount cannot be negative, got: %s", amount.Amount()), operation)
}

func FailureReasonEmptyError(ctx context.Context, operation string) error {
	return paymentValidationError(ctx, PaymentFailureReasonEmpty, "failure_reason",
		"failure reason cannot be empty", operation)
}

func CannotProcessError(ctx context.Context, currentStatus PaymentStatus, operation string) error {
	return paymentBusinessError(ctx, PaymentCannotProcess, "process", operation)
}

func CannotRefundError(ctx context.Context, currentStatus PaymentStatus, operation string) error {
	return paymentBusinessError(ctx, PaymentCannotRefund, "refund", operation)
}

func CannotCancelError(ctx context.Context, currentStatus PaymentStatus, operation string) error {
	return paymentBusinessError(ctx, PaymentCannotCancel, "cancel", operation)
}

func AlreadyProcessedError(ctx context.Context, currentStatus PaymentStatus, operation string) error {
	return paymentBusinessError(ctx, PaymentAlreadyProcessed, "process", operation)
}

func CannotMarkAsPaidError(ctx context.Context, currentStatus PaymentStatus, operation string) error {
	return paymentBusinessError(ctx, PaymentCannotProcess, "mark_paid", operation)
}

func CannotRequestRefundError(ctx context.Context, currentStatus PaymentStatus, operation string) error {
	return paymentBusinessError(ctx, PaymentCannotRefund, "request_refund", operation)
}

func RefundAmountExceededError(ctx context.Context, refundAmount, originalAmount shared.Money, operation string) error {
	return paymentValidationError(ctx, PaymentRefundAmountNegative, "refund_amount",
		fmt.Sprintf("refund amount (%s) cannot exceed original amount (%s)",
			refundAmount.Amount(), originalAmount.Amount()), operation)
}

func CancellationReasonRequiredError(ctx context.Context, operation string) error {
	return paymentValidationError(ctx, PaymentFailureReasonEmpty, "cancellation_reason",
		"cancellation reason is required", operation)
}

func ReferenceRequiredError(ctx context.Context, operation string) error {
	return paymentValidationError(ctx, PaymentReferenceEmpty, "reference",
		"payment reference is required", operation)
}

func CannotMarkOverdueError(ctx context.Context, currentStatus PaymentStatus, operation string) error {
	return paymentBusinessError(ctx, PaymentCannotProcess, "mark_overdue", operation)
}

func CannotMarkFailedError(ctx context.Context, currentStatus PaymentStatus, operation string) error {
	return paymentBusinessError(ctx, PaymentCannotProcess, "mark_failed", operation)
}

func DueDateNotReachedError(ctx context.Context, dueDate time.Time, operation string) error {
	return paymentBusinessError(ctx, PaymentCannotProcess, "mark_overdue", operation)
}

func RefundPeriodExpiredError(ctx context.Context, paidAt time.Time, operation string) error {
	return paymentBusinessError(ctx, PaymentCannotRefund, "refund", operation)
}
