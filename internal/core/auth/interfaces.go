package auth

import (
	"context"

	"clinic-vet-api/internal/shared"
)

// PasswordChecker verifies a plain password against a hash. Used by auth service for login and password change.
// Infrastructure can provide shared/password.PasswordEncoder which implements this.
type PasswordChecker interface {
	CheckPassword(hashedPassword, plainPassword string) bool
}

// TwoFactorProvider handles 2FA verification and setup. Implementation is in infrastructure (e.g. TOTP, SMS).
// Auth core only depends on this interface.
type TwoFactorProvider interface {
	// VerifyCode validates the 2FA code for the user (e.g. TOTP or backup code).
	// Returns true if valid, false otherwise. Error for unexpected failures.
	VerifyCode(ctx context.Context, userID shared.UserID, code string) (valid bool, err error)

	// GenerateSecret generates a new 2FA secret and backup codes for enabling 2FA.
	// Returns secret (e.g. TOTP secret), backup codes, and optional QR/setup URI.
	GenerateSecret(ctx context.Context, userID shared.UserID, method string) (secret string, backupCodes []string, err error)
	HashBackupCodes(codes []string) ([]string, error)

	// Disable clears 2FA state for the user (e.g. invalidate backup codes server-side if stored).
	Disable(ctx context.Context, userID shared.UserID) error
}
