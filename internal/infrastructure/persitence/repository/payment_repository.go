package repository

import (
	"clinic-vet-api/database/models"
	"clinic-vet-api/database/sqlc"
	"clinic-vet-api/internal/core/appointments"
	"clinic-vet-api/internal/core/customers"
	"clinic-vet-api/internal/core/payments"
	customErr "clinic-vet-api/internal/shared/errors"
	"clinic-vet-api/internal/shared/mapper"
	"clinic-vet-api/internal/shared/page"
	"context"
	"fmt"
	"time"

	"clinic-vet-api/internal/shared"
	"github.com/jackc/pgx/v5/pgtype"
)

type PaymentSqlcRepository struct {
	queries *sqlc.Queries
	pgMap   *mapper.SqlcFieldMapper
}

func NewPaymentSqlcRepository(queries *sqlc.Queries) payments.PaymentRepository {
	return &PaymentSqlcRepository{
		queries: queries,
		pgMap:   mapper.NewSqlcFieldMapper(),
	}
}

func (r *PaymentSqlcRepository) GetByID(ctx context.Context, id payments.PaymentID) (payments.Payment, error) {
	row, err := r.queries.FindPaymentByID(ctx, id.Int32())
	if err != nil {
		return payments.Payment{}, r.dbError(OpSelect, ErrMsgFindPaymentByID, err)
	}
	return r.toEntity(row), nil
}

func (r *PaymentSqlcRepository) GetByReference(ctx context.Context, reference string) (payments.Payment, error) {
	row, err := r.queries.FindPaymentByReference(ctx, r.pgMap.PgText.FromString(reference))
	if err != nil {
		return payments.Payment{}, r.dbError(OpSelect, ErrMsgFindPaymentByTransactionID, err)
	}
	return r.toEntity(row), nil
}

func (r *PaymentSqlcRepository) GetByIDAndCustomerID(ctx context.Context, id payments.PaymentID, customerID customers.CustomerID) (payments.Payment, error) {
	row, err := r.queries.FindPaymentByIDAndCustomerID(ctx, sqlc.FindPaymentByIDAndCustomerIDParams{
		ID:               id.Int32(),
		PaidByCustomerID: r.pgMap.PgInt4.FromInt32(customerID.Int32()),
	})
	if err != nil {
		return payments.Payment{}, r.dbError(OpSelect, ErrMsgFindPaymentByIDAndCustomerID, err)
	}
	return r.toEntity(row), nil
}

func (r *PaymentSqlcRepository) GetBySpecification(ctx context.Context, spec payments.PaymentSpecification) (page.Page[payments.Payment], error) {
	var customerID int32
	if spec.CustomerID != nil {
		customerID = spec.CustomerID.Int32()
	}

	var status models.PaymentStatus
	if spec.Status != nil {
		status = models.PaymentStatus(spec.Status.String())
	}

	var method models.PaymentMethod
	if spec.Method != nil {
		method = models.PaymentMethod(spec.Method.String())
	}

	var fromCreated, toCreated pgtype.Timestamptz
	if spec.FromCreatedAt != nil {
		fromCreated = r.pgMap.PgTimestamptz.FromTime(*spec.FromCreatedAt)
	}
	if spec.ToCreatedAt != nil {
		toCreated = r.pgMap.PgTimestamptz.FromTime(*spec.ToCreatedAt)
	}

	limit := int32(spec.Pagination.Size)
	if limit <= 0 {
		limit = page.DefaultPageSize
	}
	offset := spec.Pagination.Offset()

	params := sqlc.FindPaymentsBySpecificationParams{
		Column1: customerID,
		Column2: status,
		Column3: method,
		Column4: fromCreated,
		Column5: toCreated,
		Column6: spec.OverdueOnly,
		Limit:   limit,
		Offset:  offset,
	}

	rows, err := r.queries.FindPaymentsBySpecification(ctx, params)
	if err != nil {
		return page.Page[payments.Payment]{}, r.dbError(OpSelect, ErrMsgFindPaymentsBySpecification, err)
	}

	countParams := sqlc.CountPaymentsBySpecificationParams{
		Column1: customerID,
		Column2: status,
		Column3: method,
		Column4: fromCreated,
		Column5: toCreated,
		Column6: spec.OverdueOnly,
	}
	total, err := r.queries.CountPaymentsBySpecification(ctx, countParams)
	if err != nil {
		return page.Page[payments.Payment]{}, r.dbError(OpCount, ErrMsgFindPaymentsBySpecification, err)
	}

	items := make([]payments.Payment, len(rows))
	for i := range rows {
		items[i] = r.toEntity(rows[i])
	}

	pagReq := page.PaginationRequest{
		Page:     int32(spec.Pagination.Number),
		PageSize: limit,
	}
	return page.NewPage(items, total, pagReq), nil
}

