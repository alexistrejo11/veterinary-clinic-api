package dtos

// Auth request/response DTOs with validation. Use ShouldBindBodyWithJSON for body, ShouldBindQuery for query.

const (
	maxLenEmail     = 255
	maxLenPassword  = 128
	maxLenCode      = 32
	maxLenRole      = 50
	maxLenName      = 200
	maxLenPhone     = 30
	maxLenBio       = 500
	maxLenDevice    = 200
	maxLenUserAgent = 512
	maxLenMethod    = 20
)

// RegisterRequest is the body for POST /auth/register
// @Description Request body for public customer registration.
type RegisterRequest struct {
	Role          string  `json:"role" binding:"required,eq=customer" example:"customer"`
	Email         string  `json:"email" binding:"required,email,max=255" example:"user@example.com"`
	Password      string  `json:"password" binding:"required,min=8,max=128" example:"SecurePass123!"`
	FirstName     string  `json:"first_name" binding:"required,max=100"`
	LastName      string  `json:"last_name" binding:"required,max=100"`
	PhoneNumber   string  `json:"phone_number" binding:"omitempty,max=30"`
	DateOfBirth   string  `json:"date_of_birth" binding:"required,datetime=2006-01-02"`
	Gender        string  `json:"gender" binding:"required,oneof=male female other not_specified"`
	PhotoURL      string  `json:"photo_url" binding:"omitempty,max=500"`
	EmployeeID    *uint   `json:"employee_id,omitempty"`
	LicenseNumber *string `json:"license_number,omitempty" binding:"omitempty,max=100"`
}

// ActivateAccountRequest for GET /auth/activate (query) or POST body
// @Description User ID and activation code (e.g. from email link). Supports query params or JSON body.
type ActivateAccountRequest struct {
	UserID string `form:"user_id" json:"user_id" binding:"required,max=20"`
	Code   string `form:"code" json:"code" binding:"required,max=32"`
}

// LoginRequest is the body for POST /auth/login
// @Description Email, password, and optional two_factor_code when 2FA is enabled.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,max=128"`
}

// RefreshTokenRequest is the body for POST /auth/refresh
// @Description Refresh token string to exchange for new access and refresh tokens.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required,max=2000"`
}

// LogoutRequest is the body for POST /auth/logout
// @Description Refresh token to revoke for the current session.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required,max=2000"`
}

// VerifyTwoFactorRequest is the body for POST /auth/2fa/verify
// @Description One-time 2FA code (TOTP or backup code).
type VerifyTwoFactorRequest struct {
	ChallengeToken string `json:"challenge_token" binding:"required,max=64"`
	TwoFactorCode  string `json:"two_factor_code" binding:"required,max=16"`
}

// EnableTwoFactorRequest is the body for POST /auth/2fa/enable
// @Description 2FA method: totp.
type EnableTwoFactorRequest struct {
	Method string `json:"method" binding:"required,eq=totp"`
}

// VerifyEmailRequest is the body for POST /auth/verify-email
// @Description Email and verification code.
type VerifyEmailRequest struct {
	Email string `json:"email" binding:"required,email,max=255"`
	Code  string `json:"code" binding:"required,max=32"`
}

// RequestResetPasswordRequest is the body for POST /auth/reset-password/request
// @Description Email address to send the password reset link/code to.
type RequestResetPasswordRequest struct {
	Email string `json:"email" binding:"required,email,max=255"`
}

// ResetPasswordRequest is the body for POST /auth/reset-password/reset
// @Description Email, code received by email, and new password.
type ResetPasswordRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Code     string `json:"code" binding:"required,max=32"`
	Password string `json:"password" binding:"required,min=8,max=128"`
}

// UpdatePasswordRequest is the body for POST /auth/update-password
// @Description Current password and new password. Requires Bearer auth.
type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required,max=128"`
	NewPassword     string `json:"new_password" binding:"required,min=8,max=128"`
}

// SessionResponse returned on login, refresh, and verify-2fa
// @Description Contains access token, refresh token, and minimal user info.
type SessionResponse struct {
	AccessToken  TokenResponse    `json:"access_token"`
	RefreshToken TokenResponse    `json:"refresh_token"`
	User         UserAuthResponse `json:"user"`
}

// TokenResponse for access/refresh in responses
// @Description JWT token string, expiry time, and type (e.g. Bearer).
type TokenResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
	Type      string `json:"type"`
}

// UserAuthResponse minimal user info for auth responses
// @Description User identifier, email, and role returned in login/session payloads.
type UserAuthResponse struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// RequiresTwoFactorResponse when login needs 2FA step
// @Description Returned when login succeeds but 2FA verification is required. Client should prompt for code and call POST /auth/2fa/verify.
type RequiresTwoFactorResponse struct {
	RequiresTwoFactor bool   `json:"requires_two_factor"`
	ChallengeToken    string `json:"challenge_token"`
	Message           string `json:"message"`
}

// EnableTwoFactorResponse contains recovery codes shown only once at enrollment.
type EnableTwoFactorResponse struct {
	BackupCodes []string `json:"backup_codes"`
}
