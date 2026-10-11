package addresses

import (
	"clinic-vet-api/internal/shared"
	"clinic-vet-api/internal/shared/page"
	"context"
	"fmt"
	"strings"
)

type Address struct {
	shared.Entity[AddressID]
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

type (
	BuildingType string
	Country      string
	AddressID    shared.IntegerID
)

func NewAddressID(id uint) AddressID {
	return shared.NewBaseID(id)
}

func NewAddressIDEmpty() AddressID {
	return shared.NewBaseIDEmpty()
}

const (
	USA    Country = "USA"
	Mexico Country = "Mexico"
	Canada Country = "Canada"
)

func (c Country) IsValid() bool {
	switch c {
	case USA, Mexico, Canada:
		return true
	}
	return false
}

func ParseCountry(country string) (Country, error) {
	switch strings.ToLower(strings.TrimSpace(country)) {
	case "usa", "us", "united states", "united states of america":
		return USA, nil
	case "mexico":
		return Mexico, nil
	case "canada":
		return Canada, nil
	}
	return Country(""), fmt.Errorf("invalid country")
}

func (c Country) DisplayName() string {
	switch c {
	case USA:
		return "United States"
	case Mexico:
		return "Mexico"
	case Canada:
		return "Canada"
	}

	return "Unknown Country"
}

const (
	BuildingTypeHouse     BuildingType = "house"
	BuildingTypeApartment BuildingType = "apartment"
	BuildingTypeOffice    BuildingType = "office"
	BuildingTypeOther     BuildingType = "other"
)

func ParseBuildingType(buildingType string) (BuildingType, error) {
	switch buildingType {
	case "house":
		return BuildingTypeHouse, nil
	case "apartment":
		return BuildingTypeApartment, nil
	case "office":
		return BuildingTypeOffice, nil
	case "other":
		return BuildingTypeOther, nil
	}
	return BuildingType(""), fmt.Errorf("invalid building type")
}

func (bt BuildingType) DisplayName() string {
	switch bt {
	case BuildingTypeHouse:
		return "House"
	case BuildingTypeApartment:
		return "Apartment"
	}
	return "Unknown Building Type"
}

func (a Address) Validate(ctx context.Context) error {
	const op = "Address.Validate"
	if a.Street == "" {
		return StreetRequiredError(ctx, op)
	}
	if a.City == "" {
		return CityRequiredError(ctx, op)
	}

	if a.State == "" {
		return StateRequiredError(ctx, op)
	}
	if a.ZipCode == "" {
		return ZipRequiredError(ctx, op)
	}
	if a.Country == "" {
		return CountryRequiredError(ctx, op)
	}
	if !a.Country.IsValid() {
		return InvalidCountryError(ctx, a.Country, op)
	}

	switch a.BuildingType {
	case "", BuildingTypeHouse, BuildingTypeApartment, BuildingTypeOffice, BuildingTypeOther:
		// ok
	default:
		return InvalidBuildingTypeError(ctx, a.BuildingType, op)
	}

	return nil
}

type AddressRepository interface {
	GetByID(ctx context.Context, id AddressID) (Address, error)
	RestoreByID(ctx context.Context, id AddressID) error
	Save(ctx context.Context, address Address) (Address, error)
	BulkUpdate(ctx context.Context, addresses []Address) error
	Delete(ctx context.Context, id AddressID) error
	ExistsByID(ctx context.Context, id AddressID) (bool, error)
	GetBySpecification(ctx context.Context, spec AddressSpecification) (page.Page[Address], error)
	CountByUserID(ctx context.Context, userID shared.UserID) (int64, error)
	GetAllByUserID(ctx context.Context, userID shared.UserID) ([]Address, error)
	GetByIDAndUserID(ctx context.Context, id AddressID, userID shared.UserID) (Address, error)
}

type AddressSpecification struct {
	page.Pagination
	// UserID when set filters addresses by user; nil means no filter (e.g. admin listing).
	UserID *uint
}
