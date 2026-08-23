package mappers

import (
	"time"

	"clinic-vet-api/internal/core/customers"
	"clinic-vet-api/internal/infrastructure/http/handlers/dtos"
	"clinic-vet-api/internal/shared"
	"clinic-vet-api/internal/shared/page"
)

type CustomerMapper struct{}

func NewCustomerMapper() *CustomerMapper {
	return &CustomerMapper{}
}

func (m *CustomerMapper) ToCustomerResponse(customer customers.Customer) dtos.CustomerResponse {
	return dtos.CustomerResponse{
		ID:          customer.ID.Value(),
		FirstName:   customer.FirstName,
		LastName:    customer.LastName,
		Gender:      string(customer.Gender),
		DateOfBirth: customer.DateOfBirth,
		PhotoURL:    customer.PhotoURL,
		UserID:      customer.UserID,
		IsActive:    customer.IsActive,
		CreatedAt:   customer.CreatedAt,
		UpdatedAt:   customer.UpdatedAt,
	}
}

func (m *CustomerMapper) ToCustomerResponsePage(customerPage page.Page[customers.Customer]) page.Page[dtos.CustomerResponse] {
	return page.MapItems(customerPage, m.ToCustomerResponse)
}

func (m *CustomerMapper) RequestToSpecification(request dtos.CustomerSearchRequest) customers.CustomerSpecification {
	spec := customers.CustomerSpecification{
		IsActive:   request.IsActive,
		Pagination: request.PaginationRequest.WithDefaults().ToPagination(),
	}
	if request.Search != "" {
		spec.SearchTerm = &request.Search
	}
	return spec
}

func (m *CustomerMapper) RequestToCreateCommand(request dtos.CustomerCreateRequest) (customers.CreateCustomerCommand, error) {
	gender, err := shared.ParseGender(request.Gender)
	if err != nil {
		return customers.CreateCustomerCommand{}, err
	}

	dob, err := time.Parse("2006-01-02", request.DateOfBirth)
	if err != nil {
		return customers.CreateCustomerCommand{}, err
	}

	isActive := true
	if request.IsActive != nil {
		isActive = *request.IsActive
	}

	return customers.CreateCustomerCommand{
		FirstName:   request.FirstName,
		LastName:    request.LastName,
		Email:       request.Email,
		Gender:      gender,
		DateOfBirth: dob,
		PhotoURL:    request.PhotoURL,
		IsActive:    isActive,
		UserID:      request.UserID,
	}, nil
}

func (m *CustomerMapper) RequestToUpdateCommand(request dtos.CustomerUpdateRequest) (customers.UpdateCustomerCommand, error) {
	var gender *shared.PersonGender
	if request.Gender != nil {
		parsed, err := shared.ParseGender(*request.Gender)
		if err != nil {
			return customers.UpdateCustomerCommand{}, err
		}
		gender = &parsed
	}

	var dob *time.Time
	if request.DateOfBirth != nil {
		parsed, err := time.Parse("2006-01-02", *request.DateOfBirth)
		if err != nil {
			return customers.UpdateCustomerCommand{}, err
		}
		dob = &parsed
	}

	cmd := customers.UpdateCustomerCommand{
		ID:          customers.NewCustomerID(request.ID),
		FirstName:   request.FirstName,
		LastName:    request.LastName,
		Gender:      gender,
		DateOfBirth: dob,
		PhotoURL:    request.PhotoURL,
		UserID:      request.UserID,
		IsActive:    request.IsActive,
	}

	return cmd, nil
}