func (r *PaymentSqlcRepository) ExistsByID(ctx context.Context, id payments.PaymentID) (bool, error) {
	exists, err := r.queries.ExistsPaymentByID(ctx, id.Int32())
	if err != nil {
		return false, r.dbError(OpSelect, ErrMsgCheckPaymentExists, err)
	}
	return exists, nil
}

func (r *PaymentSqlcRepository) ExistsByReference(ctx context.Context, reference string) (bool, error) {
	exists, err := r.queries.ExistsPaymentByReference(ctx, r.pgMap.PgText.FromString(reference))
	if err != nil {
		return false, r.dbError(OpSelect, ErrMsgCheckPaymentExists, err)
	}
	return exists, nil
}

func (r *PaymentSqlcRepository) SettleAuthorized(ctx context.Context, id payments.PaymentID, receiptNumber, reason string) (payments.Payment, error) {
	row, err := r.queries.SettleAuthorizedPayment(ctx, sqlc.SettleAuthorizedPaymentParams{ID: id.Int32(), ReceiptNumber: r.pgMap.PgText.FromString(receiptNumber), Reason: r.pgMap.PgText.FromString(reason)})
	if err != nil {
		return payments.Payment{}, r.dbError(OpUpdate, ErrMsgUpdatePayment, err)
	}
	return r.toEntity(row), nil
}

func (r *PaymentSqlcRepository) CancelAuthorized(ctx context.Context, id payments.PaymentID, reason string) (payments.Payment, error) {
	row, err := r.queries.CancelAuthorizedPayment(ctx, sqlc.CancelAuthorizedPaymentParams{ID: id.Int32(), Reason: r.pgMap.PgText.FromString(reason)})
	if err != nil {
		return payments.Payment{}, r.dbError(OpUpdate, ErrMsgUpdatePayment, err)
	}
	return r.toEntity(row), nil
}

func (r *PaymentSqlcRepository) RefundSettled(ctx context.Context, id payments.PaymentID, reason string) (payments.Payment, error) {
	row, err := r.queries.RefundSettledPayment(ctx, sqlc.RefundSettledPaymentParams{ID: id.Int32(), Reason: r.pgMap.PgText.FromString(reason)})
	if err != nil {
		return payments.Payment{}, r.dbError(OpUpdate, ErrMsgUpdatePayment, err)
	}
	return r.toEntity(row), nil
}

func (r *PaymentSqlcRepository) Save(ctx context.Context, payment *payments.Payment) error {
	if payment.ID.IsZero() {
		return r.create(ctx, payment)
	}
	return r.update(ctx, payment)
}

func (r *PaymentSqlcRepository) SoftDelete(ctx context.Context, id payments.PaymentID) error {
	return r.queries.SoftDeletePayment(ctx, id.Int32())
}

func (r *PaymentSqlcRepository) HardDelete(ctx context.Context, id payments.PaymentID) error {
	return r.queries.HardDeletePayment(ctx, id.Int32())
}

func (r *PaymentSqlcRepository) CountByStatus(ctx context.Context, status payments.PaymentStatus) (int64, error) {
	count, err := r.queries.CountPaymentsByStatus(ctx, models.PaymentStatus(status.String()))
	if err != nil {
		return 0, r.dbError(OpCount, ErrMsgCountPaymentsByStatus, err)
	}
	return count, nil
}

func (r *PaymentSqlcRepository) CountByCustomerID(ctx context.Context, customerID customers.CustomerID) (int64, error) {
	count, err := r.queries.CountPaymentsByCustomerID(ctx, r.pgMap.PgInt4.FromInt32(customerID.Int32()))
	if err != nil {
		return 0, r.dbError(OpCount, ErrMsgCountPaymentsByCustomerID, err)
	}
	return count, nil
}

func (r *PaymentSqlcRepository) CountOverdue(ctx context.Context) (int64, error) {
	count, err := r.queries.CountOverduePayments(ctx)
	if err != nil {
		return 0, r.dbError(OpCount, ErrMsgCountOverduePayments, err)
	}

	return count, nil
}

