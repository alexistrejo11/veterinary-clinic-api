package users

import (
	"clinic-vet-api/internal/shared"
	"clinic-vet-api/internal/shared/password"
	"context"

	"golang.org/x/crypto/bcrypt"
)

type CommandService interface {
	CreateUser(ctx context.Context, cmd CreateUserCommand) (User, error)
	UpdatePassword(ctx context.Context, cmd UpdatePasswordCommand) error
	ResetPasswordByCode(ctx context.Context, cmd ResetPasswordByCodeCommand) error
	SetEmailVerified(ctx context.Context, id shared.UserID) error
	EnableTwoFactorAuth(ctx context.Context, id shared.UserID, method, encryptedSecret string, backupCodeHashes []string) error
	DisableTwoFactorAuth(ctx context.Context, id shared.UserID) error
	ConsumeBackupCode(ctx context.Context, id shared.UserID, code string) (bool, error)
	UpdateUserStatus(ctx context.Context, cmd UpdateUserStatusCommand) error
	UpdateProfileData(ctx context.Context, cmd UpdateProfileCommand) error
	RestoreUser(ctx context.Context, id shared.UserID) error
	DeleteUser(ctx context.Context, cmd DeleteUserCommand) error
}

type commandService struct {
	repository      UserRepository
	passwordEncoder password.PasswordEncoder
}

func NewCommandService(
	repository UserRepository, passwordEncoder password.PasswordEncoder,
) CommandService {
	return &commandService{
		repository:      repository,
		passwordEncoder: passwordEncoder,
	}
}

func (s *commandService) CreateUser(ctx context.Context, cmd CreateUserCommand) (User, error) {
	if err := s.validateUniqueEmail(ctx, cmd.Email, "CreateUser"); err != nil {
		return User{}, err
	}

	if cmd.PhoneNumber != nil {
		if err := s.validateUniquePhone(ctx, *cmd.PhoneNumber, "CreateUser"); err != nil {
			return User{}, err
		}
	}

	hashedPassword, err := s.hashPasswordWithValidation(ctx, cmd.PlainPassword)
	if err != nil {
		return User{}, err
	}

	user := s.createUserEntity(cmd, hashedPassword)

	err = s.repository.Save(ctx, &user)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (s *commandService) UpdatePassword(ctx context.Context, cmd UpdatePasswordCommand) error {
	user, err := s.repository.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	valid := s.passwordEncoder.CheckPassword(user.HashedPassword, cmd.CurrentPassword)
	if !valid {
		return PasswordMismatchError(ctx, "UpdatePassword")
	}

	if err := shared.ValidatePasswordStrength(cmd.NewPassword); err != nil {
		return err
	}

	hashedPassword, err := s.passwordEncoder.HashPassword(cmd.NewPassword)
	if err != nil {
		return err

	}

	user.UpdatePassword(hashedPassword)
	return s.repository.Save(ctx, &user)
}

func (s *commandService) ResetPasswordByCode(ctx context.Context, cmd ResetPasswordByCodeCommand) error {
	user, err := s.repository.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}
	if err := shared.ValidatePasswordStrength(cmd.NewPassword); err != nil {
		return err
	}
	hashedPassword, err := s.passwordEncoder.HashPassword(cmd.NewPassword)
	if err != nil {
		return err
	}
	user.UpdatePassword(hashedPassword)
	return s.repository.Save(ctx, &user)
}

func (s *commandService) SetEmailVerified(ctx context.Context, id shared.UserID) error {
	user, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if !user.EmailVerified {
		user.EmailVerified = true
		user.IncrementVersion()
	}
	return s.repository.Save(ctx, &user)
}

func (s *commandService) EnableTwoFactorAuth(ctx context.Context, id shared.UserID, method, encryptedSecret string, backupCodeHashes []string) error {
	user, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := user.EnableTwoFactorAuth(method, encryptedSecret, backupCodeHashes); err != nil {
		return err
	}
	return s.repository.Save(ctx, &user)
}

func (s *commandService) DisableTwoFactorAuth(ctx context.Context, id shared.UserID) error {
	user, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := user.DisableTwoFactorAuth(); err != nil {
		return err
	}
	return s.repository.Save(ctx, &user)
}

