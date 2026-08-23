package http

import (
	"context"
	"fmt"

	"clinic-vet-api/database/sqlc"
	"clinic-vet-api/internal/core/addresses"
	"clinic-vet-api/internal/core/appointments"
	"clinic-vet-api/internal/core/auth"
	"clinic-vet-api/internal/core/customers"
	"clinic-vet-api/internal/core/employees"
	"clinic-vet-api/internal/core/medical"
	"clinic-vet-api/internal/core/notifications"
	"clinic-vet-api/internal/core/payments"
	"clinic-vet-api/internal/core/pets"
	"clinic-vet-api/internal/core/users"
	"clinic-vet-api/internal/infrastructure/http/handlers"
	"clinic-vet-api/internal/infrastructure/http/handlers/mappers"
	"clinic-vet-api/internal/infrastructure/notification"
	"clinic-vet-api/internal/infrastructure/persitence/repository"
	"clinic-vet-api/internal/infrastructure/session"
	"clinic-vet-api/internal/infrastructure/token"
	"clinic-vet-api/internal/infrastructure/twofactor"
	"clinic-vet-api/internal/middleware"
	"clinic-vet-api/internal/shared/mapper"
	"clinic-vet-api/internal/shared/password"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"github.com/twilio/twilio-go"
)

// ─── Config ─────────────────────────────────────────────────────────────────

type APIConfig struct {
	Router         *gin.Engine
	Queries        *sqlc.Queries
	Validator      *validator.Validate
	AuthMiddleware *middleware.AuthMiddleware
	// Optional: required for full auth (login/refresh/sessions)
	Redis                  *redis.Client
	JWTSecret              string
	TwoFactorEncryptionKey []byte
	SMTPConfig             notification.SMTPConfig
	TwilioClient           *twilio.RestClient
	TwilioFrom             string
}

func (c *APIConfig) Validate() error {
	if c.Router == nil {
		return fmt.Errorf("router is required")
	}
	if c.Queries == nil {
		return fmt.Errorf("queries are required")
	}
	if c.Validator == nil {
		return fmt.Errorf("validator is required")
	}
	if c.AuthMiddleware == nil {
		return fmt.Errorf("auth middleware is required")
	}
	return nil
}

// ─── Bootstrap ─────────────────────────────────────────────────────────────

