// Package users defines the User entity and its related functionalities.
package users

import (
	"context"
	"errors"
	"time"

	"clinic-vet-api/internal/core/customers"
	"clinic-vet-api/internal/core/employees"
	"clinic-vet-api/internal/shared"
)

// ============================================================================
// Type Definitions
// ============================================================================

// User represents a user in the system
type User struct {
	shared.Entity[shared.UserID]
	Email          Email
	PhoneNumber    PhoneNumber
	HashedPassword string
	Role           UserRole
	Status         UserStatus
	LastLoginAt    *time.Time
	TwoFactorAuth  shared.TwoFactorAuth
	EmployeeID     *employees.EmployeeID
	CustomerID     *customers.CustomerID
	Profile        Profile

	// OAuth2 fields
	OAuthProvider   string
	OAuthProviderID *string
	OAuthToken      *OAuthToken
	EmailVerified   bool

	// Security fields
	LoginAttempts int
	LockedUntil   *time.Time
}

// OAuthToken holds OAuth2 access and refresh tokens
type OAuthToken struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    *time.Time
}

// Profile represents user profile information
type Profile struct {
	Name        string
	Gender      shared.PersonGender
	PhotoURL    string
	Bio         string
	DateOfBirth *time.Time
}

func (p Profile) Validate(ctx context.Context) error {
	if p.Name == "" {
		return InvalidProfileError(ctx, "Name is required")
	}

	if !p.Gender.IsValid() {
		return InvalidProfileError(ctx, "Invalid gender value")
	}

	if p.DateOfBirth != nil && p.DateOfBirth.After(time.Now()) {
		return InvalidProfileError(ctx, "Date of birth cannot be in the future")
	}

	if p.PhotoURL == "" {
		return InvalidProfileError(ctx, "Photo URL is required")
	}

	if len(p.Bio) > 500 {
		return InvalidProfileError(ctx, "Bio cannot exceed 500 characters")
	}

	return nil
}

// ============================================================================
// User Methods
// ============================================================================

// UpdateEmail updates the user's email address
func (u *User) UpdateEmail(newEmail Email) error {
	if u.Email.String() == newEmail.String() {
		return nil // No change needed
	}
	u.Email = newEmail
	u.IncrementVersion()
	return nil
}

// UpdatePhoneNumber updates the user's phone number
func (u *User) UpdatePhoneNumber(newPhone PhoneNumber) error {
	u.PhoneNumber = newPhone
	u.IncrementVersion()
	return nil
}

// UpdatePassword updates the user's hashed password
func (u *User) UpdatePassword(newPasswordHash string) error {
	u.HashedPassword = newPasswordHash
	u.IncrementVersion()
	return nil
}

func (u *User) UpdateProfile(profile Profile) error {
	hasChanges := false
	if u.Profile.Name != "" || u.Profile.Name != profile.Name {
		u.Profile.Name = profile.Name
		hasChanges = true
	}

	if u.Profile.Gender != profile.Gender {
		u.Profile.Gender = profile.Gender
		hasChanges = true
	}

	if u.Profile.PhotoURL != profile.PhotoURL {
		u.Profile.PhotoURL = profile.PhotoURL
		hasChanges = true
	}
	if u.Profile.Bio != profile.Bio {
		u.Profile.Bio = profile.Bio
		hasChanges = true
	}
	if u.Profile.DateOfBirth != profile.DateOfBirth {
		u.Profile.DateOfBirth = profile.DateOfBirth
		hasChanges = true
	}

	if hasChanges {
		u.IncrementVersion()
	}

	return nil
}

// Activate activates the user account
func (u *User) Activate() error {
	ctx := context.Background()
	operation := "ActivateUser"

	if u.Status == UserStatusBanned {
		return CannotActivateBannedUserError(ctx, operation)
	}
	u.Status = UserStatusActive
	u.IncrementVersion()
	return nil
}

// Deactivate deactivates the user account
func (u *User) Deactivate() error {
	u.Status = UserStatusInactive
	u.IncrementVersion()
	return nil
}

// Ban bans the user account
func (u *User) Ban() error {
	u.Status = UserStatusBanned
	u.IncrementVersion()
	return nil
}

// RecordLogin records the user's last login time
func (u *User) RecordLogin() error {
	now := time.Now()
	u.LastLoginAt = &now
	u.IncrementVersion()
	return nil
}

// EnableTwoFactorAuth enables two-factor authentication for the user
func (u *User) EnableTwoFactorAuth(method, secret string, backupCodes []string) error {
	now := time.Now()
	twoFAMethod := shared.TwoFactorMethod(method)
	if !twoFAMethod.IsValid() || twoFAMethod == shared.TwoFactorMethodBackup {
		return errors.New("invalid two-factor authentication method")
	}

	twoFA := shared.NewTwoFactorAuth(
		true,
		twoFAMethod,
		secret,
		backupCodes,
		nil,
		&now,
	)
	u.TwoFactorAuth = twoFA
	u.IncrementVersion()
	return nil
}

// DisableTwoFactorAuth disables two-factor authentication for the user
func (u *User) DisableTwoFactorAuth() error {
	u.TwoFactorAuth = shared.NewDisabledTwoFactorAuth()
	u.IncrementVersion()
	return nil
}

// ValidateLogin validates if the user can login
func (u *User) ValidateLogin(is2FA bool) error {
	ctx := context.Background()
	operation := "ValidateLogin"

	if u.Status != UserStatusActive {
		return UserNotActiveError(ctx, operation)
	}

	if u.Status == UserStatusBanned {
		return UserBannedError(ctx, operation)
	}

	if is2FA && !u.TwoFactorAuth.IsEnabled {
		return TwoFactorRequiredError(ctx, operation)
	}

	return nil
}

// AssignEmployee assigns an employee to this user account
func (u *User) AssignEmployee(employeeID employees.EmployeeID) error {
	ctx := context.Background()
	operation := "AssignEmployee"

	if u.EmployeeID != nil {
		return EmployeeAlreadyAssignedError(ctx, operation)
	}

	if u.CustomerID != nil {
		return CannotAssignEmployeeToCustomerError(ctx, operation)
	}

	u.EmployeeID = &employeeID
	u.IncrementVersion()
	return nil
}
