package appointments

import (
	"context"
	"fmt"
	"time"

	"clinic-vet-api/internal/core/customers"
	"clinic-vet-api/internal/core/employees"
	"clinic-vet-api/internal/core/notifications"
	"clinic-vet-api/internal/core/pets"
	"clinic-vet-api/internal/core/users"
	"clinic-vet-api/internal/shared/page"
)

// CommandService defines write operations for appointments
type CommandService interface {
	RequestAppointment(ctx context.Context, cmd RequestByCustomerCommand) error
	CreateAppointment(ctx context.Context, cmd CreateCommand) (Appointment, error)
	CancelAppointment(ctx context.Context, cmd CancelCommand) error
	DeleteAppointment(ctx context.Context, cmd DeleteCommand) error
	ConfirmAppointment(ctx context.Context, cmd ConfirmCommand) error
	MarkAppointmentAsNotAttend(ctx context.Context, cmd NotAttendCommand) error
	RescheduleAppointment(ctx context.Context, cmd RescheduleCommand) error
	CompleteAppointment(ctx context.Context, cmd CompleteCommand) error
	UpdateAppointmentGeneralInfo(ctx context.Context, cmd UpdateCommand) error
}

type commandService struct {
	repository          AppointmentRepository
	customerRepository  customers.CustomerRepository
	employeeRepository  employees.EmployeeRepository
	petRepository       pets.PetRepository
	domainService       AppointmentDomainService
	notificationService notifications.NotificationService
	userQuery           users.QueryService
}

// NotificationDependencies are optional to keep appointment business rules usable without delivery infrastructure.
type NotificationDependencies struct {
	Service notifications.NotificationService
	Users   users.QueryService
}

// NewCommandService creates a new command service instance
func NewCommandService(
	repository AppointmentRepository,
	customerRepository customers.CustomerRepository,
	employeeRepository employees.EmployeeRepository,
	petRepository pets.PetRepository,
	domainService AppointmentDomainService, notificationDependencies ...NotificationDependencies,
) CommandService {
	var notificationService notifications.NotificationService
	var userQuery users.QueryService
	if len(notificationDependencies) > 0 {
		notificationService = notificationDependencies[0].Service
		userQuery = notificationDependencies[0].Users
	}
	return &commandService{
		repository:          repository,
		customerRepository:  customerRepository,
		employeeRepository:  employeeRepository,
		petRepository:       petRepository,
		domainService:       domainService,
		notificationService: notificationService,
		userQuery:           userQuery,
	}
}

// =========================================================================
// Command Service implementation
// =========================================================================

// RequestAppointment creates an appointment request that needs to be approved by a vet.
// It starts as pending and can be confirmed by a vet later.
// Must be performed by customers.
func (s *commandService) RequestAppointment(ctx context.Context, cmd RequestByCustomerCommand) error {
	if err := s.ensurePetOwnedByCustomer(ctx, cmd.PetID, cmd.CustomerID); err != nil {
		return err
	}

	appointment := Appointment{
		PetID:         cmd.PetID,
		CustomerID:    cmd.CustomerID,
		ScheduledDate: cmd.RequestedDate,
		Status:        AppointmentStatusPending,
		Notes:         cmd.Notes,
		Reason:        cmd.Reason,
		Service:       cmd.Service,
	}

	// Validate appointment business rules
	if err := appointment.ValidateInsert(ctx); err != nil {
		return err
	}
	// Validate capacity (no more than 5 appointments per hour)
	if err := s.domainService.ValidateCapacity(ctx, appointment.ScheduledDate); err != nil {
		return err
	}

	if err := s.repository.Save(ctx, &appointment); err != nil {
		return err
	}

	return s.notifyReceptionists(ctx, appointment, "Appointment request", "A customer requested an appointment.")
}

// CreateAppointment creates an appointment that doesn't need approval.
// Must be performed by staff.
func (s *commandService) CreateAppointment(ctx context.Context, cmd CreateCommand) (Appointment, error) {
	if err := s.ensurePetOwnedByCustomer(ctx, cmd.PetID, cmd.CustomerID); err != nil {
		return Appointment{}, err
	}

	status := AppointmentStatusPending
	if cmd.Status != nil {
		status = *cmd.Status
	}

	appointment := Appointment{
		PetID:         cmd.PetID,
		CustomerID:    cmd.CustomerID,
		EmployeeID:    cmd.VetID,
		Service:       cmd.Service,
		ScheduledDate: cmd.Datetime,
		Status:        status,
		Notes:         cmd.Notes,
		Reason:        cmd.Reason,
	}

	// Validate appointment business rules
	if err := appointment.ValidateInsert(ctx); err != nil {
		return Appointment{}, err
	}
	if err := s.validateVeterinarianAvailability(ctx, appointment.EmployeeID, appointment.ScheduledDate, appointment.Service); err != nil {
		return Appointment{}, err
	}

	// Validate no overlapping appointments for the assigned employee
	if err := s.domainService.ValidateNoOverlapping(ctx, &appointment); err != nil {
		return Appointment{}, err
	}

	// Validate capacity (no more than 5 appointments per hour)
	if err := s.domainService.ValidateCapacity(ctx, appointment.ScheduledDate); err != nil {
		return Appointment{}, err
	}

	if err := s.repository.Save(ctx, &appointment); err != nil {
		return Appointment{}, err
	}

	if err := s.notifyCustomer(ctx, appointment, "Appointment created", "Your appointment has been created.", notifications.NotificationTypeAppointmentConfirm); err != nil {
		return Appointment{}, err
	}
	return appointment, nil
}