// Bootstrap wires all repositories, services, and handlers then registers routes.
// Redis and JWTSecret are optional; if provided, auth (login/refresh) is fully wired.
func Bootstrap(
	engine *gin.Engine,
	queries *sqlc.Queries,
	validator *validator.Validate,
	authMiddleware *middleware.AuthMiddleware,
	redis *redis.Client,
	jwtSecret string,
	twoFactorEncryptionKey []byte,
	smtpConfig notification.SMTPConfig,
	twilioClient *twilio.RestClient,
	twilioFrom string,
) error {
	config := &APIConfig{
		Router:                 engine,
		Queries:                queries,
		Validator:              validator,
		AuthMiddleware:         authMiddleware,
		Redis:                  redis,
		JWTSecret:              jwtSecret,
		TwoFactorEncryptionKey: twoFactorEncryptionKey,
		SMTPConfig:             smtpConfig,
		TwilioClient:           twilioClient,
		TwilioFrom:             twilioFrom,
	}
	if err := config.Validate(); err != nil {
		return fmt.Errorf("validate API config: %w", err)
	}

	pgMap := mapper.NewSqlcFieldMapper()

	// ─── Repositories ───────────────────────────────────────────────────────
	addrRepo := repository.NewAddressSqlcRepository(queries)
	userRepo := repository.NewSqlcUserRepository(queries, pgMap)
	customerRepo := repository.NewCustomersSqlcRepository(queries)
	employeeRepo := repository.NewEmployeeSqlcRepository(queries)
	petRepo := repository.NewPetSqlcRepository(queries)
	appointmentRepo := repository.NewSqlcAppointmentRepository(queries)
	paymentRepo := repository.NewPaymentSqlcRepository(queries)
	notificationRepo := repository.NewSqlcNotificationRepository(queries)
	medicalSessionRepo := repository.NewMedicalSessionSqlcRepository(queries)
	vaccinationRepo := repository.NewSessionVaccinationSqlcRepository(queries)
	surgeryRepo := repository.NewSessionSurgerySqlcRepository(queries)
	prescriptionRepo := repository.NewSessionPrescriptionSqlcRepository(queries)
	attachmentRepo := repository.NewSessionAttachmentSqlcRepository(queries)
	sessionServiceRepo := repository.NewSessionServiceSqlcRepository(queries)
	vaccineCatalogRepo := repository.NewVaccineCatalogSqlcRepository(queries)
	medicationRepo := repository.NewMedicationSqlcRepository(queries)
	serviceCatalogRepo := repository.NewServiceCatalogSqlcRepository(queries)

	// ─── Shared ─────────────────────────────────────────────────────────────
	passwordEncoder := password.NewPasswordEncoder()
	emailSender, err := notification.NewSMTPSender(config.SMTPConfig)
	if err != nil {
		return fmt.Errorf("create SMTP sender: %w", err)
	}
	smsSender, err := notification.NewTwilioSender(config.TwilioClient, config.TwilioFrom)
	if err != nil {
		return fmt.Errorf("create Twilio sender: %w", err)
	}
	notificationSvc := notifications.NewNotificationService(notificationRepo, emailSender, smsSender)

	// ─── User + Profile ─────────────────────────────────────────────────────
	userCommand := users.NewCommandService(userRepo, passwordEncoder)
	userQuery := users.NewUserQueryService(userRepo)

	// ─── Customer ────────────────────────────────────────────────────────────
	customerSvc := customers.NewCustomerService(customerRepo)

	// ─── Auth (optional: needs Redis + JWTSecret) ────────────────────────────
	var authSvc auth.AuthService
	if config.Redis != nil && config.JWTSecret != "" {
		tokenSvc, err := token.NewTokenService(token.Config{
			Redis: config.Redis, JWTSecret: config.JWTSecret,
		})
		if err != nil {
			return fmt.Errorf("create token service: %w", err)
		}
		sessionRepo, err := session.NewRedisRepository(config.Redis)
		if err != nil {
			return fmt.Errorf("create session repository: %w", err)
		}
		twoFactorProvider, err := twofactor.NewTOTPProvider(twofactor.ProviderConfig{
			EncryptionKey: config.TwoFactorEncryptionKey,
			Issuer:        "Clinic Vet API",
			Users:         userQuery,
		})
		if err != nil {
			return fmt.Errorf("create two-factor provider: %w", err)
		}
		authSvc = auth.NewAuthService(
			tokenSvc, userCommand, userQuery,
			sessionRepo, passwordEncoder, twoFactorProvider, notificationSvc, customerSvc,
		)
	} else {
		// Auth handler will be nil; router must allow nil auth or Bootstrap will fail
		authSvc = nil
	}

	// ─── Address ─────────────────────────────────────────────────────────────
	addressSvc := addresses.NewAddressService(addrRepo)

	// ─── Employee ────────────────────────────────────────────────────────────
	employeeSvc := employees.NewEmployeeService(employeeRepo)

	// ─── Pets ───────────────────────────────────────────────────────────────
	// Adapter: pets.CustomerRepository requires ExistsByIDAndActive; customers repo has FindByID + IsActive
	petCustomerRepo := &petCustomerRepoAdapter{repo: customerRepo}
	petSvc := pets.NewPetService(petRepo, petCustomerRepo)

	// ─── Medical ────────────────────────────────────────────────────────────
	medicalSessionSvc := medical.NewMedicalSessionApplicationService(
		medicalSessionRepo, vaccinationRepo, surgeryRepo, prescriptionRepo,
		attachmentRepo, sessionServiceRepo, petRepo, employeeRepo,
	)
	vaccinationSvc := medical.NewVaccinationApplicationService(vaccinationRepo, medicalSessionRepo, vaccineCatalogRepo)
	surgerySvc := medical.NewSurgeryApplicationService(surgeryRepo, medicalSessionRepo)
	prescriptionSvc := medical.NewPrescriptionApplicationService(prescriptionRepo, medicalSessionRepo, medicationRepo)
	attachmentSvc := medical.NewAttachmentApplicationService(attachmentRepo, medicalSessionRepo)
	sessionServiceSvc := medical.NewSessionServiceApplicationService(sessionServiceRepo, medicalSessionRepo, serviceCatalogRepo)
	vaccineCatalogSvc := medical.NewVaccineCatalogApplicationService(vaccineCatalogRepo)
	medicationCatalogSvc := medical.NewMedicationCatalogApplicationService(medicationRepo)
	serviceCatalogSvc := medical.NewServiceCatalogApplicationService(serviceCatalogRepo)

	// ─── Appointments ────────────────────────────────────────────────────────
	appointmentDomain := appointments.NewAppointmentDomainService(appointmentRepo)
	appointmentCommand := appointments.NewCommandService(
		appointmentRepo, customerRepo, employeeRepo, petRepo, appointmentDomain,
		appointments.NotificationDependencies{Service: notificationSvc, Users: userQuery},
	)
	appointmentQuery := appointments.NewAppointmentQueryHandler(
		appointmentRepo, customerRepo, employeeRepo,
	)

	// ─── Payments ────────────────────────────────────────────────────────────
	paymentQuery := payments.NewQueryService(paymentRepo)
	paymentCommand := payments.NewCommandService(paymentRepo, appointmentRepo)

	// ─── Resolvers (user ID → customer/employee ID) ───────────────────────────
	customerIDResolver := func(ctx context.Context, userID uint) (uint, error) {
		c, err := customerRepo.FindByUserID(ctx, userID)
		if err != nil {
			return 0, err
		}
		return c.ID.Value(), nil
	}
	employeeIDResolver := func(ctx context.Context, userID uint) (uint, error) {
		e, err := employeeRepo.GetByUserID(ctx, userID)
		if err != nil {
			return 0, err
		}
		return e.ID.Value(), nil
	}

	// ─── Mappers ─────────────────────────────────────────────────────────────
	addressMapper := mappers.NewAddressMapper()
	authMapper := mappers.NewAuthMapper()
	userMapper := mappers.UserHandlerMapper{}
	petMapper := mappers.NewPetMapper()
	appointmentMapper := mappers.NewAppointmentResponseMapper()
	paymentMapper := &mappers.PaymentMapper{}

	// ─── Handlers ────────────────────────────────────────────────────────────
	addressHandler := handlers.NewAddressHandler(addressSvc, config.Validator, addressMapper)

	var authHandler *handlers.AuthHandler
	if authSvc != nil {
		authHandler = handlers.NewAuthHandler(authSvc, config.Validator, authMapper)
	} else {
		authHandler = nil // needs Redis + JWTSecret (and SessionRepository) to be set
	}

	userHandler := handlers.NewUserHandler(config.Validator, userQuery, userCommand, userMapper)
	profileHandler := handlers.NewProfileHandler(userQuery, userCommand)
	customerHandler := handlers.NewCustomerHandler(customerSvc, config.Validator)
	employeeHandler := handlers.NewEmployeeHandler(employeeSvc, config.Validator)
	petHandler := handlers.NewPetHandler(petSvc, config.Validator, petMapper, customerIDResolver)
	appointmentHandler := handlers.NewAppointmentHandler(
		appointmentCommand, appointmentQuery, config.Validator,
		appointmentMapper, customerIDResolver, employeeIDResolver,
	)
	paymentHandler := handlers.NewPaymentHandler(
		paymentQuery, paymentCommand, config.Validator, paymentMapper, customerIDResolver,
	)
	notificationHandler := handlers.NewNotificationHandler(notificationSvc, config.Validator)
	homeHandler := handlers.NewHomeHandler(config.Router)

	medicalHandler := handlers.NewMedicalHandler(
		medicalSessionSvc, vaccinationSvc, surgerySvc, prescriptionSvc, attachmentSvc,
		sessionServiceSvc, vaccineCatalogSvc, medicationCatalogSvc, serviceCatalogSvc,
		config.Validator, customerIDResolver, employeeIDResolver, petSvc,
	)

	appHandlers := &AppHandlers{
		address:      addressHandler,
		auth:         authHandler,
		user:         userHandler,
		profile:      profileHandler,
		customer:     customerHandler,
		employee:     employeeHandler,
		pets:         petHandler,
		appointment:  appointmentHandler,
		payment:      paymentHandler,
		medical:      medicalHandler,
		notification: notificationHandler,
		home:         homeHandler,
	}

	if err := assertHandlers(appHandlers, config); err != nil {
		return err
	}

	apiRouter, err := NewAPIRouter(appHandlers, config)
	if err != nil {
		return fmt.Errorf("create API router: %w", err)
	}
	apiRouter.RegisterRoutes()
	return nil
}

