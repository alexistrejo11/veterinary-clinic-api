package mapper

import (
	"math/big"
	"testing"

	"clinic-vet-api/internal/shared"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestNumericMapperMoneyPreservesCents(t *testing.T) {
	amount, err := shared.NewDecimalFromString("12.34")
	if err != nil {
		t.Fatal(err)
	}

	mapper := NumericMapper{}
	numeric := mapper.FromMoney(shared.NewMoney(amount, "MXN"))
	if numeric.Int.Cmp(big.NewInt(1234)) != 0 || numeric.Exp != -2 || !numeric.Valid {
		t.Fatalf("expected 1234e-2 numeric, got %+v", numeric)
	}

	money, err := mapper.ToMoneyPtr(pgtype.Numeric{Int: big.NewInt(1234), Exp: -2, Valid: true}, "MXN")
	if err != nil {
		t.Fatal(err)
	}
	if money == nil || money.Amount().String() != "12.34" {
		t.Fatalf("expected 12.34, got %#v", money)
	}
}
