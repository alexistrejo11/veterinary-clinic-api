package repository

import (
	"clinic-vet-api/internal/core/addresses"
	"clinic-vet-api/internal/shared"
	"clinic-vet-api/internal/shared/mapper"
	"clinic-vet-api/internal/shared/page"
	"clinic-vet-api/database/sqlc"
	"context"
	"database/sql"
	"errors"
	"fmt"

	customErr "clinic-vet-api/internal/shared/errors"
)

type AddressSqlcRepository struct {
	queries *sqlc.Queries
	pgMap   *mapper.SqlcFieldMapper
}

func NewAddressSqlcRepository(queries *sqlc.Queries) addresses.AddressRepository {
	return &AddressSqlcRepository{
		queries: queries,
		pgMap:   mapper.NewSqlcFieldMapper(),
	}
}

func (r *AddressSqlcRepository) GetByID(
	ctx context.Context, id addresses.AddressID,
) (addresses.Address, error) {
	row, err := r.queries.GetAddressByID(ctx, id.Int32())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return addresses.Address{}, r.notFoundError("id", id.String())
		}
		return addresses.Address{}, r.dbError(OpSelect, ErrMsgFindAddressByID, err)
	}
	return r.toEntity(row), nil
}

func (r *AddressSqlcRepository) RestoreByID(
	ctx context.Context, id addresses.AddressID,
) error {
	if err := r.queries.RestoreAddress(ctx, id.Int32()); err != nil {
		return r.dbError(OpUpdate, ErrMsgRestoreAddress, err)
	}
	return nil
}

func (r *AddressSqlcRepository) Save(
	ctx context.Context, address addresses.Address,
) (addresses.Address, error) {
	if address.ID.IsZero() {
		return r.create(ctx, address)
	}
	return r.update(ctx, address)
}

func (r *AddressSqlcRepository) BulkUpdate(
	ctx context.Context, list []addresses.Address,
) error {
	for i := range list {
		if _, err := r.update(ctx, list[i]); err != nil {
			return err
		}
	}
	return nil
}

func (r *AddressSqlcRepository) Delete(
	ctx context.Context, id addresses.AddressID,
) error {
	if err := r.queries.SoftDeleteAddress(ctx, id.Int32()); err != nil {
		return r.dbError(OpDelete, ErrMsgSoftDeleteAddress, err)
	}
	return nil
}

func (r *AddressSqlcRepository) ExistsByID(
	ctx context.Context, id addresses.AddressID,
) (bool, error) {
	exists, err := r.queries.ExistsAddressByID(ctx, id.Int32())
	if err != nil {
		return false, r.dbError(OpSelect, "failed to check address existence", err)
	}
	return exists, nil
}

func (r *AddressSqlcRepository) GetBySpecification(
	ctx context.Context, spec addresses.AddressSpecification,
) (page.Page[addresses.Address], error) {
	userFilter := int32(0)
	if spec.UserID != nil {
		userFilter = int32(*spec.UserID)
	}
	params := sqlc.FindAddressesBySpecParams{
		Column1: userFilter,
		Limit:   int32(spec.Pagination.Size),
		Offset:  spec.Pagination.Offset(),
	}
	rows, err := r.queries.FindAddressesBySpec(ctx, params)
	if err != nil {
		return page.Page[addresses.Address]{}, r.dbError(OpSelect, "failed to find addresses by specification", err)
	}
	total, err := r.queries.CountAddressesBySpec(ctx, userFilter)
	if err != nil {
		return page.Page[addresses.Address]{}, r.dbError(OpCount, "failed to count addresses by specification", err)
	}
	items := make([]addresses.Address, len(rows))
	for i := range rows {
		items[i] = r.toEntity(rows[i])
	}
	pagReq := page.PaginationRequest{
		Page:     int32(spec.Pagination.Number),
		PageSize: int32(spec.Pagination.Size),
	}
	return page.NewPage(items, total, pagReq), nil
}

func (r *AddressSqlcRepository) CountByUserID(
	ctx context.Context, userID shared.UserID,
) (int64, error) {
	count, err := r.queries.CountAddressesByUserID(ctx, userID.Int32())
	if err != nil {
		return 0, r.dbError(OpCount, ErrMsgFindAddressByUserID, err)
	}
	return count, nil
}

func (r *AddressSqlcRepository) GetAllByUserID(
	ctx context.Context,
	userID shared.UserID,
) ([]addresses.Address, error) {
	rows, err := r.queries.GetAddressesByUserID(ctx, userID.Int32())
	if err != nil {
		return nil, r.dbError(OpSelect, ErrMsgFindAddressByUserID, err)
	}
	items := make([]addresses.Address, len(rows))
	for i := range rows {
		items[i] = r.toEntity(rows[i])
	}
	return items, nil
}

