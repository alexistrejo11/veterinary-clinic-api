package addresses

import (
	"clinic-vet-api/internal/shared"
	"clinic-vet-api/internal/shared/page"
	"context"
	"errors"
)

var (
	MAX_ADDRESS_PER_CUSTOMER = 3
)

type CreateAddressCommand struct {
	UserID              shared.UserID
	Street              string
	City                string
	State               string
	ZipCode             string
	Country             Country
	BuildingType        BuildingType
	BuildingOuterNumber string
	BuildingInnerNumber *string
	IsDefault           bool
}

type UpdateAddressCommand struct {
	ID                  AddressID
	UserID              *shared.UserID
	Street              *string
	City                *string
	State               *string
	ZipCode             *string
	Country             *Country
	BuildingType        *BuildingType
	BuildingOuterNumber *string
	BuildingInnerNumber *string
	IsDefault           *bool
}

type AddressService interface {
	CreateAddress(ctx context.Context, cmd CreateAddressCommand) (Address, error)
	UpdateAddress(ctx context.Context, cmd UpdateAddressCommand) error
	RestoreAddress(ctx context.Context, id AddressID) error
	DeleteAddress(ctx context.Context, id AddressID, userID *shared.UserID) error
	GetAddressByIDAndUserID(ctx context.Context, id AddressID, userID *shared.UserID) (Address, error)
	GetAddressesByUserID(ctx context.Context, userID shared.UserID) ([]Address, error)
	GetAddressesBySpecification(ctx context.Context, spec AddressSpecification) (page.Page[Address], error)
}

type addressService struct {
	repository AddressRepository
}

func NewAddressService(repository AddressRepository) AddressService {
	return &addressService{repository: repository}
}

func (s *addressService) CreateAddress(ctx context.Context, cmd CreateAddressCommand) (Address, error) {
	const op = "CreateAddress"
	count, err := s.repository.CountByUserID(ctx, cmd.UserID)
	if err != nil {
		return Address{}, err
	}
	if count >= int64(MAX_ADDRESS_PER_CUSTOMER) {
		return Address{}, MaxAddressesPerCustomerError(ctx, cmd.UserID, MAX_ADDRESS_PER_CUSTOMER, op)
	}

	address := Address{
		UserID:              cmd.UserID,
		Street:              cmd.Street,
		City:                cmd.City,
		State:               cmd.State,
		ZipCode:             cmd.ZipCode,
		Country:             cmd.Country,
		BuildingType:        cmd.BuildingType,
		BuildingOuterNumber: cmd.BuildingOuterNumber,
		BuildingInnerNumber: cmd.BuildingInnerNumber,
		IsDefault:           cmd.IsDefault,
	}

	if err := address.Validate(ctx); err != nil {
		return Address{}, err
	}

	created, err := s.repository.Save(ctx, address)
	if err != nil {
		return Address{}, err
	}

	return created, nil
}

func (s *addressService) UpdateAddress(ctx context.Context, cmd UpdateAddressCommand) error {
	const op = "UpdateAddress"
	address, err := s.repository.GetByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	originalWasDefault := address.IsDefault
	if cmd.Street != nil {
		address.Street = *cmd.Street
	}
	if cmd.City != nil {
		address.City = *cmd.City
	}
	if cmd.State != nil {
		address.State = *cmd.State
	}
	if cmd.ZipCode != nil {
		address.ZipCode = *cmd.ZipCode
	}
	if cmd.Country != nil {
		address.Country = *cmd.Country
	}
	if cmd.BuildingType != nil {
		address.BuildingType = *cmd.BuildingType
	}
	if cmd.BuildingOuterNumber != nil {
		address.BuildingOuterNumber = *cmd.BuildingOuterNumber
	}
	if cmd.BuildingInnerNumber != nil {
		address.BuildingInnerNumber = cmd.BuildingInnerNumber
	}

	// Manage default flag change
	if cmd.IsDefault != nil {
		address.IsDefault = *cmd.IsDefault
	}

	if err := address.Validate(ctx); err != nil {
		return err
	}

	// Do not persist an invalid state before checking whether another default exists.
	if cmd.IsDefault != nil && !*cmd.IsDefault && originalWasDefault {
		addresses, err := s.repository.GetAllByUserID(ctx, address.UserID)
		if err != nil {
			return err
		}
		hasOtherDefault := false
		for _, candidate := range addresses {
			if candidate.ID.Value() != address.ID.Value() && candidate.IsDefault {
				hasOtherDefault = true
				break
			}
		}
		if !hasOtherDefault {
			return DefaultAddressRequiredError(ctx, address.UserID, op)
		}
	}

	_, err = s.repository.Save(ctx, address)
	if err != nil {
		return err
	}

	return nil
}

func (s *addressService) RestoreAddress(ctx context.Context, id AddressID) error {
	return s.repository.RestoreByID(ctx, id)
}

func AddressNotFoundError(ctx context.Context, id AddressID) error {
	return errors.New("address not found")
}

func (s *addressService) DeleteAddress(ctx context.Context, id AddressID, optUserID *shared.UserID) error {
	if optUserID == nil {
		if address, err := s.repository.GetByID(ctx, id); err != nil {
			return err
		} else {
			return s.repository.Delete(ctx, address.ID)
		}
	}

	address, err := s.repository.GetByIDAndUserID(ctx, id, *optUserID)
	if err != nil {
		return err
	}
	return s.repository.Delete(ctx, address.ID)
}

func (s *addressService) GetAddressByIDAndUserID(ctx context.Context, id AddressID, userID *shared.UserID) (Address, error) {
	if userID == nil {
		return s.repository.GetByID(ctx, id)
	}
	return s.repository.GetByIDAndUserID(ctx, id, *userID)
}

func (s *addressService) GetAddressesByUserID(ctx context.Context, userID shared.UserID) ([]Address, error) {
	return s.repository.GetAllByUserID(ctx, userID)
}

func (s *addressService) GetAddressesBySpecification(ctx context.Context, spec AddressSpecification) (page.Page[Address], error) {
	return s.repository.GetBySpecification(ctx, spec)
}
