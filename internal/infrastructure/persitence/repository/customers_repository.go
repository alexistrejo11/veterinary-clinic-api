package repository

import (
	"clinic-vet-api/database/models"
	"clinic-vet-api/database/sqlc"
	"clinic-vet-api/internal/core/customers"
	"clinic-vet-api/internal/shared"
	"clinic-vet-api/internal/shared/mapper"
	"clinic-vet-api/internal/shared/page"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	customErr "clinic-vet-api/internal/shared/errors"
)

type CustomersSqlcRepository struct {
	queries *sqlc.Queries
	pgMap   *mapper.SqlcFieldMapper
}

func NewCustomersSqlcRepository(queries *sqlc.Queries) customers.CustomerRepository {
	return &CustomersSqlcRepository{
		queries: queries,
		pgMap:   mapper.NewSqlcFieldMapper(),
	}
}

func (r *CustomersSqlcRepository) FindBySpecification(ctx context.Context, spec customers.CustomerSpecification) (page.Page[customers.Customer], error) {
	userFilter := int32(0)
	if len(spec.UserIDs) > 0 {
		userFilter = int32(spec.UserIDs[0])
	}
	isActiveFilter := int32(-1) // -1 = any, 0 = false, 1 = true
	if spec.IsActive != nil {
		if *spec.IsActive {
			isActiveFilter = 1
		} else {
			isActiveFilter = 0
		}
	}
	searchTerm := ""
	if spec.SearchTerm != nil {
		searchTerm = *spec.SearchTerm
	}
	params := sqlc.FindCustomersBySpecParams{
		Column1: userFilter,
		Column2: isActiveFilter,
		Column3: searchTerm,
		Limit:   int32(spec.Pagination.Size),
		Offset:  spec.Pagination.Offset(),
	}
	rows, err := r.queries.FindCustomersBySpec(ctx, params)
	if err != nil {
		return page.Page[customers.Customer]{}, r.dbError(OpSelect, ErrMsgFindCustomersBySpec, err)
	}
	total, err := r.queries.CountCustomersBySpec(ctx, sqlc.CountCustomersBySpecParams{
		Column1: userFilter,
		Column2: isActiveFilter,
		Column3: searchTerm,
	})
	if err != nil {
		return page.Page[customers.Customer]{}, r.dbError(OpCount, ErrMsgFindCustomersBySpec, err)
	}
	items := make([]customers.Customer, len(rows))
	for i := range rows {
		items[i] = r.toEntity(rows[i])
	}
	pagReq := page.PaginationRequest{
		Page:     int32(spec.Pagination.Number),
		PageSize: int32(spec.Pagination.Size),
	}
	return page.NewPage(items, total, pagReq), nil
}

func (r *CustomersSqlcRepository) FindByID(ctx context.Context, id customers.CustomerID) (customers.Customer, error) {
	row, err := r.queries.GetCustomerByID(ctx, id.Int32())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return customers.Customer{}, r.notFoundError("id", id.String())
		}
		return customers.Customer{}, r.dbError(OpSelect, ErrMsgFindCustomerByID, err)
	}
	return r.toEntity(row), nil
}

func (r *CustomersSqlcRepository) FindActive(ctx context.Context, pagination page.Pagination) (page.Page[customers.Customer], error) {
	params := sqlc.FindActiveCustomersParams{
		Limit:  pagination.Limit(),
		Offset: pagination.Offset(),
	}
	rows, err := r.queries.FindActiveCustomers(ctx, params)
	if err != nil {
		return page.Page[customers.Customer]{}, r.dbError(OpSelect, ErrMsgFindCustomersBySpec, err)
	}
	total, err := r.queries.CountActiveCustomers(ctx)
	if err != nil {
		return page.Page[customers.Customer]{}, r.dbError(OpCount, ErrMsgFindCustomersBySpec, err)
	}
	items := make([]customers.Customer, len(rows))
	for i := range rows {
		items[i] = r.toEntity(rows[i])
	}
	pagReq := page.PaginationRequest{
		Page:     int32(pagination.Number),
		PageSize: pagination.Limit(),
	}
	return page.NewPage(items, total, pagReq), nil
}

func (r *CustomersSqlcRepository) ExistsByID(ctx context.Context, id customers.CustomerID) (bool, error) {
	exists, err := r.queries.ExistsCustomerByID(ctx, id.Int32())
	if err != nil {
		return false, r.dbError(OpSelect, "failed to check customer existence", err)
	}
	return exists, nil
}

func (r *CustomersSqlcRepository) Save(ctx context.Context, customer customers.Customer) (customers.Customer, error) {
	if customer.ID.IsZero() {
		return r.create(ctx, customer)
	}
	return r.update(ctx, customer)
}

func (r *CustomersSqlcRepository) SoftDelete(ctx context.Context, id customers.CustomerID) error {
	if err := r.queries.SoftDeleteCustomer(ctx, id.Int32()); err != nil {
		return r.dbError(OpDelete, "failed to soft delete customer", err)
	}
	return nil
}

func (r *CustomersSqlcRepository) HardDelete(ctx context.Context, id customers.CustomerID) error {
	if err := r.queries.HardDeleteCustomer(ctx, id.Int32()); err != nil {
		return r.dbError(OpDelete, "failed to hard delete customer", err)
	}
	return nil
}

