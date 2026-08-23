package medical

import (
	"clinic-vet-api/internal/shared/page"
	"time"
)

// ─── MedicalSessionSpecification ────────────────────────────────────────────

// MedicalSessionSpecification defines filters and pagination for session queries.
// It mirrors the CustomerSpecification pattern and can be used both as a query
// builder for the repository layer and as an in-memory predicate for tests.
type MedicalSessionSpecification struct {
	IDs            []SessionID
	PetIDs         []uint
	CustomerIDs    []uint
	EmployeeIDs    []uint
	AppointmentIDs []uint

	ClinicServices []ClinicService
	IsEmergency    *bool
	IsDeleted      *bool // nil = only active, true = only deleted, false = active

	VisitDateFrom *time.Time
	VisitDateTo   *time.Time
	FollowUpFrom  *time.Time
	FollowUpTo    *time.Time

	page.Pagination
}

// IsSatisfiedBy is an in-memory predicate, useful for unit tests.
func (s *MedicalSessionSpecification) IsSatisfiedBy(entity any) bool {
	ms, ok := entity.(MedicalSession)
	if !ok {
		return false
	}

	if len(s.IDs) > 0 && !containsSessionID(s.IDs, ms.ID) {
		return false
	}
	if len(s.PetIDs) > 0 && !containsUint(s.PetIDs, ms.PetID) {
		return false
	}
	if len(s.CustomerIDs) > 0 && !containsUint(s.CustomerIDs, ms.CustomerID) {
		return false
	}
	if len(s.EmployeeIDs) > 0 && !containsUint(s.EmployeeIDs, ms.EmployeeID) {
		return false
	}
	if len(s.ClinicServices) > 0 && !containsClinicService(s.ClinicServices, ms.ClinicService) {
		return false
	}
	if s.IsEmergency != nil && ms.IsEmergency != *s.IsEmergency {
		return false
	}
	if s.IsDeleted != nil {
		if *s.IsDeleted && !ms.IsDeleted() {
			return false
		}
		if !*s.IsDeleted && ms.IsDeleted() {
			return false
		}
	}
	if s.VisitDateFrom != nil && ms.VisitDate.Before(*s.VisitDateFrom) {
		return false
	}
	if s.VisitDateTo != nil && ms.VisitDate.After(*s.VisitDateTo) {
		return false
	}
	if s.FollowUpFrom != nil {
		if ms.FollowUpDate == nil || ms.FollowUpDate.Before(*s.FollowUpFrom) {
			return false
		}
	}
	if s.FollowUpTo != nil {
		if ms.FollowUpDate == nil || ms.FollowUpDate.After(*s.FollowUpTo) {
			return false
		}
	}
	return true
}

// ── Fluent builder methods ────────────────────────────────────────────────────

func (s *MedicalSessionSpecification) WithIDs(ids ...SessionID) *MedicalSessionSpecification {
	s.IDs = ids
	return s
}

func (s *MedicalSessionSpecification) WithPetIDs(ids ...uint) *MedicalSessionSpecification {
	s.PetIDs = ids
	return s
}

func (s *MedicalSessionSpecification) WithCustomerIDs(ids ...uint) *MedicalSessionSpecification {
	s.CustomerIDs = ids
	return s
}

func (s *MedicalSessionSpecification) WithEmployeeIDs(ids ...uint) *MedicalSessionSpecification {
	s.EmployeeIDs = ids
	return s
}

func (s *MedicalSessionSpecification) WithClinicServices(svc ...ClinicService) *MedicalSessionSpecification {
	s.ClinicServices = svc
	return s
}

func (s *MedicalSessionSpecification) WithIsEmergency(v bool) *MedicalSessionSpecification {
	s.IsEmergency = &v
	return s
}

func (s *MedicalSessionSpecification) WithIsDeleted(v bool) *MedicalSessionSpecification {
	s.IsDeleted = &v
	return s
}

func (s *MedicalSessionSpecification) WithVisitDateRange(from, to *time.Time) *MedicalSessionSpecification {
	s.VisitDateFrom = from
	s.VisitDateTo = to
	return s
}

func (s *MedicalSessionSpecification) WithFollowUpRange(from, to *time.Time) *MedicalSessionSpecification {
	s.FollowUpFrom = from
	s.FollowUpTo = to
	return s
}

func (s *MedicalSessionSpecification) WithPagination(p page.Pagination) *MedicalSessionSpecification {
	s.Pagination = p
	return s
}

// ── helpers ──────────────────────────────────────────────────────────────────

func containsSessionID(ids []SessionID, id SessionID) bool {
	for _, v := range ids {
		if v.Value() == id.Value() {
			return true
		}
	}
	return false
}

func containsUint(slice []uint, v uint) bool {
	for _, u := range slice {
		if u == v {
			return true
		}
	}
	return false
}

func containsClinicService(slice []ClinicService, v ClinicService) bool {
	for _, s := range slice {
		if s == v {
			return true
		}
	}
	return false
}

// ─── VaccinationHistorySpecification ────────────────────────────────────────

// VaccinationHistorySpecification filters the pet_vaccination_history view.
type VaccinationHistorySpecification struct {
	PetIDs            []uint
	EmployeeIDs       []uint
	VaccineCatalogIDs []uint
	DateFrom          *time.Time
	DateTo            *time.Time
	page.Pagination
}

func (s *VaccinationHistorySpecification) WithEmployeeIDs(ids ...uint) *VaccinationHistorySpecification {
	s.EmployeeIDs = ids
	return s
}

func (s *VaccinationHistorySpecification) WithPetIDs(ids ...uint) *VaccinationHistorySpecification {
	s.PetIDs = ids
	return s
}

func (s *VaccinationHistorySpecification) WithVaccineCatalogIDs(ids ...uint) *VaccinationHistorySpecification {
	s.VaccineCatalogIDs = ids
	return s
}

func (s *VaccinationHistorySpecification) WithDateRange(from, to *time.Time) *VaccinationHistorySpecification {
	s.DateFrom = from
	s.DateTo = to
	return s
}

func (s *VaccinationHistorySpecification) WithPagination(p page.Pagination) *VaccinationHistorySpecification {
	s.Pagination = p
	return s
}
