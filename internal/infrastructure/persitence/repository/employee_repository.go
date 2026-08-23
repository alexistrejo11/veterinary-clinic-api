package repository

import (
	"clinic-vet-api/database/models"
	"clinic-vet-api/database/sqlc"
	"clinic-vet-api/internal/core/employees"
	"clinic-vet-api/internal/shared"
	"clinic-vet-api/internal/shared/mapper"
	p "clinic-vet-api/internal/shared/page"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	customErr "clinic-vet-api/internal/shared/errors"
)

type EmployeeSqlcRepository struct {
	queries *sqlc.Queries
	pgMap   *mapper.SqlcFieldMapper
}

func NewEmployeeSqlcRepository(queries *sqlc.Queries) employees.EmployeeRepository {
	return &EmployeeSqlcRepository{
		queries: queries,
		pgMap:   mapper.NewSqlcFieldMapper(),
	}
}

func (r *EmployeeSqlcRepository) GetBySpecification(ctx context.Context, spec employees.EmployeeSpecification) (p.Page[employees.Employee], error) {
	pagination := p.FromSpecPagination(spec.Pagination).WithDefaults()
	limit := pagination.Limit()
	offset := pagination.Offset()
	activeFilter := int32(-1)
	if spec.IsActive != nil {
		if *spec.IsActive {
			activeFilter = 1
		} else {
			activeFilter = 0
		}
	}
	var rows []sqlc.Employee
	var total int64
	var err error
	if spec.Specialty != nil {
		rows, err = r.queries.FindEmployeesBySpeciality(ctx, sqlc.FindEmployeesBySpecialityParams{Speciality: models.VeterinarianSpeciality(*spec.Specialty), Column2: activeFilter, Limit: limit, Offset: offset})
		if err == nil {
			total, err = r.queries.CountEmployeesBySpeciality(ctx, sqlc.CountEmployeesBySpecialityParams{Speciality: models.VeterinarianSpeciality(*spec.Specialty), Column2: activeFilter})
		}
	} else {
		rows, err = r.queries.FindEmployees(ctx, sqlc.FindEmployeesParams{Column1: activeFilter, Limit: limit, Offset: offset})
		if err == nil {
			total, err = r.queries.CountEmployeesByActive(ctx, activeFilter)
		}
	}
	if err != nil {
		return p.Page[employees.Employee]{}, r.dbError(OpSelect, ErrMsgFindEmployeesBySpec, err)
	}
	items := make([]employees.Employee, len(rows))
	for i := range rows {
		items[i] = r.toEntity(rows[i])
	}
	pagReq := p.PaginationRequest{
		Page:     pagination.Page,
		PageSize: limit,
	}
	return p.NewPage(items, total, pagReq), nil
}

func (r *EmployeeSqlcRepository) GetByID(ctx context.Context, id employees.EmployeeID) (employees.Employee, error) {
	row, err := r.queries.FindEmployeeByID(ctx, id.Int32())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return employees.Employee{}, r.notFoundError("id", id.String())
		}
		return employees.Employee{}, r.dbError(OpSelect, ErrMsgFindEmployeeByID, err)
	}
	return r.toEntity(row), nil
}

func (r *EmployeeSqlcRepository) GetByUserID(ctx context.Context, userID uint) (employees.Employee, error) {
	row, err := r.queries.FindEmployeeByUserID(ctx, pgtype.Int4{Int32: int32(userID), Valid: true})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return employees.Employee{}, r.notFoundError("user_id", fmt.Sprintf("%d", userID))
		}
		return employees.Employee{}, r.dbError(OpSelect, ErrMsgFindEmployeeByID, err)
	}
	return r.toEntity(row), nil
}

func (r *EmployeeSqlcRepository) GetActive(ctx context.Context, pagination p.Pagination) (p.Page[employees.Employee], error) {
	params := sqlc.FindActiveEmployeesParams{
		Limit:  pagination.Limit(),
		Offset: pagination.Offset(),
	}
	rows, err := r.queries.FindActiveEmployees(ctx, params)
	if err != nil {
		return p.Page[employees.Employee]{}, r.dbError(OpSelect, ErrMsgFindEmployeesBySpec, err)
	}
	total, err := r.queries.CountActiveEmployees(ctx)
	if err != nil {
		return p.Page[employees.Employee]{}, r.dbError(OpCount, ErrMsgFindEmployeesBySpec, err)
	}
	items := make([]employees.Employee, len(rows))
	for i := range rows {
		items[i] = r.toEntity(rows[i])
	}
	pagReq := p.PaginationRequest{
		Page:     int32(pagination.Number),
		PageSize: pagination.Limit(),
	}
	return p.NewPage(items, total, pagReq), nil
}

