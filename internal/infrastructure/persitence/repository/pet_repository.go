package repository

import (
	"clinic-vet-api/database/sqlc"
	"clinic-vet-api/internal/core/pets"
	customErr "clinic-vet-api/internal/shared/errors"
	"clinic-vet-api/internal/shared/mapper"
	"clinic-vet-api/internal/shared/page"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

type PetSqlcRepository struct {
	queries *sqlc.Queries
	pgMap   *mapper.SqlcFieldMapper
}

func NewPetSqlcRepository(queries *sqlc.Queries) pets.PetRepository {
	return &PetSqlcRepository{
		queries: queries,
		pgMap:   mapper.NewSqlcFieldMapper(),
	}
}

// ============================================================================
// FIND OPERATIONS
// ============================================================================

func (r *PetSqlcRepository) FindByID(ctx context.Context, id pets.PetID) (pets.Pet, error) {
	sqlRow, err := r.queries.FindPetByID(ctx, id.Int32())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return pets.Pet{}, r.notFoundError("id", id.String())
		}
		return pets.Pet{}, r.dbError(OpSelect, ErrMsgFindPetByID, err)
	}
	return r.rowToEntity(sqlRow), nil
}

func (r *PetSqlcRepository) FindBySpecification(ctx context.Context, spec *pets.PetSpecification) (page.Page[pets.Pet], error) {
	if spec == nil {
		spec = &pets.PetSpecification{}
	}

	ids := make([]int32, 0, len(spec.IDs))
	for _, id := range spec.IDs {
		ids = append(ids, id.Int32())
	}

	customerIDs := make([]int32, 0, len(spec.CustomerIDs))
	for _, cid := range spec.CustomerIDs {
		customerIDs = append(customerIDs, int32(cid))
	}

	species := make([]string, 0, len(spec.Species))
	for _, sp := range spec.Species {
		species = append(species, sp.String())
	}

	genders := make([]string, 0, len(spec.Genders))
	for _, g := range spec.Genders {
		genders = append(genders, g.String())
	}

	isActive := pgtype.Bool{Valid: false}
	if spec.IsActive != nil {
		isActive = pgtype.Bool{Bool: *spec.IsActive, Valid: true}
	}

	search := ""
	if spec.SearchTerm != nil {
		search = *spec.SearchTerm
	}

	// pagination defaults
	pageNumber := spec.Pagination.Number
	if pageNumber < 1 {
		pageNumber = page.DefaultPage
	}
	pageSize := spec.Pagination.Size
	if pageSize < 1 {
		pageSize = page.DefaultPageSize
	} else if pageSize > page.MaxPageSize {
		pageSize = page.MaxPageSize
	}
	limit := int32(pageSize)
	offset := int32((pageNumber - 1) * pageSize)

	countParams := sqlc.CountPetsBySpecificationParams{
		Ids:         ids,
		CustomerIds: customerIDs,
		Species:     species,
		Genders:     genders,
		IsActive:    isActive,
		Search:      search,
	}
	totalCount, err := r.queries.CountPetsBySpecification(ctx, countParams)
	if err != nil {
		return page.Page[pets.Pet]{}, r.dbError(OpCount, ErrMsgCountPets, err)
	}

	findParams := sqlc.FindPetsBySpecificationParams{
		Ids:         ids,
		CustomerIds: customerIDs,
		Species:     species,
		Genders:     genders,
		IsActive:    isActive,
		Search:      search,
		PageLimit:   limit,
		PageOffset:  offset,
	}

	sqlRows, err := r.queries.FindPetsBySpecification(ctx, findParams)
	if err != nil {
		return page.Page[pets.Pet]{}, r.dbError(OpSelect, ErrMsgFindPetsBySpecification, err)
	}

	req := page.PaginationRequest{
		Page:          int32(pageNumber),
		PageSize:      int32(pageSize),
		SortDirection: page.SortDirection(spec.Pagination.SortDir),
		OrderBy:       spec.Pagination.OrderBy,
	}

	return page.NewPage(r.rowsToEntities(sqlRows), totalCount, req), nil
}

