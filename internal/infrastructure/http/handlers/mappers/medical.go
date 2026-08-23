package mappers

import (
	"time"

	"clinic-vet-api/internal/core/medical"
	"clinic-vet-api/internal/infrastructure/http/handlers/dtos"
)

const dateTimeLayout = time.RFC3339
const dateLayout = "2006-01-02"

// ─── Session ─────────────────────────────────────────────────────────────────

func ParseTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	return time.Parse(dateTimeLayout, s)
}

func ParseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	return time.Parse(dateLayout, s)
}

func CreateSessionRequestToCommand(req dtos.CreateSessionRequest) (medical.CreateSessionCommand, error) {
	visitDate, err := ParseTime(req.VisitDate)
	if err != nil {
		return medical.CreateSessionCommand{}, err
	}
	svc, err := medical.ParseClinicService(req.ClinicService)
	if err != nil {
		return medical.CreateSessionCommand{}, err
	}
	cmd := medical.CreateSessionCommand{
		PetID:         req.PetID,
		CustomerID:    req.CustomerID,
		EmployeeID:    req.EmployeeID,
		AppointmentID: req.AppointmentID,
		ClinicService: svc,
		VisitType:     req.VisitType,
		VisitDate:     visitDate,
		IsEmergency:   req.IsEmergency,
		Symptoms:      req.Symptoms,
		Condition:     req.Condition,
		Diagnosis:     req.Diagnosis,
		Treatment:     req.Treatment,
		Notes:         req.Notes,
		Medications:   req.Medications,
		Vitals: medical.Vitals{
			WeightKg:        req.WeightKg,
			TemperatureC:    req.TemperatureC,
			HeartRate:       req.HeartRate,
			RespiratoryRate: req.RespiratoryRate,
		},
	}
	if req.FollowUpDate != nil && *req.FollowUpDate != "" {
		t, err := ParseTime(*req.FollowUpDate)
		if err != nil {
			return medical.CreateSessionCommand{}, err
		}
		cmd.FollowUpDate = &t
	}
	return cmd, nil
}

func UpdateSessionRequestToCommand(sessionID uint, req dtos.UpdateSessionRequest) (medical.UpdateSessionCommand, error) {
	cmd := medical.UpdateSessionCommand{
		ID:            medical.NewSessionID(sessionID),
		EmployeeID:    req.EmployeeID,
		AppointmentID: req.AppointmentID,
		ClinicService: nil,
		VisitType:     req.VisitType,
		VisitDate:     nil,
		IsEmergency:   req.IsEmergency,
		Symptoms:      req.Symptoms,
		Condition:     req.Condition,
		Diagnosis:     req.Diagnosis,
		Treatment:     req.Treatment,
		Notes:         req.Notes,
		Medications:   req.Medications,
		Vitals:        nil,
		FollowUpDate:  nil,
	}
	if req.ClinicService != nil {
		svc, err := medical.ParseClinicService(*req.ClinicService)
		if err != nil {
			return medical.UpdateSessionCommand{}, err
		}
		cmd.ClinicService = &svc
	}
	if req.VisitDate != nil && *req.VisitDate != "" {
		t, err := ParseTime(*req.VisitDate)
		if err != nil {
			return medical.UpdateSessionCommand{}, err
		}
		cmd.VisitDate = &t
	}
	if req.FollowUpDate != nil && *req.FollowUpDate != "" {
		t, err := ParseTime(*req.FollowUpDate)
		if err != nil {
			return medical.UpdateSessionCommand{}, err
		}
		cmd.FollowUpDate = &t
	}
	if req.WeightKg != nil || req.TemperatureC != nil || req.HeartRate != nil || req.RespiratoryRate != nil {
		cmd.Vitals = &medical.Vitals{
			WeightKg:        req.WeightKg,
			TemperatureC:    req.TemperatureC,
			HeartRate:       req.HeartRate,
			RespiratoryRate: req.RespiratoryRate,
		}
	}
	return cmd, nil
}

func SessionSearchToSpec(req dtos.SessionSearchRequest) (*medical.MedicalSessionSpecification, error) {
	spec := &medical.MedicalSessionSpecification{}
	if req.PetID > 0 {
		spec.WithPetIDs(req.PetID)
	}
	if req.CustomerID > 0 {
		spec.WithCustomerIDs(req.CustomerID)
	}
	if req.EmployeeID > 0 {
		spec.WithEmployeeIDs(req.EmployeeID)
	}
	if req.ClinicService != "" {
		svc, err := medical.ParseClinicService(req.ClinicService)
		if err != nil {
			return nil, err
		}
		spec.WithClinicServices(svc)
	}
	if req.IsEmergency != nil {
		spec.WithIsEmergency(*req.IsEmergency)
	}
	if req.VisitDateFrom != "" || req.VisitDateTo != "" {
		var from, to *time.Time
		if req.VisitDateFrom != "" {
			t, _ := ParseDate(req.VisitDateFrom)
			from = &t
		}
		if req.VisitDateTo != "" {
			t, _ := ParseDate(req.VisitDateTo)
			to = &t
		}
		spec.WithVisitDateRange(from, to)
	}
	spec.WithPagination(req.PaginationRequest.ToPagination())
	return spec, nil
}

