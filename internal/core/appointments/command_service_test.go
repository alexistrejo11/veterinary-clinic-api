package appointments

import (
	"context"
	"errors"
	"testing"
	"time"

	"clinic-vet-api/internal/core/customers"
	"clinic-vet-api/internal/core/employees"
	"clinic-vet-api/internal/core/pets"
	"clinic-vet-api/internal/shared/page"
)

type appointmentRepositoryStub struct {
	appointment     Appointment
	saved           Appointment
	byEmployeeCalls int
	byCustomerCalls int
	saveCalls       int
}

func (r *appointmentRepositoryStub) FindByID(context.Context, AppointmentID) (Appointment, error) {
	return r.appointment, nil
}
func (r *appointmentRepositoryStub) FindByIDAndCustomerID(context.Context, AppointmentID, customers.CustomerID) (Appointment, error) {
	r.byCustomerCalls++
	return r.appointment, nil
}
func (r *appointmentRepositoryStub) FindByIDAndEmployeeID(context.Context, AppointmentID, employees.EmployeeID) (Appointment, error) {
	r.byEmployeeCalls++
	return r.appointment, nil
}
func (r *appointmentRepositoryStub) Find(context.Context, *AppointmentSpecification) (page.Page[Appointment], error) {
	return page.Page[Appointment]{}, nil
}
func (r *appointmentRepositoryStub) Count(context.Context, *AppointmentSpecification) (int64, error) {
	return 0, nil
}
func (r *appointmentRepositoryStub) ExistsByID(context.Context, AppointmentID) (bool, error) {
	return true, nil
}
func (r *appointmentRepositoryStub) Save(_ context.Context, appointment *Appointment) error {
	r.saveCalls++
	if appointment.ID.IntegerID == nil || appointment.ID.IsZero() {
		appointment.SetID(NewAppointmentID(99))
	}
	r.saved = *appointment
	return nil
}
func (r *appointmentRepositoryStub) DeleteByID(context.Context, AppointmentID, bool) error {
	return nil
}
func (r *appointmentRepositoryStub) RestoreByID(context.Context, AppointmentID) error { return nil }

type petRepositoryStub struct{ err error }

func (r petRepositoryStub) FindBySpecification(context.Context, *pets.PetSpecification) (page.Page[pets.Pet], error) {
	return page.Page[pets.Pet]{}, nil
}
func (r petRepositoryStub) FindAllByCustomerID(context.Context, uint) ([]pets.Pet, error) {
	return nil, nil
}
func (r petRepositoryStub) FindByCustomerID(context.Context, uint, page.Pagination) (page.Page[pets.Pet], error) {
	return page.Page[pets.Pet]{}, nil
}
func (r petRepositoryStub) FindByID(context.Context, pets.PetID) (pets.Pet, error) {
	return pets.Pet{}, nil
}
func (r petRepositoryStub) FindByIDAndCustomerID(context.Context, pets.PetID, uint) (pets.Pet, error) {
	return pets.Pet{}, r.err
}
func (r petRepositoryStub) ExistsByID(context.Context, pets.PetID) (bool, error) { return true, nil }
func (r petRepositoryStub) ExistsByMicrochip(context.Context, string) (bool, error) {
	return false, nil
}
func (r petRepositoryStub) ExistsDeletedByID(context.Context, pets.PetID) (bool, error) {
	return false, nil
}
func (r petRepositoryStub) Save(context.Context, pets.Pet) (pets.Pet, error)       { return pets.Pet{}, nil }
func (r petRepositoryStub) Delete(context.Context, pets.PetID, bool) error         { return nil }
func (r petRepositoryStub) RestoreByID(context.Context, pets.PetID) error          { return nil }
func (r petRepositoryStub) CountByCustomerID(context.Context, uint) (int64, error) { return 0, nil }

type appointmentDomainServiceStub struct{}

func (appointmentDomainServiceStub) ValidateNoOverlapping(context.Context, *Appointment) error {
	return nil
}
func (appointmentDomainServiceStub) ValidateCapacity(context.Context, time.Time) error { return nil }

type employeeRepositoryStub struct {
	employee employees.Employee
	err      error
}

