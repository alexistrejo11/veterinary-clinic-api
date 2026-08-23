package twofactor

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"

	"clinic-vet-api/internal/core/auth"
	"clinic-vet-api/internal/core/users"
	"clinic-vet-api/internal/shared"

	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

const encryptedSecretVersion = "v1:"

type ProviderConfig struct {
	EncryptionKey []byte
	Issuer        string
	Users         users.QueryService
}

// TOTPProvider generates encrypted TOTP secrets and verifies TOTP or hashed
// backup codes. Persisting enrollment changes is handled by the auth workflow.
type TOTPProvider struct {
	block  cipher.AEAD
	issuer string
	users  users.QueryService
}

func NewTOTPProvider(config ProviderConfig) (*TOTPProvider, error) {
	if len(config.EncryptionKey) != 32 {
		return nil, errors.New("two-factor encryption key must be 32 bytes")
	}
	if config.Users == nil {
		return nil, errors.New("user query service is required")
	}

	block, err := aes.NewCipher(config.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("create two-factor cipher: %w", err)
	}
	issuer := strings.TrimSpace(config.Issuer)
	if issuer == "" {
		issuer = "Clinic Vet API"
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create two-factor AEAD: %w", err)
	}

	return &TOTPProvider{
		block:  aead,
		issuer: issuer,
		users:  config.Users,
	}, nil
}

var _ auth.TwoFactorProvider = (*TOTPProvider)(nil)

func (p *TOTPProvider) GenerateSecret(_ context.Context, userID shared.UserID, method string) (string, []string, error) {
	if !strings.EqualFold(method, string(shared.TwoFactorMethodTOTP)) {
		return "", nil, fmt.Errorf("unsupported two-factor method: %s", method)
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      p.issuer,
		AccountName: userID.String(),
	})
	if err != nil {
		return "", nil, fmt.Errorf("generate TOTP secret: %w", err)
	}

	encrypted, err := p.EncryptSecret(key.Secret())
	if err != nil {
		return "", nil, err
	}
	backupCodes, err := GenerateBackupCodes(8)
	if err != nil {
		return "", nil, err
	}
	return encrypted, backupCodes, nil
}

func (p *TOTPProvider) VerifyCode(ctx context.Context, userID shared.UserID, code string) (bool, error) {
	user, err := p.users.GetByID(ctx, userID)
	if err != nil {
		return false, err
	}
	if !user.TwoFactorAuth.IsEnabled {
		return false, nil
	}

	for _, backupCode := range user.TwoFactorAuth.BackupCodes {
		if bcrypt.CompareHashAndPassword([]byte(backupCode), []byte(code)) == nil {
			return true, nil
		}
	}

	if user.TwoFactorAuth.Method != shared.TwoFactorMethodTOTP {
		return false, nil
	}
	secret, err := p.DecryptSecret(user.TwoFactorAuth.Secret)
	if err != nil {
		return false, err
	}
	return totp.Validate(code, secret), nil
}

// Disable has no external state to clear. The auth workflow persists the user
// state change and removes the encrypted secret and backup-code hashes.
func (p *TOTPProvider) Disable(context.Context, shared.UserID) error {
	return nil
}

func (p *TOTPProvider) EncryptSecret(secret string) (string, error) {
	if secret == "" {
		return "", errors.New("two-factor secret is required")
	}
	nonce := make([]byte, p.block.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate encryption nonce: %w", err)
	}
	ciphertext := p.block.Seal(nil, nonce, []byte(secret), nil)
	return encryptedSecretVersion + base64.RawStdEncoding.EncodeToString(append(nonce, ciphertext...)), nil
}

func (p *TOTPProvider) DecryptSecret(value string) (string, error) {
	if !strings.HasPrefix(value, encryptedSecretVersion) {
		return "", errors.New("two-factor secret is not encrypted")
	}
	payload, err := base64.RawStdEncoding.DecodeString(strings.TrimPrefix(value, encryptedSecretVersion))
	if err != nil {
		return "", fmt.Errorf("decode encrypted two-factor secret: %w", err)
	}
	if len(payload) < p.block.NonceSize() {
		return "", errors.New("encrypted two-factor secret is malformed")
	}
	plaintext, err := p.block.Open(nil, payload[:p.block.NonceSize()], payload[p.block.NonceSize():], nil)
	if err != nil {
		return "", errors.New("decrypt two-factor secret")
	}
	return string(plaintext), nil
}

func GenerateBackupCodes(count int) ([]string, error) {
	if count < 1 {
		return nil, errors.New("backup code count must be positive")
	}

	codes := make([]string, count)
	for i := range codes {
		bytes := make([]byte, 5)
		if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
			return nil, fmt.Errorf("generate backup code: %w", err)
		}
		codes[i] = strings.ToUpper(base64.RawStdEncoding.EncodeToString(bytes))
	}
	return codes, nil
}

func HashBackupCodes(codes []string) ([]string, error) {
	hashes := make([]string, len(codes))
	for i, code := range codes {
		hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("hash backup code: %w", err)
		}
		hashes[i] = string(hash)
	}
	return hashes, nil
}

func (p *TOTPProvider) HashBackupCodes(codes []string) ([]string, error) {
	return HashBackupCodes(codes)
}
