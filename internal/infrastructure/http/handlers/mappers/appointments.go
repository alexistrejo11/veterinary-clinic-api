package mappers

import (
	"clinic-vet-api/internal/core/appointments"
	"clinic-vet-api/internal/core/customers"
	"clinic-vet-api/internal/core/employees"
	"clinic-vet-api/internal/core/pets"
	"clinic-vet-api/internal/infrastructure/http/handlers/dtos"
	"clinic-vet-api/internal/shared/page"
	"time"
)

type AppointmentResponseMapper struct{}

func NewAppointmentResponseMapper() *AppointmentResponseMapper {
	return &AppointmentResponseMapper{}
}

// RequestToRequestByCustomerCommand maps customer request DTO to RequestByCustomerCommand (customerID from context).
func (m *AppointmentResponseMapper) RequestToRequestByCustomerCommand(
	req dtos.AppointmentRequestByCustomerRequest,
	customerID customers.CustomerID,
) (appointments.RequestByCustomerCommand, error) {
	service, err := appointments.ParseClinicService(req.Service)
	if err != nil {
		return appointments.RequestByCustomerCommand{}, err
	}
	reason, err := appointments.ParseVisitReason(req.Reason)
	if err != nil {
		return appointments.RequestByCustomerCommand{}, err
	}
	petID := pets.NewPetID(req.PetID)
	return appointments.NewRequestByCustomerCommand(
		petID,
		customerID,
		req.ScheduledDate,
		reason,
		service,
		req.Notes,
	), nil
}

// RequestToCreateCommand maps DTO to CreateCommand with full parsing and validation
func (m *AppointmentResponseMapper) RequestToCreateCommand(
	req dtos.AppointmentCreateRequest, optEmpID *uint,
) (appointments.CreateCommand, error) {
	// Parse service
	service, err := appointments.ParseClinicService(req.Service)
	if err != nil {
		return appointments.CreateCommand{}, err
	}

	// Parse reason
	reason, err := appointments.ParseVisitReason(req.Reason)
	if err != nil {
		return appointments.CreateCommand{}, err
	}

	// Parse status (optional)
	var status *appointments.AppointmentStatus
	if req.Status != nil {
		parsedStatus, err := appointments.ParseAppointmentStatus(*req.Status)
		if err != nil {
			return appointments.CreateCommand{}, err
		}
		status = &parsedStatus
	}

	// Parse employee ID (optional)
	var employeeID *employees.EmployeeID
	if optEmpID != nil {
		empID := employees.NewEmployeeID(*optEmpID)
		employeeID = &empID
	}

	// Create value objects
	customerID := customers.NewCustomerID(req.CustomerID)
	petID := pets.NewPetID(req.PetID)

	return appointments.NewCreateCommand(
		customerID,
		petID,
		employeeID,
		service,
		reason,
		status,
		req.Notes,
		req.ScheduledDate,
	), nil
}

// RequestToUpdateCommand maps DTO to UpdateCommand with full parsing and validation
func (m *AppointmentResponseMapper) RequestToUpdateCommand(
	req dtos.AppointmentUpdateGeneralInfoRequest, apptID uint, optEmpID *uint,
) (appointments.UpdateCommand, error) {
	// Parse service (optional)
	var service *appointments.ClinicService
	if req.Service != nil {
		parsedService, err := appointments.ParseClinicService(*req.Service)
		if err != nil {
			return appointments.UpdateCommand{}, err
		}
		service = &parsedService
	}

	appointmentID := appointments.NewAppointmentID(apptID)
	var employeeID *employees.EmployeeID
	if optEmpID != nil {
		id := employees.NewEmployeeID(*optEmpID)
		employeeID = &id
	}

	return appointments.NewUpdateCommand(
		appointmentID,
		employeeID,
		req.Reason,
		req.Notes,
		service,
	), nil
}

// RequestToRescheduleCommand maps DTO to RescheduleCommand
func (m *AppointmentResponseMapper) RequestToRescheduleCommand(
	req dtos.RescheduleAppointmentRequest, apptID uint, optEmpID *uint,
) appointments.RescheduleCommand {
	appointmentID := appointments.NewAppointmentID(apptID)
	var employeeID *employees.EmployeeID
	if optEmpID != nil {
		id := employees.NewEmployeeID(*optEmpID)
		employeeID = &id
	}

	return appointments.NewRescheduleCommand(
		appointmentID,
		employeeID,
		req.NewDateTime,
		req.Reason,
	)
}

// ToDeleteCommand maps primitive ID and hard flag to DeleteCommand
func (m *AppointmentResponseMapper) ToDeleteCommand(id uint, isHardDelete bool) appointments.DeleteCommand {
	appointmentID := appointments.NewAppointmentID(id)
	return appointments.NewDeleteCommand(appointmentID, isHardDelete)
}

