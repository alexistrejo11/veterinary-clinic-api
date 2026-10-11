package users

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// ============================================================================
// Email Value Object
// ============================================================================

// Email represents a validated email address
type Email struct {
	value string
}

// NewEmail creates a new Email value object with validation
func NewEmail(emailStr string) (Email, error) {
	ctx := context.Background()
	operation := "NewEmail"

	if emailStr == "" {
		return Email{}, EmptyEmailError(ctx, operation)
	}

	email := Email{value: strings.ToLower(emailStr)}
	if !email.isValid() {
		return Email{}, InvalidEmailFormatError(ctx, emailStr, operation)
	}

	return email, nil
}

// NewEmailNoErr creates an Email without validation (use with caution)
func NewEmailNoErr(emailStr string) Email {
	return Email{value: strings.ToLower(emailStr)}
}

func (e Email) isValid() bool {
	const emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(e.value)
}

// Value returns the email string value
func (e Email) Value() string {
	return e.value
}

// String returns the email string representation
func (e Email) String() string {
	return e.value
}

// ============================================================================
// PhoneNumber Value Object
// ============================================================================

// PhoneNumber represents a validated phone number
type PhoneNumber struct {
	value string
}

// NewPhoneNumber creates a new PhoneNumber value object with validation
func NewPhoneNumber(phone string) (PhoneNumber, error) {
	ctx := context.Background()
	operation := "NewPhoneNumber"

	if phone == "" {
		return PhoneNumber{}, EmptyPhoneNumberError(ctx, operation)
	}

	cleaned := regexp.MustCompile(`[^\d+]`).ReplaceAllString(phone, "")

	if len(cleaned) < 10 {
		return PhoneNumber{}, PhoneNumberTooShortError(ctx, operation)
	}

	return PhoneNumber{value: cleaned}, nil
}

// NewPhoneNumberNoErr creates a PhoneNumber without validation
func NewPhoneNumberNoErr(phone string) PhoneNumber {
	p, _ := NewPhoneNumber(phone)
	return p
}

// Value returns the phone number string value
func (p PhoneNumber) Value() string {
	return p.value
}

// String returns the phone number string representation
func (p PhoneNumber) String() string {
	return p.value
}

// ============================================================================
// PersonName Value Object
// ============================================================================

// PersonName represents a person's first and last name
type PersonName struct {
	FirstName string
	LastName  string
}

// NewPersonName creates a new PersonName value object with validation
func NewPersonName(firstName, lastName string) (PersonName, error) {
	ctx := context.Background()
	operation := "NewPersonName"

	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)

	if firstName == "" {
		return PersonName{}, EmptyFirstNameError(ctx, operation)
	}

	if lastName == "" {
		return PersonName{}, EmptyLastNameError(ctx, operation)
	}

	return PersonName{
		FirstName: firstName,
		LastName:  lastName,
	}, nil
}

// FullName returns the concatenated full name
func (n PersonName) FullName() string {
	return fmt.Sprintf("%s %s", n.FirstName, n.LastName)
}
