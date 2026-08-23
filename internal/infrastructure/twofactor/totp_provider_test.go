package twofactor

import (
	"context"
	"errors"
	"testing"
	"time"

	"clinic-vet-api/internal/core/customers"
	"clinic-vet-api/internal/core/employees"
	"clinic-vet-api/internal/core/users"
	"clinic-vet-api/internal/shared"
	"clinic-vet-api/internal/shared/page"

	"github.com/pquerna/otp/totp"
)

type fakeUserQuery struct {
	user users.User
}

func (q fakeUserQuery) GetByID(context.Context, shared.UserID) (users.User, error) {
	return q.user, nil
}

func (fakeUserQuery) GetByEmail(context.Context, users.Email) (users.User, error) {
	return users.User{}, errors.New("not implemented")
}

func (fakeUserQuery) GetByPhone(context.Context, users.PhoneNumber) (users.User, error) {
	return users.User{}, errors.New("not implemented")
}

func (fakeUserQuery) GetByEmployeeID(context.Context, employees.EmployeeID) (users.User, error) {
	return users.User{}, errors.New("not implemented")
}

func (fakeUserQuery) GetByCustomerID(context.Context, customers.CustomerID) (users.User, error) {
	return users.User{}, errors.New("not implemented")
}

func (fakeUserQuery) GetByOAuthProvider(context.Context, string, string) (users.User, error) {
	return users.User{}, errors.New("not implemented")
}

func (fakeUserQuery) GetSpecification(context.Context, users.UserSpecification) (page.Page[users.User], error) {
	return page.Page[users.User]{}, errors.New("not implemented")
}

func TestTOTPProviderEncryptsAndVerifiesCode(t *testing.T) {
	key := []byte("01234567890123456789012345678901")
	provider, err := NewTOTPProvider(ProviderConfig{EncryptionKey: key, Users: fakeUserQuery{}})
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}

	encryptedSecret, backupCodes, err := provider.GenerateSecret(context.Background(), shared.NewUserID(1), "totp")
	if err != nil {
		t.Fatalf("generate secret: %v", err)
	}
	if encryptedSecret == "" || encryptedSecret == "totp" || len(backupCodes) != 8 {
		t.Fatalf("unexpected enrollment data: secret=%q codes=%d", encryptedSecret, len(backupCodes))
	}

	secret, err := provider.DecryptSecret(encryptedSecret)
	if err != nil {
		t.Fatalf("decrypt secret: %v", err)
	}
	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatalf("generate code: %v", err)
	}

	provider.users = fakeUserQuery{user: users.User{
		TwoFactorAuth: shared.TwoFactorAuth{
			IsEnabled: true,
			Method:    shared.TwoFactorMethodTOTP,
			Secret:    encryptedSecret,
		},
	}}
	valid, err := provider.VerifyCode(context.Background(), shared.NewUserID(1), code)
	if err != nil {
		t.Fatalf("verify code: %v", err)
	}
	if !valid {
		t.Fatal("expected generated TOTP code to be valid")
	}
}

func TestHashBackupCodes(t *testing.T) {
	hashes, err := HashBackupCodes([]string{"ABCD123"})
	if err != nil {
		t.Fatalf("hash backup codes: %v", err)
	}
	if len(hashes) != 1 || hashes[0] == "ABCD123" {
		t.Fatalf("expected one non-plaintext hash, got %#v", hashes)
	}
}
