package medical

import (
	"context"
	"fmt"
	"time"

	"clinic-vet-api/internal/core/employees"
	"clinic-vet-api/internal/core/pets"
	"clinic-vet-api/internal/shared/page"
)

// applicationService keeps the medical aggregate's persistence dependencies in
// one place. Constructors below expose only the interfaces each endpoint needs.
type applicationService struct {
	sessions        MedicalSessionRepository
	vaccinations    SessionVaccinationRepository
	surgeries       SessionSurgeryRepository
	prescriptions   SessionPrescriptionRepository
	attachments     SessionAttachmentRepository
	services        SessionServiceRepository
	vaccines        VaccineCatalogRepository
	medications     MedicationRepository
	catalogServices ServiceCatalogRepository
	pets            pets.PetRepository
	employees       employees.EmployeeRepository
}

func NewMedicalSessionApplicationService(s MedicalSessionRepository, v SessionVaccinationRepository, su SessionSurgeryRepository, p SessionPrescriptionRepository, a SessionAttachmentRepository, ss SessionServiceRepository, pets pets.PetRepository, employees employees.EmployeeRepository) MedicalSessionService {
	return &applicationService{sessions: s, vaccinations: v, surgeries: su, prescriptions: p, attachments: a, services: ss, pets: pets, employees: employees}
}
func NewVaccinationApplicationService(v SessionVaccinationRepository, s MedicalSessionRepository, c VaccineCatalogRepository) VaccinationService {
	return &applicationService{vaccinations: v, sessions: s, vaccines: c}
}
func NewSurgeryApplicationService(su SessionSurgeryRepository, s MedicalSessionRepository) SurgeryService {
	return &applicationService{surgeries: su, sessions: s}
}
func NewPrescriptionApplicationService(p SessionPrescriptionRepository, s MedicalSessionRepository, m MedicationRepository) PrescriptionService {
	return &applicationService{prescriptions: p, sessions: s, medications: m}
}
func NewAttachmentApplicationService(a SessionAttachmentRepository, s MedicalSessionRepository) AttachmentService {
	return &applicationService{attachments: a, sessions: s}
}
func NewSessionServiceApplicationService(ss SessionServiceRepository, s MedicalSessionRepository, c ServiceCatalogRepository) SessionServiceManager {
	return &applicationService{services: ss, sessions: s, catalogServices: c}
}
func NewVaccineCatalogApplicationService(r VaccineCatalogRepository) VaccineCatalogService {
	return &applicationService{vaccines: r}
}
func NewMedicationCatalogApplicationService(r MedicationRepository) MedicationCatalogService {
	return &applicationService{medications: r}
}
func NewServiceCatalogApplicationService(r ServiceCatalogRepository) ServiceCatalogService {
	return &applicationService{catalogServices: r}
}

func (a *applicationService) validateSessionReferences(ctx context.Context, petID, customerID, employeeID uint) error {
	if _, err := a.pets.FindByIDAndCustomerID(ctx, pets.NewPetID(petID), customerID); err != nil {
		return err
	}
	ok, err := a.employees.ExistsByID(ctx, employees.NewEmployeeID(employeeID))
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("employee %d does not exist", employeeID)
	}
	return nil
}
func (a *applicationService) ensureSession(ctx context.Context, id SessionID) error {
	ok, err := a.sessions.ExistsByID(ctx, id)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("medical session %d does not exist", id.Value())
	}
	return nil
}

