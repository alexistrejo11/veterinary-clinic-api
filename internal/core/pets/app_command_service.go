package pets

import (
	"clinic-vet-api/internal/shared"
	"clinic-vet-api/internal/shared/page"
	"context"
	"errors"
)

type Service interface {
	CreatePet(ctx context.Context, cmd CreatePetCommand) (Pet, error)
	UpdatePet(ctx context.Context, cmd UpdatePetCommand) error
	DeactivatePet(ctx context.Context, id PetID) error
	ActivatePet(ctx context.Context, id PetID) error
	RestorePet(ctx context.Context, id PetID) error
	DeletePet(ctx context.Context, cmd DeletePetCommand) error

	GetPetsByCustomerID(ctx context.Context, customerID uint, pagination page.Pagination) (page.Page[Pet], error)
	GetPetByID(ctx context.Context, id PetID) (Pet, error)
	GetPetByIDAndCustomerID(ctx context.Context, id PetID, customerID uint) (Pet, error)
	GetPetBySpecification(ctx context.Context, spec *PetSpecification) (page.Page[Pet], error)
}

type petService struct {
	repository     PetRepository
	userRepository CustomerRepository
}

func NewPetService(repository PetRepository, userRepository CustomerRepository) Service {
	return &petService{repository: repository, userRepository: userRepository}
}

// ============================================================================
// Command Methods
// ============================================================================

func (s *petService) CreatePet(ctx context.Context, cmd CreatePetCommand) (Pet, error) {
	ifExists, err := s.userRepository.ExistsByIDAndActive(ctx, cmd.CustomerID)
	if err != nil {
		return Pet{}, err
	} else if !ifExists {
		return Pet{}, CustomerNotActiveError(ctx, cmd.CustomerID, "CreatePet")
	}

	pet := Pet{
		Entity:      shared.CreateEntity(NewPetID(0)),
		CustomerID:  cmd.CustomerID,
		Name:        cmd.Name,
		Species:     cmd.Species,
		Breed:       cmd.Breed,
		Age:         cmd.Age,
		Gender:      cmd.Gender,
		Photo:       cmd.Photo,
		Color:       cmd.Color,
		MicrochipID: cmd.MicrochipID,
		BloodType:   cmd.BloodType,
		IsNeutered:  cmd.IsNeutered,
		IsActive:    cmd.IsActive,
	}

	if err := pet.Validate(ctx); err != nil {
		return Pet{}, err
	}

	petCreated, err := s.repository.Save(ctx, pet)
	if err != nil {
		return Pet{}, err
	}
	return petCreated, nil
}

func (s *petService) UpdatePet(ctx context.Context, cmd UpdatePetCommand) error {
	var pet Pet
	var err error
	if cmd.CustomerID != nil {
		pet, err = s.repository.FindByIDAndCustomerID(ctx, cmd.PetID, *cmd.CustomerID)
	} else {
		pet, err = s.repository.FindByID(ctx, cmd.PetID)
	}
	if err != nil {
		return err
	}

	if cmd.Name != nil {
		pet.Name = *cmd.Name
	}
	if cmd.Photo != nil {
		pet.Photo = cmd.Photo
	}
	if cmd.Species != nil {
		pet.Species = *cmd.Species
	}
	if cmd.Breed != nil {
		pet.Breed = cmd.Breed
	}
	if cmd.Age != nil {
		pet.Age = cmd.Age
	}
	if cmd.Gender != nil {
		pet.Gender = *cmd.Gender
	}
	if cmd.Color != nil {
		pet.Color = cmd.Color
	}
	if cmd.BloodType != nil {
		pet.BloodType = cmd.BloodType
	}
	if cmd.MicrochipID != nil {
		pet.MicrochipID = cmd.MicrochipID
	}
	if cmd.IsNeutered != nil {
		pet.IsNeutered = cmd.IsNeutered
	}
	if cmd.IsActive != nil {
		pet.IsActive = *cmd.IsActive
	}

	if cmd.CustomerID != nil {
		ifExists, err := s.userRepository.ExistsByIDAndActive(ctx, *cmd.CustomerID)
		if err != nil {
			return err
		} else if !ifExists {
			return CustomerNotActiveError(ctx, *cmd.CustomerID, "UpdatePet")
		}
		pet.CustomerID = *cmd.CustomerID
	}

	if err := pet.Validate(ctx); err != nil {
		return err
	}

	_, err = s.repository.Save(ctx, pet)
	if err != nil {
		return err
	}
	return nil
}

func (s *petService) ActivatePet(ctx context.Context, id PetID) error {
	pet, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := pet.Activate(); err != nil {
		return err
	}
	_, err = s.repository.Save(ctx, pet)
	return err
}

func (s *petService) DeactivatePet(ctx context.Context, id PetID) error {
	pet, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := pet.Deactivate(); err != nil {
		return err
	}
	_, err = s.repository.Save(ctx, pet)
	return err
}

func (s *petService) RestorePet(ctx context.Context, id PetID) error {
	exists, err := s.repository.ExistsDeletedByID(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("pet not found")
	}
	return s.repository.RestoreByID(ctx, id)
}

func (s *petService) DeletePet(ctx context.Context, cmd DeletePetCommand) error {
	if cmd.OptCustomerID != nil {
		_, err := s.repository.FindByIDAndCustomerID(ctx, cmd.PetID, *cmd.OptCustomerID)
		if err != nil {
			return err
		}
	}
	return s.repository.Delete(ctx, cmd.PetID, cmd.IsHardDelete)
}

// ============================================================================
// Query Methods
// ============================================================================

func (s *petService) GetPetsByCustomerID(ctx context.Context, customerID uint, pagination page.Pagination) (page.Page[Pet], error) {
	return s.repository.FindByCustomerID(ctx, customerID, pagination)
}

func (s *petService) GetPetByID(ctx context.Context, id PetID) (Pet, error) {
	return s.repository.FindByID(ctx, id)
}

func (s *petService) GetPetByIDAndCustomerID(ctx context.Context, id PetID, customerID uint) (Pet, error) {
	return s.repository.FindByIDAndCustomerID(ctx, id, customerID)
}

func (s *petService) GetPetBySpecification(ctx context.Context, spec *PetSpecification) (page.Page[Pet], error) {
	return s.repository.FindBySpecification(ctx, spec)
}