// ToNotAttendCommand maps primitives to NotAttendCommand
func (m *AppointmentResponseMapper) ToNotAttendCommand(
	appointmentID uint,
	employeeID *uint,
) appointments.NotAttendCommand {
	apptID := appointments.NewAppointmentID(appointmentID)

	var empID *employees.EmployeeID
	if employeeID != nil {
		emp := employees.NewEmployeeID(*employeeID)
		empID = &emp
	}

	return appointments.NewNotAttendCommand(apptID, empID)
}

// ToConfirmCommand maps primitives to ConfirmCommand
func (m *AppointmentResponseMapper) ToConfirmCommand(
	appointmentID uint,
	employeeID uint,
) appointments.ConfirmCommand {
	apptID := appointments.NewAppointmentID(appointmentID)
	empID := employees.NewEmployeeID(employeeID)

	return appointments.NewConfirmCommand(apptID, empID)
}

// ToCompleteCommand maps primitives to CompleteCommand
func (m *AppointmentResponseMapper) ToCompleteCommand(
	appointmentID uint,
	employeeID *uint,
	notes string,
) appointments.CompleteCommand {
	apptID := appointments.NewAppointmentID(appointmentID)

	var empID *employees.EmployeeID
	if employeeID != nil {
		emp := employees.NewEmployeeID(*employeeID)
		empID = &emp
	}

	var notesPtr *string
	if notes != "" {
		notesPtr = &notes
	}

	return appointments.NewCompleteCommand(apptID, empID, notesPtr)
}

// ToCancelCommand maps primitives to CancelCommand
func (m *AppointmentResponseMapper) ToCancelCommand(
	appointmentID uint,
	employeeID *uint,
	customerID *uint,
	reason string,
) appointments.CancelCommand {
	apptID := appointments.NewAppointmentID(appointmentID)

	var empID *employees.EmployeeID
	if employeeID != nil {
		emp := employees.NewEmployeeID(*employeeID)
		empID = &emp
	}
	var custID *customers.CustomerID
	if customerID != nil {
		id := customers.NewCustomerID(*customerID)
		custID = &id
	}

	return appointments.NewCancelCommand(apptID, empID, custID, reason)
}

// RequestToSearchQuery maps search request DTO to GetBySpecQuery with full parsing
func (m *AppointmentResponseMapper) RequestToSearchQuery(
	r dtos.AppointmentSearchRequest,
) (appointments.GetBySpecQuery, error) {
	spec := appointments.NewAppointmentSpecification()

	// Parse and set customer ID
	if r.CustomerID > 0 {
		customerID := customers.NewCustomerID(r.CustomerID)
		spec.WithCustomerID(customerID)
	}

	// Parse and set employee ID
	if r.EmployeeID > 0 {
		employeeID := employees.NewEmployeeID(r.EmployeeID)
		spec.WithEmployeeID(employeeID)
	}

	// Parse and set pet ID
	if r.PetID > 0 {
		petID := pets.NewPetID(r.PetID)
		spec.WithPetID(petID)
	}

	// Parse and set service
	if r.Service != "" {
		service, err := appointments.ParseClinicService(r.Service)
		if err != nil {
			return appointments.GetBySpecQuery{}, err
		}
		spec.WithService(service)
	}

	// Parse and set status
	if r.Status != "" {
		status, err := appointments.ParseAppointmentStatus(r.Status)
		if err != nil {
			return appointments.GetBySpecQuery{}, err
		}
		spec.WithStatus(status)
	}

	// Parse and set date range
	if r.StartDate != "" && r.EndDate != "" {
		startDate, err := time.Parse("2006-01-02", r.StartDate)
		if err != nil {
			return appointments.GetBySpecQuery{}, err
		}
		endDate, err := time.Parse("2006-01-02", r.EndDate)
		if err != nil {
			return appointments.GetBySpecQuery{}, err
		}
		spec.WithDateRange(startDate, endDate)
	}

	// Set pagination
	spec.FromPagination(r.PaginationRequest.ToPagination())

	return appointments.NewGetBySpecQuery(spec), nil
}

func (m *AppointmentResponseMapper) ToResponse(appt appointments.Appointment) dtos.AppointmentResponse {
	var employeIDUint uint
	if appt.EmployeeID != nil {
		employeIDUint = appt.EmployeeID.Value()
	}

	return dtos.AppointmentResponse{
		ID:            appt.ID.Value(),
		PetID:         appt.PetID.Value(),
		CustomerID:    appt.CustomerID.Value(),
		EmployeeID:    employeIDUint,
		Service:       appt.Service.DisplayName(),
		Datetime:      appt.ScheduledDate.Format(time.RFC822),
		ScheduledDate: appt.ScheduledDate.Format(time.RFC822),
		Reason:        appt.Reason.DisplayName(),
		Notes:         appt.Notes,
		Status:        appt.Status.DisplayName(),
		CreatedAt:     appt.CreatedAt.Format(time.RFC822),
		UpdatedAt:     appt.UpdatedAt.Format(time.RFC822),
	}
}

func (m *AppointmentResponseMapper) ToResponsePage(
	apptPage page.Page[appointments.Appointment],
) *page.Page[dtos.AppointmentResponse] {

	responsePage := page.MapItems(apptPage, m.ToResponse)
	return &responsePage
}
