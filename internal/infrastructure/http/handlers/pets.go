package handlers

import (
	"errors"
	"strconv"

	"clinic-vet-api/internal/core/pets"
	"clinic-vet-api/internal/infrastructure/http/handlers/dtos"
	"clinic-vet-api/internal/infrastructure/http/handlers/mappers"
	"clinic-vet-api/internal/shared/http"
	"clinic-vet-api/internal/shared/page"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type PetHandler struct {
	service            pets.Service
	validator          *validator.Validate
	mapper             *mappers.PetMapper
	customerIDResolver CustomerIDResolver
}

func NewPetHandler(
	service pets.Service,
	validator *validator.Validate,
	mapper *mappers.PetMapper,
	customerIDResolver CustomerIDResolver,
) *PetHandler {
	if mapper == nil {
		mapper = mappers.NewPetMapper()
	}
	return &PetHandler{
		service:            service,
		validator:          validator,
		mapper:             mapper,
		customerIDResolver: customerIDResolver,
	}
}

// parsePetIDFromParam returns the pet id from the "id" URL param.
func parsePetIDFromParam(c *gin.Context) (pets.PetID, error) {
	id, err := http.ParseParamToUInt(c, "id")
	if err != nil {
		return pets.PetID{}, err
	}
	return pets.NewPetID(id), nil
}

// ------------------------------------------------------------
// Internal handlers
// ------------------------------------------------------------

func (s *PetHandler) getPetByIDInternal(ctx *gin.Context, petID pets.PetID, optCustomerID *uint) (any, error) {
	if optCustomerID != nil {
		return s.service.GetPetByIDAndCustomerID(ctx.Request.Context(), petID, *optCustomerID)
	}
	return s.service.GetPetByID(ctx.Request.Context(), petID)
}

func (s *PetHandler) getPetsByCustomerIDInternal(ctx *gin.Context, getCustomerID CustomerIDProvider, pagination page.Pagination) (any, error) {
	customerID, err := getCustomerID(ctx)
	if err != nil {
		return nil, err
	}
	return s.service.GetPetsByCustomerID(ctx.Request.Context(), customerID, pagination)
}

func (s *PetHandler) createPetInternal(ctx *gin.Context, req dtos.PetCreateRequest, getCustomerID CustomerIDProvider, isActive bool) (any, error) {
	customerID, err := getCustomerID(ctx)
	if err != nil {
		return nil, err
	}
	command, err := s.mapper.RequestToCreateCommand(req, customerID, isActive)
	if err != nil {
		return nil, err
	}
	pet, err := s.service.CreatePet(ctx.Request.Context(), command)
	if err != nil {
		return nil, err
	}
	return pet.ID.Value(), nil
}

func (s *PetHandler) updatePetInternal(ctx *gin.Context, req dtos.PetUpdateRequest, petID pets.PetID, optCustomerID *uint) (any, error) {
	command, err := s.mapper.RequestToUpdateCommand(req, petID, optCustomerID)
	if err != nil {
		return nil, err
	}
	err = s.service.UpdatePet(ctx.Request.Context(), command)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *PetHandler) deletePetInternal(ctx *gin.Context, petID pets.PetID, optCustomerID *uint, isHardDelete bool) (any, error) {
	command := pets.DeletePetCommand{
		PetID:         petID,
		OptCustomerID: optCustomerID,
		IsHardDelete:  isHardDelete,
	}
	err := s.service.DeletePet(ctx.Request.Context(), command)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *PetHandler) restorePetInternal(ctx *gin.Context, petID pets.PetID) (any, error) {
	err := s.service.RestorePet(ctx.Request.Context(), petID)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// ------------------------------------------------------------
// Manager level handlers
// ------------------------------------------------------------

// @Summary Get pet by ID
// @Description Requires admin or receptionist role.
// @Tags pets
// @Security BearerAuth
// @Param id path int true "Pet ID"
// @Success 200 {object} http.APIResponse{data=dtos.PetResponse}
// @Router /pets/{id} [get]
func (s *PetHandler) GetPetByID(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		petID, err := parsePetIDFromParam(ctx)
		if err != nil {
			return nil, err
		}
		return s.getPetByIDInternal(ctx, petID, nil)
	}
	http.HandleGetRequest(s.validator, "Pet", logic)(c)
}

// @Summary List customer pets
// @Description Requires admin or receptionist role.
// @Tags pets
// @Security BearerAuth
// @Param customer_id path int true "Customer ID"
// @Param page_number query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} http.APIResponse{data=[]dtos.PetResponse}
// @Router /pets/customers/{customer_id} [get]
func (s *PetHandler) GetPetsByCustomerID(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		var pageParams page.PaginationRequest
		if err := http.ShouldBindPageParams(&pageParams, ctx, s.validator); err != nil {
			return nil, err
		}
		return s.getPetsByCustomerIDInternal(ctx, CustomerIDFromParam, pageParams.ToPagination())
	}
	http.HandleRequestNoBodyWithResponder(s.validator, logic, func(c *gin.Context, res any) {
		if res == nil {
			http.Success(c, nil, "No pets found")
			return
		}
		p := res.(page.Page[pets.Pet])
		responsePage := s.mapper.ToPaginatedResponse(p)
		http.Paginated(c, &responsePage, "Pets")
	})(c)
}

