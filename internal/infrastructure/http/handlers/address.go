package handlers

import (
	"clinic-vet-api/internal/core/addresses"
	"clinic-vet-api/internal/infrastructure/http/handlers/dtos"
	"clinic-vet-api/internal/infrastructure/http/handlers/mappers"
	"clinic-vet-api/internal/shared/http"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AddressHandler struct {
	service   addresses.AddressService
	validator *validator.Validate
	mapper    *mappers.AddressMapper
}

func NewAddressHandler(service addresses.AddressService, validator *validator.Validate, mapper *mappers.AddressMapper) *AddressHandler {
	if mapper == nil {
		mapper = mappers.NewAddressMapper()
	}
	return &AddressHandler{service: service, validator: validator, mapper: mapper}
}

// ------------------------------------------------------------
// Internal handlers
// ------------------------------------------------------------

func (s *AddressHandler) getAddressByIDInternal(ctx *gin.Context, getID EntityIDProvider, getUserID OptionalIdentityProvider) (any, error) {
	entityID, err := getID(ctx)
	if err != nil {
		return nil, err
	}

	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	addressID := entityID.(addresses.AddressID)
	address, err := s.service.GetAddressByIDAndUserID(ctx.Request.Context(), addressID, userID)
	return address, nil
}

func (s *AddressHandler) getAddressesByUserIDInternal(ctx *gin.Context, getUserID IdentityProvider) (any, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	addresses, err := s.service.GetAddressesByUserID(ctx.Request.Context(), userID)
	return addresses, nil
}

func (s *AddressHandler) createAddressInternal(ctx *gin.Context, req dtos.AddressCreateRequest, getID IdentityProvider) (any, error) {
	userID, err := getID(ctx)
	if err != nil {
		return nil, err
	}

	command, err := s.mapper.RequestToCreateCommand(req, userID)
	if err != nil {
		return nil, err
	}

	address, err := s.service.CreateAddress(ctx.Request.Context(), command)
	if err != nil {
		return nil, err
	}
	return address.ID.Value(), nil
}

func (s *AddressHandler) updateAddressInternal(ctx *gin.Context, req dtos.AddressUpdateRequest, getID OptionalIdentityProvider) (any, error) {
	userID, err := getID(ctx)
	if err != nil {
		return nil, err
	}
	pathID, err := http.ParseParamToUInt(ctx, "id")
	if err != nil {
		return nil, err
	}
	if req.ID != 0 && req.ID != pathID {
		return nil, fmt.Errorf("body id must match path id")
	}
	req.ID = pathID

	command, err := s.mapper.RequestToUpdateCommand(req, userID)
	if err != nil {
		return nil, err
	}

	err = s.service.UpdateAddress(ctx.Request.Context(), command)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *AddressHandler) deleteAddressInternal(ctx *gin.Context, getID OptionalIdentityProvider) (any, error) {
	userID, err := getID(ctx)
	if err != nil {
		return nil, err
	}

	entityID, err := http.ParseParamToUInt(ctx, "id")
	if err != nil {
		return nil, err
	}

	addressID := addresses.NewAddressID(entityID)
	err = s.service.DeleteAddress(ctx.Request.Context(), addressID, userID)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ------------------------------------------------------------
// Manager level handlers
// ------------------------------------------------------------

// @Summary Get my address
// @Description Returns an address belonging to the authenticated user.
// @Tags addresses
// @Security BearerAuth
// @Param id path int true "Address ID"
// @Success 200 {object} http.APIResponse{data=dtos.AddressResponse}
// @Router /me/addresses/{id} [get]
func (s *AddressHandler) GetAddressByID(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		return s.getAddressByIDInternal(ctx, AddressIDFromParam, UserIDFromContextPtr)
	}
	http.HandleRequestNoBodyStruct(s.validator, logic)(c)
}

// @Summary Get user address
// @Description Returns an address for a user. Requires admin or receptionist role.
// @Tags addresses
// @Security BearerAuth
// @Param user_id path int true "User ID"
// @Param id path int true "Address ID"
// @Success 200 {object} http.APIResponse{data=dtos.AddressResponse}
// @Router /addresses/users/{user_id}/{id} [get]
func (s *AddressHandler) GetAddressByIDForUser(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		return s.getAddressByIDInternal(ctx, AddressIDFromParam, UserIDFromParamOptional)
	}
	http.HandleRequestNoBodyStruct(s.validator, logic)(c)
}

// @Summary Search addresses
// @Description Lists addresses. Requires admin or receptionist role.
// @Tags addresses
// @Security BearerAuth
// @Param page_number query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} http.APIResponse{data=[]dtos.AddressResponse}
// @Router /addresses [get]
func (s *AddressHandler) SearchAddresses(c *gin.Context) {
	var requestBodyData dtos.AddressSearchRequest
	if err := http.ShouldBindAndValidateBody(c, &requestBodyData, s.validator); err != nil {
		http.BadRequest(c, err)
		return
	}

	query, err := s.mapper.RequestToSpecification(requestBodyData)
	if err != nil {
		http.BadRequest(c, err)
		return
	}

	addressPage, err := s.service.GetAddressesBySpecification(c.Request.Context(), query)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	responsePage := s.mapper.ToPaginatedResponse(addressPage)
	http.Paginated(c, &responsePage, "Addresses")
}

