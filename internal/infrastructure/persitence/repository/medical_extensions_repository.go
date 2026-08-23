package repository

import (
	"clinic-vet-api/internal/core/medical"
	"clinic-vet-api/internal/shared"
	"clinic-vet-api/internal/shared/mapper"
	"clinic-vet-api/internal/shared/page"
	"clinic-vet-api/database/sqlc"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	customErr "clinic-vet-api/internal/shared/errors"
)

const (
	TableSessionVaccinations  = "session_vaccinations"
	TableSessionSurgeries     = "session_surgeries"
	TableSessionPrescriptions = "session_prescriptions"
	TableSessionAttachments   = "session_attachments"
	TableSessionServices      = "session_services"
)

// ─── SessionVaccinationRepository ───────────────────────────────────────────

type SessionVaccinationSqlcRepository struct {
	queries *sqlc.Queries
	pgMap   *mapper.SqlcFieldMapper
}

func NewSessionVaccinationSqlcRepository(queries *sqlc.Queries) medical.SessionVaccinationRepository {
	return &SessionVaccinationSqlcRepository{queries: queries, pgMap: mapper.NewSqlcFieldMapper()}
}

func (r *SessionVaccinationSqlcRepository) FindBySessionID(ctx context.Context, sessionID medical.SessionID) ([]medical.SessionVaccination, error) {
	rows, err := r.queries.FindSessionVaccinationsBySessionID(ctx, int32(sessionID.Value()))
	if err != nil {
		return nil, customErr.DatabaseError(OpSelect, TableSessionVaccinations, DriverSQL, fmt.Errorf("find by session: %w", err))
	}
	out := make([]medical.SessionVaccination, len(rows))
	for i := range rows {
		out[i] = r.vaccinationToEntity(rows[i])
	}
	return out, nil
}

func (r *SessionVaccinationSqlcRepository) FindByID(ctx context.Context, id medical.VaccinationID) (medical.SessionVaccination, error) {
	row, err := r.queries.FindSessionVaccinationByID(ctx, int32(id.Value()))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return medical.SessionVaccination{}, customErr.DBNotFoundError("id", id.String(), OpSelect, TableSessionVaccinations, DriverSQL)
		}
		return medical.SessionVaccination{}, customErr.DatabaseError(OpSelect, TableSessionVaccinations, DriverSQL, err)
	}
	return r.vaccinationToEntity(row), nil
}

func (r *SessionVaccinationSqlcRepository) FindHistoryBySpec(ctx context.Context, spec medical.VaccinationHistorySpecification) (page.Page[medical.SessionVaccination], error) {
	petIds := make([]int32, 0, len(spec.PetIDs))
	for _, p := range spec.PetIDs {
		petIds = append(petIds, int32(p))
	}
	employeeIds := make([]int32, 0, len(spec.EmployeeIDs))
	for _, employeeID := range spec.EmployeeIDs {
		employeeIds = append(employeeIds, int32(employeeID))
	}
	vcIds := make([]int32, 0, len(spec.VaccineCatalogIDs))
	for _, v := range spec.VaccineCatalogIDs {
		vcIds = append(vcIds, int32(v))
	}
	var dateFrom, dateTo pgtype.Timestamptz
	if spec.DateFrom != nil {
		dateFrom = r.pgMap.PgTimestamptz.FromTime(*spec.DateFrom)
	}
	if spec.DateTo != nil {
		dateTo = r.pgMap.PgTimestamptz.FromTime(*spec.DateTo)
	}
	limit := spec.Pagination.Limit()
	if limit <= 0 {
		limit = int32(page.DefaultPageSize)
	}
	offset := spec.Pagination.Offset()
	params := sqlc.FindVaccinationHistoryBySpecParams{
		PetIds:            petIds,
		EmployeeIds:       employeeIds,
		VaccineCatalogIds: vcIds,
		DateFrom:          dateFrom,
		DateTo:            dateTo,
		PageLimit:         limit,
		PageOffset:        offset,
	}
	rows, err := r.queries.FindVaccinationHistoryBySpec(ctx, params)
	if err != nil {
		return page.Page[medical.SessionVaccination]{}, customErr.DatabaseError(OpSelect, TableSessionVaccinations, DriverSQL, err)
	}
	countParams := sqlc.CountVaccinationHistoryBySpecParams{
		PetIds:            petIds,
		EmployeeIds:       employeeIds,
		VaccineCatalogIds: vcIds,
		DateFrom:          dateFrom,
		DateTo:            dateTo,
	}
	total, err := r.queries.CountVaccinationHistoryBySpec(ctx, countParams)
	if err != nil {
		return page.Page[medical.SessionVaccination]{}, customErr.DatabaseError(OpCount, TableSessionVaccinations, DriverSQL, err)
	}
	items := make([]medical.SessionVaccination, len(rows))
	for i := range rows {
		items[i] = r.vaccinationToEntity(rows[i])
	}
	pagReq := page.PaginationRequest{Page: int32(spec.Pagination.Number), PageSize: limit}
	return page.NewPage(items, total, pagReq), nil
}