func (r *AddressSqlcRepository) GetByIDAndUserID(
	ctx context.Context,
	id addresses.AddressID,
	userID shared.UserID,
) (addresses.Address, error) {
	row, err := r.queries.GetAddressByIDAndUserID(ctx, sqlc.GetAddressByIDAndUserIDParams{
		ID:     id.Int32(),
		UserID: userID.Int32(),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return addresses.Address{}, r.notFoundError("id", id.String())
		}
		return addresses.Address{}, r.dbError(OpSelect, ErrMsgFindAddressByID, err)
	}
	return r.toEntity(row), nil
}

func (r *AddressSqlcRepository) create(
	ctx context.Context, address addresses.Address,
) (addresses.Address, error) {
	innerNum := ""
	if address.BuildingInnerNumber != nil {
		innerNum = *address.BuildingInnerNumber
	}
	params := sqlc.CreateAddressParams{
		UserID:              address.UserID.Int32(),
		Street:              address.Street,
		City:                address.City,
		State:               address.State,
		ZipCode:             address.ZipCode,
		Country:             string(address.Country),
		BuildingType:        string(address.BuildingType),
		BuildingOuterNumber: address.BuildingOuterNumber,
		BuildingInnerNumber: innerNum,
		IsDefault:           address.IsDefault,
	}
	created, err := r.queries.CreateAddress(ctx, params)
	if err != nil {
		return addresses.Address{}, r.dbError(OpInsert, ErrMsgCreateAddress, err)
	}
	out := r.toEntity(created)
	return out, nil
}

func (r *AddressSqlcRepository) update(
	ctx context.Context, address addresses.Address,
) (addresses.Address, error) {
	innerNum := ""
	if address.BuildingInnerNumber != nil {
		innerNum = *address.BuildingInnerNumber
	}
	params := sqlc.UpdateAddressParams{
		ID:                  address.ID.Int32(),
		Street:              address.Street,
		City:                address.City,
		State:               address.State,
		ZipCode:             address.ZipCode,
		Country:             string(address.Country),
		BuildingType:        string(address.BuildingType),
		BuildingOuterNumber: address.BuildingOuterNumber,
		BuildingInnerNumber: innerNum,
		IsDefault:           address.IsDefault,
	}
	updated, err := r.queries.UpdateAddress(ctx, params)
	if err != nil {
		return addresses.Address{}, r.dbError(OpUpdate, fmt.Sprintf("%s with ID %d", ErrMsgUpdateAddress, address.ID.Value()), err)
	}
	return r.toEntity(sqlc.Address{
		ID:                  updated.ID,
		UserID:              updated.UserID,
		Street:              updated.Street,
		City:                updated.City,
		State:               updated.State,
		ZipCode:             updated.ZipCode,
		Country:             updated.Country,
		BuildingType:        updated.BuildingType,
		BuildingOuterNumber: updated.BuildingOuterNumber,
		BuildingInnerNumber: updated.BuildingInnerNumber,
		IsDefault:           updated.IsDefault,
		CreatedAt:           updated.CreatedAt,
		UpdatedAt:           updated.UpdatedAt,
		DeletedAt:           updated.DeletedAt,
		Version:             updated.Version,
	}), nil
}

func (r *AddressSqlcRepository) toEntity(row sqlc.Address) addresses.Address {
	var innerNum *string
	if row.BuildingInnerNumber != "" {
		innerNum = &row.BuildingInnerNumber
	}
	addr := addresses.Address{
		UserID:              shared.NewUserID(uint(row.UserID)),
		Street:              row.Street,
		City:                row.City,
		State:               row.State,
		ZipCode:             row.ZipCode,
		Country:             addresses.Country(row.Country),
		BuildingType:        addresses.BuildingType(row.BuildingType),
		BuildingOuterNumber: row.BuildingOuterNumber,
		BuildingInnerNumber: innerNum,
		IsDefault:           row.IsDefault,
	}
	addr.SetID(addresses.NewAddressID(uint(row.ID)))
	addr.SetTimeStamps(r.pgMap.PgTimestamptz.ToTime(row.CreatedAt), r.pgMap.PgTimestamptz.ToTime(row.UpdatedAt))
	addr.Version = int(row.Version)
	return addr
}

func (r *AddressSqlcRepository) dbError(operation, message string, err error) error {
	return customErr.DatabaseError(operation, TableAddresses, DriverSQL, fmt.Errorf("%s: %w", message, err))
}

func (r *AddressSqlcRepository) notFoundError(parameterName, parameterValue string) error {
	return customErr.DBNotFoundError(parameterName, parameterValue, OpSelect, TableAddresses, DriverSQL)
}