func (r employeeRepositoryStub) GetBySpecification(context.Context, employees.EmployeeSpecification) (page.Page[employees.Employee], error) {
	return page.Page[employees.Employee]{}, r.err
}
func (r employeeRepositoryStub) GetByID(context.Context, employees.EmployeeID) (employees.Employee, error) {
	return r.employee, r.err
}
func (r employeeRepositoryStub) GetByUserID(context.Context, uint) (employees.Employee, error) {
	return r.employee, r.err
}
func (r employeeRepositoryStub) GetActive(context.Context, page.Pagination) (page.Page[employees.Employee], error) {
	return page.Page[employees.Employee]{}, r.err
}
func (r employeeRepositoryStub) GetBySpeciality(context.Context, employees.VetSpecialty, page.Pagination) (page.Page[employees.Employee], error) {
	return page.Page[employees.Employee]{}, r.err
}
func (r employeeRepositoryStub) ExistsByID(context.Context, employees.EmployeeID) (bool, error) {
	return r.err == nil, r.err
}
func (r employeeRepositoryStub) ExistsByUserID(context.Context, uint) (bool, error) {
	return r.err == nil, r.err
}
func (r employeeRepositoryStub) Save(context.Context, *employees.Employee) error        { return r.err }
func (r employeeRepositoryStub) SoftDelete(context.Context, employees.EmployeeID) error { return r.err }
func (r employeeRepositoryStub) HardDelete(context.Context, employees.EmployeeID) error { return r.err }
func (r employeeRepositoryStub) RestoreByID(context.Context, employees.EmployeeID) error {
	return r.err
}
func (r employeeRepositoryStub) CountBySpeciality(context.Context, employees.VetSpecialty) (int64, error) {
	return 0, r.err
}
func (r employeeRepositoryStub) CountByActive(context.Context, *bool) (int64, error) { return 0, r.err }
func (r employeeRepositoryStub) CountActive(context.Context) (int64, error)          { return 0, r.err }

func TestRequestAppointmentRejectsPetNotOwnedByCustomer(t *testing.T) {
	repository := &appointmentRepositoryStub{}
	service := NewCommandService(repository, nil, nil, petRepositoryStub{err: errors.New("not owned")}, appointmentDomainServiceStub{})
	err := service.RequestAppointment(context.Background(), NewRequestByCustomerCommand(pets.NewPetID(2), customers.NewCustomerID(1), nextWeekday(), VisitReasonRoutineCheckup, ClinicServiceGeneralConsultation, nil))
	if err == nil || repository.saveCalls != 0 {
		t.Fatal("must reject an appointment for another customer's pet before saving")
	}
}

func TestVeterinarianMutationsUseAssignedAppointmentLookup(t *testing.T) {
	employeeID := employees.NewEmployeeID(7)
	appointment := Appointment{Status: AppointmentStatusConfirmed, ScheduledDate: nextWeekday()}
	appointment.SetID(NewAppointmentID(3))
	tests := []struct {
		name string
		run  func(CommandService) error
	}{
		{"update", func(s CommandService) error {
			return s.UpdateAppointmentGeneralInfo(context.Background(), NewUpdateCommand(appointment.ID, &employeeID, nil, nil, nil))
		}},
		{"reschedule", func(s CommandService) error {
			return s.RescheduleAppointment(context.Background(), NewRescheduleCommand(appointment.ID, &employeeID, nextWeekday(), nil))
		}},
		{"complete", func(s CommandService) error {
			return s.CompleteAppointment(context.Background(), NewCompleteCommand(appointment.ID, &employeeID, nil))
		}},
		{"cancel", func(s CommandService) error {
			return s.CancelAppointment(context.Background(), NewCancelCommand(appointment.ID, &employeeID, nil, ""))
		}},
		{"no-show", func(s CommandService) error {
			return s.MarkAppointmentAsNotAttend(context.Background(), NewNotAttendCommand(appointment.ID, &employeeID))
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := &appointmentRepositoryStub{appointment: appointment}
			service := NewCommandService(repository, nil, nil, petRepositoryStub{}, appointmentDomainServiceStub{})
			if err := tt.run(service); err != nil || repository.byEmployeeCalls != 1 {
				t.Fatalf("error=%v assigned lookups=%d; want nil/1", err, repository.byEmployeeCalls)
			}
		})
	}
}

