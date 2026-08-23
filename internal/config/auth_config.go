package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"time"
)

type AuthConfig struct {
	JWTSecret              string        `json:"jwt_secret"`
	JWTExpirationTime      time.Duration `json:"jwt_expiration_time"`
	RefreshTokenExpiry     time.Duration `json:"refresh_token_expiry"`
	PasswordMinLength      int           `json:"password_min_length"`
	MaxLoginAttempts       int           `json:"max_login_attempts"`
	LockoutDuration        time.Duration `json:"lockout_duration"`
	TwoFactorEncryptionKey []byte        `json:"-"`
}

func loadAuthConfig(config *AuthConfig) error {
	config.JWTSecret = os.Getenv("JWT_SECRET")
	if config.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}

	twoFactorKey, err := base64.StdEncoding.DecodeString(os.Getenv("TWO_FACTOR_ENCRYPTION_KEY"))
	if err != nil || len(twoFactorKey) != 32 {
		return fmt.Errorf("TWO_FACTOR_ENCRYPTION_KEY must be base64-encoded 32 bytes")
	}
	config.TwoFactorEncryptionKey = twoFactorKey

	config.JWTExpirationTime, err = parseDuration("JWT_EXPIRATION_TIME", "24h")
	if err != nil {
		return fmt.Errorf("invalid JWT_EXPIRATION_TIME: %w", err)
	}

	config.RefreshTokenExpiry, err = parseDuration("REFRESH_TOKEN_EXPIRY", "168h") // 7 days
	if err != nil {
		return fmt.Errorf("invalid REFRESH_TOKEN_EXPIRY: %w", err)
	}

	config.PasswordMinLength, err = parseIntWithDefault("PASSWORD_MIN_LENGTH", 8)
	if err != nil {
		return fmt.Errorf("invalid PASSWORD_MIN_LENGTH: %w", err)
	}

	config.MaxLoginAttempts, err = parseIntWithDefault("MAX_LOGIN_ATTEMPTS", 5)
	if err != nil {
		return fmt.Errorf("invalid MAX_LOGIN_ATTEMPTS: %w", err)
	}

	config.LockoutDuration, err = parseDuration("LOCKOUT_DURATION", "30m")
	if err != nil {
		return fmt.Errorf("invalid LOCKOUT_DURATION: %w", err)
	}

	return nil
}