func (r *SessionVaccinationSqlcRepository) Save(ctx context.Context, v medical.SessionVaccination) (medical.SessionVaccination, error) {
	if v.ID.Value() == 0 {
		return r.createVaccination(ctx, v)
	}
	return r.updateVaccination(ctx, v)
}

func (r *SessionVaccinationSqlcRepository) DeleteByID(ctx context.Context, id medical.VaccinationID) error {
	if err := r.queries.DeleteSessionVaccinationByID(ctx, int32(id.Value())); err != nil {
		return customErr.DatabaseError(OpDelete, TableSessionVaccinations, DriverSQL, err)
	}
	return nil
}

func (r *SessionVaccinationSqlcRepository) DeleteBySessionID(ctx context.Context, sessionID medical.SessionID) error {
	if err := r.queries.DeleteSessionVaccinationsBySessionID(ctx, int32(sessionID.Value())); err != nil {
		return customErr.DatabaseError(OpDelete, TableSessionVaccinations, DriverSQL, err)
	}
	return nil
}

func (r *SessionVaccinationSqlcRepository) createVaccination(ctx context.Context, v medical.SessionVaccination) (medical.SessionVaccination, error) {
	params := sqlc.CreateSessionVaccinationParams{
		SessionID:        int32(v.SessionID.Value()),
		VaccineCatalogID: int32(v.VaccineCatalogID.Value()),
		DoseNumber:       int32(v.DoseNumber),
		BatchNumber:      r.pgMap.PgText.FromStringPtr(v.BatchNumber),
		ExpirationDate:   r.pgMap.PgDate.FromTimePtr(v.ExpirationDate),
		SiteOfInjection:  r.pgMap.PgText.FromStringPtr(v.SiteOfInjection),
		NextDoseDate:     r.pgMap.PgDate.FromTimePtr(v.NextDoseDate),
		ReactionNotes:    r.pgMap.PgText.FromStringPtr(v.ReactionNotes),
		AdministeredBy:   r.pgMap.PgInt4.FromUintPtr(v.AdministeredBy),
	}
	created, err := r.queries.CreateSessionVaccination(ctx, params)
	if err != nil {
		return medical.SessionVaccination{}, customErr.DatabaseError(OpInsert, TableSessionVaccinations, DriverSQL, err)
	}
	return r.vaccinationToEntity(created), nil
}

func (r *SessionVaccinationSqlcRepository) updateVaccination(ctx context.Context, v medical.SessionVaccination) (medical.SessionVaccination, error) {
	params := sqlc.UpdateSessionVaccinationParams{
		ID:               int32(v.ID.Value()),
		SessionID:        int32(v.SessionID.Value()),
		VaccineCatalogID: int32(v.VaccineCatalogID.Value()),
		DoseNumber:       int32(v.DoseNumber),
		BatchNumber:      r.pgMap.PgText.FromStringPtr(v.BatchNumber),
		ExpirationDate:   r.pgMap.PgDate.FromTimePtr(v.ExpirationDate),
		SiteOfInjection:  r.pgMap.PgText.FromStringPtr(v.SiteOfInjection),
		NextDoseDate:     r.pgMap.PgDate.FromTimePtr(v.NextDoseDate),
		ReactionNotes:    r.pgMap.PgText.FromStringPtr(v.ReactionNotes),
		AdministeredBy:   r.pgMap.PgInt4.FromUintPtr(v.AdministeredBy),
	}
	updated, err := r.queries.UpdateSessionVaccination(ctx, params)
	if err != nil {
		return medical.SessionVaccination{}, customErr.DatabaseError(OpUpdate, TableSessionVaccinations, DriverSQL, err)
	}
	return r.vaccinationToEntity(updated), nil
}

