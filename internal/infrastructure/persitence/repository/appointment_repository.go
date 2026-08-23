package repository

import (
	"clinic-vet-api/database/models"
	"clinic-vet-api/database/sqlc"
	"clinic-vet-api/internal/core/appointments"
	"clinic-vet-api/internal/core/customers"
	"clinic-vet-api/internal/core/employees"
	"clinic-vet-api/internal/core/pets"
	customErr "clinic-vet-api/internal/shared/errors"
	"clinic-vet-api/internal/shared/mapper"
	p "clinic-vet-api/internal/shared/page"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

type SqlcAppointmentRepository struct {
	queries *sqlc.Queries
	pgMap   *mapper.SqlcFieldMapper
}

func NewSqlcAppointmentRepository(queries *sqlc.Queries) appointments.AppointmentRepository {
	return &SqlcAppointmentRepository{
		queries: queries,
		pgMap:   mapper.NewSqlcFieldMapper(),
	}
}

func (r *SqlcAppointmentRepository) Find(
	ctx context.Context,
	spec *appointments.AppointmentSpecification,
) (p.Page[appointments.Appointment], error) {
	pagination := p.FromSpecPagination(spec.Pagination).WithDefaults()
	params := spec.ToSearchParams()
	params.Limit = pagination.Limit()
	params.Offset = pagination.Offset()
	var service string
	if params.Service != nil {
		service = *params.Service
	}

	var status string
	if params.Status != nil {
		status = *params.Status
	}

	sqlcParams := sqlc.FindAppointmentsBySpecParams{
		Column1: r.pgMap.Primitive.Int32PtrToInt32(params.ID),                              // ID
		Column2: r.pgMap.Primitive.Int32PtrToInt32(params.CustomerID),                      // CustomerID
		Column3: r.pgMap.Primitive.Int32PtrToInt32(params.EmployeeID),                      // EmployeeID
		Column4: r.pgMap.Primitive.Int32PtrToInt32(params.PetID),                           // PetID
		Column5: service,                                                                   // Service
		Column6: status,                                                                    // Status
		Column7: pgtype.Timestamp(r.pgMap.PgTimestamptz.FromTimePtr(params.StartDate)),     // StartDate
		Column8: pgtype.Timestamp(r.pgMap.PgTimestamptz.FromTimePtr(params.EndDate)),       // EndDate
		Column9: pgtype.Timestamp(r.pgMap.PgTimestamptz.FromTimePtr(params.ScheduledDate)), // ScheduledDate
		Limit:   params.Limit,
		Offset:  params.Offset,
	}

	rows, err := r.queries.FindAppointmentsBySpec(ctx, sqlcParams)
	if err != nil {
		return p.Page[appointments.Appointment]{}, err
	}

	countParams := sqlc.CountAppointmentsBySpecParams{
		Column1: sqlcParams.Column1, // ID
		Column2: sqlcParams.Column2, // CustomerID
		Column3: sqlcParams.Column3, // EmployeeID
		Column4: sqlcParams.Column4, // PetID
		Column5: sqlcParams.Column5, // Service
		Column6: sqlcParams.Column6, // Status
		Column7: sqlcParams.Column7, // StartDate
		Column8: sqlcParams.Column8, // EndDate
		Column9: sqlcParams.Column9, // ScheduledDate
	}

	appt := r.specRowsToEntities(rows)

	total, err := r.queries.CountAppointmentsBySpec(ctx, countParams)
	if err != nil {
		return p.Page[appointments.Appointment]{}, err
	}

	return p.NewPage(appt, total, pagination), nil
}

func (r *SqlcAppointmentRepository) FindByID(
	ctx context.Context,
	id appointments.AppointmentID,
) (appointments.Appointment, error) {
	sqlRow, err := r.queries.FindAppointmentByID(ctx, id.Int32())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appointments.Appointment{}, r.notFoundError("id", id.String())
		}
		return appointments.Appointment{}, r.dbError(OpSelect, ErrMsgFindAppointmentByID, err)
	}

	return r.rowToEntity(sqlRow), nil
}

