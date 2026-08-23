// Package auth contains the authentication related entities like Session and TwoFaAuth
package auth

import (
	"time"

	"clinic-vet-api/internal/core/users"
)

// JWT claim keys used by TokenService implementations (e.g. JWT)
const (
	ClaimUserID     = "user_id"
	ClaimEmail      = "email"
	ClaimRole       = "role"
	ClaimTokenType  = "token_type"
	ClaimExp        = "exp"
	ClaimIat        = "iat"
	ClaimJTI        = "jti"
	ClaimDeviceInfo = "device_info"
	ClaimUserAgent  = "user_agent"
	ClaimIPAddress  = "ip_address"
)

// JWTAccessClaims holds standard claims for access tokens (role, email for authorization)
type JWTAccessClaims struct {
	UserID string         `json:"user_id"`
	Email  string         `json:"email"`
	Role   users.UserRole `json:"role"`
	Exp    int64          `json:"exp"`
	Iat    int64          `json:"iat"`
	JTI    string         `json:"jti,omitempty"`
}

// JWTRefreshClaims holds minimal claims for refresh tokens
type JWTRefreshClaims struct {
	UserID string `json:"user_id"`
	JTI    string `json:"jti"`
	Exp    int64  `json:"exp"`
	Iat    int64  `json:"iat"`
}

type JwtSession struct {
	UserID       string
	RefreshToken string
	DeviceInfo   string
	UserAgent    string
	IPAddress    string
	ExpiresAt    time.Time
	CreatedAt    time.Time
}

type TokenType string

const (
	TokenTypeAccessToken        TokenType = "access_token"
	TokenTypeRefreshToken       TokenType = "refresh_token"
	TokenTypeVerification       TokenType = "verification"
	TokenTypePasswordReset      TokenType = "password_reset"
	TokenTypeTwoFactorChallenge TokenType = "two_factor_challenge"
	TokenJwtAccessToken         TokenType = "jwt_access_token"
	TokenJwtRefreshToken        TokenType = "jwt_refresh_token"
)

type Token struct {
	Code      string
	Type      TokenType
	ExpiresAt time.Time
	CreatedAt time.Time
}
