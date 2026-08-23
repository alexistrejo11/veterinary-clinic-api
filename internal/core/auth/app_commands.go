package auth

import (
	"clinic-vet-api/internal/core/users"
	"clinic-vet-api/internal/shared"
)

type LoginCommand struct {
	Email      string
	Password   string
	DeviceInfo string
	UserAgent  string
	IPAddress  string

	TwoFactorCode *string
}

type RefreshTokenCommand struct {
	RefreshToken string
}

type LogoutCommand struct {
	RefreshToken string
	UserID       shared.UserID
}

type VerifyTwoFactorCommand struct {
	ChallengeToken string
	TwoFactorCode  string
}

type VerifyEmailCommand struct {
	Email string
	Code  string
}

type ResetPasswordCommand struct {
	Email    string
	Code     string
	Password string
}

type UpdatePasswordCommand struct {
	UserID          shared.UserID
	CurrentPassword string
	NewPassword     string
}

type RegisterCommand struct {
	Role     users.UserRole
	Email    string
	Password string

	// Personal information
	FirstName   string
	LastName    string
	PhoneNumber string
	DateOfBirth string
	Gender      string
	PhotoURL    string

	// Employee credentials
	EmployeeID    *uint
	LicenseNumber *string
}

func (c *RegisterCommand) HasEmployeeData() bool {
	return c.EmployeeID != nil && c.LicenseNumber != nil

}

type RequestResetPasswordCommand struct {
	Email string
}

type LogoutAllCommand struct {
	UserID shared.UserID
}

type DeleteAccountCommand struct {
	UserID shared.UserID
}

type ActivateAccountCommand struct {
	UserID shared.UserID
	Code   string
}

type EnableTwoFactorCommand struct {
	UserID shared.UserID
	Method string
}

type DisableTwoFactorCommand struct {
	UserID shared.UserID
}
