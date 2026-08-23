package handlers

import (
	"clinic-vet-api/internal/core/auth"
	"clinic-vet-api/internal/infrastructure/http/handlers/dtos"
	"clinic-vet-api/internal/infrastructure/http/handlers/mappers"
	"clinic-vet-api/internal/middleware"
	"clinic-vet-api/internal/shared/http"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	service   auth.AuthService
	validator *validator.Validate
	mapper    *mappers.AuthMapper
}

func NewAuthHandler(service auth.AuthService, validator *validator.Validate, mapper *mappers.AuthMapper) *AuthHandler {
	if mapper == nil {
		mapper = mappers.NewAuthMapper()
	}
	return &AuthHandler{service: service, validator: validator, mapper: mapper}
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a pending customer account and emails an activation code.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dtos.RegisterRequest  true  "Registration data"
// @Success      201   {object}  http.APIResponse      "Registration successful"
// @Failure      400   {object}  http.APIResponse      "Invalid request body or validation error"
// @Failure      500   {object}  http.APIResponse      "Internal server error"
// @Router       /auth/register [post]
func (ctrl *AuthHandler) Register(c *gin.Context) {
	var req dtos.RegisterRequest
	if err := http.ShouldBindAndValidateBody(c, &req, ctrl.validator); err != nil {
		http.BadRequest(c, err)
		return
	}

	command, err := ctrl.mapper.RegisterRequestToCommand(req)
	if err != nil {
		http.BadRequest(c, err)
		return
	}

	msg, err := ctrl.service.Register(c.Request.Context(), command)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	http.SuccessCreated(c, msg, "Registration successful")
}

// ActivateAccount godoc
// @Summary      Activate account
// @Description  Activates a user account using the user ID and activation code (e.g. from email link). Supports query params or JSON body.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user_id  query     string  true   "User ID"
// @Param        code     query     string  true   "Activation code"
// @Param        body     body      dtos.ActivateAccountRequest  false  "Alternative: JSON body with user_id and code"
// @Success      200      {object}  http.APIResponse  "Account activated"
// @Failure      400      {object}  http.APIResponse  "Invalid or missing parameters"
// @Failure      500      {object}  http.APIResponse  "Internal server error"
// @Router       /auth/activate [post]
func (ctrl *AuthHandler) ActivateAccount(c *gin.Context) {
	var req dtos.ActivateAccountRequest
	if err := c.ShouldBind(&req); err != nil {
		http.BadRequest(c, err)
		return
	}
	if err := ctrl.validator.Struct(&req); err != nil {
		http.BadRequest(c, err)
		return
	}

	command, err := ctrl.mapper.ActivateAccountRequestToCommand(req)
	if err != nil {
		http.BadRequest(c, err)
		return
	}

	err = ctrl.service.ActivateAccount(c.Request.Context(), command)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	http.Success(c, nil, "Account activated successfully")
}

// Login godoc
// @Summary      Login
// @Description  Authenticates with email and password. Returns an opaque challenge when TOTP verification is required.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body    body      dtos.LoginRequest  true  "Email and password"
// @Success      200     {object}  http.APIResponse{data=dtos.SessionResponse}  "Login successful (session payload with tokens and user)"
// @Failure      400     {object}  http.APIResponse   "Invalid credentials or validation error"
// @Failure      401     {object}  http.APIResponse   "Unauthorized"
// @Failure      500     {object}  http.APIResponse   "Internal server error"
// @Router       /auth/login [post]
func (ctrl *AuthHandler) Login(c *gin.Context) {
	var req dtos.LoginRequest
	if err := http.ShouldBindAndValidateBody(c, &req, ctrl.validator); err != nil {
		http.BadRequest(c, err)
		return
	}

	command := ctrl.mapper.LoginRequestToCommand(req, c)
	session, err := ctrl.service.Login(c.Request.Context(), command)
	if err != nil {
		var requires2FA *auth.RequiresTwoFactorError
		if errors.As(err, &requires2FA) {
			http.Success(c, ctrl.mapper.RequiresTwoFactorToResponse(requires2FA.ChallengeToken), "Two-factor authentication required")
			return
		}
		http.ApplicationError(c, err)
		return
	}

	http.Success(c, ctrl.mapper.SessionPayloadToResponse(session), "Login successful")
}

// RefreshToken godoc
// @Summary      Refresh access token
// @Description  Exchanges a valid refresh token for a new access token and refresh token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body    body      dtos.RefreshTokenRequest  true  "Refresh token"
// @Success      200     {object}  http.APIResponse{data=dtos.SessionResponse}  "New session payload"
// @Failure      400     {object}  http.APIResponse          "Invalid or expired refresh token"
// @Failure      401     {object}  http.APIResponse          "Unauthorized"
// @Failure      500     {object}  http.APIResponse          "Internal server error"
// @Router       /auth/refresh [post]
func (ctrl *AuthHandler) RefreshToken(c *gin.Context) {
	var req dtos.RefreshTokenRequest
	if err := http.ShouldBindAndValidateBody(c, &req, ctrl.validator); err != nil {
		http.BadRequest(c, err)
		return
	}

	command := ctrl.mapper.RefreshTokenRequestToCommand(req)
	session, err := ctrl.service.RefreshToken(c.Request.Context(), command)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	http.Success(c, ctrl.mapper.SessionPayloadToResponse(session), "Token refreshed successfully")
}

// Logout godoc
// @Summary      Logout
// @Description  Invalidates the current session (refresh token). Requires Bearer auth.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body    body      dtos.LogoutRequest  true  "Refresh token to revoke"
// @Success      200     {object}  http.APIResponse    "Logged out"
// @Failure      400     {object}  http.APIResponse    "Invalid request"
// @Failure      401     {object}  http.APIResponse   "Unauthorized"
// @Failure      500     {object}  http.APIResponse   "Internal server error"
// @Router       /auth/logout [post]
func (ctrl *AuthHandler) Logout(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		http.Unauthorized(c, err)
		return
	}

	var req dtos.LogoutRequest
	if err := http.ShouldBindAndValidateBody(c, &req, ctrl.validator); err != nil {
		http.BadRequest(c, err)
		return
	}

	command := ctrl.mapper.LogoutRequestToCommand(req)
	command.UserID = userID
	err = ctrl.service.Logout(c.Request.Context(), command)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	http.Success(c, nil, "Logout successful")
}

