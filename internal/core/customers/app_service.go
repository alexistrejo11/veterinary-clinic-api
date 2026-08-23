package customers

import (
	"clinic-vet-api/internal/shared"
	"clinic-vet-api/internal/shared/errors"
	"clinic-vet-api/internal/shared/page"
	"context"
	"time"
)

type CreateCustomerCommand struct {
	FirstName   string
	LastName    string
	Email       string
	Gender      shared.PersonGender
	DateOfBirth time.Time
	PhotoURL    string
	IsActive    bool
	UserID      *uint
}

type UpdateCustomerCommand struct {
	ID          CustomerID
	FirstName   *string
	LastName    *string
	Gender      *shared.PersonGender
	DateOfBirth *time.Time
	PhotoURL    *string
	UserID      *uint
	IsActive    *bool
}

type CustomerService interface {
	CreateCustomer(ctx context.Context, cmd CreateCustomerCommand) (Customer, error)
	ActivateCustomerByUserID(ctx context.Context, userID uint) error
	UpdateCustomer(ctx context.Context, cmd UpdateCustomerCommand) error
	RestoreCustomer(ctx context.Context, id CustomerID) error
	DeleteCustomer(ctx context.Context, id CustomerID) error
	GetCustomerByID(ctx context.Context, id CustomerID) (Customer, error)
	GetCustomersBySpecification(ctx context.Context, spec CustomerSpecification) (page.Page[Customer], error)
	GetCustomerStats(ctx context.Context) (CustomerStats, error)
}

type customerService struct {
	repository CustomerRepository
}

func NewCustomerService(repository CustomerRepository) CustomerService {
	return &customerService{repository: repository}
}

func (s *customerService) CreateCustomer(ctx context.Context, cmd CreateCustomerCommand) (Customer, error) {
	const op = "CreateCustomer"

	if cmd.UserID != nil {
		ifExists, err := s.repository.ExistsByUserID(ctx, *cmd.UserID)
		if err != nil {
			return Customer{}, err
		}
		if ifExists {
			return Customer{}, UserAlreadyAssociatedError(ctx, *cmd.UserID, op)
		}
	}

	sharedPerson := shared.Person{
		FirstName:   cmd.FirstName,
		LastName:    cmd.LastName,
		Gender:      cmd.Gender,
		DateOfBirth: cmd.DateOfBirth,
	}
	customer := Customer{
		Person:   sharedPerson,
		PhotoURL: cmd.PhotoURL,
		IsActive: cmd.IsActive,
	}
	if cmd.UserID != nil {
		customer.UserID = *cmd.UserID
	}

	if err := customer.Validate(ctx, op); err != nil {
		return Customer{}, err
	}

	customerCreated, err := s.repository.Save(ctx, customer)
	return customerCreated, err
}

func (s *customerService) ActivateCustomerByUserID(ctx context.Context, userID uint) error {
	customer, err := s.repository.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if err := customer.Activate(ctx); err != nil {
		return err
	}
	_, err = s.repository.Save(ctx, customer)
	return err
}

func (s *customerService) UpdateCustomer(ctx context.Context, cmd UpdateCustomerCommand) error {
	const op = "UpdateCustomer"
	customer, err := s.repository.FindByID(ctx, cmd.ID)

	if err != nil {
		return err
	}

	if cmd.FirstName != nil {
		customer.FirstName = *cmd.FirstName
	}
	if cmd.LastName != nil {
		customer.LastName = *cmd.LastName
	}

	if cmd.Gender != nil {
		customer.Gender = *cmd.Gender
	}
	if cmd.DateOfBirth != nil {
		customer.DateOfBirth = *cmd.DateOfBirth
	}
	if cmd.PhotoURL != nil {
		customer.PhotoURL = *cmd.PhotoURL
	}
	if cmd.UserID != nil {
		// prevent associating a user already bound to another customer
		ifExists, err := s.repository.ExistsByUserID(ctx, *cmd.UserID)
		if err != nil {
			return err
		}
		if ifExists && customer.UserID != *cmd.UserID {
			return UserAlreadyAssociatedError(ctx, *cmd.UserID, op)
		}
		customer.UserID = *cmd.UserID
	}
	if cmd.IsActive != nil {
		customer.IsActive = *cmd.IsActive
	}

	if err := customer.Validate(ctx, op); err != nil {
		return err
	}

	_, err = s.repository.Save(ctx, customer)
	if err != nil {
		return err
	}
	return nil
}

func (s *customerService) RestoreCustomer(ctx context.Context, id CustomerID) error {
	exists, err := s.repository.IsDeletedByID(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return errors.EntityNotFoundError(ctx, "customer", id.String(), "RestoreCustomer")
	}
	return s.repository.RestoreByID(ctx, id)
}

func (s *customerService) DeleteCustomer(ctx context.Context, id CustomerID) error {
	return s.repository.SoftDelete(ctx, id)
}

func (s *customerService) GetCustomerByID(ctx context.Context, id CustomerID) (Customer, error) {
	return s.repository.FindByID(ctx, id)
}

func (s *customerService) GetCustomersBySpecification(ctx context.Context, spec CustomerSpecification) (page.Page[Customer], error) {
	return s.repository.FindBySpecification(ctx, spec)
}

// CustomerStats represents high-level customer statistics.
type CustomerStats struct {
	TotalCustomers    int64 `json:"total_customers"`
	ActiveCustomers   int64 `json:"active_customers"`
	InactiveCustomers int64 `json:"inactive_customers"`
}

func (s *customerService) GetCustomerStats(ctx context.Context) (CustomerStats, error) {
	total, err := s.repository.CountAll(ctx)
	if err != nil {
		return CustomerStats{}, err
	}
	active, err := s.repository.CountActive(ctx)
	if err != nil {
		return CustomerStats{}, err
	}
	return CustomerStats{
		TotalCustomers:    total,
		ActiveCustomers:   active,
		InactiveCustomers: total - active,
	}, nil
}
