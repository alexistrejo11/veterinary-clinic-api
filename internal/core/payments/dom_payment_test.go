package payments

import (
	"context"
	"testing"

	"clinic-vet-api/internal/shared"
)

func TestAuthorizedPaymentCanBeRecordedAsPaid(t *testing.T) {
	payment := Payment{
		Amount: shared.NewMoney(shared.NewDecimalFromInt(250), "MXN"),
		Status: PaymentStatusAuthorized,
	}

	if err := payment.Pay(context.Background(), "local-reference"); err != nil {
		t.Fatalf("record authorized payment: %v", err)
	}
	if payment.Status != PaymentStatusPaid || payment.Reference != "local-reference" || payment.PaidAt == nil {
		t.Fatalf("unexpected settled ledger entry: %#v", payment)
	}
}

func TestAuthorizedPaymentCanBeCancelled(t *testing.T) {
	payment := Payment{Status: PaymentStatusAuthorized}
	if err := payment.Cancel(context.Background(), "customer cancelled"); err != nil {
		t.Fatalf("cancel authorized payment: %v", err)
	}
	if payment.Status != PaymentStatusCancelled || payment.Reason == nil || *payment.Reason != "customer cancelled" {
		t.Fatalf("unexpected cancelled ledger entry: %#v", payment)
	}
}
