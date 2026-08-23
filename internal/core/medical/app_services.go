package medical

import (
	"clinic-vet-api/internal/shared/page"
	"context"
)

// ─── MedicalSessionService ───────────────────────────────────────────────────

// MedicalSessionService is the primary application service for clinical sessions.
// It orchestrates the session aggregate and delegates to extension services when
// additional records are required (vaccinations, surgeries, etc.).
type MedicalSessionService interface {
	// CreateSession opens a new medical session.
	CreateSession(ctx context.Context, cmd CreateSessionCommand) (MedicalSession, error)

	// UpdateSession applies partial updates to an existing session.
	UpdateSession(ctx context.Context, cmd UpdateSessionCommand) (MedicalSession, error)

	// GetSessionByID returns a session without extensions loaded.
	GetSessionByID(ctx context.Context, id SessionID) (MedicalSession, error)

	// GetSessionFull returns a session with all extensions eagerly loaded.
	GetSessionFull(ctx context.Context, id SessionID) (SessionWithExtensions, error)

	// GetSessionsBySpecification returns a filtered, paginated list of sessions.
	GetSessionsBySpecification(ctx context.Context, spec MedicalSessionSpecification) (page.Page[MedicalSession], error)

	// GetSessionsByPet returns sessions for a specific pet.
	GetSessionsByPet(ctx context.Context, petID uint, p page.Pagination) (page.Page[MedicalSession], error)

	// GetSessionsByCustomer returns sessions associated with a customer.
	GetSessionsByCustomer(ctx context.Context, customerID uint, p page.Pagination) (page.Page[MedicalSession], error)

	// GetSessionStats returns aggregate counts for dashboards.
	GetSessionStats(ctx context.Context) (SessionStats, error)

	// SoftDeleteSession marks a session as deleted.
	SoftDeleteSession(ctx context.Context, id SessionID) error

	// HardDeleteSession permanently removes a session and all its extensions.
	HardDeleteSession(ctx context.Context, id SessionID) error

	// RestoreSession reverses a soft-delete.
	RestoreSession(ctx context.Context, id SessionID) error
}

// ─── VaccinationService ───────────────────────────────────────────────────────

// VaccinationService manages vaccination records and the pet vaccination schema.
type VaccinationService interface {
	// AddVaccination records a new vaccine dose on an existing session.
	AddVaccination(ctx context.Context, cmd AddVaccinationCommand) (SessionVaccination, error)

	// UpdateVaccination updates mutable fields of a vaccination record.
	UpdateVaccination(ctx context.Context, cmd UpdateVaccinationCommand) (SessionVaccination, error)

	// RemoveVaccination deletes a single vaccination record.
	RemoveVaccination(ctx context.Context, id VaccinationID) error

	// GetVaccinationsBySession returns all vaccination records for a session.
	GetVaccinationsBySession(ctx context.Context, sessionID SessionID) ([]SessionVaccination, error)
	GetVaccinationByID(ctx context.Context, id VaccinationID) (SessionVaccination, error)

	// GetVaccinationHistory returns the full vaccination history for a pet,
	// optionally filtered and paginated.
	GetVaccinationHistory(ctx context.Context, spec VaccinationHistorySpecification) (page.Page[SessionVaccination], error)

	// GetPetVaccinationSummary returns the pet's vaccination status for each
	// known vaccine, highlighting which doses are upcoming or overdue.
	GetPetVaccinationSummary(ctx context.Context, petID uint, employeeIDs []uint) (PetVaccinationSummary, error)
}

// ─── SurgeryService ───────────────────────────────────────────────────────────

// SurgeryService manages surgical detail records for a session.
type SurgeryService interface {
	AddSurgery(ctx context.Context, cmd AddSurgeryCommand) (SessionSurgery, error)
	UpdateSurgery(ctx context.Context, cmd UpdateSurgeryCommand) (SessionSurgery, error)
	RemoveSurgery(ctx context.Context, id SurgeryID) error
	GetSurgeriesBySession(ctx context.Context, sessionID SessionID) ([]SessionSurgery, error)
	GetSurgeryByID(ctx context.Context, id SurgeryID) (SessionSurgery, error)
}

