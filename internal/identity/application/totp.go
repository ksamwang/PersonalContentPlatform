package application

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"image/png"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ksamwang/PersonalContentPlatform/internal/identity/domain"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

var (
	ErrTOTPInvalid     = errors.New("invalid TOTP or recovery code")
	ErrTOTPUnavailable = errors.New("TOTP login is unavailable")
	ErrTOTPRateLimited = errors.New("too many TOTP attempts")
)

type TOTPEnrollment struct {
	Secret     string `json:"secret"`
	OTPAuthURI string `json:"otpauth_uri"`
	QRDataURL  string `json:"qr_data_url"`
}

type TOTPStatus struct {
	Enabled bool `json:"enabled"`
}

func (s *Service) TOTPStatus(ctx context.Context, userID uuid.UUID) (TOTPStatus, error) {
	enabled, err := s.repo.HasTOTP(ctx, userID)
	return TOTPStatus{Enabled: enabled}, err
}

func (s *Service) BeginTOTPEnrollment(email string) (TOTPEnrollment, error) {
	key, err := totp.Generate(totp.GenerateOpts{Issuer: "Personal Content Platform", AccountName: email, Period: 30, SecretSize: 20, Digits: otp.DigitsSix, Algorithm: otp.AlgorithmSHA1})
	if err != nil {
		return TOTPEnrollment{}, err
	}
	image, err := key.Image(240, 240)
	if err != nil {
		return TOTPEnrollment{}, err
	}
	var output bytes.Buffer
	if err = png.Encode(&output, image); err != nil {
		return TOTPEnrollment{}, err
	}
	return TOTPEnrollment{
		Secret:     key.Secret(),
		OTPAuthURI: key.URL(),
		QRDataURL:  "data:image/png;base64," + base64.StdEncoding.EncodeToString(output.Bytes()),
	}, nil
}

func (s *Service) EnableTOTP(ctx context.Context, userID uuid.UUID, secret, code string) ([]string, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" || !validateTOTP(secret, code, time.Now()) {
		return nil, ErrTOTPInvalid
	}
	ciphertext, err := s.encryptTOTPSecret(secret)
	if err != nil {
		return nil, err
	}
	codes, hashes, err := s.newRecoveryCodes(8)
	if err != nil {
		return nil, err
	}
	if err = s.repo.SaveTOTP(ctx, userID, ciphertext); err != nil {
		return nil, err
	}
	if err = s.repo.ReplaceRecoveryCodes(ctx, userID, hashes); err != nil {
		return nil, err
	}
	return codes, nil
}

func (s *Service) DisableTOTP(ctx context.Context, userID uuid.UUID, code string) error {
	valid, err := s.verifyTOTPOrRecovery(ctx, userID, code)
	if err != nil {
		return err
	}
	if !valid {
		return ErrTOTPInvalid
	}
	return s.repo.DeleteTOTP(ctx, userID)
}

func (s *Service) LoginWithTOTP(ctx context.Context, email, code, client, userAgent string) (domain.Principal, string, error) {
	email, err := normalizeEmail(email)
	if err != nil {
		return domain.Principal{}, "", ErrUnauthorized
	}
	limitKey := email + "|" + strings.TrimSpace(client)
	if !s.totpLimit.Allow(limitKey, time.Now()) {
		return domain.Principal{}, "", ErrTOTPRateLimited
	}
	u, err := s.repo.UserByEmail(ctx, email)
	if err != nil || u.Status != "active" {
		s.totpLimit.Fail(limitKey, time.Now())
		return domain.Principal{}, "", ErrUnauthorized
	}
	policy, err := s.repo.AuthPolicyByUser(ctx, u.ID, s.sessionTTL)
	if err != nil || !policy.TOTPLoginEnabled {
		return domain.Principal{}, "", ErrTOTPUnavailable
	}
	valid, err := s.verifyTOTPOrRecovery(ctx, u.ID, code)
	if err != nil || !valid {
		s.totpLimit.Fail(limitKey, time.Now())
		return domain.Principal{}, "", ErrUnauthorized
	}
	s.totpLimit.Success(limitKey)
	return s.createSession(ctx, u.ID, policy.SessionTTL, userAgent)
}

