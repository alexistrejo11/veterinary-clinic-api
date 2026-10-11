// Package appointments defines the Appointment API module
package appointments

import (
	"clinic-vet-api/internal/core/customers"
	"clinic-vet-api/internal/core/employees"
	"clinic-vet-api/internal/core/pets"
	"clinic-vet-api/internal/shared"
	"context"
	"slices"
	"time"
)

type Appointment struct {
	shared.Entity[AppointmentID]
	Service            ClinicService
	ScheduledDate      time.Time
	Status             AppointmentStatus
	Reason             VisitReason
	Notes              *string
	CancellationReason *string
	CustomerID         customers.CustomerID
	EmployeeID         *employees.EmployeeID
	PetID              pets.PetID
}

// SetAsRequest marks the appointment as a request that needs approval
func (a *Appointment) SetAsRequest() {
	a.Status = AppointmentStatusPending
}

// ValidateInsert validates appointment business rules before insertion
func (a *Appointment) ValidateInsert(ctx context.Context) error {
	operation := "ValidateInsert"

	// Validate scheduled date
	if err := validateScheduledDate(ctx, a.ScheduledDate); err != nil {
		return err
	}

	// Validate service
	if !a.Service.IsValid() {
		return InvalidServiceError(ctx, a.Service, operation)
	}

	// Validate reason
	if !a.Reason.IsValid() {
		return InvalidReasonError(ctx, string(a.Reason), operation)
	}

	// Validate notes length if provided
	if a.Notes != nil && len(*a.Notes) > 1000 {
		return NotesTooLongError(ctx, operation)
	}

	// Validate required IDs
	if a.CustomerID.IsZero() {
		return appointmentValidationError(ctx, "APPOINTMENT_INVALID_CUSTOMER", "customer_id",
			"customer ID is required", operation)
	}

	if a.PetID.IsZero() {
		return appointmentValidationError(ctx, "APPOINTMENT_INVALID_PET", "pet_id",
			"pet ID is required", operation)
	}

	return nil
}

func (a *Appointment) Update(
	ctx context.Context,
	notes *string,
	service *ClinicService,
	reason *string,
) error {
	operation := "UpdateAppointment"
	if notes != nil {
		if len(*notes) > 1000 {
			return NotesTooLongError(ctx, operation)
		}
		a.Notes = notes
	}

	if service != nil {
		if !service.IsValid() {
			return InvalidServiceError(ctx, *service, operation)
		}
		a.Service = *service
	}

	if reason != nil {
		parsedReason, err := ParseVisitReason(*reason)
		if err != nil {
			return InvalidReasonError(ctx, *reason, operation)
		}
		a.Reason = parsedReason
	}

	a.IncrementVersion()
	return nil
}

func (a *Appointment) Reschedule(ctx context.Context, newDate time.Time) error {
	operation := "RescheduleAppointment"

	if err := validateScheduledDate(ctx, newDate); err != nil {
		return err
	}

	if !a.canBeRescheduled() {
		return CannotRescheduleError(ctx, a.Status, operation)
	}

	a.ScheduledDate = newDate
	a.Status = AppointmentStatusRescheduled
	a.IncrementVersion()

	return nil
}

func (a *Appointment) Cancel(ctx context.Context, reason string) error {
	operation := "CancelAppointment"

	if !a.Status.CanBeCancelled() {
		return CannotCancelError(ctx, a.Status, operation)
	}

	a.Status = AppointmentStatusCancelled
	if reason != "" {
		a.CancellationReason = &reason
	}
	a.IncrementVersion()
	return nil
}

func (a *Appointment) Complete(ctx context.Context, notes *string) error {
	operation := "CompleteAppointment"

	if !a.canTransitionTo(AppointmentStatusCompleted) {
		return CannotCompleteError(ctx, a.Status, operation)
	}
	if notes != nil {
		if len(*notes) > 1000 {
			return NotesTooLongError(ctx, operation)
		}
		a.Notes = notes
	}

	a.Status = AppointmentStatusCompleted
	a.IncrementVersion()
	return nil
}

