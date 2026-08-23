-- name: FindPaymentByID :one
SELECT payments.* FROM payments
WHERE id = $1 AND deleted_at IS NULL;

-- name: FindPaymentByReference :one
SELECT payments.* FROM payments
WHERE reference = $1 AND deleted_at IS NULL;

-- name: FindPaymentByIDAndCustomerID :one
SELECT payments.* FROM payments
WHERE id = $1 AND paid_by_customer_id = $2 AND deleted_at IS NULL;

-- name: CountPaymentsByCustomerID :one
SELECT COUNT(*) FROM payments
WHERE paid_by_customer_id = $1 AND deleted_at IS NULL;

-- name: CountPaymentsByStatus :one
SELECT COUNT(*) FROM payments
WHERE status = $1 AND deleted_at IS NULL;

-- name: CountOverduePayments :one
SELECT COUNT(*) FROM payments
WHERE due_date < CURRENT_TIMESTAMP 
  AND status NOT IN ('paid', 'refunded', 'cancelled') 
  AND deleted_at IS NULL;

-- name: ExistsPaymentByID :one
SELECT COUNT(*) > 0 FROM payments
WHERE id = $1 AND deleted_at IS NULL;

-- name: ExistsPaymentByReference :one
SELECT COUNT(*) > 0 FROM payments
WHERE reference = $1 AND deleted_at IS NULL;

-- name: ExistsPendingPaymentByCustomerID :one
SELECT COUNT(*) > 0 FROM payments
WHERE paid_by_customer_id = $1 
  AND status = 'pending' 
  AND deleted_at IS NULL;

-- name: CreatePayment :one
INSERT INTO payments (
    amount, currency, status, method, reference, description, appointment_id, reason,
    due_date, paid_at, refunded_at, paid_by_customer_id,
    med_session_id, invoice_id, refund_amount, failure_reason
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16
)
RETURNING payments.*;

-- name: UpdatePayment :one
UPDATE payments SET
    amount = $2,
    currency = $3,
    status = $4,
    method = $5,
    reference = $6,
    description = $7,
    appointment_id = $8,
    reason = $9,
    due_date = $10,
    paid_at = $11,
    refunded_at = $12,
    paid_by_customer_id = $13,
    med_session_id = $14,
    invoice_id = $15,
    refund_amount = $16,
    failure_reason = $17,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING payments.*;

-- name: SettleAuthorizedPayment :one
UPDATE payments SET
    status = 'paid',
    paid_at = CURRENT_TIMESTAMP,
    receipt_number = $2,
    reason = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND status IN ('authorized', 'pending', 'overdue') AND deleted_at IS NULL
RETURNING payments.*;

-- name: CancelAuthorizedPayment :one
UPDATE payments SET
    status = 'cancelled',
    reason = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND status IN ('authorized', 'pending', 'overdue') AND deleted_at IS NULL
RETURNING payments.*;

-- name: RefundSettledPayment :one
UPDATE payments SET
    status = 'refunded',
    refunded_at = CURRENT_TIMESTAMP,
    refund_amount = amount,
    reason = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND status = 'paid' AND deleted_at IS NULL
RETURNING payments.*;

-- name: SoftDeletePayment :exec
UPDATE payments SET
    is_active = FALSE,
    deleted_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: HardDeletePayment :exec
DELETE FROM payments WHERE id = $1;

-- name: TotalRevenueByDateRange :one
SELECT COALESCE(SUM(amount), 0) FROM payments
WHERE status = 'paid'
  AND paid_at BETWEEN $1 AND $2
  AND deleted_at IS NULL;

-- name: FindPaymentsBySpecification :many
SELECT payments.* FROM payments
WHERE deleted_at IS NULL
  AND ($1::int IS NULL OR paid_by_customer_id = $1)
  AND ($2::payment_status IS NULL OR status = $2)
  AND ($3::payment_method IS NULL OR method = $3)
  AND ($4::timestamptz IS NULL OR created_at >= $4)
  AND ($5::timestamptz IS NULL OR created_at <= $5)
  AND (NOT $6::boolean OR (due_date < CURRENT_TIMESTAMP AND status NOT IN ('paid','refunded','cancelled')))
ORDER BY created_at DESC
LIMIT $7 OFFSET $8;

-- name: CountPaymentsBySpecification :one
SELECT COUNT(*) FROM payments
WHERE deleted_at IS NULL
  AND ($1::int IS NULL OR paid_by_customer_id = $1)
  AND ($2::payment_status IS NULL OR status = $2)
  AND ($3::payment_method IS NULL OR method = $3)
  AND ($4::timestamptz IS NULL OR created_at >= $4)
  AND ($5::timestamptz IS NULL OR created_at <= $5)
  AND (NOT $6::boolean OR (due_date < CURRENT_TIMESTAMP AND status NOT IN ('paid','refunded','cancelled')));