func (r *PetSqlcRepository) FindAllByCustomerID(ctx context.Context, customerID uint) ([]pets.Pet, error) {
	sqlRows, err := r.queries.FindAllPetsByCustomerID(ctx, int32(customerID))
	if err != nil {
		return nil, r.dbError(OpSelect, ErrMsgFindPetsByCustomerID, err)
	}
	return r.rowsToEntities(sqlRows), nil
}

func (r *PetSqlcRepository) FindByCustomerID(ctx context.Context, customerID uint, pagination page.Pagination) (page.Page[pets.Pet], error) {
	params := sqlc.FindPetsByCustomerIDParams{
		CustomerID: int32(customerID),
		Limit:      pagination.Limit(),
		Offset:     pagination.Offset(),
	}
	sqlRows, err := r.queries.FindPetsByCustomerID(ctx, params)
	if err != nil {
		return page.Page[pets.Pet]{}, r.dbError(OpSelect, ErrMsgFindPetsByCustomerID, err)
	}

	req := page.PaginationRequest{
		Page:          int32(pagination.Number),
		PageSize:      int32(pagination.Size),
		SortDirection: page.SortDirection(pagination.SortDir),
		OrderBy:       pagination.OrderBy,
	}

	total, err := r.queries.CountPetsByCustomerID(ctx, int32(customerID))
	if err != nil {
		return page.Page[pets.Pet]{}, r.dbError(OpCount, ErrMsgCountPets, err)
	}
	petPage := page.NewPage(r.rowsToEntities(sqlRows), total, req)
	return petPage, nil
}

func (r *PetSqlcRepository) FindByIDAndCustomerID(ctx context.Context, id pets.PetID, customerID uint) (pets.Pet, error) {
	params := sqlc.FindPetByIDAndCustomerIDParams{
		ID:         id.Int32(),
		CustomerID: int32(customerID),
	}

	sqlRow, err := r.queries.FindPetByIDAndCustomerID(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return pets.Pet{}, r.notFoundError("id and customer_id", fmt.Sprintf("%s and %d", id.String(), customerID))
		}
		return pets.Pet{}, r.dbError(OpSelect, ErrMsgFindPetByIDAndCustomerID, err)
	}

	return r.rowToEntity(sqlRow), nil
}

func (r *PetSqlcRepository) ExistsByID(ctx context.Context, id pets.PetID) (bool, error) {
	return r.queries.ExistsPetByID(ctx, id.Int32())
}

func (r *PetSqlcRepository) ExistsByMicrochip(ctx context.Context, microchip string) (bool, error) {
	return r.queries.ExistsPetByMicrochip(ctx, r.pgMap.PgText.FromString(microchip))
}

func (r *PetSqlcRepository) ExistsDeletedByID(ctx context.Context, id pets.PetID) (bool, error) {
	return r.queries.ExistsDeletedPetByID(ctx, id.Int32())
}

// ============================================================================
// SAVE, DELETE & RESTORE OPERATIONS
// ============================================================================

func (r *PetSqlcRepository) Save(ctx context.Context, pet pets.Pet) (pets.Pet, error) {
	if pet.ID.IsZero() {
		return r.create(ctx, pet)
	}
	return r.update(ctx, pet)
}

func (r *PetSqlcRepository) Delete(ctx context.Context, petID pets.PetID, isHard bool) error {
	if isHard {
		return r.hardDelete(ctx, petID)
	}
	return r.softDelete(ctx, petID)
}

func (r *PetSqlcRepository) RestoreByID(ctx context.Context, petID pets.PetID) error {
	return r.queries.RestorePet(ctx, petID.Int32())
}

func (r *PetSqlcRepository) CountByCustomerID(ctx context.Context, customerID uint) (int64, error) {
	return r.queries.CountPetsByCustomerID(ctx, int32(customerID))
}

func (r *PetSqlcRepository) create(ctx context.Context, pet pets.Pet) (pets.Pet, error) {
	sqlRow, err := r.queries.CreatePet(ctx, pet.ToCreateParams())
	if err != nil {
		return pets.Pet{}, r.dbError(OpInsert, ErrMsgCreatePet, err)
	}
	return r.rowToEntity(sqlRow), nil
}