// LogoutAll godoc
// @Summary      Logout from all devices
// @Description  Invalidates all refresh tokens for the authenticated user. Requires Bearer auth.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  http.APIResponse  "Logged out from all devices"
// @Failure      401  {object}  http.APIResponse  "Unauthorized"
// @Failure      500  {object}  http.APIResponse  "Internal server error"
// @Router       /auth/logout-all [post]
func (ctrl *AuthHandler) LogoutAll(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		http.Unauthorized(c, err)
		return
	}
	err = ctrl.service.LogoutAll(c.Request.Context(), auth.LogoutAllCommand{UserID: userID})
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	http.Success(c, nil, "Logged out from all devices")
}

// VerifyTwoFactor godoc
// @Summary      Verify 2FA code
// @Description  Completes login using an opaque two-factor challenge and a TOTP or backup code.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body    body      dtos.VerifyTwoFactorRequest  true  "2FA code"
// @Success      200     {object}  http.APIResponse{data=dtos.SessionResponse}  "Session with tokens"
// @Failure      400     {object}  http.APIResponse            "Invalid code"
// @Failure      401     {object}  http.APIResponse            "Unauthorized"
// @Failure      500     {object}  http.APIResponse            "Internal server error"
// @Router       /auth/2fa/verify [post]
func (ctrl *AuthHandler) VerifyTwoFactor(c *gin.Context) {
	var req dtos.VerifyTwoFactorRequest
	if err := http.ShouldBindAndValidateBody(c, &req, ctrl.validator); err != nil {
		http.BadRequest(c, err)
		return
	}

	command := ctrl.mapper.VerifyTwoFactorRequestToCommand(req)
	session, err := ctrl.service.VerifyTwoFactor(c.Request.Context(), command)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	http.Success(c, ctrl.mapper.SessionPayloadToResponse(session), "Two factor verified successfully")
}

// EnableTwoFactor godoc
// @Summary      Enable two-factor authentication
// @Description  Enables TOTP 2FA for the authenticated user and returns recovery codes once. Requires Bearer auth.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body    body      dtos.EnableTwoFactorRequest  true  "2FA method"
// @Success      200     {object}  http.APIResponse            "2FA enabled"
// @Failure      400     {object}  http.APIResponse            "Invalid request"
// @Failure      401     {object}  http.APIResponse            "Unauthorized"
// @Failure      500     {object}  http.APIResponse            "Internal server error"
// @Router       /auth/2fa/enable [post]
func (ctrl *AuthHandler) EnableTwoFactor(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		http.Unauthorized(c, err)
		return
	}

	var req dtos.EnableTwoFactorRequest
	if err := http.ShouldBindAndValidateBody(c, &req, ctrl.validator); err != nil {
		http.BadRequest(c, err)
		return
	}

	command := ctrl.mapper.EnableTwoFactorRequestToCommand(req, userID)
	enrollment, err := ctrl.service.EnableTwoFactor(c.Request.Context(), command)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	http.Success(c, dtos.EnableTwoFactorResponse{BackupCodes: enrollment.BackupCodes}, "Two factor enabled successfully")
}