func (r *SessionVaccinationSqlcRepository) vaccinationToEntity(row sqlc.SessionVaccination) medical.SessionVaccination {
	v := medical.SessionVaccination{
		ID:               medical.NewVaccinationID(uint(row.ID)),
		SessionID:        medical.NewSessionID(uint(row.SessionID)),
		VaccineCatalogID: medical.NewVaccineCatalogID(uint(row.VaccineCatalogID)),
		DoseNumber:       int(row.DoseNumber),
		BatchNumber:      r.pgMap.PgText.ToStringPtr(row.BatchNumber),
		ExpirationDate:   r.pgMap.PgDate.ToTimePtr(row.ExpirationDate),
		SiteOfInjection:  r.pgMap.PgText.ToStringPtr(row.SiteOfInjection),
		NextDoseDate:     r.pgMap.PgDate.ToTimePtr(row.NextDoseDate),
		ReactionNotes:    r.pgMap.PgText.ToStringPtr(row.ReactionNotes),
		CreatedAt:        r.pgMap.PgTimestamptz.ToTime(row.CreatedAt),
	}
	if row.AdministeredBy.Valid {
		u := uint(row.AdministeredBy.Int32)
		v.AdministeredBy = &u
	}
	return v
}

// ─── SessionSurgeryRepository ───────────────────────────────────────────────

type SessionSurgerySqlcRepository struct {
	queries *sqlc.Queries
	pgMap   *mapper.SqlcFieldMapper
}

func NewSessionSurgerySqlcRepository(queries *sqlc.Queries) medical.SessionSurgeryRepository {
	return &SessionSurgerySqlcRepository{queries: queries, pgMap: mapper.NewSqlcFieldMapper()}
}

func (r *SessionSurgerySqlcRepository) FindBySessionID(ctx context.Context, sessionID medical.SessionID) ([]medical.SessionSurgery, error) {
	rows, err := r.queries.FindSessionSurgeriesBySessionID(ctx, int32(sessionID.Value()))
	if err != nil {
		return nil, customErr.DatabaseError(OpSelect, TableSessionSurgeries, DriverSQL, err)
	}
	out := make([]medical.SessionSurgery, len(rows))
	for i := range rows {
		out[i] = r.surgeryToEntity(rows[i])
	}
	return out, nil
}

func (r *SessionSurgerySqlcRepository) FindByID(ctx context.Context, id medical.SurgeryID) (medical.SessionSurgery, error) {
	row, err := r.queries.FindSessionSurgeryByID(ctx, int32(id.Value()))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return medical.SessionSurgery{}, customErr.DBNotFoundError("id", id.String(), OpSelect, TableSessionSurgeries, DriverSQL)
		}
		return medical.SessionSurgery{}, customErr.DatabaseError(OpSelect, TableSessionSurgeries, DriverSQL, err)
	}
	return r.surgeryToEntity(row), nil
}

func (r *SessionSurgerySqlcRepository) Save(ctx context.Context, s medical.SessionSurgery) (medical.SessionSurgery, error) {
	if s.ID.Value() == 0 {
		return r.createSurgery(ctx, s)
	}
	return r.updateSurgery(ctx, s)
}

func (r *SessionSurgerySqlcRepository) DeleteByID(ctx context.Context, id medical.SurgeryID) error {
	return r.queries.DeleteSessionSurgeryByID(ctx, int32(id.Value()))
}

func (r *SessionSurgerySqlcRepository) DeleteBySessionID(ctx context.Context, sessionID medical.SessionID) error {
	return r.queries.DeleteSessionSurgeriesBySessionID(ctx, int32(sessionID.Value()))
}