func TestCustomerCancellationUsesCustomerScopedLookup(t *testing.T) {
	appointment := Appointment{Status: AppointmentStatusPending}
	appointment.SetID(NewAppointmentID(3))
	repository := &appointmentRepositoryStub{appointment: appointment}
	service := NewCommandService(repository, nil, nil, petRepositoryStub{}, appointmentDomainServiceStub{})
	customerID := customers.NewCustomerID(4)
	if err := service.CancelAppointment(context.Background(), NewCancelCommand(appointment.ID, nil, &customerID, "")); err != nil || repository.byCustomerCalls != 1 || repository.saveCalls != 1 {
		t.Fatalf("error=%v customer lookup/save=%d/%d; want nil/1/1", err, repository.byCustomerCalls, repository.saveCalls)
	}
}

func TestCreateAppointmentReturnsPersistedID(t *testing.T) {
	repository := &appointmentRepositoryStub{}
	service := NewCommandService(repository, nil, nil, petRepositoryStub{}, appointmentDomainServiceStub{})
	created, err := service.CreateAppointment(context.Background(), NewCreateCommand(
		customers.NewCustomerID(1), pets.NewPetID(2), nil, ClinicServiceGeneralConsultation,
		VisitReasonRoutineCheckup, nil, nil, nextWeekday(),
	))
	if err != nil || created.ID.Value() != 99 {
		t.Fatalf("error=%v id=%d; want nil/99", err, created.ID.Value())
	}
}

func TestCancellationReasonAndCompletionNotesAreSaved(t *testing.T) {
	employeeID := employees.NewEmployeeID(7)
	appointment := Appointment{Status: AppointmentStatusConfirmed, ScheduledDate: nextWeekday(), EmployeeID: &employeeID}
	appointment.SetID(NewAppointmentID(3))
	repository := &appointmentRepositoryStub{appointment: appointment}
	service := NewCommandService(repository, nil, nil, petRepositoryStub{}, appointmentDomainServiceStub{})
	if err := service.CancelAppointment(context.Background(), NewCancelCommand(appointment.ID, &employeeID, nil, "owner request")); err != nil {
		t.Fatal(err)
	}
	if repository.saved.CancellationReason == nil || *repository.saved.CancellationReason != "owner request" {
		t.Fatal("cancellation reason was not saved")
	}
	repository.appointment.Status = AppointmentStatusConfirmed
	notes := "exam completed"
	if err := service.CompleteAppointment(context.Background(), NewCompleteCommand(appointment.ID, &employeeID, &notes)); err != nil {
		t.Fatal(err)
	}
	if repository.saved.Notes == nil || *repository.saved.Notes != notes {
		t.Fatal("completion notes were not saved")
	}
}

func TestConfirmAppointmentRejectsInactiveVeterinarian(t *testing.T) {
	appointment := Appointment{Status: AppointmentStatusPending, ScheduledDate: nextWeekday(), Service: ClinicServiceGeneralConsultation}
	appointment.SetID(NewAppointmentID(3))
	repository := &appointmentRepositoryStub{appointment: appointment}
	service := NewCommandService(repository, nil, employeeRepositoryStub{employee: employees.Employee{IsActive: false}}, petRepositoryStub{}, appointmentDomainServiceStub{})
	if err := service.ConfirmAppointment(context.Background(), NewConfirmCommand(appointment.ID, employees.NewEmployeeID(7))); err == nil {
		t.Fatal("must reject confirmation with an inactive veterinarian")
	}
	if repository.saveCalls != 0 {
		t.Fatal("inactive veterinarian must not be assigned")
	}
}

func nextWeekday() time.Time {
	base := time.Now().AddDate(0, 0, 7)
	date := time.Date(base.Year(), base.Month(), base.Day(), 12, 0, 0, 0, base.Location())
	for date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
		date = date.AddDate(0, 0, 1)
	}
	return date
}