func (r *EmployeeSqlcRepository) GetBySpeciality(ctx context.Context, speciality employees.VetSpecialty, pagination p.Pagination) (p.Page[employees.Employee], error) {
	params := sqlc.FindEmployeesBySpecialityParams{
		Speciality: models.VeterinarianSpeciality(speciality),
		Column2:    -1,
		Limit:      pagination.Limit(),
		Offset:     pagination.Offset(),
	}
	rows, err := r.queries.FindEmployeesBySpeciality(ctx, params)
	if err != nil {
		return p.Page[employees.Employee]{}, r.dbError(OpSelect, ErrMsgFindEmployeesBySpec, err)
	}
	total, err := r.queries.CountEmployeesBySpeciality(ctx, sqlc.CountEmployeesBySpecialityParams{Speciality: models.VeterinarianSpeciality(speciality), Column2: -1})
	if err != nil {
		return p.Page[employees.Employee]{}, r.dbError(OpCount, ErrMsgFindEmployeesBySpec, err)
	}
	items := make([]employees.Employee, len(rows))
	for i := range rows {
		items[i] = r.toEntity(rows[i])
	}
	pagReq := p.PaginationRequest{
		Page:     int32(pagination.Number),
		PageSize: pagination.Limit(),
	}
	return p.NewPage(items, total, pagReq), nil
}

func (r *EmployeeSqlcRepository) ExistsByID(ctx context.Context, id employees.EmployeeID) (bool, error) {
	exists, err := r.queries.ExistsEmployeeByID(ctx, id.Int32())
	if err != nil {
		return false, r.dbError(OpSelect, "failed to check employee existence", err)
	}
	return exists, nil
}

func (r *EmployeeSqlcRepository) ExistsByUserID(ctx context.Context, userID uint) (bool, error) {
	exists, err := r.queries.ExistsEmployeeByUserID(ctx, pgtype.Int4{Int32: int32(userID), Valid: true})
	if err != nil {
		return false, r.dbError(OpSelect, "failed to check employee by user ID", err)
	}
	return exists, nil
}

func (r *EmployeeSqlcRepository) Save(ctx context.Context, employee *employees.Employee) error {
	if employee.ID.IsZero() {
		return r.create(ctx, employee)
	}
	return r.update(ctx, employee)
}

func (r *EmployeeSqlcRepository) SoftDelete(ctx context.Context, id employees.EmployeeID) error {
	if err := r.queries.SoftDeleteEmployee(ctx, id.Int32()); err != nil {
		return r.dbError(OpDelete, "failed to soft delete employee", err)
	}
	return nil
}

func (r *EmployeeSqlcRepository) HardDelete(ctx context.Context, id employees.EmployeeID) error {
	if err := r.queries.HardDeleteEmployee(ctx, id.Int32()); err != nil {
		return r.dbError(OpDelete, "failed to hard delete employee", err)
	}
	return nil
}

func (r *EmployeeSqlcRepository) RestoreByID(ctx context.Context, id employees.EmployeeID) error {
	if _, err := r.queries.RestoreEmployee(ctx, id.Int32()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return r.notFoundError("id", id.String())
		}
		return r.dbError(OpUpdate, "failed to restore employee", err)
	}
	return nil
}

func (r *EmployeeSqlcRepository) CountBySpeciality(ctx context.Context, speciality employees.VetSpecialty) (int64, error) {
	count, err := r.queries.CountEmployeesBySpeciality(ctx, sqlc.CountEmployeesBySpecialityParams{Speciality: models.VeterinarianSpeciality(speciality), Column2: -1})
	if err != nil {
		return 0, r.dbError(OpCount, "failed to count employees by speciality", err)
	}
	return count, nil
}

func (r *EmployeeSqlcRepository) CountByActive(ctx context.Context, active *bool) (int64, error) {
	filter := int32(-1)
	if active != nil {
		if *active {
			filter = 1
		} else {
			filter = 0
		}
	}
	count, err := r.queries.CountEmployeesByActive(ctx, filter)
	if err != nil {
		return 0, r.dbError(OpCount, "failed to count employees by active status", err)
	}
	return count, nil
}