func (s *Service) verifyTOTPOrRecovery(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	ciphertext, err := s.repo.TOTPSecretByUser(ctx, userID)
	if err != nil {
		return false, ErrTOTPUnavailable
	}
	secret, err := s.decryptTOTPSecret(ciphertext)
	if err != nil {
		return false, err
	}
	if validateTOTP(secret, code, time.Now()) {
		return true, nil
	}
	normalized := normalizeRecoveryCode(code)
	if len(normalized) != 12 {
		return false, nil
	}
	return s.repo.ConsumeRecoveryCode(ctx, userID, s.recoveryCodeHash(normalized))
}

func validateTOTP(secret, code string, now time.Time) bool {
	code = strings.TrimSpace(code)
	if len(code) != 6 {
		return false
	}
	valid, err := totp.ValidateCustom(code, secret, now, totp.ValidateOpts{Period: 30, Skew: 1, Digits: otp.DigitsSix, Algorithm: otp.AlgorithmSHA1})
	return err == nil && valid
}

func (s *Service) encryptTOTPSecret(secret string) ([]byte, error) {
	block, err := aes.NewCipher(s.totpEncryptionKey())
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}
	return aead.Seal(nonce, nonce, []byte(secret), nil), nil
}

func (s *Service) decryptTOTPSecret(ciphertext []byte) (string, error) {
	block, err := aes.NewCipher(s.totpEncryptionKey())
	if err != nil {
		return "", err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(ciphertext) < aead.NonceSize() {
		return "", fmt.Errorf("invalid TOTP secret")
	}
	plain, err := aead.Open(nil, ciphertext[:aead.NonceSize()], ciphertext[aead.NonceSize():], nil)
	return string(plain), err
}

func (s *Service) totpEncryptionKey() []byte {
	sum := sha256.Sum256([]byte("personal-content-platform:totp:" + s.pepper))
	return sum[:]
}

func (s *Service) newRecoveryCodes(count int) ([]string, [][]byte, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	codes := make([]string, 0, count)
	hashes := make([][]byte, 0, count)
	for range count {
		raw := make([]byte, 12)
		seed := make([]byte, 12)
		if _, err := rand.Read(seed); err != nil {
			return nil, nil, err
		}
		for i := range raw {
			raw[i] = alphabet[int(seed[i])%len(alphabet)]
		}
		normalized := string(raw)
		codes = append(codes, normalized[:4]+"-"+normalized[4:8]+"-"+normalized[8:])
		hashes = append(hashes, s.recoveryCodeHash(normalized))
	}
	return codes, hashes, nil
}

func normalizeRecoveryCode(code string) string {
	return strings.ToUpper(strings.NewReplacer("-", "", " ", "").Replace(strings.TrimSpace(code)))
}

func (s *Service) recoveryCodeHash(code string) []byte {
	mac := hmac.New(sha256.New, []byte(s.pepper))
	_, _ = mac.Write([]byte("totp-recovery:" + code))
	return mac.Sum(nil)
}

type totpAttempt struct {
	Failures int
	Blocked  time.Time
}

type totpAttemptLimiter struct {
	mu       sync.Mutex
	attempts map[string]totpAttempt
}

func newTOTPAttemptLimiter() *totpAttemptLimiter {
	return &totpAttemptLimiter{attempts: make(map[string]totpAttempt)}
}

func (l *totpAttemptLimiter) Allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	value := l.attempts[key]
	if !value.Blocked.IsZero() && now.Before(value.Blocked) {
		return false
	}
	if !value.Blocked.IsZero() {
		delete(l.attempts, key)
	}
	return true
}

func (l *totpAttemptLimiter) Fail(key string, now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	value := l.attempts[key]
	value.Failures++
	if value.Failures >= 5 {
		value.Blocked = now.Add(5 * time.Minute)
	}
	l.attempts[key] = value
}

func (l *totpAttemptLimiter) Success(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, key)
}
