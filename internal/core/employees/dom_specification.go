package employees

import "clinic-vet-api/internal/shared/page"

type EmployeeSpecification struct {
	Specialty *VetSpecialty
	IsActive  *bool
	page.Pagination
}