func (r *SessionSurgerySqlcRepository) createSurgery(ctx context.Context, s medical.SessionSurgery) (medical.SessionSurgery, error) {
	params := sqlc.CreateSessionSurgeryParams{
		SessionID:       int32(s.SessionID.Value()),
		ProcedureName:   s.ProcedureName,
		AnesthesiaType:  r.pgMap.PgText.FromStringPtr(s.AnesthesiaType),
		AnesthesiaAgent: r.pgMap.PgText.FromStringPtr(s.AnesthesiaAgent),
		PreOpNotes:      r.pgMap.PgText.FromStringPtr(s.PreOpNotes),
		IntraOpNotes:    r.pgMap.PgText.FromStringPtr(s.IntraOpNotes),
		PostOpNotes:     r.pgMap.PgText.FromStringPtr(s.PostOpNotes),
		Outcome:         pgtype.Text{String: string(s.Outcome), Valid: s.Outcome != ""},
		SurgeonID:       r.pgMap.PgInt4.FromUintPtr(s.SurgeonID),
	}
	if s.DurationMinutes != nil {
		params.DurationMinutes = pgtype.Int4{Int32: int32(*s.DurationMinutes), Valid: true}
	} else {
		params.DurationMinutes = pgtype.Int4{Valid: false}
	}
	created, err := r.queries.CreateSessionSurgery(ctx, params)
	if err != nil {
		return medical.SessionSurgery{}, customErr.DatabaseError(OpInsert, TableSessionSurgeries, DriverSQL, err)
	}
	return r.surgeryToEntity(created), nil
}

func (r *SessionSurgerySqlcRepository) updateSurgery(ctx context.Context, s medical.SessionSurgery) (medical.SessionSurgery, error) {
	params := sqlc.UpdateSessionSurgeryParams{
		ID:              int32(s.ID.Value()),
		SessionID:       int32(s.SessionID.Value()),
		ProcedureName:   s.ProcedureName,
		AnesthesiaType:  r.pgMap.PgText.FromStringPtr(s.AnesthesiaType),
		AnesthesiaAgent: r.pgMap.PgText.FromStringPtr(s.AnesthesiaAgent),
		PreOpNotes:      r.pgMap.PgText.FromStringPtr(s.PreOpNotes),
		IntraOpNotes:    r.pgMap.PgText.FromStringPtr(s.IntraOpNotes),
		PostOpNotes:     r.pgMap.PgText.FromStringPtr(s.PostOpNotes),
		Outcome:         pgtype.Text{String: string(s.Outcome), Valid: s.Outcome != ""},
		SurgeonID:       r.pgMap.PgInt4.FromUintPtr(s.SurgeonID),
	}
	if s.DurationMinutes != nil {
		params.DurationMinutes = pgtype.Int4{Int32: int32(*s.DurationMinutes), Valid: true}
	} else {
		params.DurationMinutes = pgtype.Int4{Valid: false}
	}
	updated, err := r.queries.UpdateSessionSurgery(ctx, params)
	if err != nil {
		return medical.SessionSurgery{}, customErr.DatabaseError(OpUpdate, TableSessionSurgeries, DriverSQL, err)
	}
	return r.surgeryToEntity(updated), nil
}

func (r *SessionSurgerySqlcRepository) surgeryToEntity(row sqlc.SessionSurgery) medical.SessionSurgery {
	s := medical.SessionSurgery{
		ID:              medical.NewSurgeryID(uint(row.ID)),
		SessionID:       medical.NewSessionID(uint(row.SessionID)),
		ProcedureName:   row.ProcedureName,
		AnesthesiaType:  r.pgMap.PgText.ToStringPtr(row.AnesthesiaType),
		AnesthesiaAgent: r.pgMap.PgText.ToStringPtr(row.AnesthesiaAgent),
		PreOpNotes:      r.pgMap.PgText.ToStringPtr(row.PreOpNotes),
		IntraOpNotes:    r.pgMap.PgText.ToStringPtr(row.IntraOpNotes),
		PostOpNotes:     r.pgMap.PgText.ToStringPtr(row.PostOpNotes),
		CreatedAt:       r.pgMap.PgTimestamptz.ToTime(row.CreatedAt),
	}
	if row.DurationMinutes.Valid {
		d := int(row.DurationMinutes.Int32)
		s.DurationMinutes = &d
	}
	if row.Outcome.Valid {
		s.Outcome = medical.SurgeryOutcome(row.Outcome.String)
	}
	if row.SurgeonID.Valid {
		u := uint(row.SurgeonID.Int32)
		s.SurgeonID = &u
	}
	return s
}

// ─── SessionPrescriptionRepository ───────────────────────────────────────────

type SessionPrescriptionSqlcRepository struct {
	queries *sqlc.Queries
	pgMap   *mapper.SqlcFieldMapper
}