func (r *SqlcAppointmentRepository) FindByIDAndCustomerID(
	ctx context.Context,
	id appointments.AppointmentID,
	customerID customers.CustomerID,
) (appointments.Appointment, error) {
	params := sqlc.FindAppointmentByIDAndCustomerIDParams{
		ID:         id.Int32(),
		CustomerID: customerID.Int32(),
	}

	sqlRow, err := r.queries.FindAppointmentByIDAndCustomerID(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appointments.Appointment{}, r.notFoundError("id and customer_id", fmt.Sprintf("%s and %s", id.String(), customerID.String()))
		}
		return appointments.Appointment{}, r.dbError(OpSelect, ErrMsgFindAppointmentByID, err)
	}

	return r.rowToEntity(sqlRow), nil
}

func (r *SqlcAppointmentRepository) FindByIDAndEmployeeID(
	ctx context.Context,
	id appointments.AppointmentID,
	employeeID employees.EmployeeID,
) (appointments.Appointment, error) {
	params := sqlc.FindAppointmentByIDAndEmployeeIDParams{
		ID:         id.Int32(),
		EmployeeID: r.pgMap.PgInt4.FromInt32(employeeID.Int32()),
	}

	sqlRow, err := r.queries.FindAppointmentByIDAndEmployeeID(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appointments.Appointment{}, r.notFoundError("id and employee_id", fmt.Sprintf("%s and %s", id.String(), employeeID.String()))
		}
		return appointments.Appointment{}, r.dbError(OpSelect, ErrMsgFindAppointmentByID, err)
	}
	return r.rowToEntity(sqlRow), nil
}

func (r *SqlcAppointmentRepository) ExistsByID(
	ctx context.Context,
	id appointments.AppointmentID,
) (bool, error) {
	exists, err := r.queries.ExistsAppointmentID(ctx, id.Int32())
	if err != nil {
		return false, r.dbError(OpSelect, ErrMsgCheckAppointmentExists, err)
	}
	return exists, nil
}

func (r *SqlcAppointmentRepository) Save(
	ctx context.Context,
	appointment *appointments.Appointment,
) error {
	if appointment.ID.IntegerID == nil || appointment.ID.IsZero() {
		return r.create(ctx, appointment)
	}
	return r.update(ctx, appointment)
}

func (r *SqlcAppointmentRepository) DeleteByID(
	ctx context.Context,
	id appointments.AppointmentID,
	hardDelete bool,
) error {
	if hardDelete {
		return r.hardDeleteByID(ctx, id)
	}
	return r.softDeleteByID(ctx, id)
}

func (r *SqlcAppointmentRepository) RestoreByID(ctx context.Context, id appointments.AppointmentID) error {
	if err := r.queries.RestoreAppointment(ctx, id.Int32()); err != nil {
		return r.dbError(OpUpdate, ErrMsgRestoreAppointment, err)
	}
	return nil
}

func (r *SqlcAppointmentRepository) Count(ctx context.Context, spec *appointments.AppointmentSpecification) (int64, error) {
	params := spec.ToSearchParams()
	var service string
	if params.Service != nil {
		service = *params.Service
	}
	var status string
	if params.Status != nil {
		status = *params.Status
	}
	sqlcParams := sqlc.CountAppointmentsBySpecParams{
		Column1: r.pgMap.Primitive.Int32PtrToInt32(params.ID),                              // ID
		Column2: r.pgMap.Primitive.Int32PtrToInt32(params.CustomerID),                      // CustomerID
		Column3: r.pgMap.Primitive.Int32PtrToInt32(params.EmployeeID),                      // EmployeeID
		Column4: r.pgMap.Primitive.Int32PtrToInt32(params.PetID),                           // PetID
		Column5: service,                                                                   // Service
		Column6: status,                                                                    // Status
		Column7: pgtype.Timestamp(r.pgMap.PgTimestamptz.FromTimePtr(params.StartDate)),     // StartDate
		Column8: pgtype.Timestamp(r.pgMap.PgTimestamptz.FromTimePtr(params.EndDate)),       // EndDate
		Column9: pgtype.Timestamp(r.pgMap.PgTimestamptz.FromTimePtr(params.ScheduledDate)), // ScheduledDate
	}
	total, err := r.queries.CountAppointmentsBySpec(ctx, sqlcParams)
	if err != nil {
		return 0, r.dbError(OpCount, ErrMsgCountAppointments, err)
	}
	return total, nil
}

