package repository

import (
	"testing"
	"time"

	"clinic-vet-api/internal/core/appointments"
	"clinic-vet-api/internal/core/customers"
	"clinic-vet-api/internal/core/pets"
)

func TestAppointmentUpdateParamsAllowOptionalFields(t *testing.T) {
	appointment := appointments.Appointment{
		CustomerID:    customers.NewCustomerID(1),
		PetID:         pets.NewPetID(2),
		ScheduledDate: time.Now(),
		Status:        appointments.AppointmentStatusCancelled,
		Service:       appointments.ClinicServiceGeneralConsultation,
		Reason:        appointments.VisitReasonRoutineCheckup,
	}
	appointment.SetID(appointments.NewAppointmentID(3))

	params := (&SqlcAppointmentRepository{}).toUpdateParams(&appointment)
	if params.EmployeeID.Valid || params.Notes.Valid || params.CancellationReason.Valid {
		t.Fatal("optional appointment fields must remain NULL when absent")
	}
}
