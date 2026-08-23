package payments

import (
	"clinic-vet-api/internal/shared"
	"fmt"
	"slices"
)

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusAuthorized    PaymentStatus = "authorized"
	PaymentStatusPending       PaymentStatus = "pending"
	PaymentStatusPaid          PaymentStatus = "paid"
	PaymentStatusFailed        PaymentStatus = "failed"
	PaymentStatusCancelled     PaymentStatus = "cancelled"
	PaymentStatusRefunded      PaymentStatus = "refunded"
	PaymentStatusOverdue       PaymentStatus = "overdue"
	PaymentStatusPartiallyPaid PaymentStatus = "partially_paid"
	PaymentStatusDisputed      PaymentStatus = "disputed"
)

// PaymentStatus constants and methods
var (
	ValidPaymentStatuses = []PaymentStatus{
		PaymentStatusAuthorized,
		PaymentStatusPending,
		PaymentStatusPaid,
		PaymentStatusFailed,
		PaymentStatusCancelled,
		PaymentStatusRefunded,
		PaymentStatusOverdue,
		PaymentStatusPartiallyPaid,
		PaymentStatusDisputed,
	}

	paymentStatusMap = map[string]PaymentStatus{
		"authorized":     PaymentStatusAuthorized,
		"pending":        PaymentStatusPending,
		"paid":           PaymentStatusPaid,
		"failed":         PaymentStatusFailed,
		"cancelled":      PaymentStatusCancelled,
		"canceled":       PaymentStatusCancelled,
		"refunded":       PaymentStatusRefunded,
		"overdue":        PaymentStatusOverdue,
		"partially_paid": PaymentStatusPartiallyPaid,
		"partially paid": PaymentStatusPartiallyPaid,
		"disputed":       PaymentStatusDisputed,
	}

	paymentStatusDisplayNames = map[PaymentStatus]string{
		PaymentStatusAuthorized:    "Authorized",
		PaymentStatusPending:       "Pending",
		PaymentStatusPaid:          "Paid",
		PaymentStatusFailed:        "Failed",
		PaymentStatusCancelled:     "Cancelled",
		PaymentStatusRefunded:      "Refunded",
		PaymentStatusOverdue:       "Overdue",
		PaymentStatusPartiallyPaid: "Partially Paid",
		PaymentStatusDisputed:      "Disputed",
	}

	paymentStatusFlow = map[PaymentStatus][]PaymentStatus{
		PaymentStatusAuthorized: {
			PaymentStatusPaid,
			PaymentStatusCancelled,
		},
		PaymentStatusPending: {
			PaymentStatusPaid,
			PaymentStatusFailed,
			PaymentStatusCancelled,
			PaymentStatusPartiallyPaid,
			PaymentStatusDisputed,
		},
		PaymentStatusPartiallyPaid: {
			PaymentStatusPaid,
			PaymentStatusFailed,
			PaymentStatusCancelled,
			PaymentStatusDisputed,
		},
		PaymentStatusPaid: {
			PaymentStatusRefunded,
			PaymentStatusDisputed,
		},
		PaymentStatusFailed: {
			PaymentStatusPending,
			PaymentStatusCancelled,
		},
		PaymentStatusDisputed: {
			PaymentStatusPaid,
			PaymentStatusRefunded,
			PaymentStatusCancelled,
		},
	}
)

func (ps PaymentStatus) IsValid() bool {
	_, exists := paymentStatusMap[string(ps)]
	return exists
}

func ParsePaymentStatus(status string) (PaymentStatus, error) {
	normalized := shared.NormalizeInput(status)
	if val, exists := paymentStatusMap[normalized]; exists {
		return val, nil
	}
	return "", fmt.Errorf("invalid payment status: %s", status)
}

func MustParsePaymentStatus(status string) PaymentStatus {
	parsed, err := ParsePaymentStatus(status)
	if err != nil {
		panic(err)
	}
	return parsed
}

func (ps PaymentStatus) String() string {
	return string(ps)
}