// CancelAppointment cancels an existing appointment
func (s *commandService) CancelAppointment(ctx context.Context, cmd CancelCommand) error {
	if cmd.CustomerID != nil {
		appointment, err := s.repository.FindByIDAndCustomerID(ctx, cmd.ID, *cmd.CustomerID)
		if err != nil {
			return err
		}
		if err := appointment.Cancel(ctx, cmd.Reason); err != nil {
			return err
		}
		if err := s.repository.Save(ctx, &appointment); err != nil {
			return err
		}
		return s.notifyCustomer(ctx, appointment, "Appointment cancelled", "Your appointment has been cancelled.", notifications.NotificationTypeAppointmentCancel)
	}

	appointment, err := s.getApptByEmployeeOpt(ctx, cmd.ID, cmd.EmployeeID)
	if err != nil {
		return err
	}

	if err := appointment.Cancel(ctx, cmd.Reason); err != nil {
		return err
	}

	if err := s.repository.Save(ctx, &appointment); err != nil {
		return err
	}

	return s.notifyCustomer(ctx, appointment, "Appointment cancelled", "Your appointment has been cancelled.", notifications.NotificationTypeAppointmentCancel)
}

// ConfirmAppointment confirms a pending appointment
func (s *commandService) ConfirmAppointment(ctx context.Context, cmd ConfirmCommand) error {
	appointment, err := s.repository.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}
	if err := s.validateVeterinarianAvailability(ctx, &cmd.EmployeeID, appointment.ScheduledDate, appointment.Service); err != nil {
		return err
	}
	appointment.EmployeeID = &cmd.EmployeeID
	if err := s.domainService.ValidateNoOverlapping(ctx, &appointment); err != nil {
		return err
	}

	if err := appointment.Confirm(ctx, cmd.EmployeeID); err != nil {
		return err
	}

	if err := s.repository.Save(ctx, &appointment); err != nil {
		return err
	}

	return s.notifyCustomer(ctx, appointment, "Appointment confirmed", "Your appointment has been confirmed.", notifications.NotificationTypeAppointmentConfirm)
}

// DeleteAppointment removes an appointment
func (s *commandService) DeleteAppointment(ctx context.Context, cmd DeleteCommand) error {
	appointment, err := s.repository.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	if active := appointment.IsActive(); active {
		return fmt.Errorf("cannot delete active appointment")
	}

	return s.repository.DeleteByID(ctx, cmd.ID, cmd.IsHardDelete)
}

// MarkAppointmentAsNotAttend marks an appointment as not attended
func (s *commandService) MarkAppointmentAsNotAttend(ctx context.Context, cmd NotAttendCommand) error {
	appointment, err := s.getApptByEmployeeOpt(ctx, cmd.ID, cmd.EmployeeID)
	if err != nil {
		return err
	}

	if err := appointment.MarkAsNotPresented(ctx); err != nil {
		return err
	}

	if err := s.repository.Save(ctx, &appointment); err != nil {
		return err
	}

	return s.notifyCustomer(ctx, appointment, "Appointment no-show", "Your appointment was marked as not attended.", notifications.NotificationTypeAlert)
}

// CompleteAppointment marks an appointment as completed
func (s *commandService) CompleteAppointment(ctx context.Context, cmd CompleteCommand) error {
	appointment, err := s.getApptByEmployeeOpt(ctx, cmd.ID, cmd.EmployeeID)
	if err != nil {
		return err
	}

	if err := appointment.Complete(ctx, cmd.Notes); err != nil {
		return err
	}

	if err := s.repository.Save(ctx, &appointment); err != nil {
		return err
	}

	return s.notifyCustomer(ctx, appointment, "Appointment completed", "Your appointment has been completed.", notifications.NotificationTypeInfo)
}

// UpdateAppointmentGeneralInfo updates general information of an appointment
func (s *commandService) UpdateAppointmentGeneralInfo(ctx context.Context, cmd UpdateCommand) error {
	appointment, err := s.getApptByEmployeeOpt(ctx, cmd.ID, cmd.EmployeeID)
	if err != nil {
		return err
	}

	if err := appointment.Update(ctx, cmd.Notes, cmd.Service, cmd.Reason); err != nil {
		return err
	}

	if err := s.repository.Save(ctx, &appointment); err != nil {
		return err
	}

	return s.notifyCustomer(ctx, appointment, "Appointment updated", "Your appointment details have been updated.", notifications.NotificationTypeInfo)
}