func (a *applicationService) CreateSession(ctx context.Context, c CreateSessionCommand) (MedicalSession, error) {
	if err := a.validateSessionReferences(ctx, c.PetID, c.CustomerID, c.EmployeeID); err != nil {
		return MedicalSession{}, err
	}
	x := MedicalSession{ID: NewSessionID(0), PetID: c.PetID, CustomerID: c.CustomerID, EmployeeID: c.EmployeeID, AppointmentID: c.AppointmentID, ClinicService: c.ClinicService, VisitType: c.VisitType, VisitDate: c.VisitDate, IsEmergency: c.IsEmergency, Symptoms: c.Symptoms, Condition: c.Condition, Diagnosis: c.Diagnosis, Treatment: c.Treatment, Notes: c.Notes, Vitals: c.Vitals, Medications: c.Medications, FollowUpDate: c.FollowUpDate}
	if err := x.Validate(); err != nil {
		return MedicalSession{}, err
	}
	return a.sessions.Save(ctx, x)
}
func (a *applicationService) UpdateSession(ctx context.Context, c UpdateSessionCommand) (MedicalSession, error) {
	x, err := a.sessions.FindByID(ctx, c.ID)
	if err != nil {
		return MedicalSession{}, err
	}
	if c.EmployeeID != nil {
		x.EmployeeID = *c.EmployeeID
	}
	if c.AppointmentID != nil {
		x.AppointmentID = c.AppointmentID
	}
	if c.ClinicService != nil {
		x.ClinicService = *c.ClinicService
	}
	if c.VisitType != nil {
		x.VisitType = *c.VisitType
	}
	if c.VisitDate != nil {
		x.VisitDate = *c.VisitDate
	}
	if c.IsEmergency != nil {
		x.IsEmergency = *c.IsEmergency
	}
	if c.Symptoms != nil {
		x.Symptoms = c.Symptoms
	}
	if c.Condition != nil {
		x.Condition = c.Condition
	}
	if c.Diagnosis != nil {
		x.Diagnosis = c.Diagnosis
	}
	if c.Treatment != nil {
		x.Treatment = c.Treatment
	}
	if c.Notes != nil {
		x.Notes = c.Notes
	}
	if c.Medications != nil {
		x.Medications = c.Medications
	}
	if c.FollowUpDate != nil {
		x.FollowUpDate = c.FollowUpDate
	}
	if c.Vitals != nil {
		x.Vitals = *c.Vitals
	}
	if err := a.validateSessionReferences(ctx, x.PetID, x.CustomerID, x.EmployeeID); err != nil {
		return MedicalSession{}, err
	}
	if err := x.Validate(); err != nil {
		return MedicalSession{}, err
	}
	return a.sessions.Save(ctx, x)
}
func (a *applicationService) GetSessionByID(ctx context.Context, id SessionID) (MedicalSession, error) {
	return a.sessions.FindByID(ctx, id)
}
func (a *applicationService) GetSessionFull(ctx context.Context, id SessionID) (SessionWithExtensions, error) {
	x, err := a.sessions.FindByID(ctx, id)
	if err != nil {
		return SessionWithExtensions{}, err
	}
	v, err := a.vaccinations.FindBySessionID(ctx, id)
	if err != nil {
		return SessionWithExtensions{}, err
	}
	su, err := a.surgeries.FindBySessionID(ctx, id)
	if err != nil {
		return SessionWithExtensions{}, err
	}
	p, err := a.prescriptions.FindBySessionID(ctx, id)
	if err != nil {
		return SessionWithExtensions{}, err
	}
	at, err := a.attachments.FindBySessionID(ctx, id)
	if err != nil {
		return SessionWithExtensions{}, err
	}
	ss, err := a.services.FindBySessionID(ctx, id)
	if err != nil {
		return SessionWithExtensions{}, err
	}
	return SessionWithExtensions{Session: x, Vaccinations: v, Surgeries: su, Prescriptions: p, Attachments: at, Services: ss}, nil
}
func (a *applicationService) GetSessionsBySpecification(ctx context.Context, s MedicalSessionSpecification) (page.Page[MedicalSession], error) {
	return a.sessions.FindBySpecification(ctx, s)
}
func (a *applicationService) GetSessionsByPet(ctx context.Context, id uint, p page.Pagination) (page.Page[MedicalSession], error) {
	return a.sessions.FindByPetID(ctx, id, p)
}
func (a *applicationService) GetSessionsByCustomer(ctx context.Context, id uint, p page.Pagination) (page.Page[MedicalSession], error) {
	return a.sessions.FindByCustomerID(ctx, id, p)
}
func (a *applicationService) GetSessionStats(ctx context.Context) (SessionStats, error) {
	t, e := a.sessions.CountAll(ctx)
	if e != nil {
		return SessionStats{}, e
	}
	active, e := a.sessions.CountActive(ctx)
	if e != nil {
		return SessionStats{}, e
	}
	by, e := a.sessions.CountByClinicService(ctx)
	if e != nil {
		return SessionStats{}, e
	}
	em := true
	p, e := a.sessions.FindBySpecification(ctx, MedicalSessionSpecification{IsEmergency: &em, Pagination: page.Pagination{Number: 1, Size: 1}})
	if e != nil {
		return SessionStats{}, e
	}
	return SessionStats{Total: t, Active: active, Emergency: int64(p.Metadata.TotalCount), ByClinicService: by}, nil
}
func (a *applicationService) SoftDeleteSession(ctx context.Context, id SessionID) error {
	return a.sessions.SoftDelete(ctx, id)
}
func (a *applicationService) HardDeleteSession(ctx context.Context, id SessionID) error {
	return a.sessions.HardDelete(ctx, id)
}
func (a *applicationService) RestoreSession(ctx context.Context, id SessionID) error {
	return a.sessions.RestoreByID(ctx, id)
}