// assertHandlers verifies required handlers and config for routing.
func assertHandlers(h *AppHandlers, c *APIConfig) error {
	if h.address == nil {
		return fmt.Errorf("address handler is required")
	}
	if h.auth == nil {
		return fmt.Errorf("auth handler is required (set Redis and JWTSecret in config)")
	}
	if h.user == nil {
		return fmt.Errorf("user handler is required")
	}
	if h.profile == nil {
		return fmt.Errorf("profile handler is required")
	}
	if h.customer == nil {
		return fmt.Errorf("customer handler is required")
	}
	if h.employee == nil {
		return fmt.Errorf("employee handler is required")
	}
	if h.pets == nil {
		return fmt.Errorf("pets handler is required")
	}
	if h.appointment == nil {
		return fmt.Errorf("appointment handler is required")
	}
	if h.payment == nil {
		return fmt.Errorf("payment handler is required")
	}
	if h.home == nil {
		return fmt.Errorf("home handler is required")
	}
	if c.Router == nil || c.AuthMiddleware == nil {
		return fmt.Errorf("router and auth middleware are required")
	}
	return nil
}

// petCustomerRepoAdapter adapts customers.CustomerRepository to pets.CustomerRepository (ExistsByIDAndActive).
type petCustomerRepoAdapter struct {
	repo customers.CustomerRepository
}

func (a *petCustomerRepoAdapter) ExistsByIDAndActive(ctx context.Context, customerID uint) (bool, error) {
	c, err := a.repo.FindByID(ctx, customers.NewCustomerID(customerID))
	if err != nil {
		return false, err
	}
	return c.IsActive, nil
}