func (r *PaymentSqlcRepository) TotalRevenueByDateRange(ctx context.Context, startDate, endDate time.Time) (float64, error) {
	total, err := r.queries.TotalRevenueByDateRange(ctx, sqlc.TotalRevenueByDateRangeParams{
		PaidAt:   r.pgMap.PgTimestamptz.FromTime(startDate),
		PaidAt_2: r.pgMap.PgTimestamptz.FromTime(endDate),
	})
	if err != nil {
		return 0, r.dbError(OpSelect, ErrMsgTotalRevenueByDateRange, err)
	}
	return total.(float64), nil
}

func (r *PaymentSqlcRepository) ExistsPendingByCustomerID(ctx context.Context, customerID customers.CustomerID) (bool, error) {
	exists, err := r.queries.ExistsPendingPaymentByCustomerID(ctx, r.pgMap.PgInt4.FromInt32(customerID.Int32()))
	if err != nil {
		return false, r.dbError(OpSelect, ErrMsgCheckPaymentExists, err)
	}
	return exists, nil
}

func (r *PaymentSqlcRepository) create(ctx context.Context, payment *payments.Payment) error {
	params := sqlc.CreatePaymentParams{
		Amount:      r.pgMap.PgNumeric.ToNumeric(payment.Amount),
		Currency:    payment.Amount.Currency(),
		Status:      models.PaymentStatus(payment.Status.String()),
		Method:      models.PaymentMethod(payment.Method.String()),
		Reference:   r.pgMap.PgText.FromString(payment.Reference),
		Description: r.pgMap.PgText.FromStringPtr(payment.Description),
		AppointmentID: func() pgtype.Int4 {
			if payment.AppointmentID != nil {
				return r.pgMap.PgInt4.FromInt32(payment.AppointmentID.Int32())
			}
			return pgtype.Int4{}
		}(),
		Reason:           r.pgMap.PgText.FromStringPtr(payment.Reason),
		DueDate:          r.pgMap.PgTimestamptz.FromTimePtr(payment.DueDate),
		PaidAt:           r.pgMap.PgTimestamptz.FromTimePtr(payment.PaidAt),
		RefundedAt:       r.pgMap.PgTimestamptz.FromTimePtr(payment.RefundedAt),
		PaidByCustomerID: r.pgMap.PgInt4.FromCustomerIDPtr(&payment.PaidFromCustomer),
		MedSessionID:     pgtype.Int4{Valid: false},
		InvoiceID:        r.pgMap.PgText.FromStringPtr(payment.InvoiceID),
		RefundAmount: func() pgtype.Numeric {
			if payment.RefundAmount != nil {
				return r.pgMap.PgNumeric.FromMoney(*payment.RefundAmount)
			}
			return pgtype.Numeric{Valid: false}
		}(),
		FailureReason: r.pgMap.PgText.FromStringPtr(payment.FailureReason),
	}
	created, err := r.queries.CreatePayment(ctx, params)
	if err != nil {
		return r.dbError(OpInsert, ErrMsgCreatePayment, err)
	}
	payment.SetID(payments.NewPaymentID(uint(created.ID)))
	payment.SetTimeStamps(r.pgMap.PgTimestamptz.ToTime(created.CreatedAt), r.pgMap.PgTimestamptz.ToTime(created.UpdatedAt))
	payment.IsActive = created.IsActive
	return nil
}