// @Summary Create customer pet
// @Description Requires admin or receptionist role.
// @Tags pets
// @Security BearerAuth
// @Param customer_id path int true "Customer ID"
// @Param is_active query bool false "Whether the pet is active" default(true)
// @Param body body dtos.PetCreateRequest true "Pet"
// @Success 201 {object} http.APIResponse
// @Router /pets/customers/{customer_id} [post]
func (s *PetHandler) CreatePet(c *gin.Context) {
	logic := func(ctx *gin.Context, req dtos.PetCreateRequest) (any, error) {
		isActive := true
		if q := ctx.Query("is_active"); q != "" {
			var err error
			isActive, err = strconv.ParseBool(q)
			if err != nil {
				return nil, err
			}
		}
		return s.createPetInternal(ctx, req, CustomerIDFromParam, isActive)
	}
	http.HandleCreateRequest(s.validator, "Pet", logic)(c)
}

// @Summary Update pet
// @Description Requires admin or receptionist role.
// @Tags pets
// @Security BearerAuth
// @Param id path int true "Pet ID"
// @Param body body dtos.PetUpdateRequest true "Pet"
// @Success 200 {object} http.APIResponse
// @Router /pets/{id} [put]
func (s *PetHandler) UpdatePet(c *gin.Context) {
	logic := func(ctx *gin.Context, req dtos.PetUpdateRequest) (any, error) {
		petID, err := parsePetIDFromParam(ctx)
		if err != nil {
			return nil, err
		}
		optCustomerID, _ := CustomerIDFromParamOptional(ctx)
		return s.updatePetInternal(ctx, req, petID, optCustomerID)
	}
	http.HandleUpdateRequest(s.validator, "Pet", logic)(c)
}

// @Summary Delete pet
// @Description Requires admin or receptionist role. Set hard to permanently delete.
// @Tags pets
// @Security BearerAuth
// @Param id path int true "Pet ID"
// @Param hard query bool false "Permanently delete"
// @Success 200 {object} http.APIResponse
// @Router /pets/{id} [delete]
func (s *PetHandler) DeletePet(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		petID, err := parsePetIDFromParam(ctx)
		if err != nil {
			return nil, err
		}
		isHard := ctx.Query("hard") == "true"
		return s.deletePetInternal(ctx, petID, nil, isHard)
	}
	http.HandleDeleteRequest(s.validator, "Pet", logic)(c)
}

// @Summary Restore pet
// @Description Requires admin or receptionist role.
// @Tags pets
// @Security BearerAuth
// @Param id path int true "Pet ID"
// @Success 200 {object} http.APIResponse
// @Router /pets/{id}/restore [post]
func (s *PetHandler) RestorePet(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		petID, err := parsePetIDFromParam(ctx)
		if err != nil {
			return nil, err
		}
		return s.restorePetInternal(ctx, petID)
	}
	http.HandleRequestNoBodyWithResponder(s.validator, logic, func(c *gin.Context, _ any) {
		http.Updated(c, nil, "Pet")
	})(c)
}

// @Summary Search pets
// @Description Requires admin or receptionist role.
// @Tags pets
// @Security BearerAuth
// @Param page_number query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} http.APIResponse{data=[]dtos.PetResponse}
// @Router /pets [get]
func (s *PetHandler) SearchPets(c *gin.Context) {
	var requestBodyData dtos.PetSearchRequest
	if err := http.ShouldBindAndValidateBody(c, &requestBodyData, s.validator); err != nil {
		http.BadRequest(c, err)
		return
	}
	query, err := s.mapper.RequestToSpecification(requestBodyData)
	if err != nil {
		http.BadRequest(c, err)
		return
	}
	petPage, err := s.service.GetPetBySpecification(c.Request.Context(), &query)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}
	responsePage := s.mapper.ToPaginatedResponse(petPage)
	http.Paginated(c, &responsePage, "Pets")
}

