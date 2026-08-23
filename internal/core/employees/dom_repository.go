package employees

import (
	p "clinic-vet-api/internal/shared/page"
	"context"
)

type EmployeeRepository interface {
	GetBySpecification(ctx context.Context, spec EmployeeSpecification) (p.Page[Employee], error)
	GetByID(ctx context.Context, id EmployeeID) (Employee, error)
	GetByUserID(ctx context.Context, userID uint) (Employee, error)
	GetActive(ctx context.Context, pagination p.Pagination) (p.Page[Employee], error)
	GetBySpeciality(ctx context.Context, speciality VetSpecialty, pagination p.Pagination) (p.Page[Employee], error)

	ExistsByID(ctx context.Context, id EmployeeID) (bool, error)
	ExistsByUserID(ctx context.Context, userID uint) (bool, error)

	Save(ctx context.Context, employee *Employee) error
	SoftDelete(ctx context.Context, id EmployeeID) error
	HardDelete(ctx context.Context, id EmployeeID) error
	RestoreByID(ctx context.Context, id EmployeeID) error

	CountBySpeciality(ctx context.Context, speciality VetSpecialty) (int64, error)
	CountByActive(ctx context.Context, active *bool) (int64, error)
	CountActive(ctx context.Context) (int64, error)
}
