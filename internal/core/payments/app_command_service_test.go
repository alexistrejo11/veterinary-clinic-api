package payments

import (
	"context"
	"testing"
	"time"

	"clinic-vet-api/internal/core/customers"
	"clinic-vet-api/internal/shared"
	"clinic-vet-api/internal/shared/page"
)

type paymentRepositoryStub struct {
	payment Payment
	saved   bool
}

func (r *paymentRepositoryStub) GetByID(context.Context, PaymentID) (Payment, error) {
	return r.payment, nil
}
func (*paymentRepositoryStub) GetByReference(context.Context, string) (Payment, error) {
	return Payment{}, context.Canceled
}
func (*paymentRepositoryStub) GetByIDAndCustomerID(context.Context, PaymentID, customers.CustomerID) (Payment, error) {
	return Payment{}, context.Canceled
}
func (*paymentRepositoryStub) GetBySpecification(context.Context, PaymentSpecification) (page.Page[Payment], error) {
	return page.Page[Payment]{}, nil
}
func (*paymentRepositoryStub) ExistsByID(context.Context, PaymentID) (bool, error) { return true, nil }
func (*paymentRepositoryStub) ExistsByReference(context.Context, string) (bool, error) {
	return false, nil
}
func (*paymentRepositoryStub) ExistsPendingByCustomerID(context.Context, customers.CustomerID) (bool, error) {
	return false, nil
}
func (r *paymentRepositoryStub) Save(_ context.Context, payment *Payment) error {
	r.payment = *payment
	r.saved = true
	return nil
}
func (*paymentRepositoryStub) SettleAuthorized(context.Context, PaymentID, string, string) (Payment, error) {
	return Payment{}, nil
}
func (*paymentRepositoryStub) CancelAuthorized(context.Context, PaymentID, string) (Payment, error) {
	return Payment{}, nil
}
func (*paymentRepositoryStub) RefundSettled(context.Context, PaymentID, string) (Payment, error) {
	return Payment{}, nil
}
func (*paymentRepositoryStub) SoftDelete(context.Context, PaymentID) error { return nil }
func (*paymentRepositoryStub) HardDelete(context.Context, PaymentID) error { return nil }
func (*paymentRepositoryStub) CountByStatus(context.Context, PaymentStatus) (int64, error) {
	return 0, nil
}
func (*paymentRepositoryStub) CountByCustomerID(context.Context, customers.CustomerID) (int64, error) {
	return 0, nil
}
func (*paymentRepositoryStub) CountOverdue(context.Context) (int64, error) { return 0, nil }
func (*paymentRepositoryStub) TotalRevenueByDateRange(context.Context, time.Time, time.Time) (float64, error) {
	return 0, nil
}

func TestUpdatePaymentAllowsAuthorizedLedgerEntry(t *testing.T) {
	repository := &paymentRepositoryStub{payment: Payment{
		Amount: shared.NewMoney(shared.NewDecimalFromInt(100), "MXN"),
		Status: PaymentStatusAuthorized,
		Method: PaymentMethodCash,
	}}
	service := NewCommandService(repository, nil)

	err := service.UpdatePayment(context.Background(), UpdatePaymentCommand{
		ID:       NewPaymentID(1),
		Amount:   125,
		Currency: "MXN",
	})
	if err != nil {
		t.Fatalf("update authorized payment: %v", err)
	}
	if !repository.saved || repository.payment.Amount.Amount().Float64() != 125 {
		t.Fatalf("updated payment = %#v, saved=%t", repository.payment, repository.saved)
	}
}