// ─── Vaccination ─────────────────────────────────────────────────────────────

func AddVaccinationRequestToCommand(req dtos.AddVaccinationRequest) (medical.AddVaccinationCommand, error) {
	cmd := medical.AddVaccinationCommand{
		SessionID:        medical.NewSessionID(req.SessionID),
		VaccineCatalogID: medical.NewVaccineCatalogID(req.VaccineCatalogID),
		BatchNumber:      req.BatchNumber,
		DoseNumber:       req.DoseNumber,
		SiteOfInjection:  req.SiteOfInjection,
		NextDoseDate:     nil,
		ReactionNotes:    req.ReactionNotes,
		AdministeredBy:   req.AdministeredBy,
	}
	if req.ExpirationDate != nil && *req.ExpirationDate != "" {
		t, err := ParseDate(*req.ExpirationDate)
		if err != nil {
			return medical.AddVaccinationCommand{}, err
		}
		cmd.ExpirationDate = &t
	}
	if req.NextDoseDate != nil && *req.NextDoseDate != "" {
		t, err := ParseDate(*req.NextDoseDate)
		if err != nil {
			return medical.AddVaccinationCommand{}, err
		}
		cmd.NextDoseDate = &t
	}
	return cmd, nil
}

func UpdateVaccinationRequestToCommand(vaccinationID uint, req dtos.UpdateVaccinationRequest) (medical.UpdateVaccinationCommand, error) {
	cmd := medical.UpdateVaccinationCommand{
		ID:              medical.NewVaccinationID(vaccinationID),
		BatchNumber:     req.BatchNumber,
		SiteOfInjection: req.SiteOfInjection,
		ReactionNotes:   req.ReactionNotes,
	}
	if req.NextDoseDate != nil && *req.NextDoseDate != "" {
		t, err := ParseDate(*req.NextDoseDate)
		if err != nil {
			return medical.UpdateVaccinationCommand{}, err
		}
		cmd.NextDoseDate = &t
	}
	return cmd, nil
}

// ─── Surgery ────────────────────────────────────────────────────────────────

func AddSurgeryRequestToCommand(req dtos.AddSurgeryRequest) (medical.AddSurgeryCommand, error) {
	outcome, err := medical.ParseSurgeryOutcome(req.Outcome)
	if err != nil {
		return medical.AddSurgeryCommand{}, err
	}
	return medical.AddSurgeryCommand{
		SessionID:       medical.NewSessionID(req.SessionID),
		ProcedureName:   req.ProcedureName,
		AnesthesiaType:  req.AnesthesiaType,
		AnesthesiaAgent: req.AnesthesiaAgent,
		PreOpNotes:      req.PreOpNotes,
		IntraOpNotes:    req.IntraOpNotes,
		PostOpNotes:     req.PostOpNotes,
		DurationMinutes: req.DurationMinutes,
		Outcome:         outcome,
		SurgeonID:       req.SurgeonID,
	}, nil
}

func UpdateSurgeryRequestToCommand(surgeryID uint, req dtos.UpdateSurgeryRequest) (medical.UpdateSurgeryCommand, error) {
	cmd := medical.UpdateSurgeryCommand{
		ID:              medical.NewSurgeryID(surgeryID),
		AnesthesiaType:  req.AnesthesiaType,
		AnesthesiaAgent: req.AnesthesiaAgent,
		PreOpNotes:      req.PreOpNotes,
		IntraOpNotes:    req.IntraOpNotes,
		PostOpNotes:     req.PostOpNotes,
		DurationMinutes: req.DurationMinutes,
		Outcome:         nil,
	}
	if req.Outcome != nil {
		o, err := medical.ParseSurgeryOutcome(*req.Outcome)
		if err != nil {
			return medical.UpdateSurgeryCommand{}, err
		}
		cmd.Outcome = &o
	}
	return cmd, nil
}

// ─── Prescription ───────────────────────────────────────────────────────────

