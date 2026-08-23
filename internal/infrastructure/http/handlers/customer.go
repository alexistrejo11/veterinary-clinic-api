package handlers

import (
	"clinic-vet-api/internal/core/customers"
	"clinic-vet-api/internal/infrastructure/http/handlers/dtos"
	"clinic-vet-api/internal/infrastructure/http/handlers/mappers"
	"clinic-vet-api/internal/shared/http"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type CustomerHandler struct {
	service   customers.CustomerService
	validator *validator.Validate
	mapper    *mappers.CustomerMapper
}

func NewCustomerHandler(service customers.CustomerService, validator *validator.Validate) *CustomerHandler {
	return &CustomerHandler{service: service, validator: validator, mapper: mappers.NewCustomerMapper()}
}

// GetCustomerByID godoc
// @Summary      Get customer by ID
// @Description  Returns a single customer by ID. Requires admin or receptionist role.
// @Tags         customers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Customer ID"
// @Success      200  {object}  http.APIResponse{data=dtos.CustomerResponse}  "Customer found"
// @Failure      400  {object}  http.APIResponse  "Invalid ID"
// @Failure      401  {object}  http.APIResponse  "Unauthorized"
// @Failure      404  {object}  http.APIResponse  "Customer not found"
// @Failure      500  {object}  http.APIResponse  "Internal server error"
// @Router       /customers/{id} [get]
func (s *CustomerHandler) GetCustomerByID(c *gin.Context) {
	customerIDUInt, err := http.ParseParamToUInt(c, "id")
	if err != nil {
		http.BadRequest(c, err)
		return
	}

	customerID := customers.NewCustomerID(customerIDUInt)
	customer, err := s.service.GetCustomerByID(c.Request.Context(), customerID)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	customerResponse := s.mapper.ToCustomerResponse(customer)
	http.Found(c, customerResponse, "Customer")
}

// SearchCustomers godoc
// @Summary      Search customers
// @Description  Paginated search/filter of customers. Requires admin or receptionist role.
// @Tags         customers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body   body      dtos.CustomerSearchRequest  true  "Search criteria and pagination"
// @Success      200     {object}  http.APIResponse{data=[]dtos.CustomerResponse}  "Paginated list of customers"
// @Failure      400     {object}  http.APIResponse           "Validation error"
// @Failure      401     {object}  http.APIResponse           "Unauthorized"
// @Failure      500     {object}  http.APIResponse           "Internal server error"
// @Router       /customers [get]
func (s *CustomerHandler) SearchCustomers(c *gin.Context) {
	var requestBodyData dtos.CustomerSearchRequest
	if err := http.ShouldBindAndValidateQuery(c, &requestBodyData, s.validator); err != nil {
		http.BadRequest(c, err)
		return
	}
	spec := s.mapper.RequestToSpecification(requestBodyData)
	customerPage, err := s.service.GetCustomersBySpecification(c.Request.Context(), spec)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}
	responsePage := s.mapper.ToCustomerResponsePage(customerPage)
	http.Paginated(c, &responsePage, "Customers")
}

// CreateCustomer godoc
// @Summary      Create customer
// @Description  Creates a new customer. Requires admin or receptionist role.
// @Tags         customers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body   body      dtos.CustomerCreateRequest  true  "Customer data"
// @Success      201    {object}  http.APIResponse            "Customer created"
// @Failure      400    {object}  http.APIResponse            "Validation error"
// @Failure      401    {object}  http.APIResponse            "Unauthorized"
// @Failure      500    {object}  http.APIResponse            "Internal server error"
// @Router       /customers [post]
func (s *CustomerHandler) CreateCustomer(c *gin.Context) {
	var requestBodyData dtos.CustomerCreateRequest
	if err := http.ShouldBindAndValidateBody(c, &requestBodyData, s.validator); err != nil {
		http.BadRequest(c, err)
		return
	}

	command, err := s.mapper.RequestToCreateCommand(requestBodyData)
	if err != nil {
		http.BadRequest(c, err)
		return
	}

	customer, err := s.service.CreateCustomer(c.Request.Context(), command)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	http.Created(c, customer.ID.Value(), "Customer")
}

// UpdateCustomer godoc
// @Summary      Update customer
// @Description  Updates an existing customer by ID. Requires admin or receptionist role.
// @Tags         customers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body   body      dtos.CustomerUpdateRequest  true  "Customer data to update"
// @Success      200    {object}  http.APIResponse           "Customer updated"
// @Failure      400    {object}  http.APIResponse           "Validation error"
// @Failure      401    {object}  http.APIResponse           "Unauthorized"
// @Failure      500    {object}  http.APIResponse           "Internal server error"
// @Router       /customers/{id} [put]
func (s *CustomerHandler) UpdateCustomer(c *gin.Context) {
	pathID, err := http.ParseParamToUInt(c, "id")
	if err != nil {
		http.BadRequest(c, err)
		return
	}
	var requestBodyData dtos.CustomerUpdateRequest
	if err := http.ShouldBindAndValidateBody(c, &requestBodyData, s.validator); err != nil {
		http.BadRequest(c, err)
		return
	}
	if requestBodyData.ID != pathID {
		http.BadRequest(c, fmt.Errorf("body id must match path id"))
		return
	}

	command, err := s.mapper.RequestToUpdateCommand(requestBodyData)
	if err != nil {
		http.BadRequest(c, err)
		return
	}
	err = s.service.UpdateCustomer(c.Request.Context(), command)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	http.Updated(c, nil, "Customer")
}

// DeleteCustomer godoc
// @Summary      Delete customer
// @Description  Soft-deletes a customer by ID. Requires admin or receptionist role.
// @Tags         customers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Customer ID"
// @Success      200  {object}  http.APIResponse  "Customer deleted"
// @Failure      400  {object}  http.APIResponse  "Invalid ID"
// @Failure      401  {object}  http.APIResponse  "Unauthorized"
// @Failure      500  {object}  http.APIResponse  "Internal server error"
// @Router       /customers/{id} [delete]
func (s *CustomerHandler) DeleteCustomer(c *gin.Context) {
	entityID, err := http.ParseParamToUInt(c, "id")
	if err != nil {
		http.BadRequest(c, err)
		return
	}

	customerID := customers.NewCustomerID(entityID)
	err = s.service.DeleteCustomer(c.Request.Context(), customerID)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	http.Success(c, nil, "Customer")
}

// RestoreCustomer godoc
// @Summary      Restore customer
// @Description  Restores a soft-deleted customer. Requires admin or receptionist role.
// @Tags         customers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Customer ID"
// @Success      200  {object}  http.APIResponse  "Customer restored"
// @Failure      400  {object}  http.APIResponse  "Invalid ID"
// @Failure      401  {object}  http.APIResponse  "Unauthorized"
// @Failure      500  {object}  http.APIResponse  "Internal server error"
// @Router       /customers/{id}/restore [post]
func (s *CustomerHandler) RestoreCustomer(c *gin.Context) {
	entityID, err := http.ParseParamToUInt(c, "id")
	if err != nil {
		http.BadRequest(c, err)
		return
	}

	customerID := customers.NewCustomerID(entityID)
	err = s.service.RestoreCustomer(c.Request.Context(), customerID)
	if err != nil {
		http.ApplicationError(c, err)
		return
	}

	http.Success(c, nil, "Customer")
}