func (s *commandService) ConsumeBackupCode(ctx context.Context, id shared.UserID, code string) (bool, error) {
	user, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return false, err
	}
	for index, hash := range user.TwoFactorAuth.BackupCodes {
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(code)) != nil {
			continue
		}
		user.TwoFactorAuth.BackupCodes = append(user.TwoFactorAuth.BackupCodes[:index], user.TwoFactorAuth.BackupCodes[index+1:]...)
		user.TwoFactorAuth.RecordUsage()
		if err := s.repository.Save(ctx, &user); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (s *commandService) UpdateUserStatus(ctx context.Context, cmd UpdateUserStatusCommand) error {
	user, err := s.repository.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	switch cmd.Status {
	case UserStatusActive:
		if err := user.Activate(); err != nil {
			return err
		}
	case UserStatusInactive:
		if err := user.Deactivate(); err != nil {
			return err
		}
	case UserStatusBanned:
		if err := user.Ban(); err != nil {
			return err
		}
	default:
		return InvalidUserStatusError(ctx, cmd.Status.DisplayName(), "UpdateUserStatus")
	}

	return s.repository.Save(ctx, &user)
}

func (s *commandService) UpdatePhoneNumber(ctx context.Context, cmd UpdatePhoneCommand) error {
	user, err := s.repository.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	if err := s.validateUniquePhone(ctx, cmd.Phone, "UpdatePhoneNumber"); err != nil {
		return err
	}

	user.UpdatePhoneNumber(cmd.Phone)
	return s.repository.Save(ctx, &user)
}

func (h *commandService) UpdateEmail(ctx context.Context, cmd UpdateEmailCommand) error {
	user, err := h.repository.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	if err := h.validateUniqueEmail(ctx, cmd.Email, "UpdateEmail"); err != nil {
		return err
	}

	user.UpdateEmail(cmd.Email)
	return h.repository.Save(ctx, &user)
}

func (s *commandService) RestoreUser(ctx context.Context, id shared.UserID) error {
	if deleted, err := s.repository.IsDeletedByID(ctx, id); err != nil {
		return err
	} else if !deleted {
		return UserNotDeletedError(ctx, id, "RestoreUser")
	}

	return s.repository.RestoreByID(ctx, id)
}

func (s *commandService) DeleteUser(ctx context.Context, cmd DeleteUserCommand) error {
	if _, err := s.repository.FindByID(ctx, cmd.ID); err != nil {
		return err
	}

	if cmd.IsHardDelete {
		return s.repository.HardDelete(ctx, cmd.ID)
	}

	return s.repository.SoftDelete(ctx, cmd.ID)

}

// validateUniqueEmail checks if email is already in use
func (s *commandService) validateUniqueEmail(ctx context.Context, email Email, operation string) error {
	exists, err := s.repository.ExistsByEmail(ctx, email)
	if err != nil {
		return err
	}
	if exists {
		return EmailAlreadyExistsError(ctx, email, operation)
	}
	return nil
}

// validateUniquePhone checks if phone number is already in use
func (s *commandService) validateUniquePhone(ctx context.Context, phone PhoneNumber, operation string) error {
	exists, err := s.repository.ExistsByPhone(ctx, phone)
	if err != nil {
		return err
	}
	if exists {
		return PhoneAlreadyExistsError(ctx, phone, operation)
	}
	return nil
}

// hashPasswordWithValidation validates password strength and returns hashed password
func (s *commandService) hashPasswordWithValidation(ctx context.Context, plainPassword string) (string, error) {
	// Validate password strength
	if err := shared.ValidatePasswordStrength(plainPassword); err != nil {
		return "", err
	}

	// Hash password
	hashedPassword, err := s.passwordEncoder.HashPassword(plainPassword)
	if err != nil {
		return "", err
	}

	return hashedPassword, nil
}

func (s *commandService) UpdateProfileData(ctx context.Context, cmd UpdateProfileCommand) error {
	user, err := s.repository.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	Profile := Profile{
		Name:        cmd.Name,
		Gender:      cmd.Gender,
		PhotoURL:    cmd.PhotoURL,
		Bio:         cmd.Bio,
		DateOfBirth: cmd.DateOfBirth,
	}

	if err := user.UpdateProfile(Profile); err != nil {
		return err
	}

	return s.repository.Save(ctx, &user)
}

// createUserEntity creates a new User entity from command
func (s *commandService) createUserEntity(cmd CreateUserCommand, hashedPassword string) User {
	phoneNumber := PhoneNumber{}
	if cmd.PhoneNumber != nil {
		phoneNumber = *cmd.PhoneNumber
	}

	return User{
		Entity:         shared.CreateEntity(shared.UserID{}),
		Email:          cmd.Email,
		PhoneNumber:    phoneNumber,
		HashedPassword: hashedPassword,
		Role:           cmd.Role,
		Status:         cmd.Status,
		TwoFactorAuth:  shared.NewDisabledTwoFactorAuth(),
		Profile:        Profile{},
		EmailVerified:  false,
		LoginAttempts:  0,
	}
}