func (r *SqlcAppointmentRepository) create(ctx context.Context, appointment *appointments.Appointment) error {
	params := r.toCreateParams(appointment)

	created, err := r.queries.CreateAppointment(ctx, params)
	if err != nil {
		return r.dbError(OpInsert, ErrMsgCreateAppointment, err)
	}
	*appointment = r.rowToEntity(created)

	return nil
}

func (r *SqlcAppointmentRepository) update(ctx context.Context, appointment *appointments.Appointment) error {
	params := r.toUpdateParams(appointment)

	updated, err := r.queries.UpdateAppointment(ctx, params)
	if err != nil {
		return r.dbError(OpUpdate, fmt.Sprintf("%s with ID %d", ErrMsgUpdateAppointment, appointment.ID.Value()), err)
	}
	*appointment = r.rowToEntity(updated)

	return nil
}

func (r *SqlcAppointmentRepository) softDeleteByID(ctx context.Context, id appointments.AppointmentID) error {
	if err := r.queries.SoftDeleteAppointment(ctx, id.Int32()); err != nil {
		return r.dbError(OpUpdate, ErrMsgSoftDeleteAppointment, err)
	}
	return nil
}

func (r *SqlcAppointmentRepository) hardDeleteByID(ctx context.Context, id appointments.AppointmentID) error {
	if err := r.queries.HardDeleteAppointment(ctx, id.Int32()); err != nil {
		return r.dbError(OpDelete, ErrMsgHardDeleteAppointment, err)
	}
	return nil
}

// ============================================================================
// Error handling helpers
// ============================================================================

// dbError creates a standardized database operation error
func (r *SqlcAppointmentRepository) dbError(operation, message string, err error) error {
	return customErr.DatabaseError(operation, TableAppts, DriverSQL, fmt.Errorf("%s: %v", message, err))
}

// notFoundError creates a standardized entity not found error
func (r *SqlcAppointmentRepository) notFoundError(parameterName, parameterValue string) error {
	return customErr.DBNotFoundError(parameterName, parameterValue, OpSelect, TableAppts, DriverSQL)
}

// wrapConversionError wraps appointment conversion errors
func (r *SqlcAppointmentRepository) wrapConversionError(err error) error {
	return customErr.WrapError(context.Background(), err, OpSelect, TableAppts, DriverSQL, ErrMsgConvertToAppointment)
}

// ============================================================================
// Mapping helpers
// ============================================================================

func (r *SqlcAppointmentRepository) rowToEntity(row sqlc.Appointment) appointments.Appointment {
	var employeeID *employees.EmployeeID
	if row.EmployeeID.Valid {
		employeeIDValue := employees.NewEmployeeID(uint(row.EmployeeID.Int32))
		employeeID = &employeeIDValue
	}

	var notes *string
	if row.Notes.Valid {
		notes = &row.Notes.String
	}
	var cancellationReason *string
	if row.CancellationReason.Valid {
		cancellationReason = &row.CancellationReason.String
	}

	appointment := appointments.Appointment{
		CustomerID:         customers.NewCustomerID(uint(row.CustomerID)),
		EmployeeID:         employeeID,
		PetID:              pets.NewPetID(uint(row.PetID)),
		ScheduledDate:      row.ScheduledDate.Time,
		Status:             appointments.AppointmentStatus(row.Status),
		Service:            appointments.ClinicService(row.ClinicService),
		Reason:             appointments.VisitReason(row.Reason.String),
		Notes:              notes,
		CancellationReason: cancellationReason,
	}
	appointment.SetID(appointments.NewAppointmentID(uint(row.ID)))
	appointment.SetTimeStamps(row.CreatedAt.Time, row.UpdatedAt.Time)

	return appointment
}