func (r *EmployeeSqlcRepository) CountActive(ctx context.Context) (int64, error) {
	count, err := r.queries.CountActiveEmployees(ctx)
	if err != nil {
		return 0, r.dbError(OpCount, "failed to count active employees", err)
	}
	return count, nil
}

func (r *EmployeeSqlcRepository) create(ctx context.Context, employee *employees.Employee) error {
	scheduleJSON, _ := json.Marshal(employee.Schedule)
	params := sqlc.CreateEmployeeParams{
		FirstName:         employee.FirstName,
		LastName:          employee.LastName,
		Photo:             employee.Photo,
		LicenseNumber:     employee.LicenseNumber,
		Speciality:        models.VeterinarianSpeciality(employee.Specialty),
		YearsOfExperience: int32(employee.YearsExperience),
		IsActive:          employee.IsActive,
		UserID:            r.pgMap.PgInt4.FromUint(employee.UserID),
		Column9:           scheduleJSON,
		Gender:            models.PersonGender(employee.Gender),
		DateOfBirth:       r.pgMap.PgDate.FromTime(employee.DateOfBirth),
	}
	created, err := r.queries.CreateEmployee(ctx, params)
	if err != nil {
		return r.dbError(OpInsert, ErrMsgCreateEmployee, err)
	}
	employee.SetID(employees.NewEmployeeID(uint(created.ID)))
	employee.SetTimeStamps(r.pgMap.PgTimestamptz.ToTime(created.CreatedAt), r.pgMap.PgTimestamptz.ToTime(created.UpdatedAt))
	return nil
}

func (r *EmployeeSqlcRepository) update(ctx context.Context, employee *employees.Employee) error {
	var scheduleJSON []byte
	if employee.Schedule != nil {
		scheduleJSON, _ = json.Marshal(employee.Schedule)
	}
	params := sqlc.UpdateEmployeeParams{
		ID:                employee.ID.Int32(),
		FirstName:         employee.FirstName,
		LastName:          employee.LastName,
		Photo:             employee.Photo,
		LicenseNumber:     employee.LicenseNumber,
		Speciality:        models.VeterinarianSpeciality(employee.Specialty),
		YearsOfExperience: int32(employee.YearsExperience),
		IsActive:          employee.IsActive,
		UserID:            r.pgMap.PgInt4.FromUint(employee.UserID),
		Column10:          scheduleJSON,
		Gender:            models.PersonGender(employee.Gender),
		DateOfBirth:       r.pgMap.PgDate.FromTime(employee.DateOfBirth),
	}
	_, err := r.queries.UpdateEmployee(ctx, params)
	if err != nil {
		return r.dbError(OpUpdate, fmt.Sprintf("%s with ID %d", ErrMsgUpdateEmployee, employee.ID.Value()), err)
	}
	// Refresh timestamps from DB if needed; for simplicity we don't re-fetch
	employee.IncrementVersion()
	return nil
}

func (r *EmployeeSqlcRepository) toEntity(row sqlc.Employee) employees.Employee {
	specialty, _ := employees.ParseVetSpecialty(string(row.Speciality))
	emp := employees.Employee{
		Photo:           row.Photo,
		LicenseNumber:   row.LicenseNumber,
		Specialty:       specialty,
		YearsExperience: int(row.YearsOfExperience),
		IsActive:        row.IsActive,
		UserID:          r.pgMap.PgInt4.ToUint(row.UserID),
		ConsultationFee: nil,
	}
	emp.SetID(employees.NewEmployeeID(uint(row.ID)))
	emp.FirstName = row.FirstName
	emp.LastName = row.LastName
	emp.Gender = shared.PersonGender(row.Gender)
	emp.DateOfBirth = r.pgMap.PgDate.ToTime(row.DateOfBirth)
	emp.SetTimeStamps(r.pgMap.PgTimestamptz.ToTime(row.CreatedAt), r.pgMap.PgTimestamptz.ToTime(row.UpdatedAt))
	if len(row.ScheduleJson) > 0 {
		var s employees.EmployeeSchedule
		if err := json.Unmarshal(row.ScheduleJson, &s); err == nil {
			emp.Schedule = &s
		}
	}
	return emp
}

func (r *EmployeeSqlcRepository) dbError(operation, message string, err error) error {
	return customErr.DatabaseError(operation, TableEmployees, DriverSQL, fmt.Errorf("%s: %w", message, err))
}

func (r *EmployeeSqlcRepository) notFoundError(parameterName, parameterValue string) error {
	return customErr.DBNotFoundError(parameterName, parameterValue, OpSelect, TableEmployees, DriverSQL)
}