func (ps PaymentStatus) DisplayName() string {
	if displayName, exists := paymentStatusDisplayNames[ps]; exists {
		return displayName
	}
	return "Unknown Status"
}

func (ps PaymentStatus) Values() []PaymentStatus {
	return ValidPaymentStatuses
}

func (ps PaymentStatus) CanTransitionTo(newStatus PaymentStatus) bool {
	if allowedTransitions, exists := paymentStatusFlow[ps]; exists {
		return slices.Contains(allowedTransitions, newStatus)
	}
	return false
}

func (ps PaymentStatus) IsFinal() bool {
	finalStatuses := []PaymentStatus{
		PaymentStatusPaid,
		PaymentStatusCancelled,
		PaymentStatusRefunded,
	}
	return slices.Contains(finalStatuses, ps)
}

func (ps PaymentStatus) RequiresAction() bool {
	return ps == PaymentStatusPending ||
		ps == PaymentStatusOverdue ||
		ps == PaymentStatusDisputed ||
		ps == PaymentStatusPartiallyPaid
}

func (ps PaymentStatus) IsSuccessful() bool {
	return ps == PaymentStatusPaid || ps == PaymentStatusPartiallyPaid
}

// PaymentMethod represents the method used for payment
type PaymentMethod string

const (
	PaymentMethodCash          PaymentMethod = "cash"
	PaymentMethodCreditCard    PaymentMethod = "credit_card"
	PaymentMethodDebitCard     PaymentMethod = "debit_card"
	PaymentMethodBankTransfer  PaymentMethod = "bank_transfer"
	PaymentMethodPayPal        PaymentMethod = "paypal"
	PaymentMethodStripe        PaymentMethod = "stripe"
	PaymentMethodCheck         PaymentMethod = "check"
	PaymentMethodMobilePayment PaymentMethod = "mobile_payment"
	PaymentMethodInsurance     PaymentMethod = "insurance"
	PaymentMethodCareCredit    PaymentMethod = "care_credit"
)

// PaymentMethod constants and methods
var (
	ValidPaymentMethods = []PaymentMethod{
		PaymentMethodCash,
		PaymentMethodCreditCard,
		PaymentMethodDebitCard,
		PaymentMethodBankTransfer,
		PaymentMethodPayPal,
		PaymentMethodStripe,
		PaymentMethodCheck,
		PaymentMethodMobilePayment,
		PaymentMethodInsurance,
		PaymentMethodCareCredit,
	}

	paymentMethodMap = map[string]PaymentMethod{
		"cash":           PaymentMethodCash,
		"credit_card":    PaymentMethodCreditCard,
		"credit card":    PaymentMethodCreditCard,
		"credit":         PaymentMethodCreditCard,
		"debit_card":     PaymentMethodDebitCard,
		"debit card":     PaymentMethodDebitCard,
		"debit":          PaymentMethodDebitCard,
		"bank_transfer":  PaymentMethodBankTransfer,
		"bank transfer":  PaymentMethodBankTransfer,
		"transfer":       PaymentMethodBankTransfer,
		"paypal":         PaymentMethodPayPal,
		"stripe":         PaymentMethodStripe,
		"check":          PaymentMethodCheck,
		"cheque":         PaymentMethodCheck,
		"mobile_payment": PaymentMethodMobilePayment,
		"mobile payment": PaymentMethodMobilePayment,
		"mobile":         PaymentMethodMobilePayment,
		"insurance":      PaymentMethodInsurance,
		"care_credit":    PaymentMethodCareCredit,
		"care credit":    PaymentMethodCareCredit,
		"carecredit":     PaymentMethodCareCredit,
	}

	paymentMethodDisplayNames = map[PaymentMethod]string{
		PaymentMethodCash:          "Cash",
		PaymentMethodCreditCard:    "Credit Card",
		PaymentMethodDebitCard:     "Debit Card",
		PaymentMethodBankTransfer:  "Bank Transfer",
		PaymentMethodPayPal:        "PayPal",
		PaymentMethodStripe:        "Stripe",
		PaymentMethodCheck:         "Check",
		PaymentMethodMobilePayment: "Mobile Payment",
		PaymentMethodInsurance:     "Insurance",
		PaymentMethodCareCredit:    "CareCredit",
	}

	paymentMethodCategories = map[PaymentMethod]string{
		PaymentMethodCash:          "cash",
		PaymentMethodCreditCard:    "card",
		PaymentMethodDebitCard:     "card",
		PaymentMethodBankTransfer:  "electronic",
		PaymentMethodPayPal:        "electronic",
		PaymentMethodStripe:        "electronic",
		PaymentMethodCheck:         "check",
		PaymentMethodMobilePayment: "electronic",
		PaymentMethodInsurance:     "insurance",
		PaymentMethodCareCredit:    "financing",
	}
)