func NewSessionPrescriptionSqlcRepository(queries *sqlc.Queries) medical.SessionPrescriptionRepository {
	return &SessionPrescriptionSqlcRepository{queries: queries, pgMap: mapper.NewSqlcFieldMapper()}
}

func (r *SessionPrescriptionSqlcRepository) FindBySessionID(ctx context.Context, sessionID medical.SessionID) ([]medical.SessionPrescription, error) {
	rows, err := r.queries.FindSessionPrescriptionsBySessionID(ctx, int32(sessionID.Value()))
	if err != nil {
		return nil, customErr.DatabaseError(OpSelect, TableSessionPrescriptions, DriverSQL, err)
	}
	out := make([]medical.SessionPrescription, len(rows))
	for i := range rows {
		out[i] = r.prescriptionToEntity(rows[i])
	}
	return out, nil
}

func (r *SessionPrescriptionSqlcRepository) FindByID(ctx context.Context, id medical.PrescriptionID) (medical.SessionPrescription, error) {
	row, err := r.queries.FindSessionPrescriptionByID(ctx, int32(id.Value()))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return medical.SessionPrescription{}, customErr.DBNotFoundError("id", id.String(), OpSelect, TableSessionPrescriptions, DriverSQL)
		}
		return medical.SessionPrescription{}, customErr.DatabaseError(OpSelect, TableSessionPrescriptions, DriverSQL, err)
	}
	return r.prescriptionToEntity(row), nil
}

func (r *SessionPrescriptionSqlcRepository) FindActivePrescriptionsByPet(ctx context.Context, petID uint, employeeIDs []uint, p page.Pagination) (page.Page[medical.SessionPrescription], error) {
	limit := p.Limit()
	if limit <= 0 {
		limit = int32(page.DefaultPageSize)
	}
	offset := p.Offset()
	employeeIds := make([]int32, 0, len(employeeIDs))
	for _, employeeID := range employeeIDs {
		employeeIds = append(employeeIds, int32(employeeID))
	}
	params := sqlc.FindActivePrescriptionsByPetIDParams{PetID: int32(petID), EmployeeIds: employeeIds, Limit: limit, Offset: offset}
	rows, err := r.queries.FindActivePrescriptionsByPetID(ctx, params)
	if err != nil {
		return page.Page[medical.SessionPrescription]{}, customErr.DatabaseError(OpSelect, TableSessionPrescriptions, DriverSQL, err)
	}
	total, err := r.queries.CountActivePrescriptionsByPetID(ctx, sqlc.CountActivePrescriptionsByPetIDParams{PetID: int32(petID), EmployeeIds: employeeIds})
	if err != nil {
		return page.Page[medical.SessionPrescription]{}, customErr.DatabaseError(OpCount, TableSessionPrescriptions, DriverSQL, err)
	}
	items := make([]medical.SessionPrescription, len(rows))
	for i := range rows {
		items[i] = r.prescriptionToEntity(rows[i])
	}
	return page.NewPage(items, total, page.PaginationRequest{Page: int32(p.Number), PageSize: limit}), nil
}

func (r *SessionPrescriptionSqlcRepository) Save(ctx context.Context, p medical.SessionPrescription) (medical.SessionPrescription, error) {
	if p.ID.Value() == 0 {
		return r.createPrescription(ctx, p)
	}
	return r.updatePrescription(ctx, p)
}

func (r *SessionPrescriptionSqlcRepository) DeleteByID(ctx context.Context, id medical.PrescriptionID) error {
	return r.queries.DeleteSessionPrescriptionByID(ctx, int32(id.Value()))
}

func (r *SessionPrescriptionSqlcRepository) DeleteBySessionID(ctx context.Context, sessionID medical.SessionID) error {
	return r.queries.DeleteSessionPrescriptionsBySessionID(ctx, int32(sessionID.Value()))
}