func (r *PaymentSqlcRepository) update(ctx context.Context, payment *payments.Payment) error {
	params := sqlc.UpdatePaymentParams{
		ID:          payment.ID.Int32(),
		Amount:      r.pgMap.PgNumeric.ToNumeric(payment.Amount),
		Currency:    payment.Amount.Currency(),
		Status:      models.PaymentStatus(payment.Status.String()),
		Method:      models.PaymentMethod(payment.Method.String()),
		Reference:   r.pgMap.PgText.FromString(payment.Reference),
		Description: r.pgMap.PgText.FromStringPtr(payment.Description),
		AppointmentID: func() pgtype.Int4 {
			if payment.AppointmentID != nil {
				return r.pgMap.PgInt4.FromInt32(payment.AppointmentID.Int32())
			}
			return pgtype.Int4{}
		}(),
		Reason:           r.pgMap.PgText.FromStringPtr(payment.Reason),
		DueDate:          r.pgMap.PgTimestamptz.FromTimePtr(payment.DueDate),
		PaidAt:           r.pgMap.PgTimestamptz.FromTimePtr(payment.PaidAt),
		RefundedAt:       r.pgMap.PgTimestamptz.FromTimePtr(payment.RefundedAt),
		PaidByCustomerID: r.pgMap.PgInt4.FromCustomerIDPtr(&payment.PaidFromCustomer),
		MedSessionID:     pgtype.Int4{Valid: false},
		InvoiceID:        r.pgMap.PgText.FromStringPtr(payment.InvoiceID),
		RefundAmount: func() pgtype.Numeric {
			if payment.RefundAmount != nil {
				return r.pgMap.PgNumeric.FromMoney(*payment.RefundAmount)
			}
			return pgtype.Numeric{Valid: false}
		}(),
		FailureReason: r.pgMap.PgText.FromStringPtr(payment.FailureReason),
	}
	updated, err := r.queries.UpdatePayment(ctx, params)
	if err != nil {
		return r.dbError(OpUpdate, ErrMsgUpdatePayment, err)
	}
	// Refresh aggregate from DB row
	*payment = r.toEntity(updated)
	return nil
}

func (r *PaymentSqlcRepository) toEntity(row sqlc.Payment) payments.Payment {
	amountMoney := r.pgMap.PgNumeric.ToMoney(row.Amount, row.Currency)

	reference := r.pgMap.PgText.ToString(row.Reference)
	receiptNumber := r.pgMap.PgText.ToStringPtr(row.ReceiptNumber)
	reason := r.pgMap.PgText.ToStringPtr(row.Reason)
	var description *string = r.pgMap.PgText.ToStringPtr(row.Description)
	var dueDate *time.Time = r.pgMap.PgTimestamptz.ToTimePtr(row.DueDate)
	var paidAt *time.Time = r.pgMap.PgTimestamptz.ToTimePtr(row.PaidAt)
	var refundedAt *time.Time = r.pgMap.PgTimestamptz.ToTimePtr(row.RefundedAt)
	var invoiceID *string = r.pgMap.PgText.ToStringPtr(row.InvoiceID)
	var failureReason *string = r.pgMap.PgText.ToStringPtr(row.FailureReason)

	var refundAmount *shared.Money
	if m, err := r.pgMap.PgNumeric.ToMoneyPtr(row.RefundAmount, row.Currency); err == nil {
		refundAmount = m
	}

	var paidFrom customers.CustomerID
	if cid := r.pgMap.PgInt4.ToCustomerIDPtr(row.PaidByCustomerID); cid != nil {
		paidFrom = *cid
	}

	var deletedAt *time.Time = r.pgMap.PgTimestamptz.ToTimePtr(row.DeletedAt)
	var appointmentID *appointments.AppointmentID
	if row.AppointmentID.Valid {
		id := appointments.NewAppointmentID(uint(row.AppointmentID.Int32))
		appointmentID = &id
	}

	entity := payments.Payment{
		Amount:           amountMoney,
		Status:           payments.PaymentStatus(row.Status),
		Method:           payments.PaymentMethod(row.Method),
		Reference:        reference,
		ReceiptNumber:    receiptNumber,
		Reason:           reason,
		Description:      description,
		DueDate:          dueDate,
		PaidAt:           paidAt,
		RefundedAt:       refundedAt,
		IsActive:         row.IsActive,
		PaidFromCustomer: paidFrom,
		AppointmentID:    appointmentID,
		InvoiceID:        invoiceID,
		RefundAmount:     refundAmount,
		FailureReason:    failureReason,
		DeletedAt:        deletedAt,
	}
	entity.SetID(payments.NewPaymentID(uint(row.ID)))
	entity.SetTimeStamps(r.pgMap.PgTimestamptz.ToTime(row.CreatedAt), r.pgMap.PgTimestamptz.ToTime(row.UpdatedAt))
	return entity
}

func (r *PaymentSqlcRepository) dbError(operation, message string, err error) error {
	return customErr.DatabaseError(operation, TablePayments, DriverSQL, fmt.Errorf("%s: %w", message, err))
}

func (r *PaymentSqlcRepository) notFoundError(parameterName, parameterValue string) error {
	return customErr.DBNotFoundError(parameterName, parameterValue, OpSelect, TablePayments, DriverSQL)
}
