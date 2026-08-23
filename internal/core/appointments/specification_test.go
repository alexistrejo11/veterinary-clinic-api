package appointments

import "testing"

func TestSpecificationPreservesRequestedPagination(t *testing.T) {
	spec := NewAppointmentSpecification().WithPagination(3, 25, "", "ASC")
	params := spec.ToSearchParams()
	if params.Limit != 25 || params.Offset != 50 {
		t.Fatalf("got limit=%d offset=%d; want limit=25 offset=50", params.Limit, params.Offset)
	}
}
