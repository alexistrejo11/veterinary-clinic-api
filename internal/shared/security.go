package shared

import (
	"errors"
	"time"
)

// Shared functions for security that some package may need

// TwoFactorAuth represents two-factor authentication settings for a user
type TwoFactorAuth struct {
	IsEnabled   bool
	Method      TwoFactorMethod
	Secret      string
	BackupCodes []string
	EnabledAt   *time.Time
	LastUsedAt  *time.Time
}

// TwoFactorMethod represents the type of 2FA method
type TwoFactorMethod string

const (
	TwoFactorMethodTOTP    TwoFactorMethod = "totp"   // Time-based OTP
	TwoFactorMethodSMS     TwoFactorMethod = "sms"    // SMS codes
	TwoFactorMethodEmail   TwoFactorMethod = "email"  // Email codes
	TwoFactorMethodBackup  TwoFactorMethod = "backup" // Backup codes
	TwoFactorMethodUnknown TwoFactorMethod = "unknown"
)

// NewTwoFactorAuth creates a new TwoFactorAuth value object
func NewTwoFactorAuth(enabled bool, method TwoFactorMethod, secret string,
	backupCodes []string, enabledAt, lastUsedAt *time.Time,
) TwoFactorAuth {
	return TwoFactorAuth{
		IsEnabled:   enabled,
		Method:      method,
		Secret:      secret,
		BackupCodes: backupCodes,
		EnabledAt:   enabledAt,
		LastUsedAt:  lastUsedAt,
	}
}

// NewDisabledTwoFactorAuth creates a disabled 2FA instance
func NewDisabledTwoFactorAuth() TwoFactorAuth {
	return TwoFactorAuth{
		IsEnabled:   false,
		Method:      TwoFactorMethodUnknown,
		Secret:      "",
		BackupCodes: nil,
		EnabledAt:   nil,
		LastUsedAt:  nil,
	}
}

// RecordUsage records when 2FA was last used
func (t *TwoFactorAuth) RecordUsage() {
	now := time.Now()
	t.LastUsedAt = &now
}

func (m TwoFactorMethod) IsValid() bool {
	switch m {
	case TwoFactorMethodTOTP, TwoFactorMethodSMS, TwoFactorMethodEmail, TwoFactorMethodBackup:
		return true
	default:
		return false
	}
}

func (m TwoFactorMethod) String() string {
	return string(m)
}

func ValidatePasswordStrength(plainPassword string) error {
	if len(plainPassword) < 8 || len(plainPassword) > 128 {
		return errors.New("password must be between 8 and 128 characters long")
	}

	var hasLower, hasUpper, hasDigit, hasSpecial bool
	for _, character := range plainPassword {
		switch {
		case character >= 'a' && character <= 'z':
			hasLower = true
		case character >= 'A' && character <= 'Z':
			hasUpper = true
		case character >= '0' && character <= '9':
			hasDigit = true
		case character == '@' || character == '$' || character == '!' || character == '%' || character == '*' || character == '?' || character == '&':
			hasSpecial = true
		default:
			return errors.New("password contains unsupported characters")
		}
	}
	if !hasLower || !hasUpper || !hasDigit || !hasSpecial {
		return errors.New("password must contain at least one uppercase letter, one lowercase letter, one number, and one special character")
	}

	return nil
}
