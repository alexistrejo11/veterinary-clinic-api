package auth

import (
	"context"
	"testing"
	"time"

	"clinic-vet-api/internal/core/customers"
	"clinic-vet-api/internal/core/notifications"
	"clinic-vet-api/internal/core/users"
	"clinic-vet-api/internal/shared"
	sharedlog "clinic-vet-api/internal/shared/log"

	"go.uber.org/zap"
)

type registrationUserCommandStub struct{ users.CommandService }

type logoutSessionRepositoryStub struct {
	SessionRepository
	session JwtSession
	deleted bool
}

func (s *logoutSessionRepositoryStub) GetByRefreshToken(context.Context, string) (JwtSession, error) {
	return s.session, nil
}

func (s *logoutSessionRepositoryStub) DeleteUserSession(context.Context, uint, string) error {
	s.deleted = true
	return nil
}

type logoutTokenServiceStub struct {
	TokenService
	revoked bool
}

func (s *logoutTokenServiceStub) RevokeToken(context.Context, string, TokenType) error {
	s.revoked = true
	return nil
}

func (s registrationUserCommandStub) CreateUser(_ context.Context, command users.CreateUserCommand) (users.User, error) {
	user := users.User{Email: command.Email, Role: command.Role, Status: command.Status}
	user.SetID(shared.NewUserID(42))
	return user, nil
}

type registrationCustomerServiceStub struct {
	customers.CustomerService
	command customers.CreateCustomerCommand
}

func (s *registrationCustomerServiceStub) CreateCustomer(_ context.Context, command customers.CreateCustomerCommand) (customers.Customer, error) {
	s.command = command
	return customers.Customer{}, nil
}

type registrationTokenServiceStub struct{ TokenService }

func (registrationTokenServiceStub) CreateToken(context.Context, map[string]any, TokenType) (Token, error) {
	return Token{Code: "activation-code", Type: TokenTypeVerification, CreatedAt: time.Now()}, nil
}

type registrationNotificationServiceStub struct {
	notifications.NotificationService
	sent int
}

type activationUserCommandStub struct {
	users.CommandService
	statusSet     bool
	emailVerified bool
}

func (s *activationUserCommandStub) UpdateUserStatus(context.Context, users.UpdateUserStatusCommand) error {
	s.statusSet = true
	return nil
}

func (s *activationUserCommandStub) SetEmailVerified(context.Context, shared.UserID) error {
	s.emailVerified = true
	return nil
}

type activationUserQueryStub struct {
	users.QueryService
	user users.User
}

func (s activationUserQueryStub) GetByID(context.Context, shared.UserID) (users.User, error) {
	return s.user, nil
}

type activationCustomerServiceStub struct {
	customers.CustomerService
	activated bool
}

func (s *activationCustomerServiceStub) ActivateCustomerByUserID(context.Context, uint) error {
	s.activated = true
	return nil
}

type activationTokenServiceStub struct {
	TokenService
	revoked bool
}

func (s *activationTokenServiceStub) VerifyToken(context.Context, string, TokenType) (bool, error) {
	return true, nil
}

func (s *activationTokenServiceStub) GetTokenClaims(context.Context, string, TokenType) (map[string]any, error) {
	return map[string]any{ClaimUserID: "42"}, nil
}

func (s *activationTokenServiceStub) RevokeToken(context.Context, string, TokenType) error {
	s.revoked = true
	return nil
}

func (s *registrationNotificationServiceStub) Send(context.Context, *notifications.Notification) error {
	s.sent++
	return nil
}

func TestLogoutRejectsRefreshTokenFromAnotherUser(t *testing.T) {
	previousLogger := sharedlog.App
	sharedlog.App = zap.NewNop()
	t.Cleanup(func() { sharedlog.App = previousLogger })

	sessions := &logoutSessionRepositoryStub{session: JwtSession{UserID: "2"}}
	tokens := &logoutTokenServiceStub{}
	service := authService{sessionRepository: sessions, tokenService: tokens}

	err := service.Logout(context.Background(), LogoutCommand{
		RefreshToken: "refresh-token",
		UserID:       shared.NewUserID(1),
	})
	if err == nil {
		t.Fatal("expected logout to reject a refresh token owned by another user")
	}
	if sessions.deleted || tokens.revoked {
		t.Fatal("logout must not revoke or delete another user's session")
	}
}

func TestActivateAccountActivatesCustomerProfile(t *testing.T) {
	user := users.User{Email: mustEmail(t, "customer@example.com"), Role: users.UserRoleCustomer, Status: users.UserStatusPending}
	user.SetID(shared.NewUserID(42))
	userCommand := &activationUserCommandStub{}
	customerService := &activationCustomerServiceStub{}
	tokenService := &activationTokenServiceStub{}
	service := NewAuthService(
		tokenService, userCommand, activationUserQueryStub{user: user},
		nil, nil, nil, nil, customerService,
	)

	if err := service.ActivateAccount(context.Background(), ActivateAccountCommand{UserID: shared.NewUserID(42), Code: "activation-code"}); err != nil {
		t.Fatalf("activate account: %v", err)
	}
	if !userCommand.statusSet || !userCommand.emailVerified || !customerService.activated || !tokenService.revoked {
		t.Fatalf("activation side effects: status=%t email=%t customer=%t token=%t", userCommand.statusSet, userCommand.emailVerified, customerService.activated, tokenService.revoked)
	}
}

func mustEmail(t *testing.T, value string) users.Email {
	t.Helper()
	email, err := users.NewEmail(value)
	if err != nil {
		t.Fatalf("create email: %v", err)
	}
	return email
}

func TestRegisterCreatesInactiveCustomerProfile(t *testing.T) {
	customers := &registrationCustomerServiceStub{}
	notifications := &registrationNotificationServiceStub{}
	service := NewAuthService(
		registrationTokenServiceStub{}, registrationUserCommandStub{}, nil,
		nil, nil, nil, notifications, customers,
	)

	message, err := service.Register(context.Background(), RegisterCommand{
		Role:        users.UserRoleCustomer,
		Email:       "customer@example.com",
		Password:    "StrongPassword123!",
		FirstName:   "Ada",
		LastName:    "Lovelace",
		DateOfBirth: "1990-12-10",
		Gender:      "female",
		PhotoURL:    "https://example.com/ada.jpg",
	})
	if err != nil || message != CustomerMessageSuccess {
		t.Fatalf("register = %q, %v", message, err)
	}
	if customers.command.UserID == nil || *customers.command.UserID != 42 || customers.command.IsActive {
		t.Fatalf("customer profile command = %#v", customers.command)
	}
	if customers.command.FirstName != "Ada" || customers.command.LastName != "Lovelace" || customers.command.Gender != shared.GenderFemale {
		t.Fatalf("customer identity fields = %#v", customers.command)
	}
	if notifications.sent != 1 {
		t.Fatalf("activation notifications = %d, want 1", notifications.sent)
	}
}