// DisableTwoFactor godoc
// @Summary      Disable two-factor authentication
// @Description  Disables 2FA for the authenticated user. Requires Bearer auth.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  http.APIResponse  "2FA disabled"
// @Failure      401  {object}  http.APIResponse  "Unauthorized"
// @Failure      500  {object}  http.APIResponse  "Internal server error"
// @Router       /auth/2fa/disable [post]
func (ctrl *AuthHandler) DisableTwoFactor(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		http.Unauthorized(c, err)
		return
	}

	err = ctrl.service.DisableTwoFactor(c.Request.Context(), auth.DisableTwoFactorCommand{UserID: userID})
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	http.Success(c, nil, "Two factor disabled successfully")
}

// VerifyEmail POST /auth/verify-email — body: VerifyEmailRequest
func (ctrl *AuthHandler) VerifyEmail(c *gin.Context) {
	var req dtos.VerifyEmailRequest
	if err := http.ShouldBindAndValidateBody(c, &req, ctrl.validator); err != nil {
		http.BadRequest(c, err)
		return
	}

	command := ctrl.mapper.VerifyEmailRequestToCommand(req)
	err := ctrl.service.VerifyEmail(c.Request.Context(), command)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	http.Success(c, nil, "Email verified successfully")
}

// RequestResetPassword godoc
// @Summary      Request password reset
// @Description  Sends a password reset code to the given email.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body    body      dtos.RequestResetPasswordRequest  true  "Email"
// @Success      200     {object}  http.APIResponse                  "If email exists, reset link sent"
// @Failure      400     {object}  http.APIResponse                  "Invalid request"
// @Failure      401     {object}  http.APIResponse                  "Unauthorized"
// @Failure      500     {object}  http.APIResponse                  "Internal server error"
// @Router       /auth/reset-password/request [post]
func (ctrl *AuthHandler) RequestResetPassword(c *gin.Context) {
	var req dtos.RequestResetPasswordRequest
	if err := http.ShouldBindAndValidateBody(c, &req, ctrl.validator); err != nil {
		http.BadRequest(c, err)
		return
	}

	command := ctrl.mapper.RequestResetPasswordRequestToCommand(req)
	err := ctrl.service.RequestResetPassword(c.Request.Context(), command)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	http.Success(c, nil, "If the email exists, a reset link has been sent")
}

// ResetPassword godoc
// @Summary      Reset password
// @Description  Resets the user password using email and the code received by email.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body    body      dtos.ResetPasswordRequest  true  "Email and reset code"
// @Success      200     {object}  http.APIResponse            "Password reset"
// @Failure      400     {object}  http.APIResponse            "Invalid or expired code"
// @Failure      401     {object}  http.APIResponse            "Unauthorized"
// @Failure      500     {object}  http.APIResponse            "Internal server error"
// @Router       /auth/reset-password/reset [post]
func (ctrl *AuthHandler) ResetPassword(c *gin.Context) {
	var req dtos.ResetPasswordRequest
	if err := http.ShouldBindAndValidateBody(c, &req, ctrl.validator); err != nil {
		http.BadRequest(c, err)
		return
	}

	command := ctrl.mapper.ResetPasswordRequestToCommand(req)
	err := ctrl.service.ResetPassword(c.Request.Context(), command)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	http.Success(c, nil, "Password reset successfully")
}

// UpdatePassword POST /auth/update-password — body: UpdatePasswordRequest; requires auth
func (ctrl *AuthHandler) UpdatePassword(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		http.Unauthorized(c, err)
		return
	}

	var req dtos.UpdatePasswordRequest
	if err := http.ShouldBindAndValidateBody(c, &req, ctrl.validator); err != nil {
		http.BadRequest(c, err)
		return
	}

	command := ctrl.mapper.UpdatePasswordRequestToCommand(req, userID)
	err = ctrl.service.UpdatePassword(c.Request.Context(), command)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	http.Success(c, nil, "Password updated successfully")
}

// DeleteMyAccount DELETE /auth/me or POST /auth/delete-account — requires auth; user from context
func (ctrl *AuthHandler) DeleteMyAccount(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		http.Unauthorized(c, err)
		return
	}

	err = ctrl.service.DeleteAccount(c.Request.Context(), userID)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	http.Success(c, nil, "Account deleted successfully")
}