// ------------------------------------------------------------
// Customer level handlers ("my" endpoints)
// ------------------------------------------------------------

func (s *PetHandler) getMyCustomerID(ctx *gin.Context) (uint, error) {
	if s.customerIDResolver == nil {
		return 0, errors.New("customer resolver not configured")
	}
	return CustomerIDFromContext(ctx, s.customerIDResolver)
}

// @Summary Get my pet
// @Description Requires customer role.
// @Tags pets
// @Security BearerAuth
// @Param id path int true "Pet ID"
// @Success 200 {object} http.APIResponse{data=dtos.PetResponse}
// @Router /me/pets/{id} [get]
func (s *PetHandler) GetMyPet(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		customerID, err := s.getMyCustomerID(ctx)
		if err != nil {
			return nil, err
		}
		petID, err := parsePetIDFromParam(ctx)
		if err != nil {
			return nil, err
		}
		return s.getPetByIDInternal(ctx, petID, &customerID)
	}
	http.HandleGetRequest(s.validator, "Pet", logic)(c)
}

// @Summary List my pets
// @Description Requires customer role.
// @Tags pets
// @Security BearerAuth
// @Param page_number query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} http.APIResponse{data=[]dtos.PetResponse}
// @Router /me/pets [get]
func (s *PetHandler) GetMyPets(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		var pageParams page.PaginationRequest
		if err := http.ShouldBindPageParams(&pageParams, ctx, s.validator); err != nil {
			return nil, err
		}
		getCustomerID := func(c *gin.Context) (uint, error) { return s.getMyCustomerID(c) }
		return s.getPetsByCustomerIDInternal(ctx, getCustomerID, pageParams.ToPagination())
	}
	http.HandleRequestNoBodyWithResponder(s.validator, logic, func(c *gin.Context, res any) {
		if res == nil {
			http.Success(c, nil, "No pets found")
			return
		}
		p := res.(page.Page[pets.Pet])
		responsePage := s.mapper.ToPaginatedResponse(p)
		http.Paginated(c, &responsePage, "Pets")
	})(c)
}

// @Summary Create my pet
// @Description Requires customer role.
// @Tags pets
// @Security BearerAuth
// @Param body body dtos.PetCreateRequest true "Pet"
// @Success 201 {object} http.APIResponse
// @Router /me/pets [post]
func (s *PetHandler) CreatePetForMe(c *gin.Context) {
	logic := func(ctx *gin.Context, req dtos.PetCreateRequest) (any, error) {
		getCustomerID := func(c *gin.Context) (uint, error) { return s.getMyCustomerID(c) }
		return s.createPetInternal(ctx, req, getCustomerID, true)
	}
	http.HandleCreateRequest(s.validator, "Pet", logic)(c)
}

// @Summary Update my pet
// @Description Requires customer role.
// @Tags pets
// @Security BearerAuth
// @Param id path int true "Pet ID"
// @Param body body dtos.PetUpdateRequest true "Pet"
// @Success 200 {object} http.APIResponse
// @Router /me/pets/{id} [put]
func (s *PetHandler) UpdateMyPet(c *gin.Context) {
	logic := func(ctx *gin.Context, req dtos.PetUpdateRequest) (any, error) {
		customerID, err := s.getMyCustomerID(ctx)
		if err != nil {
			return nil, err
		}
		petID, err := parsePetIDFromParam(ctx)
		if err != nil {
			return nil, err
		}
		return s.updatePetInternal(ctx, req, petID, &customerID)
	}
	http.HandleUpdateRequest(s.validator, "Pet", logic)(c)
}

// @Summary Delete my pet
// @Description Requires customer role.
// @Tags pets
// @Security BearerAuth
// @Param id path int true "Pet ID"
// @Success 200 {object} http.APIResponse
// @Router /me/pets/{id} [delete]
func (s *PetHandler) DeleteMyPet(c *gin.Context) {
	logic := func(ctx *gin.Context) (any, error) {
		customerID, err := s.getMyCustomerID(ctx)
		if err != nil {
			return nil, err
		}
		petID, err := parsePetIDFromParam(ctx)
		if err != nil {
			return nil, err
		}
		return s.deletePetInternal(ctx, petID, &customerID, false)
	}
	http.HandleDeleteRequest(s.validator, "Pet", logic)(c)
}
