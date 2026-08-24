package handlers

import (
	"clinic-vet-api/internal/core/users"
	"clinic-vet-api/internal/infrastructure/http/handlers/dtos"
	"clinic-vet-api/internal/infrastructure/http/handlers/mappers"
	"clinic-vet-api/internal/middleware"
	"clinic-vet-api/internal/shared"
	"clinic-vet-api/internal/shared/errors"
	"clinic-vet-api/internal/shared/http"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type UserHandler struct {
	validator      *validator.Validate
	queryService   users.QueryService
	commandService users.CommandService
	mapper         mappers.UserHandlerMapper
}

func NewUserHandler(
	validator *validator.Validate,
	queryService users.QueryService,
	commandService users.CommandService,
	mapper mappers.UserHandlerMapper,
) *UserHandler {
	return &UserHandler{
		validator:      validator,
		queryService:   queryService,
		commandService: commandService,
		mapper:         mapper,
	}
}

// GetUserByID godoc
// @Summary      Get user by ID
// @Description  Returns a single user by ID. Requires admin role.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  http.APIResponse{data=dtos.UserResponse}  "User found"
// @Failure      400  {object}  http.APIResponse  "Invalid ID"
// @Failure      401  {object}  http.APIResponse  "Unauthorized"
// @Failure      404  {object}  http.APIResponse  "User not found"
// @Failure      500  {object}  http.APIResponse  "Internal server error"
// @Router       /users/{id} [get]
func (ctrl *UserHandler) GetUserByID(c *gin.Context) {
	userIDUInt, err := http.ParseParamToUInt(c, "id")
	if err != nil {
		http.BadRequest(c, errors.RequestURLParamError(err, "id", c.Param("id")))
		return
	}

	userID := shared.NewUserID(userIDUInt)
	user, err := ctrl.queryService.GetByID(c.Request.Context(), userID)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	userResponse := ctrl.mapper.ToResponse(user)
	http.Found(c, userResponse, "User")
}

func (ctrl *UserHandler) GetUserByEmail(c *gin.Context) {
	emailStr := c.Param("email")
	if emailStr == "" {
		http.BadRequest(c, errors.RequestURLParamError(fmt.Errorf("email is required"), "email", ""))
		return
	}

	email, err := users.NewEmail(emailStr)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	user, err := ctrl.queryService.GetByEmail(c.Request.Context(), email)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	userResponse := ctrl.mapper.ToResponse(user)
	http.Found(c, userResponse, "User")
}

func (ctrl *UserHandler) GetUserByPhone(ctx *gin.Context) {
	phoneStr := ctx.Query("phone")
	if phoneStr == "" {
		http.BadRequest(ctx, errors.RequestURLParamError(fmt.Errorf("phone is required"), "phone", ""))
		return
	}

	phone, err := users.NewPhoneNumber(phoneStr)
	if err != nil {
		http.ApplicationError(ctx, err)
		return
	}

	user, err := ctrl.queryService.GetByPhone(ctx.Request.Context(), phone)
	if err != nil {
		http.ApplicationError(ctx, err)
		return
	}

	userResponse := ctrl.mapper.ToResponse(user)
	http.Found(ctx, userResponse, "User")
}

// SearchUsers godoc
// @Summary      Search users
// @Description  Paginated search and filter of users. Requires admin role.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page_number  query     int     false  "Page number"     default(1)
// @Param        page_size    query     int     false  "Page size"       default(10)
// @Param        email        query     string  false  "Filter by email"
// @Param        role         query     string  false  "Filter by role"
// @Param        status       query     string  false  "Filter by status"
// @Success      200          {object}  http.APIResponse{data=[]dtos.UserResponse}  "Paginated list of users"
// @Failure      400          {object}  http.APIResponse  "Invalid query params"
// @Failure      401          {object}  http.APIResponse  "Unauthorized"
// @Failure      500          {object}  http.APIResponse  "Internal server error"
// @Router       /users [get]
func (ctrl *UserHandler) SearchUsers(ctx *gin.Context) {
	var req dtos.UserSearchRequest
	if err := http.ShouldBindAndValidateQuery(ctx, &req, ctrl.validator); err != nil {
		http.BadRequest(ctx, err)
		return
	}

	spec, err := ctrl.mapper.UserSearchRequestToSpecification(req)
	if err != nil {
		http.ApplicationError(ctx, err)
		return
	}

	users, err := ctrl.queryService.GetSpecification(ctx.Request.Context(), *spec)
	if err != nil {
		http.ApplicationError(ctx, err)
		return
	}

	userResponses := ctrl.mapper.ToResponsePage(users)
	http.Paginated(ctx, &userResponses, "Users")
}

// CreateUser godoc
// @Summary      Create user
// @Description  Creates a new user. Requires admin role.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body   body      dtos.CreateUserRequest  true  "User data"
// @Success      201    {object}  http.APIResponse        "User created"
// @Failure      400    {object}  http.APIResponse        "Validation error"
// @Failure      401    {object}  http.APIResponse        "Unauthorized"
// @Failure      500    {object}  http.APIResponse        "Internal server error"
// @Router       /users [post]
func (ctrl *UserHandler) CreateUser(c *gin.Context) {
	var requestData dtos.CreateUserRequest
	if err := http.ShouldBindAndValidateBody(c, &requestData, ctrl.validator); err != nil {
		http.BadRequest(c, err)
		return
	}

	createUserCommand, err := ctrl.mapper.ToCreateUserCommand(requestData)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	user, err := ctrl.commandService.CreateUser(c.Request.Context(), createUserCommand)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	http.Created(c, user.ID, "User created successfully")
}

// UpdateUserStatus godoc
// @Summary      Update user status
// @Description  Updates the status of a user (e.g. activate/deactivate). Requires admin role.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id     path      int                           true  "User ID"
// @Param        body   body      dtos.UpdateUserStatusRequest  true  "New status"
// @Success      200    {object}  http.APIResponse  "Status updated"
// @Failure      400    {object}  http.APIResponse  "Invalid ID or status"
// @Failure      401    {object}  http.APIResponse  "Unauthorized"
// @Failure      500    {object}  http.APIResponse  "Internal server error"
// @Router       /users/{id}/status [post]
func (ctrl *UserHandler) UpdateUserStatus(c *gin.Context) {
	userID, err := http.ParseParamToUInt(c, "id")
	if err != nil {
		http.BadRequest(c, errors.RequestURLParamError(err, "id", c.Param("id")))
		return
	}

	var requestData dtos.UpdateUserStatusRequest
	if err := http.ShouldBindAndValidateBody(c, &requestData, ctrl.validator); err != nil {
		http.BadRequest(c, err)
		return
	}

	status, err := users.ParseUserStatus(requestData.Status)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	command := users.UpdateUserStatusCommand{
		ID:     shared.NewUserID(userID),
		Status: status,
	}

	err = ctrl.commandService.UpdateUserStatus(c.Request.Context(), command)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	http.Success(c, nil, "User status updated successfully")
}

// DeleteUser godoc
// @Summary      Delete user
// @Description  Soft-deletes a user by ID. Requires admin role.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  http.APIResponse  "User deleted"
// @Failure      400  {object}  http.APIResponse  "Invalid ID"
// @Failure      401  {object}  http.APIResponse  "Unauthorized"
// @Failure      500  {object}  http.APIResponse  "Internal server error"
// @Router       /users/{id} [delete]
func (ctrl *UserHandler) DeleteUser(c *gin.Context) {
	userIDUint, err := http.ParseParamToUInt(c, "id")
	if err != nil {
		http.BadRequest(c, errors.RequestURLParamError(err, "id", c.Param("id")))
		return
	}

	userID := shared.NewUserID(userIDUint)
	deleteCommand := users.DeleteUserCommand{
		ID:           userID,
		IsHardDelete: false,
	}

	err = ctrl.commandService.DeleteUser(c.Request.Context(), deleteCommand)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	http.Success(c, nil, "User Succesfully Deleted")
}

// RestoreUser godoc
// @Summary      Restore user
// @Description  Restores a soft-deleted user. Requires admin role.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  http.APIResponse  "User restored"
// @Failure      400  {object}  http.APIResponse  "Invalid ID"
// @Failure      401  {object}  http.APIResponse  "Unauthorized"
// @Failure      500  {object}  http.APIResponse  "Internal server error"
// @Router       /users/{id}/restore [post]
func (ctrl *UserHandler) RestoreUser(c *gin.Context) {
	userIDUint, err := http.ParseParamToUInt(c, "id")
	if err != nil {
		http.BadRequest(c, errors.RequestURLParamError(err, "id", c.Param("id")))
		return
	}

	userID := shared.NewUserID(userIDUint)
	err = ctrl.commandService.RestoreUser(c.Request.Context(), userID)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	http.Success(c, nil, "User Succesfully Restore")
}

// ------------------------------------------------------------
// Profile Routes
// ------------------------------------------------------------

type ProfileHandler struct {
	queryService   users.QueryService
	commandService users.CommandService
}

func NewProfileHandler(queryService users.QueryService, commandService users.CommandService) *ProfileHandler {
	return &ProfileHandler{queryService: queryService, commandService: commandService}
}

// GetProfile godoc
// @Summary      Get my profile
// @Description  Returns the authenticated user's profile. Requires Bearer auth.
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  http.APIResponse{data=dtos.ProfileResponse}  "Profile"
// @Failure      401  {object}  http.APIResponse  "Unauthorized"
// @Failure      500  {object}  http.APIResponse  "Internal server error"
// @Router       /profile [get]
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	user, err := h.queryService.GetByID(c.Request.Context(), userID)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	profile := dtos.ProfileResponse{
		ID:            user.ID.String(),
		Email:         user.Email.Value(),
		Phone:         user.PhoneNumber.Value(),
		Role:          user.Role.DisplayName(),
		Status:        user.Status.DisplayName(),
		Name:          user.Profile.Name,
		Gender:        user.Profile.Gender.DisplayName(),
		DateOfBirth:   user.Profile.DateOfBirth,
		ProfilePicUrl: user.Profile.PhotoURL,
		Bio:           user.Profile.Bio,
	}

	http.Success(c, profile, "Profile retrieved successfully")
}