func (r *SessionPrescriptionSqlcRepository) createPrescription(ctx context.Context, p medical.SessionPrescription) (medical.SessionPrescription, error) {
	params := sqlc.CreateSessionPrescriptionParams{
		SessionID:    int32(p.SessionID.Value()),
		MedicationID: int32(p.MedicationID.Value()),
		Dosage:       p.Dosage,
		Frequency:    p.Frequency,
		Route:        r.pgMap.PgText.FromStringPtr(p.Route),
		Instructions: r.pgMap.PgText.FromStringPtr(p.Instructions),
		StartDate:    r.pgMap.PgDate.FromTime(p.StartDate),
	}
	if p.DurationDays != nil {
		params.DurationDays = pgtype.Int4{Int32: int32(*p.DurationDays), Valid: true}
	}
	created, err := r.queries.CreateSessionPrescription(ctx, params)
	if err != nil {
		return medical.SessionPrescription{}, customErr.DatabaseError(OpInsert, TableSessionPrescriptions, DriverSQL, err)
	}
	return r.prescriptionToEntity(created), nil
}

func (r *SessionPrescriptionSqlcRepository) updatePrescription(ctx context.Context, p medical.SessionPrescription) (medical.SessionPrescription, error) {
	params := sqlc.UpdateSessionPrescriptionParams{
		ID:           int32(p.ID.Value()),
		SessionID:    int32(p.SessionID.Value()),
		MedicationID: int32(p.MedicationID.Value()),
		Dosage:       p.Dosage,
		Frequency:    p.Frequency,
		Route:        r.pgMap.PgText.FromStringPtr(p.Route),
		Instructions: r.pgMap.PgText.FromStringPtr(p.Instructions),
		StartDate:    r.pgMap.PgDate.FromTime(p.StartDate),
	}
	if p.DurationDays != nil {
		params.DurationDays = pgtype.Int4{Int32: int32(*p.DurationDays), Valid: true}
	} else {
		params.DurationDays = pgtype.Int4{Valid: false}
	}
	updated, err := r.queries.UpdateSessionPrescription(ctx, params)
	if err != nil {
		return medical.SessionPrescription{}, customErr.DatabaseError(OpUpdate, TableSessionPrescriptions, DriverSQL, err)
	}
	return r.prescriptionToEntity(updated), nil
}

func (r *SessionPrescriptionSqlcRepository) prescriptionToEntity(row sqlc.SessionPrescription) medical.SessionPrescription {
	p := medical.SessionPrescription{
		ID:           medical.NewPrescriptionID(uint(row.ID)),
		SessionID:    medical.NewSessionID(uint(row.SessionID)),
		MedicationID: medical.NewMedicationID(uint(row.MedicationID)),
		Dosage:       row.Dosage,
		Frequency:    row.Frequency,
		StartDate:    r.pgMap.PgDate.ToTime(row.StartDate),
		CreatedAt:    r.pgMap.PgTimestamptz.ToTime(row.CreatedAt),
		Route:        r.pgMap.PgText.ToStringPtr(row.Route),
		Instructions: r.pgMap.PgText.ToStringPtr(row.Instructions),
		EndDate:      r.pgMap.PgDate.ToTimePtr(row.EndDate),
	}
	if row.DurationDays.Valid {
		d := int(row.DurationDays.Int32)
		p.DurationDays = &d
	}
	return p
}

// ─── SessionAttachmentRepository ─────────────────────────────────────────────

type SessionAttachmentSqlcRepository struct {
	queries *sqlc.Queries
	pgMap   *mapper.SqlcFieldMapper
}

func NewSessionAttachmentSqlcRepository(queries *sqlc.Queries) medical.SessionAttachmentRepository {
	return &SessionAttachmentSqlcRepository{queries: queries, pgMap: mapper.NewSqlcFieldMapper()}
}

func (r *SessionAttachmentSqlcRepository) FindBySessionID(ctx context.Context, sessionID medical.SessionID) ([]medical.SessionAttachment, error) {
	rows, err := r.queries.FindSessionAttachmentsBySessionID(ctx, int32(sessionID.Value()))
	if err != nil {
		return nil, customErr.DatabaseError(OpSelect, TableSessionAttachments, DriverSQL, err)
	}
	out := make([]medical.SessionAttachment, len(rows))
	for i := range rows {
		out[i] = r.attachmentToEntity(rows[i])
	}
	return out, nil
}

func (r *SessionAttachmentSqlcRepository) FindByID(ctx context.Context, id medical.AttachmentID) (medical.SessionAttachment, error) {
	row, err := r.queries.FindSessionAttachmentByID(ctx, int32(id.Value()))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return medical.SessionAttachment{}, customErr.DBNotFoundError("id", id.String(), OpSelect, TableSessionAttachments, DriverSQL)
		}
		return medical.SessionAttachment{}, customErr.DatabaseError(OpSelect, TableSessionAttachments, DriverSQL, err)
	}
	return r.attachmentToEntity(row), nil
}