func AddPrescriptionRequestToCommand(req dtos.AddPrescriptionRequest) (medical.AddPrescriptionCommand, error) {
	startDate, err := ParseDate(req.StartDate)
	if err != nil {
		return medical.AddPrescriptionCommand{}, err
	}
	return medical.AddPrescriptionCommand{
		SessionID:    medical.NewSessionID(req.SessionID),
		MedicationID: medical.NewMedicationID(req.MedicationID),
		Dosage:       req.Dosage,
		Frequency:    req.Frequency,
		DurationDays: req.DurationDays,
		Route:        req.Route,
		Instructions: req.Instructions,
		StartDate:    startDate,
	}, nil
}

func UpdatePrescriptionRequestToCommand(prescriptionID uint, req dtos.UpdatePrescriptionRequest) medical.UpdatePrescriptionCommand {
	return medical.UpdatePrescriptionCommand{
		ID:           medical.NewPrescriptionID(prescriptionID),
		Dosage:       req.Dosage,
		Frequency:    req.Frequency,
		DurationDays: req.DurationDays,
		Route:        req.Route,
		Instructions: req.Instructions,
	}
}

// ─── Attachment ──────────────────────────────────────────────────────────────

func AddAttachmentRequestToCommand(req dtos.AddAttachmentRequest) (medical.AddAttachmentCommand, error) {
	ft, err := medical.ParseAttachmentFileType(req.FileType)
	if err != nil {
		return medical.AddAttachmentCommand{}, err
	}
	return medical.AddAttachmentCommand{
		SessionID:   medical.NewSessionID(req.SessionID),
		FileType:    ft,
		FileURL:     req.FileURL,
		Description: req.Description,
		UploadedBy:  req.UploadedBy,
	}, nil
}

// ─── Session service ───────────────────────────────────────────────────────

func AddSessionServiceRequestToCommand(req dtos.AddSessionServiceRequest) medical.AddSessionServiceCommand {
	return medical.AddSessionServiceCommand{
		SessionID:        medical.NewSessionID(req.SessionID),
		ServiceCatalogID: medical.NewServiceCatalogID(req.ServiceCatalogID),
		Quantity:         req.Quantity,
		PriceApplied:     req.PriceApplied,
		Notes:            req.Notes,
	}
}

// ─── Catalogs ───────────────────────────────────────────────────────────────

func CreateVaccineCatalogRequestToCommand(req dtos.CreateVaccineCatalogRequest) medical.CreateVaccineCatalogCommand {
	return medical.CreateVaccineCatalogCommand{
		Name:          req.Name,
		Manufacturer:  req.Manufacturer,
		Species:       req.Species,
		DiseaseTarget: req.DiseaseTarget,
		TotalDoses:    req.TotalDoses,
		ScheduleDays:  req.ScheduleDays,
		Notes:         req.Notes,
	}
}

func CreateMedicationCatalogRequestToCommand(req dtos.CreateMedicationCatalogRequest) medical.CreateMedicationCommand {
	return medical.CreateMedicationCommand{
		Name:                 req.Name,
		ActiveIngredient:     req.ActiveIngredient,
		Presentation:         req.Presentation,
		Unit:                 req.Unit,
		RequiresPrescription: req.RequiresPrescription,
		SpeciesWarnings:      req.SpeciesWarnings,
	}
}

func CreateServiceCatalogRequestToCommand(req dtos.CreateServiceCatalogRequest) (medical.CreateServiceCatalogCommand, error) {
	cat, err := medical.ParseServiceCategory(req.Category)
	if err != nil {
		return medical.CreateServiceCatalogCommand{}, err
	}
	return medical.CreateServiceCatalogCommand{
		Name:            req.Name,
		Category:        cat,
		Description:     req.Description,
		BasePrice:       req.BasePrice,
		DurationMinutes: req.DurationMinutes,
		RequiresFasting: req.RequiresFasting,
	}, nil
}

func VaccinationHistoryRequestToSpec(req dtos.VaccinationHistoryRequest) *medical.VaccinationHistorySpecification {
	spec := &medical.VaccinationHistorySpecification{}
	if req.PetID > 0 {
		spec.WithPetIDs(req.PetID)
	}
	if req.VaccineID > 0 {
		spec.WithVaccineCatalogIDs(req.VaccineID)
	}
	if req.DateFrom != "" || req.DateTo != "" {
		var from, to *time.Time
		if req.DateFrom != "" {
			t, _ := ParseDate(req.DateFrom)
			from = &t
		}
		if req.DateTo != "" {
			t, _ := ParseDate(req.DateTo)
			to = &t
		}
		spec.WithDateRange(from, to)
	}
	spec.WithPagination(req.PaginationRequest.ToPagination())
	return spec
}
