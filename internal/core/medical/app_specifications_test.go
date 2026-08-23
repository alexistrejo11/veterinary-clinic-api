package medical

import "testing"

func TestMedicalSessionSpecificationRestrictsEmployeeAssignments(t *testing.T) {
	specification := MedicalSessionSpecification{EmployeeIDs: []uint{7}}

	if specification.IsSatisfiedBy(MedicalSession{EmployeeID: 8}) {
		t.Fatal("session assigned to another veterinarian must not satisfy the specification")
	}
	if !specification.IsSatisfiedBy(MedicalSession{EmployeeID: 7}) {
		t.Fatal("session assigned to the veterinarian must satisfy the specification")
	}
}