func (pm PaymentMethod) IsValid() bool {
	_, exists := paymentMethodMap[string(pm)]
	return exists
}

func ParsePaymentMethod(method string) (PaymentMethod, error) {
	normalized := shared.NormalizeInput(method)
	if val, exists := paymentMethodMap[normalized]; exists {
		return val, nil
	}
	return "", fmt.Errorf("invalid payment method: %s", method)
}

func MustParsePaymentMethod(method string) PaymentMethod {
	parsed, err := ParsePaymentMethod(method)
	if err != nil {
		panic(err)
	}
	return parsed
}

func (pm PaymentMethod) String() string {
	return string(pm)
}

func (pm PaymentMethod) DisplayName() string {
	if displayName, exists := paymentMethodDisplayNames[pm]; exists {
		return displayName
	}
	return "Unknown Method"
}

func (pm PaymentMethod) Values() []PaymentMethod {
	return ValidPaymentMethods
}

func (pm PaymentMethod) RequiresOnlineProcessing() bool {
	onlineMethods := []PaymentMethod{
		PaymentMethodCreditCard,
		PaymentMethodDebitCard,
		PaymentMethodPayPal,
		PaymentMethodStripe,
		PaymentMethodBankTransfer,
		PaymentMethodMobilePayment,
	}
	return slices.Contains(onlineMethods, pm)
}

func (pm PaymentMethod) IsCard() bool {
	return pm == PaymentMethodCreditCard || pm == PaymentMethodDebitCard
}

func (pm PaymentMethod) IsElectronic() bool {
	return pm.RequiresOnlineProcessing() ||
		pm == PaymentMethodBankTransfer ||
		pm == PaymentMethodMobilePayment
}

func (pm PaymentMethod) Category() string {
	if category, exists := paymentMethodCategories[pm]; exists {
		return category
	}
	return "other"
}

func (pm PaymentMethod) RequiresSignature() bool {
	return pm.IsCard() || pm == PaymentMethodCheck
}

func (pm PaymentMethod) IsInsurance() bool {
	return pm == PaymentMethodInsurance || pm == PaymentMethodCareCredit
}

func GetAllPaymentStatuses() []PaymentStatus {
	return ValidPaymentStatuses
}

func GetAllPaymentMethods() []PaymentMethod {
	return ValidPaymentMethods
}

func GetSuccessfulPaymentStatuses() []PaymentStatus {
	return []PaymentStatus{
		PaymentStatusPaid,
		PaymentStatusPartiallyPaid,
	}
}

func GetPendingPaymentStatuses() []PaymentStatus {
	return []PaymentStatus{
		PaymentStatusPending,
		PaymentStatusOverdue,
		PaymentStatusDisputed,
	}
}

func GetElectronicPaymentMethods() []PaymentMethod {
	return []PaymentMethod{
		PaymentMethodCreditCard,
		PaymentMethodDebitCard,
		PaymentMethodBankTransfer,
		PaymentMethodPayPal,
		PaymentMethodStripe,
		PaymentMethodMobilePayment,
	}
}

func GetManualPaymentMethods() []PaymentMethod {
	return []PaymentMethod{
		PaymentMethodCash,
		PaymentMethodCheck,
	}
}