func (a *applicationService) AddVaccination(ctx context.Context, c AddVaccinationCommand) (SessionVaccination, error) {
	if err := a.ensureSession(ctx, c.SessionID); err != nil {
		return SessionVaccination{}, err
	}
	v, err := a.vaccines.FindByID(ctx, c.VaccineCatalogID)
	if err != nil {
		return SessionVaccination{}, err
	}
	if !v.IsActive {
		return SessionVaccination{}, fmt.Errorf("vaccine is inactive")
	}
	x := SessionVaccination{ID: NewVaccinationID(0), SessionID: c.SessionID, VaccineCatalogID: c.VaccineCatalogID, BatchNumber: c.BatchNumber, DoseNumber: c.DoseNumber, ExpirationDate: c.ExpirationDate, SiteOfInjection: c.SiteOfInjection, NextDoseDate: c.NextDoseDate, ReactionNotes: c.ReactionNotes, AdministeredBy: c.AdministeredBy}
	if err := x.Validate(); err != nil {
		return SessionVaccination{}, err
	}
	return a.vaccinations.Save(ctx, x)
}
func (a *applicationService) UpdateVaccination(ctx context.Context, c UpdateVaccinationCommand) (SessionVaccination, error) {
	x, e := a.vaccinations.FindByID(ctx, c.ID)
	if e != nil {
		return x, e
	}
	if c.BatchNumber != nil {
		x.BatchNumber = c.BatchNumber
	}
	if c.SiteOfInjection != nil {
		x.SiteOfInjection = c.SiteOfInjection
	}
	if c.NextDoseDate != nil {
		x.NextDoseDate = c.NextDoseDate
	}
	if c.ReactionNotes != nil {
		x.ReactionNotes = c.ReactionNotes
	}
	return a.vaccinations.Save(ctx, x)
}
func (a *applicationService) RemoveVaccination(ctx context.Context, id VaccinationID) error {
	return a.vaccinations.DeleteByID(ctx, id)
}
func (a *applicationService) GetVaccinationsBySession(ctx context.Context, id SessionID) ([]SessionVaccination, error) {
	if e := a.ensureSession(ctx, id); e != nil {
		return nil, e
	}
	return a.vaccinations.FindBySessionID(ctx, id)
}
func (a *applicationService) GetVaccinationByID(ctx context.Context, id VaccinationID) (SessionVaccination, error) {
	return a.vaccinations.FindByID(ctx, id)
}
func (a *applicationService) GetVaccinationHistory(ctx context.Context, s VaccinationHistorySpecification) (page.Page[SessionVaccination], error) {
	return a.vaccinations.FindHistoryBySpec(ctx, s)
}
func (a *applicationService) GetPetVaccinationSummary(ctx context.Context, petID uint, employeeIDs []uint) (PetVaccinationSummary, error) {
	sessions, e := a.sessions.FindBySpecification(ctx, MedicalSessionSpecification{
		PetIDs:      []uint{petID},
		EmployeeIDs: employeeIDs,
		Pagination:  page.Pagination{Number: 1, Size: page.MaxPageSize},
	})
	if e != nil {
		return PetVaccinationSummary{}, e
	}
	latest := map[uint]time.Time{}
	for _, s := range sessions.Items {
		vs, e := a.vaccinations.FindBySessionID(ctx, s.ID)
		if e != nil {
			return PetVaccinationSummary{}, e
		}
		for _, v := range vs {
			if s.VisitDate.After(latest[v.VaccineCatalogID.Value()]) {
				latest[v.VaccineCatalogID.Value()] = s.VisitDate
			}
		}
	}
	catalog, e := a.vaccines.FindAll(ctx, page.Pagination{Number: 1, Size: page.MaxPageSize})
	if e != nil {
		return PetVaccinationSummary{}, e
	}
	out := PetVaccinationSummary{PetID: petID}
	for _, v := range catalog.Items {
		last, ok := latest[v.ID.Value()]
		if !ok {
			continue
		}
		var lp interface{} = last
		st := VaccinationStatus{VaccineName: v.Name, LastDoseDate: &lp}
		if len(v.ScheduleDays) > 1 {
			next := last.AddDate(0, 0, v.ScheduleDays[len(v.ScheduleDays)-1])
			var np interface{} = next
			st.NextDoseDate = &np
			st.IsOverdue = next.Before(time.Now())
		}
		out.Vaccinations = append(out.Vaccinations, st)
	}
	return out, nil
}

