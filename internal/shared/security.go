package shared

import (
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"strings"
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

// GenerateNewTwoFactorAuth generates a new TOTP 2FA setup
func GenerateNewTwoFactorAuth() (TwoFactorAuth, error) {
	secret, err := generateTOTPSecret()
	if err != nil {
		return NewDisabledTwoFactorAuth(), err
	}

	backupCodes := generateBackupCodes(8)
	now := time.Now()

	return TwoFactorAuth{
		IsEnabled:   true,
		Method:      TwoFactorMethodTOTP,
		Secret:      secret,
		BackupCodes: backupCodes,
		EnabledAt:   &now,
		LastUsedAt:  nil,
	}, nil
}

// ValidateCode validates a 2FA code
func (t TwoFactorAuth) ValidateCode(code string) (bool, error) {
	if !t.IsEnabled {
		return false, errors.New("2FA is not enabled")
	}

	if code == "" {
		return false, errors.New("code cannot be empty")
	}

	// Check if it's a backup code
	if t.isBackupCode(code) {
		return true, t.consumeBackupCode(code)
	}

	// Validate based on method
	switch t.Method {
	case TwoFactorMethodTOTP:
		return t.validateTOTP(code)
	case TwoFactorMethodSMS:
		return t.validateSMS(code)
	case TwoFactorMethodEmail:
		return t.validateEmail(code)
	default:
		return false, fmt.Errorf("unsupported 2FA method: %s", t.Method)
	}
}

// IsBackupCode checks if the code is a valid backup code
func (t TwoFactorAuth) IsBackupCode(code string) bool {
	return t.isBackupCode(code)
}

// HasBackupCodesRemaining checks if there are backup codes available
func (t TwoFactorAuth) HasBackupCodesRemaining() bool {
	return len(t.BackupCodes) > 0
}

// RemainingBackupCodes returns the number of remaining backup codes
func (t TwoFactorAuth) RemainingBackupCodes() int {
	return len(t.BackupCodes)
}

// GenerateNewBackupCodes generates new backup codes and replaces old ones
func (t *TwoFactorAuth) GenerateNewBackupCodes(count int) error {
	if !t.IsEnabled {
		return errors.New("cannot generate backup codes for disabled 2FA")
	}

	if count < 1 || count > 20 {
		return errors.New("backup code count must be between 1 and 20")
	}

	t.BackupCodes = generateBackupCodes(count)
	return nil
}

// Disable disables 2FA and clears all sensitive data
func (t *TwoFactorAuth) Disable() {
	t.IsEnabled = false
	t.Method = TwoFactorMethodUnknown
	t.Secret = ""
	t.BackupCodes = nil
	t.EnabledAt = nil
	t.LastUsedAt = nil
}

// RecordUsage records when 2FA was last used
func (t *TwoFactorAuth) RecordUsage() {
	now := time.Now()
	t.LastUsedAt = &now
}

// Private helper methods

func (t TwoFactorAuth) isBackupCode(code string) bool {
	for _, backupCode := range t.BackupCodes {
		if backupCode == code {
			return true
		}
	}
	return false
}

func (t *TwoFactorAuth) consumeBackupCode(code string) error {
	for i, backupCode := range t.BackupCodes {
		if backupCode == code {
			// Remove the used backup code
			t.BackupCodes = append(t.BackupCodes[:i], t.BackupCodes[i+1:]...)
			return nil
		}
	}
	return errors.New("backup code not found")
}

func (t TwoFactorAuth) validateTOTP(code string) (bool, error) {
	// In a real implementation, you would use a TOTP library like:
	// github.com/pquerna/otp/totp
	// This is a simplified version

	if len(code) != 6 {
		return false, errors.New("TOTP code must be 6 digits")
	}

	// Basic validation - in real implementation, use proper TOTP validation
	// For now, just return true for demonstration
	// actual validation would be: totp.Validate(code, t.Secret)

	return true, nil
}

func (t TwoFactorAuth) validateSMS(code string) (bool, error) {
	// SMS code validation logic
	if len(code) != 6 {
		return false, errors.New("SMS code must be 6 digits")
	}

	// In real implementation, you would verify against stored/expected code
	return true, nil
}

func (t TwoFactorAuth) validateEmail(code string) (bool, error) {
	// Email code validation logic
	if len(code) != 6 {
		return false, errors.New("email code must be 6 digits")
	}

	// In real implementation, you would verify against stored/expected code
	return true, nil
}

// Utility functions

func generateTOTPSecret() (string, error) {
	// Generate 20 random bytes for TOTP secret
	bytes := make([]byte, 20)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate TOTP secret: %w", err)
	}

	// Base32 encode without padding
	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes)
	return secret, nil
}

func generateBackupCodes(count int) []string {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		// Generate 8-character backup codes in format XXXX-XXXX
		part1 := randomString(4, "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
		part2 := randomString(4, "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
		codes[i] = fmt.Sprintf("%s-%s", part1, part2)
	}
	return codes
}

func randomString(length int, charset string) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		panic(err) // Should not happen in practice
	}

	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}

// TwoFactorMethod utilities

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

func ParseTwoFactorMethod(method string) (TwoFactorMethod, error) {
	normalized := strings.ToLower(strings.TrimSpace(method))
	switch normalized {
	case "totp":
		return TwoFactorMethodTOTP, nil
	case "sms":
		return TwoFactorMethodSMS, nil
	case "email":
		return TwoFactorMethodEmail, nil
	case "backup":
		return TwoFactorMethodBackup, nil
	default:
		return TwoFactorMethodUnknown, fmt.Errorf("invalid 2FA method: %s", method)
	}
}

func (t TwoFactorAuth) Validate() error {
	if t.IsEnabled {
		if !t.Method.IsValid() {
			return errors.New("invalid 2FA method when enabled")
		}
		if t.Secret == "" && t.Method != TwoFactorMethodBackup {
			return errors.New("secret is required for non-backup methods")
		}
		if t.EnabledAt == nil {
			return errors.New("enabledAt timestamp is required when enabled")
		}
		if t.EnabledAt.After(time.Now()) {
			return errors.New("enabledAt cannot be in the future")
		}
		if t.LastUsedAt != nil && t.LastUsedAt.After(time.Now()) {
			return errors.New("lastUsedAt cannot be in the future")
		}
	}
	return nil
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
