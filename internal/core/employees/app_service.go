package employees

import (
	"clinic-vet-api/internal/shared"
	"clinic-vet-api/internal/shared/page"
	"context"
	"fmt"
	"strings"
	"time"
)

// =========================================================================
// Command definitions for employee operations
// =========================================================================

type ScheduleDataCommand struct {
	Day           string
	EntryTime     int
	DepartureTime int
	StartBreak    int
	EndBreak      int
}

type CreateEmployeeCommand struct {
	// Personal details
	FirstName   string
	LastName    string
	Gender      shared.PersonGender
	DateOfBirth time.Time

	Photo string
	// Professional details
	LicenseNumber   string
	YearsExperience int32
	IsActive        bool
	Specialty       VetSpecialty
	Schedule        ScheduleDataCommand
}

type UpdateEmployeeCommand struct {
	ID              *EmployeeID
	FirstName       *string
	LastName        *string
	Gender          *shared.PersonGender
	DateOfBirth     *time.Time
	Photo           *string
	LicenseNumber   string
	YearsExperience *int32
	IsActive        *bool
	Specialty       *VetSpecialty
	Schedule        *ScheduleDataCommand
}

type EmployeeStats struct {
	TotalEmployees  int64
	ActiveEmployees int64
	Specialties     map[VetSpecialty]int64
}

// =========================================================================
// Service definitions for employee operations
// =========================================================================

type EmployeeService interface {
	// Commands
	CreateEmployee(ctx context.Context, cmd CreateEmployeeCommand) (Employee, error)
	UpdateEmployee(ctx context.Context, cmd UpdateEmployeeCommand) error
	RestoreEmployee(ctx context.Context, id EmployeeID) error
	DeleteEmployee(ctx context.Context, id EmployeeID) error

	// Queries
	GetEmployeeByID(ctx context.Context, id EmployeeID) (Employee, error)
	GetEmployeesBySpecification(ctx context.Context, spec EmployeeSpecification) (page.Page[Employee], error)
	GetActiveEmployees(ctx context.Context, pagination page.Pagination) (page.Page[Employee], error)
	GetEmployeesBySpecialty(ctx context.Context, specialty VetSpecialty, pagination page.Pagination) (page.Page[Employee], error)
	GetEmployeeStats(ctx context.Context) (EmployeeStats, error)
}

// =========================================================================
// Service implementation for employee operations
// =========================================================================

type employeeService struct {
	repository EmployeeRepository
}

func NewEmployeeService(repository EmployeeRepository) EmployeeService {
	return &employeeService{repository: repository}
}

func (s *employeeService) CreateEmployee(ctx context.Context, cmd CreateEmployeeCommand) (Employee, error) {
	operation := "CreateEmployee"

	person := shared.Person{
		FirstName:   cmd.FirstName,
		LastName:    cmd.LastName,
		Gender:      cmd.Gender,
		DateOfBirth: cmd.DateOfBirth,
	}

	emp := Employee{
		Person:          person,
		Photo:           cmd.Photo,
		LicenseNumber:   cmd.LicenseNumber,
		Specialty:       cmd.Specialty,
		YearsExperience: int(cmd.YearsExperience),
		IsActive:        cmd.IsActive,
	}

	day, err := parseWeekday(cmd.Schedule.Day)
	if err != nil {
		return Employee{}, err
	}
	// Build schedule from command.
	schedule := &EmployeeSchedule{
		WorkDays: []WorkDaySchedule{
			{
				Day:       day,
				StartHour: cmd.Schedule.EntryTime,
				EndHour:   cmd.Schedule.DepartureTime,
				Breaks: Break{
					StartHour: cmd.Schedule.StartBreak,
					EndHour:   cmd.Schedule.EndBreak,
				},
			},
		},
	}
	emp.Schedule = schedule

	if err := emp.Validate(ctx, operation); err != nil {
		return Employee{}, err
	}

	if err := s.repository.Save(ctx, &emp); err != nil {
		return Employee{}, err
	}

	return emp, nil
}

