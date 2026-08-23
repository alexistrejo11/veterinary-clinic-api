package medical

import (
	"clinic-vet-api/internal/shared/page"
	"context"
)

// ─── MedicalSessionRepository ────────────────────────────────────────────────

// MedicalSessionRepository defines persistence operations for the core session
// aggregate. Extension sub-entities are persisted through their own repositories.
type MedicalSessionRepository interface {
	// FindBySpecification returns a paginated list of sessions matching the spec.
	FindBySpecification(ctx context.Context, spec MedicalSessionSpecification) (page.Page[MedicalSession], error)

	// FindByID returns the session identified by id.
	// Extensions (vaccinations, surgeries, etc.) are NOT loaded; use the
	// dedicated loader methods on the application service instead.
	FindByID(ctx context.Context, id SessionID) (MedicalSession, error)

	// FindByPetID returns all active sessions for a pet, most recent first.
	FindByPetID(ctx context.Context, petID uint, p page.Pagination) (page.Page[MedicalSession], error)

	// FindByCustomerID returns all active sessions linked to a customer.
	FindByCustomerID(ctx context.Context, customerID uint, p page.Pagination) (page.Page[MedicalSession], error)

	// ExistsByID reports whether a session with the given id exists (not deleted).
	ExistsByID(ctx context.Context, id SessionID) (bool, error)

	// Save creates or updates a session and returns the persisted entity.
	Save(ctx context.Context, session MedicalSession) (MedicalSession, error)

	// SoftDelete marks the session as deleted.
	SoftDelete(ctx context.Context, id SessionID) error

	// HardDelete permanently removes the session and all its extensions (cascade).
	HardDelete(ctx context.Context, id SessionID) error

	// RestoreByID clears the deleted_at timestamp.
	RestoreByID(ctx context.Context, id SessionID) error

	// IsDeletedByID reports whether the session is soft-deleted.
	IsDeletedByID(ctx context.Context, id SessionID) (bool, error)

	// CountAll returns the total number of sessions (including deleted).
	CountAll(ctx context.Context) (int64, error)

	// CountActive returns the number of non-deleted sessions.
	CountActive(ctx context.Context) (int64, error)

	// CountByClinicService returns the count of sessions per service type.
	CountByClinicService(ctx context.Context) (map[ClinicService]int64, error)
}

// ─── SessionVaccinationRepository ───────────────────────────────────────────

type SessionVaccinationRepository interface {
	// FindBySessionID returns all vaccinations for a given session.
	FindBySessionID(ctx context.Context, sessionID SessionID) ([]SessionVaccination, error)

	// FindByID returns a single vaccination record.
	FindByID(ctx context.Context, id VaccinationID) (SessionVaccination, error)

	// FindHistoryBySpec returns a paginated vaccination history for filtering
	// across sessions (wraps the pet_vaccination_history view).
	FindHistoryBySpec(ctx context.Context, spec VaccinationHistorySpecification) (page.Page[SessionVaccination], error)

	// Save creates or updates a vaccination record.
	Save(ctx context.Context, v SessionVaccination) (SessionVaccination, error)

	// DeleteByID permanently removes a vaccination record.
	DeleteByID(ctx context.Context, id VaccinationID) error

	// DeleteBySessionID removes all vaccinations linked to a session.
	DeleteBySessionID(ctx context.Context, sessionID SessionID) error
}

// ─── SessionSurgeryRepository ────────────────────────────────────────────────

type SessionSurgeryRepository interface {
	FindBySessionID(ctx context.Context, sessionID SessionID) ([]SessionSurgery, error)
	FindByID(ctx context.Context, id SurgeryID) (SessionSurgery, error)
	Save(ctx context.Context, s SessionSurgery) (SessionSurgery, error)
	DeleteByID(ctx context.Context, id SurgeryID) error
	DeleteBySessionID(ctx context.Context, sessionID SessionID) error
}

// ─── SessionPrescriptionRepository ──────────────────────────────────────────

type SessionPrescriptionRepository interface {
	FindBySessionID(ctx context.Context, sessionID SessionID) ([]SessionPrescription, error)
	FindByID(ctx context.Context, id PrescriptionID) (SessionPrescription, error)
	FindActivePrescriptionsByPet(ctx context.Context, petID uint, employeeIDs []uint, p page.Pagination) (page.Page[SessionPrescription], error)
	Save(ctx context.Context, p SessionPrescription) (SessionPrescription, error)
	DeleteByID(ctx context.Context, id PrescriptionID) error
	DeleteBySessionID(ctx context.Context, sessionID SessionID) error
}

// ─── SessionAttachmentRepository ─────────────────────────────────────────────

type SessionAttachmentRepository interface {
	FindBySessionID(ctx context.Context, sessionID SessionID) ([]SessionAttachment, error)
	FindByID(ctx context.Context, id AttachmentID) (SessionAttachment, error)
	Save(ctx context.Context, a SessionAttachment) (SessionAttachment, error)
	DeleteByID(ctx context.Context, id AttachmentID) error
	DeleteBySessionID(ctx context.Context, sessionID SessionID) error
}

// ─── SessionServiceRepository ────────────────────────────────────────────────

type SessionServiceRepository interface {
	FindBySessionID(ctx context.Context, sessionID SessionID) ([]SessionService, error)
	FindByID(ctx context.Context, id SessionServiceID) (SessionService, error)
	Save(ctx context.Context, s SessionService) (SessionService, error)
	DeleteByID(ctx context.Context, id SessionServiceID) error
	DeleteBySessionID(ctx context.Context, sessionID SessionID) error
}

// ─── Catalog repositories ────────────────────────────────────────────────────

type VaccineCatalogRepository interface {
	FindAll(ctx context.Context, p page.Pagination) (page.Page[VaccineCatalog], error)
	FindByID(ctx context.Context, id VaccineCatalogID) (VaccineCatalog, error)
	FindBySpecies(ctx context.Context, species string, p page.Pagination) (page.Page[VaccineCatalog], error)
	Save(ctx context.Context, v VaccineCatalog) (VaccineCatalog, error)
	DeleteByID(ctx context.Context, id VaccineCatalogID) error
}

type MedicationRepository interface {
	FindAll(ctx context.Context, p page.Pagination) (page.Page[Medication], error)
	FindByID(ctx context.Context, id MedicationID) (Medication, error)
	Search(ctx context.Context, term string, p page.Pagination) (page.Page[Medication], error)
	Save(ctx context.Context, m Medication) (Medication, error)
	DeleteByID(ctx context.Context, id MedicationID) error
}

type ServiceCatalogRepository interface {
	FindAll(ctx context.Context, p page.Pagination) (page.Page[ServiceCatalog], error)
	FindByID(ctx context.Context, id ServiceCatalogID) (ServiceCatalog, error)
	FindByCategory(ctx context.Context, cat ServiceCategory, p page.Pagination) (page.Page[ServiceCatalog], error)
	Save(ctx context.Context, s ServiceCatalog) (ServiceCatalog, error)
	DeleteByID(ctx context.Context, id ServiceCatalogID) error
}
