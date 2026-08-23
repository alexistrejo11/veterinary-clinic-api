package auth

import (
	"context"
	"fmt"

	"clinic-vet-api/internal/shared"
	"clinic-vet-api/internal/shared/errors"
)

const (
	entityAuth    = "auth"
	entitySession = "session"
)

// Field names for validation errors
const (
	FieldEmail     = "email"
	FieldPassword  = "password"
	FieldToken     = "token"
	FieldCode      = "code"
	FieldTwoFactor = "two_factor_code"
)

// =========================================================================
// Authentication / credentials errors
// =========================================================================

// InvalidCredentialsError returns a generic error for wrong email or password (no user leak)
func InvalidCredentialsError(ctx context.Context, operation string) error {
	return errors.BusinessRuleError(ctx, "invalid_credentials",
		entityAuth, "email or password is incorrect", operation)
}

// UserNotFoundForLoginError returns when user is not found (use same message as invalid credentials for security)
func UserNotFoundForLoginError(ctx context.Context, operation string) error {
	return InvalidCredentialsError(ctx, operation)
}

// PasswordMismatchError when current password is wrong (e.g. change password)
func PasswordMismatchError(ctx context.Context, operation string) error {
	return errors.ValidationError(ctx, entityAuth, FieldPassword,
		"current password is incorrect", operation)
}

// =========================================================================
// Token errors
// =========================================================================

// InvalidTokenError when token is malformed or invalid
func InvalidTokenError(ctx context.Context, operation string) error {
	return errors.ValidationError(ctx, entityAuth, FieldToken,
		"token is invalid or expired", operation)
}

// TokenExpiredError when token has expired
func TokenExpiredError(ctx context.Context, operation string) error {
	return errors.BusinessRuleError(ctx, "token_expired",
		entityAuth, "token has expired", operation)
}

// TokenRevokedError when token was revoked (e.g. logout)
func TokenRevokedError(ctx context.Context, operation string) error {
	return errors.BusinessRuleError(ctx, "token_revoked",
		entityAuth, "token has been revoked", operation)
}

// RefreshTokenRequiredError when refresh token is missing or wrong type
func RefreshTokenRequiredError(ctx context.Context, operation string) error {
	return errors.ValidationError(ctx, entityAuth, FieldToken,
		"valid refresh token is required", operation)
}

// =========================================================================
// Two-factor errors
// =========================================================================

// TwoFactorRequiredError when login requires 2FA step
func TwoFactorRequiredError(ctx context.Context, operation string) error {
	return errors.BusinessRuleError(ctx, "2fa_required",
		entityAuth, "two-factor authentication is required to complete login", operation)
}

// TwoFactorInvalidCodeError when 2FA code is wrong
func TwoFactorInvalidCodeError(ctx context.Context, operation string) error {
	return errors.ValidationError(ctx, entityAuth, FieldTwoFactor,
		"two-factor code is invalid or expired", operation)
}

// TwoFactorAlreadyEnabledError when user tries to enable 2FA again
func TwoFactorAlreadyEnabledError(ctx context.Context, operation string) error {
	return errors.BusinessRuleError(ctx, "2fa_already_enabled",
		entityAuth, "two-factor authentication is already enabled", operation)
}

// TwoFactorNotEnabledError when user tries to disable or verify 2FA but it is off
func TwoFactorNotEnabledError(ctx context.Context, operation string) error {
	return errors.BusinessRuleError(ctx, "2fa_not_enabled",
		entityAuth, "two-factor authentication is not enabled", operation)
}

// ErrRequiresTwoFactor is returned by Login when credentials are valid but 2FA must be completed.
// HTTP layer can use errors.As to respond with the opaque pre-authentication challenge.
var ErrRequiresTwoFactor = &RequiresTwoFactorError{}

// RequiresTwoFactorError holds user ID for the 2FA step
type RequiresTwoFactorError struct {
	ChallengeToken string
}

func (e *RequiresTwoFactorError) Error() string {
	return "two-factor authentication is required to complete login"
}

// =========================================================================
// Session / account state errors
// =========================================================================

// UserNotActiveError when account is not active (e.g. pending activation)
func UserNotActiveError(ctx context.Context, operation string) error {
	return errors.BusinessRuleError(ctx, "user_not_active",
		entityAuth, "user account is not active", operation)
}

// UserBannedError when account is banned
func UserBannedError(ctx context.Context, operation string) error {
	return errors.BusinessRuleError(ctx, "user_banned",
		entityAuth, "user account is banned", operation)
}

// AccountLockedError when too many failed attempts
func AccountLockedError(ctx context.Context, operation string) error {
	return errors.BusinessRuleError(ctx, "account_locked",
		entityAuth, "account is temporarily locked due to failed attempts", operation)
}

// =========================================================================
// Verification / email / password reset
// =========================================================================

// InvalidVerificationCodeError for email or password reset codes
func InvalidVerificationCodeError(ctx context.Context, operation string) error {
	return errors.ValidationError(ctx, entityAuth, FieldCode,
		"verification code is invalid or expired", operation)
}

// EmailAlreadyVerifiedError when user tries to verify again
func EmailAlreadyVerifiedError(ctx context.Context, operation string) error {
	return errors.BusinessRuleError(ctx, "email_already_verified",
		entityAuth, "email is already verified", operation)
}

// =========================================================================
// Registration errors
// =========================================================================

// InvalidRoleError when registration role is not allowed
func InvalidRoleError(ctx context.Context, role string, operation string) error {
	return errors.ValidationError(ctx, entityAuth, "role",
		fmt.Sprintf("invalid role for registration: %s", role), operation)
}

// EmailAlreadyExistsError when registering with existing email
func EmailAlreadyExistsAuthError(ctx context.Context, email string, operation string) error {
	return errors.BusinessRuleError(ctx, "email_already_exists",
		entityAuth, fmt.Sprintf("email %s is already in use", email), operation)
}

// EmployeeDataRequiredError when registering as employee without ID/license
func EmployeeDataRequiredError(ctx context.Context, operation string) error {
	return errors.ValidationError(ctx, entityAuth, "employee_id",
		"employee ID and license number are required for employee registration", operation)
}

// =========================================================================
// Not found / unauthorized (use shared where applicable)
// =========================================================================

// SessionNotFoundError when session or refresh token is not found
func SessionNotFoundError(ctx context.Context, operation string) error {
	return errors.EntityNotFoundError(ctx, entitySession, "", operation)
}

// UnauthorizedError creates unauthorized access error (delegate to shared for consistency)
func UnauthorizedError(ctx context.Context, operation, resource string) error {
	return errors.UnauthorizedError(ctx, operation, resource)
}

// ForbiddenError creates forbidden access error
func ForbiddenError(ctx context.Context, operation, resource, reason string) error {
	return errors.ForbiddenError(ctx, operation, resource, reason)
}

// UserNotFoundError when user is not found by ID (e.g. for 2FA or password)
func UserNotFoundError(ctx context.Context, userID shared.UserID, operation string) error {
	return errors.EntityNotFoundError(ctx, "user", userID.String(), operation)
}
