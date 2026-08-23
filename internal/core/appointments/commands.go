package appointments

import (
	"clinic-vet-api/internal/core/customers"
	"clinic-vet-api/internal/core/employees"
	"clinic-vet-api/internal/core/pets"
	"time"
)

// =========================================================================
// Command definitions for appointment operations
// =========================================================================

// RequestByCustomerCommand represents a customer's request to schedule an appointment
type RequestByCustomerCommand struct {
	PetID         pets.PetID
	CustomerID    customers.CustomerID
	RequestedDate time.Time
	Reason        VisitReason
	Service       ClinicService
	Notes         *string
}

// CreateCommand represents a command to create a new appointment
type CreateCommand struct {
	CustomerID customers.CustomerID
	PetID      pets.PetID
	VetID      *employees.EmployeeID
	Service    ClinicService
	Datetime   time.Time
	Status     *AppointmentStatus
	Reason     VisitReason
	Notes      *string
}

// CancelCommand represents a command to cancel an appointment
type CancelCommand struct {
	ID         AppointmentID
	EmployeeID *employees.EmployeeID
	CustomerID *customers.CustomerID
	Reason     string
}

// CompleteCommand represents a command to mark an appointment as completed
type CompleteCommand struct {
	ID         AppointmentID
	EmployeeID *employees.EmployeeID
	Notes      *string
}

// ConfirmCommand represents a command to confirm a pending appointment
type ConfirmCommand struct {
	ID         AppointmentID
	EmployeeID employees.EmployeeID
}

// UpdateCommand represents a command to update appointment information
type UpdateCommand struct {
	ID         AppointmentID
	EmployeeID *employees.EmployeeID
	Reason     *string
	Notes      *string
	Service    *ClinicService
}

// NotAttendCommand represents a command to mark an appointment as not attended
type NotAttendCommand struct {
	ID         AppointmentID
	EmployeeID *employees.EmployeeID
}

// RescheduleCommand represents a command to reschedule an appointment
type RescheduleCommand struct {
	ID          AppointmentID
	EmployeeID  *employees.EmployeeID
	NewDateTime time.Time
	Reason      *string
}

// DeleteCommand represents a command to delete an appointment
type DeleteCommand struct {
	ID           AppointmentID
	IsHardDelete bool
}

// =========================================================================
// Command constructors - Accept value objects directly (no parsing)
// =========================================================================

// NewCreateCommand creates a new CreateCommand instance
func NewCreateCommand(
	customerID customers.CustomerID,
	petID pets.PetID,
	vetID *employees.EmployeeID,
	service ClinicService,
	reason VisitReason,
	status *AppointmentStatus,
	notes *string,
	dateTime time.Time,
) CreateCommand {
	return CreateCommand{
		CustomerID: customerID,
		PetID:      petID,
		VetID:      vetID,
		Service:    service,
		Reason:     reason,
		Status:     status,
		Notes:      notes,
		Datetime:   dateTime,
	}
}

// NewCancelCommand creates a new CancelCommand instance
func NewCancelCommand(
	id AppointmentID,
	employeeID *employees.EmployeeID,
	customerID *customers.CustomerID,
	reason string,
) CancelCommand {
	return CancelCommand{
		ID:         id,
		EmployeeID: employeeID,
		CustomerID: customerID,
		Reason:     reason,
	}
}

// NewCompleteCommand creates a new CompleteCommand instance
func NewCompleteCommand(
	id AppointmentID,
	employeeID *employees.EmployeeID,
	notes *string,
) CompleteCommand {
	return CompleteCommand{
		ID:         id,
		EmployeeID: employeeID,
		Notes:      notes,
	}
}

// NewConfirmCommand creates a new ConfirmCommand instance
func NewConfirmCommand(
	id AppointmentID,
	employeeID employees.EmployeeID,
) ConfirmCommand {
	return ConfirmCommand{
		ID:         id,
		EmployeeID: employeeID,
	}
}

// NewDeleteCommand creates a new DeleteCommand instance
func NewDeleteCommand(id AppointmentID, isHardDelete bool) DeleteCommand {
	return DeleteCommand{ID: id, IsHardDelete: isHardDelete}
}

// NewNotAttendCommand creates a new NotAttendCommand instance
func NewNotAttendCommand(
	id AppointmentID,
	employeeID *employees.EmployeeID,
) NotAttendCommand {
	return NotAttendCommand{
		ID:         id,
		EmployeeID: employeeID,
	}
}

// NewRescheduleCommand creates a new RescheduleCommand instance
func NewRescheduleCommand(
	id AppointmentID,
	employeeID *employees.EmployeeID,
	newDateTime time.Time,
	reason *string,
) RescheduleCommand {
	return RescheduleCommand{
		ID:          id,
		EmployeeID:  employeeID,
		NewDateTime: newDateTime,
		Reason:      reason,
	}
}

// NewUpdateCommand creates a new UpdateCommand instance
func NewUpdateCommand(
	id AppointmentID,
	employeeID *employees.EmployeeID,
	reason *string,
	notes *string,
	service *ClinicService,
) UpdateCommand {
	return UpdateCommand{
		ID:         id,
		EmployeeID: employeeID,
		Service:    service,
		Reason:     reason,
		Notes:      notes,
	}
}

// NewRequestByCustomerCommand creates a new RequestByCustomerCommand instance
func NewRequestByCustomerCommand(
	petID pets.PetID,
	customerID customers.CustomerID,
	requestedDate time.Time,
	reason VisitReason,
	service ClinicService,
	notes *string,
) RequestByCustomerCommand {
	return RequestByCustomerCommand{
		PetID:         petID,
		CustomerID:    customerID,
		RequestedDate: requestedDate,
		Reason:        reason,
		Service:       service,
		Notes:         notes,
	}
}
