package mappers

import (
	"clinic-vet-api/internal/core/auth"
	"clinic-vet-api/internal/core/users"
	"clinic-vet-api/internal/infrastructure/http/handlers/dtos"
	"clinic-vet-api/internal/shared"
	"clinic-vet-api/internal/shared/http"

	"github.com/gin-gonic/gin"
)

// AuthMapper maps auth DTOs to auth commands and session payload to response DTOs.
type AuthMapper struct{}

func NewAuthMapper() *AuthMapper {
	return &AuthMapper{}
}

func (m *AuthMapper) RegisterRequestToCommand(req dtos.RegisterRequest) (auth.RegisterCommand, error) {
	role, err := users.ParseUserRole(req.Role)
	if err != nil {
		return auth.RegisterCommand{}, err
	}
	cmd := auth.RegisterCommand{
		Role:          role,
		Email:         req.Email,
		Password:      req.Password,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		PhoneNumber:   req.PhoneNumber,
		DateOfBirth:   req.DateOfBirth,
		Gender:        req.Gender,
		PhotoURL:      req.PhotoURL,
		EmployeeID:    req.EmployeeID,
		LicenseNumber: req.LicenseNumber,
	}
	return cmd, nil
}

func (m *AuthMapper) ActivateAccountRequestToCommand(req dtos.ActivateAccountRequest) (auth.ActivateAccountCommand, error) {
	userID, err := shared.ParseUserIDFromString(req.UserID)
	if err != nil {
		return auth.ActivateAccountCommand{}, err
	}
	return auth.ActivateAccountCommand{
		UserID: userID,
		Code:   req.Code,
	}, nil
}

func (m *AuthMapper) LoginRequestToCommand(req dtos.LoginRequest, c *gin.Context) auth.LoginCommand {
	meta := http.NewLoginMetadata(c)
	return auth.LoginCommand{
		Email:      req.Email,
		Password:   req.Password,
		DeviceInfo: meta.DeviceInfo,
		UserAgent:  meta.UserAgent,
		IPAddress:  meta.IP,
	}
}

// RefreshToken: single-field command, map DTO -> command
func (m *AuthMapper) RefreshTokenRequestToCommand(req dtos.RefreshTokenRequest) auth.RefreshTokenCommand {
	return auth.RefreshTokenCommand{RefreshToken: req.RefreshToken}
}

func (m *AuthMapper) LogoutRequestToCommand(req dtos.LogoutRequest) auth.LogoutCommand {
	return auth.LogoutCommand{RefreshToken: req.RefreshToken}
}

func (m *AuthMapper) VerifyTwoFactorRequestToCommand(req dtos.VerifyTwoFactorRequest) auth.VerifyTwoFactorCommand {
	return auth.VerifyTwoFactorCommand{
		ChallengeToken: req.ChallengeToken,
		TwoFactorCode:  req.TwoFactorCode,
	}
}

func (m *AuthMapper) EnableTwoFactorRequestToCommand(req dtos.EnableTwoFactorRequest, userID shared.UserID) auth.EnableTwoFactorCommand {
	return auth.EnableTwoFactorCommand{
		UserID: userID,
		Method: req.Method,
	}
}

func (m *AuthMapper) VerifyEmailRequestToCommand(req dtos.VerifyEmailRequest) auth.VerifyEmailCommand {
	return auth.VerifyEmailCommand{
		Email: req.Email,
		Code:  req.Code,
	}
}

func (m *AuthMapper) RequestResetPasswordRequestToCommand(req dtos.RequestResetPasswordRequest) auth.RequestResetPasswordCommand {
	return auth.RequestResetPasswordCommand{Email: req.Email}
}

func (m *AuthMapper) ResetPasswordRequestToCommand(req dtos.ResetPasswordRequest) auth.ResetPasswordCommand {
	return auth.ResetPasswordCommand{
		Email:    req.Email,
		Code:     req.Code,
		Password: req.Password,
	}
}

func (m *AuthMapper) UpdatePasswordRequestToCommand(req dtos.UpdatePasswordRequest, userID shared.UserID) auth.UpdatePasswordCommand {
	return auth.UpdatePasswordCommand{
		UserID:          userID,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	}
}

func (m *AuthMapper) SessionPayloadToResponse(payload auth.SessionPayload) dtos.SessionResponse {
	return dtos.SessionResponse{
		AccessToken:  tokenToResponse(payload.AccessToken),
		RefreshToken: tokenToResponse(payload.RefreshToken),
		User:         userToAuthResponse(payload.User),
	}
}

func tokenToResponse(t auth.Token) dtos.TokenResponse {
	return dtos.TokenResponse{
		Token:     t.Code,
		ExpiresAt: t.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		Type:      string(t.Type),
	}
}

func userToAuthResponse(u users.User) dtos.UserAuthResponse {
	return dtos.UserAuthResponse{
		ID:    u.ID.Value(),
		Email: u.Email.String(),
		Role:  u.Role.String(),
	}
}

func (m *AuthMapper) RequiresTwoFactorToResponse(challengeToken string) dtos.RequiresTwoFactorResponse {
	return dtos.RequiresTwoFactorResponse{
		RequiresTwoFactor: true,
		ChallengeToken:    challengeToken,
		Message:           "Two-factor authentication is required to complete login",
	}
}