func (r *SessionAttachmentSqlcRepository) Save(ctx context.Context, a medical.SessionAttachment) (medical.SessionAttachment, error) {
	if a.ID.Value() == 0 {
		return r.createAttachment(ctx, a)
	}
	return r.updateAttachment(ctx, a)
}

func (r *SessionAttachmentSqlcRepository) DeleteByID(ctx context.Context, id medical.AttachmentID) error {
	return r.queries.DeleteSessionAttachmentByID(ctx, int32(id.Value()))
}

func (r *SessionAttachmentSqlcRepository) DeleteBySessionID(ctx context.Context, sessionID medical.SessionID) error {
	return r.queries.DeleteSessionAttachmentsBySessionID(ctx, int32(sessionID.Value()))
}

func (r *SessionAttachmentSqlcRepository) createAttachment(ctx context.Context, a medical.SessionAttachment) (medical.SessionAttachment, error) {
	params := sqlc.CreateSessionAttachmentParams{
		SessionID:   int32(a.SessionID.Value()),
		FileType:    string(a.FileType),
		FileUrl:     a.FileURL,
		Description: r.pgMap.PgText.FromStringPtr(a.Description),
		UploadedBy:  r.pgMap.PgInt4.FromUintPtr(a.UploadedBy),
	}
	created, err := r.queries.CreateSessionAttachment(ctx, params)
	if err != nil {
		return medical.SessionAttachment{}, customErr.DatabaseError(OpInsert, TableSessionAttachments, DriverSQL, err)
	}
	return r.attachmentToEntity(created), nil
}

func (r *SessionAttachmentSqlcRepository) updateAttachment(ctx context.Context, a medical.SessionAttachment) (medical.SessionAttachment, error) {
	params := sqlc.UpdateSessionAttachmentParams{
		ID:          int32(a.ID.Value()),
		SessionID:   int32(a.SessionID.Value()),
		FileType:    string(a.FileType),
		FileUrl:     a.FileURL,
		Description: r.pgMap.PgText.FromStringPtr(a.Description),
		UploadedBy:  r.pgMap.PgInt4.FromUintPtr(a.UploadedBy),
	}
	updated, err := r.queries.UpdateSessionAttachment(ctx, params)
	if err != nil {
		return medical.SessionAttachment{}, customErr.DatabaseError(OpUpdate, TableSessionAttachments, DriverSQL, err)
	}
	return r.attachmentToEntity(updated), nil
}

func (r *SessionAttachmentSqlcRepository) attachmentToEntity(row sqlc.SessionAttachment) medical.SessionAttachment {
	a := medical.SessionAttachment{
		ID:          medical.NewAttachmentID(uint(row.ID)),
		SessionID:   medical.NewSessionID(uint(row.SessionID)),
		FileType:    medical.AttachmentFileType(row.FileType),
		FileURL:     row.FileUrl,
		Description: r.pgMap.PgText.ToStringPtr(row.Description),
		CreatedAt:   r.pgMap.PgTimestamptz.ToTime(row.CreatedAt),
	}
	if row.UploadedBy.Valid {
		u := uint(row.UploadedBy.Int32)
		a.UploadedBy = &u
	}
	return a
}

// ─── SessionServiceRepository ────────────────────────────────────────────────

type SessionServiceSqlcRepository struct {
	queries *sqlc.Queries
	pgMap   *mapper.SqlcFieldMapper
}

func NewSessionServiceSqlcRepository(queries *sqlc.Queries) medical.SessionServiceRepository {
	return &SessionServiceSqlcRepository{queries: queries, pgMap: mapper.NewSqlcFieldMapper()}
}

func (r *SessionServiceSqlcRepository) FindBySessionID(ctx context.Context, sessionID medical.SessionID) ([]medical.SessionService, error) {
	rows, err := r.queries.FindSessionServicesBySessionID(ctx, int32(sessionID.Value()))
	if err != nil {
		return nil, customErr.DatabaseError(OpSelect, TableSessionServices, DriverSQL, err)
	}
	out := make([]medical.SessionService, len(rows))
	for i := range rows {
		out[i] = r.serviceToEntity(rows[i])
	}
	return out, nil
}