// ─── PrescriptionService ──────────────────────────────────────────────────────

// PrescriptionService manages medication prescriptions attached to sessions.
type PrescriptionService interface {
	AddPrescription(ctx context.Context, cmd AddPrescriptionCommand) (SessionPrescription, error)
	UpdatePrescription(ctx context.Context, cmd UpdatePrescriptionCommand) (SessionPrescription, error)
	RemovePrescription(ctx context.Context, id PrescriptionID) error
	GetPrescriptionsBySession(ctx context.Context, sessionID SessionID) ([]SessionPrescription, error)
	GetPrescriptionByID(ctx context.Context, id PrescriptionID) (SessionPrescription, error)

	// GetActivePrescriptionsByPet returns current (not-yet-ended) prescriptions
	// across all sessions for a pet.
	GetActivePrescriptionsByPet(ctx context.Context, petID uint, employeeIDs []uint, p page.Pagination) (page.Page[SessionPrescription], error)
}

// ─── AttachmentService ────────────────────────────────────────────────────────

// AttachmentService manages file attachments linked to sessions.
type AttachmentService interface {
	AddAttachment(ctx context.Context, cmd AddAttachmentCommand) (SessionAttachment, error)
	RemoveAttachment(ctx context.Context, id AttachmentID) error
	GetAttachmentsBySession(ctx context.Context, sessionID SessionID) ([]SessionAttachment, error)
	GetAttachmentByID(ctx context.Context, id AttachmentID) (SessionAttachment, error)
}

// ─── SessionServiceManager ────────────────────────────────────────────────────

// SessionServiceManager manages billable service items applied during a session.
type SessionServiceManager interface {
	AddService(ctx context.Context, cmd AddSessionServiceCommand) (SessionService, error)
	RemoveService(ctx context.Context, id SessionServiceID) error
	GetServicesBySession(ctx context.Context, sessionID SessionID) ([]SessionService, error)
	GetSessionServiceByID(ctx context.Context, id SessionServiceID) (SessionService, error)
}

// ─── Catalog services ─────────────────────────────────────────────────────────

// VaccineCatalogService manages the master vaccine list.
type VaccineCatalogService interface {
	CreateVaccine(ctx context.Context, cmd CreateVaccineCatalogCommand) (VaccineCatalog, error)
	GetVaccineByID(ctx context.Context, id VaccineCatalogID) (VaccineCatalog, error)
	ListVaccines(ctx context.Context, p page.Pagination) (page.Page[VaccineCatalog], error)
	ListVaccinesBySpecies(ctx context.Context, species string, p page.Pagination) (page.Page[VaccineCatalog], error)
	DeactivateVaccine(ctx context.Context, id VaccineCatalogID) error
}

// MedicationCatalogService manages the medication catalog.
type MedicationCatalogService interface {
	CreateMedication(ctx context.Context, cmd CreateMedicationCommand) (Medication, error)
	GetMedicationByID(ctx context.Context, id MedicationID) (Medication, error)
	ListMedications(ctx context.Context, p page.Pagination) (page.Page[Medication], error)
	SearchMedications(ctx context.Context, term string, p page.Pagination) (page.Page[Medication], error)
	DeactivateMedication(ctx context.Context, id MedicationID) error
}

// ServiceCatalogService manages the clinic's service/price catalog.
type ServiceCatalogService interface {
	CreateService(ctx context.Context, cmd CreateServiceCatalogCommand) (ServiceCatalog, error)
	GetServiceByID(ctx context.Context, id ServiceCatalogID) (ServiceCatalog, error)
	ListServices(ctx context.Context, p page.Pagination) (page.Page[ServiceCatalog], error)
	ListServicesByCategory(ctx context.Context, cat ServiceCategory, p page.Pagination) (page.Page[ServiceCatalog], error)
	DeactivateService(ctx context.Context, id ServiceCatalogID) error
}
