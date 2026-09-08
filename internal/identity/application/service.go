package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/ksamwang/PersonalContentPlatform/internal/identity/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/identity/ports"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/id"
	"net/mail"
	"strings"
	"time"
)

var ErrUnauthorized = errors.New("unauthorized")
var ErrSetupComplete = errors.New("initial setup already completed")

type Service struct {
	repo       ports.Repository
	pepper     string
	sessionTTL time.Duration
}

func NewService(repo ports.Repository, pepper string, sessionTTL time.Duration) *Service {
	return &Service{repo: repo, pepper: pepper, sessionTTL: sessionTTL}
}

func (s *Service) Setup(ctx context.Context, email, displayName, password string) (domain.Principal, string, error) {
	exists, err := s.repo.HasUsers(ctx)
	if err != nil {
		return domain.Principal{}, "", err
	}
	if exists {
		return domain.Principal{}, "", ErrSetupComplete
	}
	email, err = normalizeEmail(email)
	if err != nil {
		return domain.Principal{}, "", err
	}
	if len(password) < 10 {
		return domain.Principal{}, "", fmt.Errorf("password must contain at least 10 characters")
	}
	hash, err := hashPassword(password, s.pepper)
	if err != nil {
		return domain.Principal{}, "", err
	}
	u := domain.User{ID: id.New(), Email: email, DisplayName: strings.TrimSpace(displayName), PasswordHash: hash, Status: "active"}
	if u.DisplayName == "" {
		u.DisplayName = email
	}
	principal, err := s.repo.CreateOwner(ctx, u, "personal", "Personal Workspace")
	if err != nil {
		return domain.Principal{}, "", err
	}
	token, tokenHash, err := newToken()
	if err != nil {
		return domain.Principal{}, "", err
	}
	if err = s.repo.CreateSession(ctx, u.ID, tokenHash, time.Now().Add(s.sessionTTL), ""); err != nil {
		return domain.Principal{}, "", err
	}
	return principal, token, nil
}
func (s *Service) Login(ctx context.Context, email, password, userAgent string) (domain.Principal, string, error) {
	email, err := normalizeEmail(email)
	if err != nil {
		return domain.Principal{}, "", ErrUnauthorized
	}
	u, err := s.repo.UserByEmail(ctx, email)
	if err != nil || u.Status != "active" || !verifyPassword(u.PasswordHash, password, s.pepper) {
		return domain.Principal{}, "", ErrUnauthorized
	}
	policy, err := s.repo.AuthPolicyByUser(ctx, u.ID, s.sessionTTL)
	if err != nil || !policy.PasswordLoginEnabled {
		return domain.Principal{}, "", ErrUnauthorized
	}
	token, hash, err := newToken()
	if err != nil {
		return domain.Principal{}, "", err
	}
	if err = s.repo.CreateSession(ctx, u.ID, hash, time.Now().Add(policy.SessionTTL), userAgent); err != nil {
		return domain.Principal{}, "", err
	}
	principal, err := s.repo.PrincipalBySessionHash(ctx, hash)
	return principal, token, err
}
func (s *Service) Authenticate(ctx context.Context, token string) (domain.Principal, error) {
	hash, err := tokenHash(token)
	if err != nil {
		return domain.Principal{}, ErrUnauthorized
	}
	p, err := s.repo.PrincipalBySessionHash(ctx, hash)
	if err != nil {
		return domain.Principal{}, ErrUnauthorized
	}
	return p, nil
}
func (s *Service) Logout(ctx context.Context, token string) error {
	hash, err := tokenHash(token)
	if err != nil {
		return nil
	}
	return s.repo.DeleteSession(ctx, hash)
}
func normalizeEmail(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	address, err := mail.ParseAddress(value)
	if err != nil || address.Address != value {
		return "", fmt.Errorf("invalid email")
	}
	return value, nil
}