func (r *SqlcAppointmentRepository) specRowsToEntities(rows []sqlc.FindAppointmentsBySpecRow) []appointments.Appointment {
	// TODO: Improve mapping

	appts := make([]appointments.Appointment, len(rows))
	for i, row := range rows {
		appts[i] = appointments.Appointment{
			CustomerID: customers.NewCustomerID(uint(row.CustomerID)),
			EmployeeID: func() *employees.EmployeeID {
				if row.EmployeeID.Valid {
					id := employees.NewEmployeeID(uint(row.EmployeeID.Int32))
					return &id
				} else {
					return nil
				}
			}(),
			PetID:         pets.NewPetID(uint(row.PetID)),
			ScheduledDate: row.ScheduledDate.Time,
			Status:        appointments.AppointmentStatus(row.Status),
			Service:       appointments.ClinicService(row.ClinicService),
			Reason:        appointments.VisitReason(row.Reason.String),
			Notes: func() *string {
				if row.Notes.Valid {
					return &row.Notes.String
				} else {
					return nil
				}
			}(),
			CancellationReason: func() *string {
				if row.CancellationReason.Valid {
					return &row.CancellationReason.String
				}
				return nil
			}(),
		}
		appts[i].SetID(appointments.NewAppointmentID(uint(row.ID)))
		appts[i].SetTimeStamps(row.CreatedAt.Time, row.UpdatedAt.Time)
	}
	return appts
}

func (r *SqlcAppointmentRepository) toCreateParams(appt *appointments.Appointment) sqlc.CreateAppointmentParams {
	params := sqlc.CreateAppointmentParams{
		CustomerID:    int32(appt.CustomerID.Value()),
		PetID:         int32(appt.PetID.Value()),
		ScheduledDate: pgtype.Timestamptz{Time: appt.ScheduledDate, Valid: true},
		Status:        models.AppointmentStatus(string(appt.Status)),
		ClinicService: models.ClinicService(string(appt.Service.String())),
		Reason:        pgtype.Text{String: appt.Reason.String(), Valid: appt.Reason != ""},
	}

	if appt.EmployeeID != nil {
		params.EmployeeID = pgtype.Int4{
			Int32: int32(appt.EmployeeID.Value()),
			Valid: true,
		}
	} else {
		params.EmployeeID = pgtype.Int4{Valid: false}
	}

	if appt.Notes != nil {
		params.Notes = pgtype.Text{
			String: *appt.Notes,
			Valid:  true,
		}
	} else {
		params.Notes = pgtype.Text{Valid: false}
	}
	if appt.CancellationReason != nil {
		params.CancellationReason = pgtype.Text{String: *appt.CancellationReason, Valid: true}
	}

	return params
}

func (r *SqlcAppointmentRepository) toUpdateParams(appt *appointments.Appointment) sqlc.UpdateAppointmentParams {
	var employeeID int32
	if appt.EmployeeID != nil {
		employeeID = int32(appt.EmployeeID.Value())
	}
	var notes string
	if appt.Notes != nil {
		notes = *appt.Notes
	}
	var cancellationReason string
	if appt.CancellationReason != nil {
		cancellationReason = *appt.CancellationReason
	}
	return sqlc.UpdateAppointmentParams{
		ID:                 int32(appt.ID.Value()),
		CustomerID:         int32(appt.CustomerID.Value()),
		EmployeeID:         pgtype.Int4{Int32: employeeID, Valid: appt.EmployeeID != nil},
		PetID:              int32(appt.PetID.Value()),
		ScheduledDate:      pgtype.Timestamptz{Time: appt.ScheduledDate, Valid: true},
		Notes:              pgtype.Text{String: notes, Valid: appt.Notes != nil},
		Reason:             pgtype.Text{String: appt.Reason.String(), Valid: appt.Reason != ""},
		CancellationReason: pgtype.Text{String: cancellationReason, Valid: appt.CancellationReason != nil},
		Status:             models.AppointmentStatus(string(appt.Status)),
	}
}
