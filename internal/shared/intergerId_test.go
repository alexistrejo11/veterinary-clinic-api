package shared

import "testing"

func TestBaseIDValueAndString(t *testing.T) {
	id := NewUserID(42)
	if id.Value() != 42 {
		t.Fatalf("expected ID value 42, got %d", id.Value())
	}
	if id.String() != "42" {
		t.Fatalf("expected ID string 42, got %q", id.String())
	}
}