func (s *employeeService) UpdateEmployee(ctx context.Context, cmd UpdateEmployeeCommand) error {
	const operation = "UpdateEmployee"

	if cmd.ID == nil {
		return fmt.Errorf("employee ID is required for update")
	}

	emp, err := s.repository.GetByID(ctx, *cmd.ID)
	if err != nil {
		return err
	}

	if cmd.FirstName != nil {
		emp.FirstName = *cmd.FirstName
	}
	if cmd.LastName != nil {
		emp.LastName = *cmd.LastName
	}
	if cmd.Gender != nil {
		emp.Gender = *cmd.Gender
	}
	if cmd.DateOfBirth != nil {
		emp.DateOfBirth = *cmd.DateOfBirth
	}
	if cmd.Photo != nil {
		emp.Photo = *cmd.Photo
	}
	if cmd.LicenseNumber != "" {
		emp.LicenseNumber = cmd.LicenseNumber
	}
	if cmd.YearsExperience != nil {
		emp.YearsExperience = int(*cmd.YearsExperience)
	}
	if cmd.IsActive != nil {
		emp.IsActive = *cmd.IsActive
	}
	if cmd.Specialty != nil {
		emp.Specialty = *cmd.Specialty
	}

	// Update schedule if provided
	if cmd.Schedule != nil {
		day, err := parseWeekday(cmd.Schedule.Day)
		if err != nil {
			return err
		}
		schedule := &EmployeeSchedule{
			WorkDays: []WorkDaySchedule{
				{
					Day:       day,
					StartHour: cmd.Schedule.EntryTime,
					EndHour:   cmd.Schedule.DepartureTime,
					Breaks: Break{
						StartHour: cmd.Schedule.StartBreak,
						EndHour:   cmd.Schedule.EndBreak,
					},
				},
			},
		}
		emp.Schedule = schedule
	}

	if err := emp.Validate(ctx, operation); err != nil {
		return err
	}

	return s.repository.Save(ctx, &emp)
}

func (s *employeeService) RestoreEmployee(ctx context.Context, id EmployeeID) error {
	return s.repository.RestoreByID(ctx, id)
}

func (s *employeeService) DeleteEmployee(ctx context.Context, id EmployeeID) error {
	return s.repository.SoftDelete(ctx, id)
}

func (s *employeeService) GetEmployeeByID(ctx context.Context, id EmployeeID) (Employee, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *employeeService) GetEmployeesBySpecification(ctx context.Context, spec EmployeeSpecification) (page.Page[Employee], error) {
	return s.repository.GetBySpecification(ctx, spec)
}

func parseWeekday(value string) (time.Weekday, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "sunday":
		return time.Sunday, nil
	case "monday":
		return time.Monday, nil
	case "tuesday":
		return time.Tuesday, nil
	case "wednesday":
		return time.Wednesday, nil
	case "thursday":
		return time.Thursday, nil
	case "friday":
		return time.Friday, nil
	case "saturday":
		return time.Saturday, nil
	default:
		return 0, fmt.Errorf("invalid schedule day %q", value)
	}
}

func (s *employeeService) GetActiveEmployees(ctx context.Context, pagination page.Pagination) (page.Page[Employee], error) {
	return s.repository.GetActive(ctx, pagination)
}

func (s *employeeService) GetEmployeesBySpecialty(ctx context.Context, specialty VetSpecialty, pagination page.Pagination) (page.Page[Employee], error) {
	return s.repository.GetBySpeciality(ctx, specialty, pagination)
}

func (s *employeeService) GetEmployeeStats(ctx context.Context) (EmployeeStats, error) {
	stats := EmployeeStats{
		Specialties: make(map[VetSpecialty]int64),
	}

	var total int64
	for _, spec := range ValidVetSpecialties {
		count, err := s.repository.CountBySpeciality(ctx, spec)
		if err != nil {
			return EmployeeStats{}, err
		}
		if count > 0 {
			stats.Specialties[spec] = count
		}
		total += count
	}

	active, err := s.repository.CountActive(ctx)
	if err != nil {
		return EmployeeStats{}, err
	}

	stats.TotalEmployees = total
	stats.ActiveEmployees = active

	return stats, nil
}