func (r *SessionServiceSqlcRepository) FindByID(ctx context.Context, id medical.SessionServiceID) (medical.SessionService, error) {
	row, err := r.queries.FindSessionServiceByID(ctx, int32(id.Value()))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return medical.SessionService{}, customErr.DBNotFoundError("id", id.String(), OpSelect, TableSessionServices, DriverSQL)
		}
		return medical.SessionService{}, customErr.DatabaseError(OpSelect, TableSessionServices, DriverSQL, err)
	}
	return r.serviceToEntity(row), nil
}

func (r *SessionServiceSqlcRepository) Save(ctx context.Context, s medical.SessionService) (medical.SessionService, error) {
	if s.ID.Value() == 0 {
		return r.createService(ctx, s)
	}
	return r.updateService(ctx, s)
}

func (r *SessionServiceSqlcRepository) DeleteByID(ctx context.Context, id medical.SessionServiceID) error {
	return r.queries.DeleteSessionServiceByID(ctx, int32(id.Value()))
}

func (r *SessionServiceSqlcRepository) DeleteBySessionID(ctx context.Context, sessionID medical.SessionID) error {
	return r.queries.DeleteSessionServicesBySessionID(ctx, int32(sessionID.Value()))
}

func (r *SessionServiceSqlcRepository) createService(ctx context.Context, s medical.SessionService) (medical.SessionService, error) {
	params := sqlc.CreateSessionServiceParams{
		SessionID:        int32(s.SessionID.Value()),
		ServiceCatalogID: int32(s.ServiceCatalogID.Value()),
		Quantity:         r.pgMap.PgNumeric.FromDecimal(shared.NewDecimalFromFloat(s.Quantity)),
		Notes:            r.pgMap.PgText.FromStringPtr(s.Notes),
	}
	if s.PriceApplied != nil {
		params.PriceApplied = r.pgMap.PgNumeric.FromDecimal(shared.NewDecimalFromFloat(*s.PriceApplied))
	} else {
		params.PriceApplied = pgtype.Numeric{Valid: false}
	}
	created, err := r.queries.CreateSessionService(ctx, params)
	if err != nil {
		return medical.SessionService{}, customErr.DatabaseError(OpInsert, TableSessionServices, DriverSQL, err)
	}
	return r.serviceToEntity(created), nil
}

func (r *SessionServiceSqlcRepository) updateService(ctx context.Context, s medical.SessionService) (medical.SessionService, error) {
	params := sqlc.UpdateSessionServiceParams{
		ID:               int32(s.ID.Value()),
		SessionID:        int32(s.SessionID.Value()),
		ServiceCatalogID: int32(s.ServiceCatalogID.Value()),
		Quantity:         r.pgMap.PgNumeric.FromDecimal(shared.NewDecimalFromFloat(s.Quantity)),
		Notes:            r.pgMap.PgText.FromStringPtr(s.Notes),
	}
	if s.PriceApplied != nil {
		params.PriceApplied = r.pgMap.PgNumeric.FromDecimal(shared.NewDecimalFromFloat(*s.PriceApplied))
	} else {
		params.PriceApplied = pgtype.Numeric{Valid: false}
	}
	updated, err := r.queries.UpdateSessionService(ctx, params)
	if err != nil {
		return medical.SessionService{}, customErr.DatabaseError(OpUpdate, TableSessionServices, DriverSQL, err)
	}
	return r.serviceToEntity(updated), nil
}

func (r *SessionServiceSqlcRepository) serviceToEntity(row sqlc.SessionService) medical.SessionService {
	s := medical.SessionService{
		ID:               medical.NewSessionServiceID(uint(row.ID)),
		SessionID:        medical.NewSessionID(uint(row.SessionID)),
		ServiceCatalogID: medical.NewServiceCatalogID(uint(row.ServiceCatalogID)),
		Quantity:         r.pgMap.PgNumeric.ToDecimal(row.Quantity).Float64(),
		Notes:            r.pgMap.PgText.ToStringPtr(row.Notes),
		CreatedAt:        r.pgMap.PgTimestamptz.ToTime(row.CreatedAt),
	}
	if row.PriceApplied.Valid {
		f := r.pgMap.PgNumeric.ToDecimal(row.PriceApplied).Float64()
		s.PriceApplied = &f
	}
	return s
}
