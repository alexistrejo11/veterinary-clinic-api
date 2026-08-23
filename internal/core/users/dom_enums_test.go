package users

import "testing"

func TestValidUserRolesMatchCanonicalAuthorizationModel(t *testing.T) {
	want := map[UserRole]bool{
		UserRoleAdmin:        true,
		UserRoleVeterinarian: true,
		UserRoleReceptionist: true,
		UserRoleCustomer:     true,
	}

	if len(ValidUserRoles) != len(want) {
		t.Fatalf("expected %d valid roles, got %d", len(want), len(ValidUserRoles))
	}

	for _, role := range ValidUserRoles {
		if !want[role] {
			t.Fatalf("unexpected valid role %q", role)
		}
	}

	for _, legacyRole := range []UserRole{"manager", "employee", "superadmin"} {
		if legacyRole.IsValid() {
			t.Errorf("legacy role %q must not be valid", legacyRole)
		}
	}
}