func (a *Appointment) MarkAsNotPresented(ctx context.Context) error {
	operation := "MarkAppointmentNotPresented"

	if !a.canTransitionTo(AppointmentStatusNotPresented) {
		return CannotMarkNotPresentedError(ctx, a.Status, operation)
	}

	a.Status = AppointmentStatusNotPresented
	a.IncrementVersion()
	return nil
}

func (a *Appointment) Confirm(ctx context.Context, employeeID employees.EmployeeID) error {
	operation := "ConfirmAppointment"

	if !a.canTransitionTo(AppointmentStatusConfirmed) {
		return CannotConfirmError(ctx, a.Status, operation)
	}

	a.EmployeeID = &employeeID
	a.Status = AppointmentStatusConfirmed
	a.IncrementVersion()
	return nil
}

func (a *Appointment) canBeRescheduled() bool {
	reschedulableStatuses := []AppointmentStatus{
		AppointmentStatusPending,
		AppointmentStatusConfirmed,
		AppointmentStatusRescheduled,
	}

	return slices.Contains(reschedulableStatuses, a.Status)
}

func (a *Appointment) canTransitionTo(newStatus AppointmentStatus) bool {
	validTransitions := map[AppointmentStatus][]AppointmentStatus{
		AppointmentStatusPending: {
			AppointmentStatusConfirmed,
			AppointmentStatusCancelled,
			AppointmentStatusRescheduled,
		},
		AppointmentStatusConfirmed: {
			AppointmentStatusCompleted,
			AppointmentStatusCancelled,
			AppointmentStatusNotPresented,
			AppointmentStatusRescheduled,
		},
		AppointmentStatusRescheduled: {
			AppointmentStatusConfirmed,
			AppointmentStatusCancelled,
		},
		AppointmentStatusCompleted:    {},
		AppointmentStatusCancelled:    {},
		AppointmentStatusNotPresented: {},
	}

	allowedTransitions, exists := validTransitions[a.Status]
	if !exists {
		return false
	}

	return slices.Contains(allowedTransitions, newStatus)
}

func (a *Appointment) IsActive() bool {
	now := time.Now()
	return (a.Status == AppointmentStatusConfirmed || a.Status == AppointmentStatusRescheduled) &&
		a.ScheduledDate.After(now)
}

func (a *Appointment) CanBeDeleted(ctx context.Context) error {
	operation := "DeleteAppointment"

	if a.Status == AppointmentStatusCompleted {
		return CannotDeleteError(ctx, a.Status, operation)
	}
	return nil
}

func validateScheduledDate(ctx context.Context, date time.Time) error {
	operation := "ValidateScheduledDate"

	if date.IsZero() {
		return ScheduledDateInvalidError(ctx, "scheduled date cannot be zero", operation)
	}

	now := time.Now()
	if date.Before(now) {
		return ScheduledDateInvalidError(ctx, "scheduled date cannot be in the past", operation)
	}

	minDate := now.AddDate(0, 0, MinAllowedDaysToSchedule)
	if date.Before(minDate) {
		return ScheduledDateInvalidError(ctx, "appointments must be scheduled at least 3 days in advance", operation)
	}

	maxDate := now.AddDate(0, 0, MaxAllowedDaysToSchedule)
	if date.After(maxDate) {
		return ScheduledDateInvalidError(ctx, "appointments cannot be scheduled more than 30 days in advance", operation)
	}

	if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
		return ScheduledDateInvalidError(ctx, "appointments cannot be scheduled on weekends", operation)
	}

	hour := date.Hour()
	if hour < 9 || hour >= 18 {
		return ScheduledDateInvalidError(ctx, "appointments can only be scheduled during business hours (9 AM to 6 PM)", operation)
	}

	return nil
}