func (r *PetSqlcRepository) update(ctx context.Context, pet pets.Pet) (pets.Pet, error) {
	sqlRow, err := r.queries.UpdatePet(ctx, pet.ToUpdateParams())
	if err != nil {
		return pets.Pet{}, r.dbError(OpUpdate, ErrMsgUpdatePet, err)
	}
	return r.rowToEntity(sqlRow), nil
}

func (r *PetSqlcRepository) hardDelete(ctx context.Context, petID pets.PetID) error {
	return r.queries.HardDeletePet(ctx, petID.Int32())
}

func (r *PetSqlcRepository) softDelete(ctx context.Context, petID pets.PetID) error {
	if err := r.queries.SoftDeletePet(ctx, petID.Int32()); err != nil {
		return r.dbError(OpDelete, ErrMsgSoftDeletePet, err)
	}
	return nil
}

// ============================================================================
// ERROR HANDLING METHODS
// ============================================================================

func (r *PetSqlcRepository) dbError(operation, message string, err error) error {
	return customErr.DatabaseError(operation, TablePets, DriverSQL, fmt.Errorf("%s: %v", message, err))
}

func (r *PetSqlcRepository) notFoundError(parameterName, parameterValue string) error {
	return customErr.DBNotFoundError(parameterName, parameterValue, OpSelect, TablePets, DriverSQL)
}

func (r *PetSqlcRepository) wrapConversionError(err error) error {
	return customErr.WrapError(context.Background(), err, OpSelect, TablePets, DriverSQL, ErrMsgConvertPetToDomain)
}

func (r *PetSqlcRepository) rowToEntity(row sqlc.Pet) pets.Pet {
	pet := pets.Pet{
		Name:                  row.Name,
		Photo:                 r.pgMap.PgText.ToStringPtr(row.Photo),
		Species:               pets.PetSpecies(row.Species),
		Breed:                 r.pgMap.PgText.ToStringPtr(row.Breed),
		Age:                   r.pgMap.PgInt2.ToIntPtr(row.Age),
		Gender:                pets.PetGender(r.pgMap.PgText.ToString(row.Gender)),
		Color:                 r.pgMap.PgText.ToStringPtr(row.Color),
		MicrochipID:           r.pgMap.PgText.ToStringPtr(row.Microchip),
		BloodType:             r.pgMap.PgText.ToStringPtr(row.BloodType),
		IsNeutered:            r.pgMap.PgBool.ToBoolPtr(row.IsNeutered),
		CustomerID:            uint(row.CustomerID),
		IsActive:              row.IsActive,
		Allergies:             r.pgMap.PgText.ToStringPtr(row.Allergies),
		CurrentMedications:    r.pgMap.PgText.ToStringPtr(row.CurrentMedications),
		SpecialNeeds:          r.pgMap.PgText.ToStringPtr(row.SpecialNeeds),
		FeedingInstructions:   r.pgMap.PgText.ToStringPtr(row.FeedingInstructions),
		BehavioralNotes:       r.pgMap.PgText.ToStringPtr(row.BehavioralNotes),
		VeterinaryContact:     r.pgMap.PgText.ToStringPtr(row.VeterinaryContact),
		EmergencyContactName:  r.pgMap.PgText.ToStringPtr(row.EmergencyContactName),
		EmergencyContactPhone: r.pgMap.PgText.ToStringPtr(row.EmergencyContactPhone),
	}
	pet.SetID(pets.NewPetID(uint(row.ID)))
	pet.SetTimeStamps(r.pgMap.PgTimestamptz.ToTime(row.CreatedAt), r.pgMap.PgTimestamptz.ToTime(row.UpdatedAt))
	return pet
}

func (r *PetSqlcRepository) rowsToEntities(rows []sqlc.Pet) []pets.Pet {
	entities := make([]pets.Pet, len(rows))
	for i, row := range rows {
		entities[i] = r.rowToEntity(row)
	}
	return entities
}