// @Summary List user addresses
// @Description Lists addresses for a user. Requires admin or receptionist role.
// @Tags addresses
// @Security BearerAuth
// @Param user_id path int true "User ID"
// @Success 200 {object} http.APIResponse{data=[]dtos.AddressResponse}
// @Router /addresses/users/{user_id} [get]
func (s *AddressHandler) GetAddressesByUserID(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		return s.getAddressesByUserIDInternal(ctx, UserIDFromParam)
	}
	http.HandleRequestNoBodyStruct(s.validator, logic)(c)
}

// @Summary Create user address
// @Description Creates an address for a user. Requires admin or receptionist role.
// @Tags addresses
// @Security BearerAuth
// @Param user_id path int true "User ID"
// @Param body body dtos.AddressCreateRequest true "Address"
// @Success 201 {object} http.APIResponse
// @Router /addresses/users/{user_id} [post]
func (s *AddressHandler) CreateAddress(c *gin.Context) {
	logic := func(ctx *gin.Context, req dtos.AddressCreateRequest) (any, error) {
		return s.createAddressInternal(ctx, req, UserIDFromParam)
	}
	http.HandleRequest(s.validator, logic)(c)
}

// @Summary Update user address
// @Description Updates an address for a user. Requires admin or receptionist role.
// @Tags addresses
// @Security BearerAuth
// @Param user_id path int true "User ID"
// @Param id path int true "Address ID"
// @Param body body dtos.AddressUpdateRequest true "Address"
// @Success 200 {object} http.APIResponse
// @Router /addresses/users/{user_id}/{id} [put]
func (s *AddressHandler) UpdateAddress(c *gin.Context) {
	logic := func(ctx *gin.Context, req dtos.AddressUpdateRequest) (any, error) {
		return s.updateAddressInternal(ctx, req, UserIDFromParamOptional)
	}
	http.HandleRequest(s.validator, logic)(c)
}

// @Summary Delete user address
// @Description Deletes an address for a user. Requires admin or receptionist role.
// @Tags addresses
// @Security BearerAuth
// @Param user_id path int true "User ID"
// @Param id path int true "Address ID"
// @Success 200 {object} http.APIResponse
// @Router /addresses/users/{user_id}/{id} [delete]
func (s *AddressHandler) DeleteAddress(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		return s.deleteAddressInternal(ctx, UserIDFromParamOptional)
	}
	http.HandleRequestNoBodyStruct(s.validator, logic)(c)
}

// ------------------------------------------------------------
// Customer level handlers
// ------------------------------------------------------------

// @Summary Create my address
// @Tags addresses
// @Security BearerAuth
// @Param body body dtos.AddressCreateRequest true "Address"
// @Success 201 {object} http.APIResponse
// @Router /me/addresses [post]
func (s *AddressHandler) CreateAddressForMe(c *gin.Context) {
	logic := func(ctx *gin.Context, req dtos.AddressCreateRequest) (any, error) {
		return s.createAddressInternal(ctx, req, UserIDFromContext)
	}
	http.HandleRequest(s.validator, logic)(c)
}

// @Summary List my addresses
// @Tags addresses
// @Security BearerAuth
// @Success 200 {object} http.APIResponse{data=[]dtos.AddressResponse}
// @Router /me/addresses [get]
func (s *AddressHandler) GetMyAddresses(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		return s.getAddressesByUserIDInternal(ctx, UserIDFromContext)
	}
	http.HandleRequestNoBodyStruct(s.validator, logic)(c)
}

// @Summary Update my address
// @Tags addresses
// @Security BearerAuth
// @Param id path int true "Address ID"
// @Param body body dtos.AddressUpdateRequest true "Address"
// @Success 200 {object} http.APIResponse
// @Router /me/addresses/{id} [put]
func (s *AddressHandler) UpdateMyAddress(c *gin.Context) {
	logic := func(ctx *gin.Context, req dtos.AddressUpdateRequest) (any, error) {
		return s.updateAddressInternal(ctx, req, UserIDFromContextPtr)
	}
	http.HandleRequest(s.validator, logic)(c)
}

// @Summary Delete my address
// @Tags addresses
// @Security BearerAuth
// @Param id path int true "Address ID"
// @Success 200 {object} http.APIResponse
// @Router /me/addresses/{id} [delete]
func (s *AddressHandler) DeleteMyAddress(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		return s.deleteAddressInternal(ctx, UserIDFromContextPtr)
	}
	http.HandleRequestNoBodyStruct(s.validator, logic)(c)
}