func (s *commandService) notifyCustomer(ctx context.Context, appointment Appointment, subject, message string, notificationType notifications.NotificationType) error {
	if s.notificationService == nil || s.userQuery == nil {
		return nil
	}
	user, err := s.userQuery.GetByCustomerID(ctx, appointment.CustomerID)
	if err != nil {
		return err
	}
	// Delivery failures are persisted by NotificationService and must not make an
	// already-saved appointment operation appear to have failed.
	_ = s.notificationService.Send(ctx, &notifications.Notification{UserID: user.ID.String(), UserEmail: user.Email.String(), Subject: subject, Message: message, NType: notificationType, Channel: notifications.NotificationChannelEmail})
	return nil
}

func (s *commandService) notifyReceptionists(ctx context.Context, appointment Appointment, subject, message string) error {
	if s.notificationService == nil || s.userQuery == nil || s.employeeRepository == nil {
		return nil
	}
	for pageNumber := 1; ; pageNumber++ {
		employeesPage, err := s.employeeRepository.GetActive(ctx, page.Pagination{Number: pageNumber, Size: page.MaxPageSize})
		if err != nil {
			return err
		}
		for _, employee := range employeesPage.Items {
			user, err := s.userQuery.GetByEmployeeID(ctx, employee.ID)
			if err != nil {
				return err
			}
			if user.Role != users.UserRoleReceptionist {
				continue
			}
			_ = s.notificationService.Send(ctx, &notifications.Notification{UserID: user.ID.String(), UserEmail: user.Email.String(), Subject: subject, Message: message, NType: notifications.NotificationTypeInfo, Channel: notifications.NotificationChannelEmail})
		}
		if !employeesPage.Metadata.HasNextPage {
			return nil
		}
	}
}

// RescheduleAppointment changes the scheduled date of an appointment
func (s *commandService) RescheduleAppointment(ctx context.Context, cmd RescheduleCommand) error {
	appointment, err := s.getApptByEmployeeOpt(ctx, cmd.ID, cmd.EmployeeID)
	if err != nil {
		return err
	}

	// Validate no overlapping appointments for the employee at new time
	tempAppointment := appointment
	tempAppointment.ScheduledDate = cmd.NewDateTime
	if err := s.validateVeterinarianAvailability(ctx, tempAppointment.EmployeeID, tempAppointment.ScheduledDate, tempAppointment.Service); err != nil {
		return err
	}
	if err := s.domainService.ValidateNoOverlapping(ctx, &tempAppointment); err != nil {
		return err
	}

	// Validate capacity at new time
	if err := s.domainService.ValidateCapacity(ctx, cmd.NewDateTime); err != nil {
		return err
	}

	if err := appointment.Reschedule(ctx, cmd.NewDateTime); err != nil {
		return err
	}

	if err := s.repository.Save(ctx, &appointment); err != nil {
		return err
	}

	return s.notifyCustomer(ctx, appointment, "Appointment rescheduled", "Your appointment has been rescheduled.", notifications.NotificationTypeAppointmentConfirm)
}

func (s *commandService) ensurePetOwnedByCustomer(
	ctx context.Context,
	petID pets.PetID,
	customerID customers.CustomerID,
) error {
	_, err := s.petRepository.FindByIDAndCustomerID(ctx, petID, customerID.Value())
	return err
}

func (s *commandService) validateVeterinarianAvailability(ctx context.Context, employeeID *employees.EmployeeID, scheduledDate time.Time, service ClinicService) error {
	if employeeID == nil || s.employeeRepository == nil {
		return nil
	}
	employee, err := s.employeeRepository.GetByID(ctx, *employeeID)
	if err != nil {
		return err
	}
	if !employee.IsActive {
		return fmt.Errorf("assigned veterinarian is inactive")
	}
	if employee.Schedule == nil {
		return nil
	}
	endTime := scheduledDate.Add(time.Duration(service.DefaultDuration()) * time.Minute)
	for _, workDay := range employee.Schedule.WorkDays {
		if workDay.Day != scheduledDate.Weekday() {
			continue
		}
		if scheduledDate.Hour() < workDay.StartHour || endTime.Hour() > workDay.EndHour || (endTime.Hour() == workDay.EndHour && endTime.Minute() > 0) {
			break
		}
		breakStart := time.Date(scheduledDate.Year(), scheduledDate.Month(), scheduledDate.Day(), workDay.Breaks.StartHour, 0, 0, 0, scheduledDate.Location())
		breakEnd := time.Date(scheduledDate.Year(), scheduledDate.Month(), scheduledDate.Day(), workDay.Breaks.EndHour, 0, 0, 0, scheduledDate.Location())
		if workDay.Breaks.EndHour <= workDay.Breaks.StartHour || endTime.Before(breakStart) || scheduledDate.After(breakEnd) || scheduledDate.Equal(breakEnd) {
			return nil
		}
		break
	}
	return fmt.Errorf("assigned veterinarian is not available at the requested time")
}

func (s *commandService) getApptByEmployeeOpt(
	ctx context.Context,
	appointID AppointmentID,
	empIdOpt *employees.EmployeeID,
) (Appointment, error) {
	if empIdOpt == nil {
		return s.repository.FindByID(ctx, appointID)
	}

	return s.repository.FindByIDAndEmployeeID(ctx, appointID, *empIdOpt)
}