func (a *applicationService) AddSurgery(ctx context.Context, c AddSurgeryCommand) (SessionSurgery, error) {
	if e := a.ensureSession(ctx, c.SessionID); e != nil {
		return SessionSurgery{}, e
	}
	x := SessionSurgery{ID: NewSurgeryID(0), SessionID: c.SessionID, ProcedureName: c.ProcedureName, AnesthesiaType: c.AnesthesiaType, AnesthesiaAgent: c.AnesthesiaAgent, PreOpNotes: c.PreOpNotes, IntraOpNotes: c.IntraOpNotes, PostOpNotes: c.PostOpNotes, DurationMinutes: c.DurationMinutes, Outcome: c.Outcome, SurgeonID: c.SurgeonID}
	if e := x.Validate(); e != nil {
		return x, e
	}
	return a.surgeries.Save(ctx, x)
}
func (a *applicationService) UpdateSurgery(ctx context.Context, c UpdateSurgeryCommand) (SessionSurgery, error) {
	x, e := a.surgeries.FindByID(ctx, c.ID)
	if e != nil {
		return x, e
	}
	if c.AnesthesiaType != nil {
		x.AnesthesiaType = c.AnesthesiaType
	}
	if c.AnesthesiaAgent != nil {
		x.AnesthesiaAgent = c.AnesthesiaAgent
	}
	if c.PreOpNotes != nil {
		x.PreOpNotes = c.PreOpNotes
	}
	if c.IntraOpNotes != nil {
		x.IntraOpNotes = c.IntraOpNotes
	}
	if c.PostOpNotes != nil {
		x.PostOpNotes = c.PostOpNotes
	}
	if c.DurationMinutes != nil {
		x.DurationMinutes = c.DurationMinutes
	}
	if c.Outcome != nil {
		x.Outcome = *c.Outcome
	}
	return a.surgeries.Save(ctx, x)
}
func (a *applicationService) RemoveSurgery(ctx context.Context, id SurgeryID) error {
	return a.surgeries.DeleteByID(ctx, id)
}
func (a *applicationService) GetSurgeriesBySession(ctx context.Context, id SessionID) ([]SessionSurgery, error) {
	if e := a.ensureSession(ctx, id); e != nil {
		return nil, e
	}
	return a.surgeries.FindBySessionID(ctx, id)
}
func (a *applicationService) GetSurgeryByID(ctx context.Context, id SurgeryID) (SessionSurgery, error) {
	return a.surgeries.FindByID(ctx, id)
}