func (r *CustomersSqlcRepository) IsDeletedByID(ctx context.Context, id customers.CustomerID) (bool, error) {
	result, err := r.queries.IsDeletedCustomerByID(ctx, id.Int32())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, r.dbError(OpSelect, "failed to check if customer is deleted", err)
	}
	deleted, _ := result.(bool)
	return deleted, nil
}

func (r *CustomersSqlcRepository) ExistsByUserID(ctx context.Context, userID uint) (bool, error) {
	_, err := r.queries.GetCustomerByUserID(ctx, pgtype.Int4{Int32: int32(userID), Valid: true})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, r.dbError(OpSelect, "failed to check customer by user ID", err)
	}
	return true, nil
}

func (r *CustomersSqlcRepository) FindByUserID(ctx context.Context, userID uint) (customers.Customer, error) {
	row, err := r.queries.GetCustomerByUserID(ctx, pgtype.Int4{Int32: int32(userID), Valid: true})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return customers.Customer{}, r.notFoundError("user_id", fmt.Sprintf("%d", userID))
		}
		return customers.Customer{}, r.dbError(OpSelect, "failed to find customer by user ID", err)
	}
	return r.toEntity(row), nil
}

func (r *CustomersSqlcRepository) RestoreByID(ctx context.Context, id customers.CustomerID) error {
	if err := r.queries.RestoreCustomer(ctx, id.Int32()); err != nil {
		return r.dbError(OpUpdate, "failed to restore customer", err)
	}
	return nil
}

func (r *CustomersSqlcRepository) CountAll(ctx context.Context) (int64, error) {
	count, err := r.queries.CountAllCustomers(ctx)
	if err != nil {
		return 0, r.dbError(OpCount, "failed to count all customers", err)
	}
	return count, nil
}

func (r *CustomersSqlcRepository) CountActive(ctx context.Context) (int64, error) {
	count, err := r.queries.CountActiveCustomers(ctx)
	if err != nil {
		return 0, r.dbError(OpCount, "failed to count active customers", err)
	}
	return count, nil
}

func (r *CustomersSqlcRepository) create(ctx context.Context, customer customers.Customer) (customers.Customer, error) {
	params := sqlc.CreateCustomerParams{
		Photo:       customer.PhotoURL,
		FirstName:   customer.FirstName,
		LastName:    customer.LastName,
		Gender:      models.PersonGender(customer.Gender),
		UserID:      r.pgMap.PgInt4.FromUint(customer.UserID),
		IsActive:    customer.IsActive,
		DateOfBirth: r.pgMap.PgDate.FromTime(customer.DateOfBirth),
	}
	created, err := r.queries.CreateCustomer(ctx, params)
	if err != nil {
		return customers.Customer{}, r.dbError(OpInsert, ErrMsgCreateCustomer, err)
	}
	return r.toEntity(created), nil
}

func (r *CustomersSqlcRepository) update(ctx context.Context, customer customers.Customer) (customers.Customer, error) {
	params := sqlc.UpdateCustomerParams{
		ID:          customer.ID.Int32(),
		Photo:       customer.PhotoURL,
		FirstName:   customer.FirstName,
		LastName:    customer.LastName,
		Gender:      models.PersonGender(customer.Gender),
		UserID:      r.pgMap.PgInt4.FromUint(customer.UserID),
		IsActive:    customer.IsActive,
		DateOfBirth: r.pgMap.PgDate.FromTime(customer.DateOfBirth),
	}
	if err := r.queries.UpdateCustomer(ctx, params); err != nil {
		return customers.Customer{}, r.dbError(OpUpdate, fmt.Sprintf("%s with ID %d", ErrMsgUpdateCustomer, customer.ID.Value()), err)
	}
	// Re-fetch to return full entity with timestamps
	updated, err := r.queries.GetCustomerByID(ctx, customer.ID.Int32())
	if err != nil {
		return customers.Customer{}, r.dbError(OpSelect, ErrMsgFindCustomerByID, err)
	}
	return r.toEntity(updated), nil
}

func (r *CustomersSqlcRepository) toEntity(row sqlc.Customer) customers.Customer {
	c := customers.Customer{
		PhotoURL: row.Photo,
		UserID:   r.pgMap.PgInt4.ToUint(row.UserID),
		IsActive: row.IsActive,
		Pets:     nil,
	}
	c.SetID(customers.NewCustomerID(uint(row.ID)))
	c.FirstName = row.FirstName
	c.LastName = row.LastName
	c.Gender = shared.PersonGender(row.Gender)
	c.DateOfBirth = r.pgMap.PgDate.ToTime(row.DateOfBirth)
	createdAt := time.Time{}
	if row.CreatedAt.Valid {
		createdAt = row.CreatedAt.Time
	}
	updatedAt := time.Time{}
	if row.UpdatedAt.Valid {
		updatedAt = row.UpdatedAt.Time
	}
	c.SetTimeStamps(createdAt, updatedAt)
	return c
}

func (r *CustomersSqlcRepository) dbError(operation, message string, err error) error {
	return customErr.DatabaseError(operation, TableCustomers, DriverSQL, fmt.Errorf("%s: %w", message, err))
}

func (r *CustomersSqlcRepository) notFoundError(parameterName, parameterValue string) error {
	return customErr.DBNotFoundError(parameterName, parameterValue, OpSelect, TableCustomers, DriverSQL)
}