// UpdateProfile godoc
// @Summary      Update my profile
// @Description  Updates the authenticated user's profile (name, gender, date of birth, photo, bio). Requires Bearer auth.
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body   body      dtos.UpdateProfileRequest  true  "Profile fields to update"
// @Success      200    {object}  http.APIResponse           "Profile updated"
// @Failure      400    {object}  http.APIResponse           "Validation error"
// @Failure      401    {object}  http.APIResponse           "Unauthorized"
// @Failure      500    {object}  http.APIResponse           "Internal server error"
// @Router       /profile [put]
func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	var requestData dtos.UpdateProfileRequest
	if err := c.ShouldBindJSON(&requestData); err != nil {
		errors.RequestBodyDataError(err)
		return
	}

	genderVO, err := shared.ParseGender(requestData.Gender)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}
	userIDVO, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	cmd := users.UpdateProfileCommand{
		ID:          userIDVO,
		Name:        requestData.Name,
		Gender:      genderVO,
		DateOfBirth: requestData.DateOfBirth,
		PhotoURL:    requestData.PhotoURL,
		Bio:         requestData.Bio,
	}

	if err := h.commandService.UpdateProfileData(c.Request.Context(), cmd); err != nil {
		http.ApplicationError(c, err)
		return
	}

	http.Success(c, nil, "Profile updated successfully")
}