func (a *applicationService) AddPrescription(ctx context.Context, c AddPrescriptionCommand) (SessionPrescription, error) {
	if e := a.ensureSession(ctx, c.SessionID); e != nil {
		return SessionPrescription{}, e
	}
	m, e := a.medications.FindByID(ctx, c.MedicationID)
	if e != nil {
		return SessionPrescription{}, e
	}
	if !m.IsActive {
		return SessionPrescription{}, fmt.Errorf("medication is inactive")
	}
	x := SessionPrescription{ID: NewPrescriptionID(0), SessionID: c.SessionID, MedicationID: c.MedicationID, Dosage: c.Dosage, Frequency: c.Frequency, DurationDays: c.DurationDays, Route: c.Route, Instructions: c.Instructions, StartDate: c.StartDate}
	if e := x.Validate(); e != nil {
		return x, e
	}
	return a.prescriptions.Save(ctx, x)
}
func (a *applicationService) UpdatePrescription(ctx context.Context, c UpdatePrescriptionCommand) (SessionPrescription, error) {
	x, e := a.prescriptions.FindByID(ctx, c.ID)
	if e != nil {
		return x, e
	}
	if c.Dosage != nil {
		x.Dosage = *c.Dosage
	}
	if c.Frequency != nil {
		x.Frequency = *c.Frequency
	}
	if c.DurationDays != nil {
		x.DurationDays = c.DurationDays
	}
	if c.Route != nil {
		x.Route = c.Route
	}
	if c.Instructions != nil {
		x.Instructions = c.Instructions
	}
	return a.prescriptions.Save(ctx, x)
}
func (a *applicationService) RemovePrescription(ctx context.Context, id PrescriptionID) error {
	return a.prescriptions.DeleteByID(ctx, id)
}
func (a *applicationService) GetPrescriptionsBySession(ctx context.Context, id SessionID) ([]SessionPrescription, error) {
	if e := a.ensureSession(ctx, id); e != nil {
		return nil, e
	}
	return a.prescriptions.FindBySessionID(ctx, id)
}
func (a *applicationService) GetPrescriptionByID(ctx context.Context, id PrescriptionID) (SessionPrescription, error) {
	return a.prescriptions.FindByID(ctx, id)
}
func (a *applicationService) GetActivePrescriptionsByPet(ctx context.Context, id uint, employeeIDs []uint, p page.Pagination) (page.Page[SessionPrescription], error) {
	return a.prescriptions.FindActivePrescriptionsByPet(ctx, id, employeeIDs, p)
}

func (a *applicationService) AddAttachment(ctx context.Context, c AddAttachmentCommand) (SessionAttachment, error) {
	if e := a.ensureSession(ctx, c.SessionID); e != nil {
		return SessionAttachment{}, e
	}
	x := SessionAttachment{ID: NewAttachmentID(0), SessionID: c.SessionID, FileType: c.FileType, FileURL: c.FileURL, Description: c.Description, UploadedBy: c.UploadedBy}
	if e := x.Validate(); e != nil {
		return x, e
	}
	return a.attachments.Save(ctx, x)
}
func (a *applicationService) RemoveAttachment(ctx context.Context, id AttachmentID) error {
	return a.attachments.DeleteByID(ctx, id)
}
func (a *applicationService) GetAttachmentsBySession(ctx context.Context, id SessionID) ([]SessionAttachment, error) {
	if e := a.ensureSession(ctx, id); e != nil {
		return nil, e
	}
	return a.attachments.FindBySessionID(ctx, id)
}
func (a *applicationService) GetAttachmentByID(ctx context.Context, id AttachmentID) (SessionAttachment, error) {
	return a.attachments.FindByID(ctx, id)
}
func (a *applicationService) AddService(ctx context.Context, c AddSessionServiceCommand) (SessionService, error) {
	if e := a.ensureSession(ctx, c.SessionID); e != nil {
		return SessionService{}, e
	}
	s, e := a.catalogServices.FindByID(ctx, c.ServiceCatalogID)
	if e != nil {
		return SessionService{}, e
	}
	if !s.IsActive {
		return SessionService{}, fmt.Errorf("service is inactive")
	}
	x := SessionService{ID: NewSessionServiceID(0), SessionID: c.SessionID, ServiceCatalogID: c.ServiceCatalogID, Quantity: c.Quantity, PriceApplied: c.PriceApplied, Notes: c.Notes}
	if e := x.Validate(); e != nil {
		return x, e
	}
	return a.services.Save(ctx, x)
}
func (a *applicationService) RemoveService(ctx context.Context, id SessionServiceID) error {
	return a.services.DeleteByID(ctx, id)
}
func (a *applicationService) GetServicesBySession(ctx context.Context, id SessionID) ([]SessionService, error) {
	if e := a.ensureSession(ctx, id); e != nil {
		return nil, e
	}
	return a.services.FindBySessionID(ctx, id)
}
func (a *applicationService) GetSessionServiceByID(ctx context.Context, id SessionServiceID) (SessionService, error) {
	return a.services.FindByID(ctx, id)
}

