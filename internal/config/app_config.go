// Package config handles application configuration loading and validation
package config

import (
	"clinic-vet-api/database/sqlc"
	"clinic-vet-api/internal/infrastructure/http"
	"clinic-vet-api/internal/infrastructure/notification"
	"clinic-vet-api/internal/middleware"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

// AppSettings holds all application configuration
type AppSettings struct {
	// Server Configuration
	Server ServerConfig `json:"server"`

	// Database Configuration
	Database DatabaseConfig `json:"database"`

	// Redis Configuration
	Redis RedisConfig `json:"redis"`

	// Authentication Configuration
	Auth AuthConfig `json:"auth"`

	// External Services Configuration
	Services ServicesConfig `json:"services"`

	// Rate Limiting Configuration
	RateLimit RateLimitConfig `json:"rate_limit"`

	// CORS Configuration
	CORS CORSConfig `json:"cors"`

	// Application Configuration
	App AppConfig `json:"app"`
}

type ServerConfig struct {
	Port         string        `json:"port"`
	Host         string        `json:"host"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	Environment  string        `json:"environment"`
}

type ServicesConfig struct {
	Twilio TwilioConfig `json:"twilio"`
	Email  EmailConfig  `json:"email"`
}

type AppConfig struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	Debug         bool   `json:"debug"`
	LogLevel      string `json:"log_level"`
	EnableSwagger bool   `json:"enable_swagger"`
}

// LoadSettings loads and validates all application settings
func LoadSettings() (*AppSettings, error) {
	// Load .env when present (Compose injects env vars; the file is not in the image).
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		fmt.Printf("Warning: Could not load .env file: %v\n", err)
	}

	settings := &AppSettings{}

	// Load and validate each configuration section
	if err := loadServerConfig(&settings.Server); err != nil {
		return nil, fmt.Errorf("server config error: %w", err)
	}

	if err := loadDatabaseConfig(&settings.Database); err != nil {
		return nil, fmt.Errorf("database config error: %w", err)
	}

	if err := loadRedisConfig(&settings.Redis); err != nil {
		return nil, fmt.Errorf("redis config error: %w", err)
	}

	if err := loadAuthConfig(&settings.Auth); err != nil {
		return nil, fmt.Errorf("auth config error: %w", err)
	}

	if err := loadServicesConfig(&settings.Services); err != nil {
		return nil, fmt.Errorf("services config error: %w", err)
	}

	loadRateLimitConfig(&settings.RateLimit)
	loadCORSConfig(&settings.CORS)
	loadAppConfig(&settings.App)

	return settings, nil
}

func loadServerConfig(config *ServerConfig) error {
	config.Port = os.Getenv("SERVER_PORT")
	if config.Port == "" {
		return fmt.Errorf("SERVER_PORT is required")
	}
	config.Host = getEnvWithDefault("SERVER_HOST", "0.0.0.0")
	config.Environment = getEnvWithDefault("ENVIRONMENT", "development")

	readTimeout, err := parseDuration("SERVER_READ_TIMEOUT", "30s")
	if err != nil {
		return fmt.Errorf("invalid SERVER_READ_TIMEOUT: %w", err)
	}
	config.ReadTimeout = readTimeout

	writeTimeout, err := parseDuration("SERVER_WRITE_TIMEOUT", "30s")
	if err != nil {
		return fmt.Errorf("invalid SERVER_WRITE_TIMEOUT: %w", err)
	}
	config.WriteTimeout = writeTimeout

	return nil
}

func loadServicesConfig(config *ServicesConfig) error {
	// Twilio
	config.Twilio.AccountSID = os.Getenv("TWILIO_ACCOUNT_SID")
	config.Twilio.AuthToken = os.Getenv("TWILIO_AUTH_TOKEN")
	config.Twilio.FromPhone = os.Getenv("TWILIO_FROM_PHONE")

	if config.Twilio.AccountSID == "" || config.Twilio.AuthToken == "" {
		return fmt.Errorf("TWILIO_ACCOUNT_SID and TWILIO_AUTH_TOKEN are required")
	}

	// Email
	config.Email.SMTPHost = os.Getenv("SMTP_HOST")
	config.Email.SMTPUsername = os.Getenv("SMTP_USERNAME")
	config.Email.SMTPPassword = os.Getenv("SMTP_PASSWORD")
	config.Email.FromEmail = os.Getenv("FROM_EMAIL")
	config.Email.FromName = getEnvWithDefault("FROM_NAME", "Clinic Vet API")

	if config.Email.SMTPHost == "" {
		return fmt.Errorf("SMTP_HOST is required")
	}

	var err error
	config.Email.SMTPPort, err = parseIntWithDefault("SMTP_PORT", 587)
	if err != nil {
		return fmt.Errorf("invalid SMTP_PORT: %w", err)
	}

	return nil
}

func loadAppConfig(config *AppConfig) {
	config.Name = getEnvWithDefault("APP_NAME", "Clinic Vet API")
	config.Version = getEnvWithDefault("APP_VERSION", "1.0.0")
	config.Debug = parseBoolWithDefault("DEBUG", false)
	config.LogLevel = getEnvWithDefault("LOG_LEVEL", "info")
	config.EnableSwagger = parseBoolWithDefault("ENABLE_SWAGGER", true)
}

func (s *AppSettings) Validate() error {
	var errors []string

	// Validate server config
	if s.Server.Port == "" {
		errors = append(errors, "server port cannot be empty")
	}

	// Validate database config
	if s.Database.URL == "" {
		errors = append(errors, "database URL is required")
	}

	// Validate auth config
	if s.Auth.JWTSecret == "" {
		errors = append(errors, "JWT secret is required")
	}
	if len(s.Auth.JWTSecret) < 32 {
		errors = append(errors, "JWT secret should be at least 32 characters long")
	}

	// Validate services
	if s.Services.Twilio.AccountSID == "" {
		errors = append(errors, "Twilio Account SID is required")
	}
	if s.Services.Email.SMTPHost == "" {
		errors = append(errors, "SMTP host is required")
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// IsDevelopment returns true if running in development mode
func (s *AppSettings) IsDevelopment() bool {
	return s.Server.Environment == "development" || s.Server.Environment == "dev"
}

// IsProduction returns true if running in production mode
func (s *AppSettings) IsProduction() bool {
	return s.Server.Environment == "production" || s.Server.Environment == "prod"
}

// GetServerAddr returns the full server address
func (s *AppSettings) GetServerAddr() string {
	return s.Server.Host + ":" + s.Server.Port
}

func BootstrapAPIModules(
	routerGroup *gin.Engine,
	queries *sqlc.Queries,
	validator *validator.Validate,
	redis *redis.Client,
	settings *AppSettings,
) error {
	if settings == nil {
		return fmt.Errorf("application settings are required")
	}
	authMiddleware := middleware.NewAuthMiddleware(settings.Auth.JWTSecret)
	emailConfig := notification.SMTPConfig{
		Host:      settings.Services.Email.SMTPHost,
		Port:      settings.Services.Email.SMTPPort,
		Username:  settings.Services.Email.SMTPUsername,
		Password:  settings.Services.Email.SMTPPassword,
		FromEmail: settings.Services.Email.FromEmail,
	}
	if err := http.Bootstrap(
		routerGroup,
		queries,
		validator,
		authMiddleware,
		redis,
		settings.Auth.JWTSecret,
		settings.Auth.TwoFactorEncryptionKey,
		emailConfig,
		GetTwilioClient(),
		settings.Services.Twilio.FromPhone,
	); err != nil {
		return fmt.Errorf("failed to bootstrap API modules: %w", err)
	}
	return nil
}
