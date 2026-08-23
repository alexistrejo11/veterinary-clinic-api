package mapper

import (
	"clinic-vet-api/internal/core/addresses"
	"clinic-vet-api/internal/core/customers"
	"clinic-vet-api/internal/core/employees"
	"clinic-vet-api/internal/core/pets"
	"clinic-vet-api/internal/core/users"
	"clinic-vet-api/internal/shared"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type SqlcFieldMapper struct {
	Primitive     PrimitiveMapper
	PgText        PgTextMapper
	PgInt2        PgInt2Mapper
	PgBool        PgBoolMapper
	PgInt4        PgInt4Mapper
	PgDate        PgDateMapper
	PgTimestamptz PgTimestamptzMapper
	PgNumeric     NumericMapper
}

type PgInt4Mapper struct{}

type PgDateMapper struct{}

type NumericMapper struct{}

func NewSqlcFieldMapper() *SqlcFieldMapper {
	return &SqlcFieldMapper{
		Primitive:     PrimitiveMapper{},
		PgText:        PgTextMapper{},
		PgInt4:        PgInt4Mapper{},
		PgDate:        PgDateMapper{},
		PgTimestamptz: PgTimestamptzMapper{},
		PgNumeric:     NumericMapper{},
		PgInt2:        PgInt2Mapper{},
		PgBool:        PgBoolMapper{},
	}
}

type PgInt2Mapper struct{}

func (m *PgInt2Mapper) ToIntPtr(pgType pgtype.Int2) *int {
	if pgType.Valid {
		i := int(pgType.Int16)
		return &i
	}
	return nil
}

func (m *PgInt2Mapper) FromInt(i *int) pgtype.Int2 {
	if i != nil {
		return pgtype.Int2{Int16: int16(*i), Valid: true}
	}
	return pgtype.Int2{Valid: false}
}

func (m *PgInt2Mapper) ToInt(pgType pgtype.Int2) int {
	if pgType.Valid {
		return int(pgType.Int16)
	}
	return 0
}

func (m *PgInt2Mapper) FromIntValue(i int) pgtype.Int2 {
	if i != 0 {
		return pgtype.Int2{Int16: int16(i), Valid: true}
	}
	return pgtype.Int2{Valid: false}
}

type PgBoolMapper struct{}

func (m *PgBoolMapper) ToBoolPtr(pgType pgtype.Bool) *bool {
	if pgType.Valid {
		return &pgType.Bool
	}
	return nil
}

func (m *PgBoolMapper) FromBool(b *bool) pgtype.Bool {
	if b != nil {
		return pgtype.Bool{Bool: *b, Valid: true}
	}
	return pgtype.Bool{Valid: false}
}

func (m *PgBoolMapper) FromBoolValue(b bool) pgtype.Bool {
	return pgtype.Bool{Bool: b, Valid: true}
}

// Pg Type Mapper
type PgTextMapper struct{}

func (m *PgTextMapper) ToStringPtr(pgType pgtype.Text) *string {
	if pgType.Valid {
		return &pgType.String
	}
	return nil
}

func (m *PgTextMapper) ToString(pgType pgtype.Text) string {
	if pgType.Valid {
		return pgType.String
	}
	return ""
}

func (m *PgTextMapper) FromStringPtr(s *string) pgtype.Text {
	if s != nil {
		return pgtype.Text{String: *s, Valid: true}
	}
	return pgtype.Text{Valid: false}
}

func (m *PgTextMapper) FromString(s string) pgtype.Text {
	if s != "" {
		return pgtype.Text{String: s, Valid: true}
	}
	return pgtype.Text{Valid: false}
}

func (m *PgTextMapper) ToPhoneNumberPtr(t pgtype.Text) *users.PhoneNumber {
	if t.Valid && t.String != "" {
		phone := users.NewPhoneNumberNoErr(t.String)
		return &phone
	}
	return nil
}

func (m *PgInt4Mapper) ToInt32Ptr(pgType pgtype.Int4) *int32 {
	if pgType.Valid {
		return &pgType.Int32
	}
	return nil
}

func (m *PgInt4Mapper) FromUintPtr(i *uint) pgtype.Int4 {
	if i != nil {
		return pgtype.Int4{Int32: int32(*i), Valid: true}
	}
	return pgtype.Int4{Valid: false}
}

func (m *PgInt4Mapper) FromUint(i uint) pgtype.Int4 {
	if i != 0 {
		return pgtype.Int4{Int32: int32(i), Valid: true}
	}
	return pgtype.Int4{Valid: false}
}

func (m *PgInt4Mapper) ToUint(pgType pgtype.Int4) uint {
	if pgType.Valid {
		return uint(pgType.Int32)
	}
	return 0
}

func (m *PgInt4Mapper) FromInt32Ptr(i *int32) pgtype.Int4 {
	if i != nil {
		return pgtype.Int4{Int32: *i, Valid: true}
	}
	return pgtype.Int4{Valid: false}
}

func (m *PgInt4Mapper) FromInt32(i int32) pgtype.Int4 {
	if i != 0 {
		return pgtype.Int4{Int32: i, Valid: true}
	}
	return pgtype.Int4{Valid: false}
}

func (m *PgInt4Mapper) ToUserIDPtr(pgType pgtype.Int4) *shared.UserID {
	if pgType.Valid {
		id := shared.NewUserID(uint(pgType.Int32))
		return &id
	}
	return nil
}

func (m *PgInt4Mapper) FromUserIDPtr(id *shared.UserID) pgtype.Int4 {
	if id != nil {
		return pgtype.Int4{Int32: int32(id.Value()), Valid: true}
	}
	return pgtype.Int4{Valid: false}
}

func (m *PgInt4Mapper) ToCustomerIDPtr(pgType pgtype.Int4) *customers.CustomerID {
	if pgType.Valid {
		id := customers.NewCustomerID(uint(pgType.Int32))
		return &id
	}
	return nil
}

func (m *PgInt4Mapper) ToAddressID(pgType pgtype.Int4) addresses.AddressID {
	if pgType.Valid {
		return addresses.NewAddressID(uint(pgType.Int32))
	}
	return addresses.NewAddressID(0)
}

func (m *PgInt4Mapper) FromAddressID(id addresses.AddressID) pgtype.Int4 {
	if !id.IsZero() {
		return pgtype.Int4{Int32: int32(id.Value()), Valid: true}
	}
	return pgtype.Int4{Valid: false}
}

func (m *PgInt4Mapper) ToEmployeeIDPtr(pgType pgtype.Int4) *employees.EmployeeID {
	if pgType.Valid {
		id := employees.NewEmployeeID(uint(pgType.Int32))
		return &id
	}
	return nil
}

func (m *PgInt4Mapper) ToEmployeeID(pgType pgtype.Int4) employees.EmployeeID {
	if pgType.Valid {
		return employees.NewEmployeeID(uint(pgType.Int32))
	}
	return employees.EmployeeID{}
}

func (m *PgInt4Mapper) FromCustomerIDPtr(id *customers.CustomerID) pgtype.Int4 {
	if id != nil {
		return pgtype.Int4{Int32: int32(id.Value()), Valid: true}
	}
	return pgtype.Int4{Valid: false}
}

func (m *PgInt4Mapper) ToPetID(pgType pgtype.Int4) pets.PetID {
	if pgType.Valid {
		return pets.NewPetID(uint(pgType.Int32))
	}
	return pets.PetID{}
}

func (m *PgDateMapper) ToTimePtr(pgType pgtype.Date) *time.Time {
	if pgType.Valid {
		return &pgType.Time
	}
	return nil
}

func (m *PgDateMapper) ToTime(pgType pgtype.Date) time.Time {
	if pgType.Valid {
		return pgType.Time
	}
	return time.Time{}
}

func (m *PgDateMapper) FromTimePtr(t *time.Time) pgtype.Date {
	if t != nil && *t != (time.Time{}) {
		return pgtype.Date{Time: *t, Valid: true}
	}
	return pgtype.Date{Valid: false}
}

func (m *PgDateMapper) FromTime(t time.Time) pgtype.Date {
	if t != (time.Time{}) {
		return pgtype.Date{Time: t, Valid: true}
	}
	return pgtype.Date{Valid: false}
}

func (m *SqlcFieldMapper) StringPtrToPgText(s *string) pgtype.Text {
	if s != nil {
		return pgtype.Text{String: *s, Valid: true}
	}
	return pgtype.Text{Valid: false}
}

func (m *SqlcFieldMapper) StringToPgText(s string) pgtype.Text {
	if s != "" {
		return pgtype.Text{String: s, Valid: true}
	}
	return pgtype.Text{Valid: false}
}

func (m *SqlcFieldMapper) TimeToPgDate(t time.Time) pgtype.Date {
	if t != (time.Time{}) {
		return pgtype.Date{Time: t, Valid: true}
	}
	return pgtype.Date{Valid: false}
}

func (m *SqlcFieldMapper) TimePtrToPgDate(t *time.Time) pgtype.Date {
	if t != nil && *t != (time.Time{}) {
		return pgtype.Date{Time: *t, Valid: true}
	}
	return pgtype.Date{Valid: false}
}

func (m *SqlcFieldMapper) PgTimeStampToTime(t pgtype.Timestamptz) time.Time {
	if t.Valid {
		return t.Time
	}
	return time.Time{}
}

func (m *SqlcFieldMapper) TimePtrToPgTypestamp(t *time.Time) pgtype.Timestamptz {
	if t != nil && *t != (time.Time{}) {
		return pgtype.Timestamptz{Time: *t, Valid: true}
	}
	return pgtype.Timestamptz{Valid: false}
}

func (m *SqlcFieldMapper) UintToPgInt4(i uint) pgtype.Int4 {
	if i != 0 {
		return pgtype.Int4{Int32: int32(i), Valid: true}
	}
	return pgtype.Int4{Valid: false}
}

func (m *SqlcFieldMapper) PetIDPtrToInt32(id *pets.PetID) int32 {
	if id != nil {
		return int32(id.Value())
	}
	return 0
}

func (m *SqlcFieldMapper) EmployeeIDPtrToInt32(id *employees.EmployeeID) int32 {
	if id != nil {
		return int32(id.Value())
	}
	return 0
}

func (m *SqlcFieldMapper) UserIDPtrToInt32(id *shared.UserID) pgtype.Int4 {
	if id != nil {
		return pgtype.Int4{Int32: int32(id.Value()), Valid: true}
	}
	return pgtype.Int4{Valid: false}
}

func (m *SqlcFieldMapper) StringPtrToString(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

func (m *SqlcFieldMapper) Int32ToUserIDPtr(i int32) *shared.UserID {
	if i != 0 {
		id := shared.NewUserID(uint(i))
		return &id
	}
	return nil
}

type PgTimestamptzMapper struct{}

func (r *PgTimestamptzMapper) FromTime(t time.Time) pgtype.Timestamptz {
	if t != (time.Time{}) {
		return pgtype.Timestamptz{Time: t, Valid: true}
	}
	return pgtype.Timestamptz{Valid: false}
}

func (r *PgTimestamptzMapper) FromTimePtr(t *time.Time) pgtype.Timestamptz {
	if t != nil && *t != (time.Time{}) {
		return pgtype.Timestamptz{Time: *t, Valid: true}
	}
	return pgtype.Timestamptz{Valid: false}
}

func (r *PgTimestamptzMapper) ToTime(t pgtype.Timestamptz) time.Time {
	if t.Valid {
		return t.Time
	}
	return time.Time{}
}

func (r *PgTimestamptzMapper) ToTimePtr(t pgtype.Timestamptz) *time.Time {
	if t.Valid {
		return &t.Time
	}
	return nil
}

// TODO: Check if this is correct
func (r *NumericMapper) ToMoney(num pgtype.Numeric, currency string) shared.Money {
	if !num.Valid {
		return shared.Money{}
	}

	floatValue, err := r.numericToFloat64(num)
	if err != nil {
		return shared.Money{}
	}

	decimal := shared.NewDecimalFromFloat(floatValue)
	return shared.NewMoney(decimal, currency)
}

func (r *NumericMapper) numericToFloat64(num pgtype.Numeric) (float64, error) {
	if !num.Valid {
		return 0, nil
	}

	float8, err := num.Float64Value()
	if err != nil {
		return 0, fmt.Errorf("error converting numeric to float64: %w", err)
	}

	if !float8.Valid {
		return 0, nil
	}

	return float8.Float64, nil
}

func (r *NumericMapper) ToMoneyPtr(num pgtype.Numeric, currency string) (*shared.Money, error) {
	if !num.Valid {
		return nil, nil
	}

	decimal, err := r.numericToDecimal(num)
	if err != nil {
		return nil, err
	}

	if decimal.IsZero() {
		return nil, nil
	}

	money := shared.NewMoney(decimal, currency)
	return &money, nil
}

func (r *NumericMapper) numericToDecimal(num pgtype.Numeric) (shared.Decimal, error) {
	if !num.Valid || num.Int == nil {
		return shared.Decimal{}, nil
	}

	value := num.Int.String()
	sign := ""
	if strings.HasPrefix(value, "-") {
		sign = "-"
		value = strings.TrimPrefix(value, "-")
	}

	if num.Exp >= 0 {
		value += strings.Repeat("0", int(num.Exp))
	} else {
		fractionDigits := int(-num.Exp)
		if len(value) <= fractionDigits {
			value = "0." + strings.Repeat("0", fractionDigits-len(value)) + value
		} else {
			value = value[:len(value)-fractionDigits] + "." + value[len(value)-fractionDigits:]
		}
	}

	decimal, err := shared.NewDecimalFromString(sign + value)
	if err != nil {
		return shared.Decimal{}, fmt.Errorf("converting numeric to decimal: %w", err)
	}
	return decimal, nil
}

func (r *NumericMapper) FromMoneyPtr(money *shared.Money) pgtype.Numeric {
	if money != nil {
		amount := money.Amount()
		return pgtype.Numeric{
			Int:   big.NewInt(amount.Int()),
			Exp:   -2,
			Valid: true,
		}
	}
	return pgtype.Numeric{Valid: false}
}

func (r *NumericMapper) FromMoney(money shared.Money) pgtype.Numeric {
	amount := money.Amount()
	return pgtype.Numeric{
		Int:   big.NewInt(amount.Int()),
		Exp:   -2,
		Valid: true,
	}
}

func (r *NumericMapper) ToNumeric(money shared.Money) pgtype.Numeric {
	amount := money.Amount()
	return pgtype.Numeric{
		Int:   big.NewInt(amount.Int()),
		Exp:   -2,
		Valid: true,
	}
}

func (r *NumericMapper) ToDecimalPtr(num pgtype.Numeric) *shared.Decimal {
	if num.Valid {
		var amount float64
		if err := num.Scan(&amount); err == nil {
			dec := shared.NewDecimalFromFloat(amount)
			return &dec
		}
	}
	return nil
}
func (r *NumericMapper) ToDecimal(num pgtype.Numeric) shared.Decimal {
	if num.Valid {
		var amount float64
		if err := num.Scan(&amount); err == nil {
			return shared.NewDecimalFromFloat(amount)
		}
	}
	return shared.Decimal{}
}

func (r *NumericMapper) FromDecimalPtr(dec *shared.Decimal) pgtype.Numeric {
	if dec != nil {
		return pgtype.Numeric{
			Int:   big.NewInt(int64(dec.Float64())),
			Valid: true,
		}
	}
	return pgtype.Numeric{Valid: false}
}

func (r *NumericMapper) FromDecimal(dec shared.Decimal) pgtype.Numeric {
	return pgtype.Numeric{
		Int:   big.NewInt(int64(dec.Float64())),
		Valid: true,
	}
}