func (a *applicationService) CreateVaccine(ctx context.Context, c CreateVaccineCatalogCommand) (VaccineCatalog, error) {
	if c.Name == "" || c.TotalDoses < 1 {
		return VaccineCatalog{}, fmt.Errorf("vaccine name and total doses are required")
	}
	return a.vaccines.Save(ctx, VaccineCatalog{ID: NewVaccineCatalogID(0), Name: c.Name, Manufacturer: c.Manufacturer, Species: c.Species, DiseaseTarget: c.DiseaseTarget, TotalDoses: c.TotalDoses, ScheduleDays: c.ScheduleDays, Notes: c.Notes, IsActive: true})
}
func (a *applicationService) GetVaccineByID(ctx context.Context, id VaccineCatalogID) (VaccineCatalog, error) {
	return a.vaccines.FindByID(ctx, id)
}
func (a *applicationService) ListVaccines(ctx context.Context, p page.Pagination) (page.Page[VaccineCatalog], error) {
	return a.vaccines.FindAll(ctx, p)
}
func (a *applicationService) ListVaccinesBySpecies(ctx context.Context, s string, p page.Pagination) (page.Page[VaccineCatalog], error) {
	return a.vaccines.FindBySpecies(ctx, s, p)
}
func (a *applicationService) DeactivateVaccine(ctx context.Context, id VaccineCatalogID) error {
	x, e := a.vaccines.FindByID(ctx, id)
	if e != nil {
		return e
	}
	x.IsActive = false
	_, e = a.vaccines.Save(ctx, x)
	return e
}
func (a *applicationService) CreateMedication(ctx context.Context, c CreateMedicationCommand) (Medication, error) {
	if c.Name == "" {
		return Medication{}, fmt.Errorf("medication name is required")
	}
	return a.medications.Save(ctx, Medication{ID: NewMedicationID(0), Name: c.Name, ActiveIngredient: c.ActiveIngredient, Presentation: c.Presentation, Unit: c.Unit, RequiresPrescription: c.RequiresPrescription, SpeciesWarnings: c.SpeciesWarnings, IsActive: true})
}
func (a *applicationService) GetMedicationByID(ctx context.Context, id MedicationID) (Medication, error) {
	return a.medications.FindByID(ctx, id)
}
func (a *applicationService) ListMedications(ctx context.Context, p page.Pagination) (page.Page[Medication], error) {
	return a.medications.FindAll(ctx, p)
}
func (a *applicationService) SearchMedications(ctx context.Context, t string, p page.Pagination) (page.Page[Medication], error) {
	return a.medications.Search(ctx, t, p)
}
func (a *applicationService) DeactivateMedication(ctx context.Context, id MedicationID) error {
	x, e := a.medications.FindByID(ctx, id)
	if e != nil {
		return e
	}
	x.IsActive = false
	_, e = a.medications.Save(ctx, x)
	return e
}
func (a *applicationService) CreateService(ctx context.Context, c CreateServiceCatalogCommand) (ServiceCatalog, error) {
	if c.Name == "" {
		return ServiceCatalog{}, fmt.Errorf("service name is required")
	}
	if _, e := ParseServiceCategory(string(c.Category)); e != nil {
		return ServiceCatalog{}, e
	}
	return a.catalogServices.Save(ctx, ServiceCatalog{ID: NewServiceCatalogID(0), Name: c.Name, Category: c.Category, Description: c.Description, BasePrice: c.BasePrice, DurationMinutes: c.DurationMinutes, RequiresFasting: c.RequiresFasting, IsActive: true})
}
func (a *applicationService) GetServiceByID(ctx context.Context, id ServiceCatalogID) (ServiceCatalog, error) {
	return a.catalogServices.FindByID(ctx, id)
}
func (a *applicationService) ListServices(ctx context.Context, p page.Pagination) (page.Page[ServiceCatalog], error) {
	return a.catalogServices.FindAll(ctx, p)
}
func (a *applicationService) ListServicesByCategory(ctx context.Context, c ServiceCategory, p page.Pagination) (page.Page[ServiceCatalog], error) {
	return a.catalogServices.FindByCategory(ctx, c, p)
}
func (a *applicationService) DeactivateService(ctx context.Context, id ServiceCatalogID) error {
	x, e := a.catalogServices.FindByID(ctx, id)
	if e != nil {
		return e
	}
	x.IsActive = false
	_, e = a.catalogServices.Save(ctx, x)
	return e
}
