package application

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/ksamwang/PersonalContentPlatform/internal/identity/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/identity/ports"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/id"
	"net/http"
	"time"
)

type Passkeys struct {
	repo ports.Repository
	web  *webauthn.WebAuthn
}

func NewPasskeys(repo ports.Repository, rpID string, origins []string) (*Passkeys, error) {
	web, err := webauthn.New(&webauthn.Config{RPID: rpID, RPDisplayName: "Personal Content Platform", RPOrigins: origins})
	if err != nil {
		return nil, err
	}
	return &Passkeys{repo: repo, web: web}, nil
}
func (p *Passkeys) BeginRegistration(ctx context.Context, userID uuid.UUID) (any, uuid.UUID, error) {
	policy, err := p.repo.AuthPolicyByUser(ctx, userID, 30*24*time.Hour)
	if err != nil || !policy.PasskeyEnabled {
		return nil, uuid.Nil, fmt.Errorf("passkey is disabled")
	}
	u, err := p.repo.UserByID(ctx, userID)
	if err != nil {
		return nil, uuid.Nil, err
	}
	options, session, err := p.web.BeginRegistration(u)
	if err != nil {
		return nil, uuid.Nil, err
	}
	data, err := json.Marshal(session)
	if err != nil {
		return nil, uuid.Nil, err
	}
	ceremonyID := id.New()
	if err = p.repo.SaveCeremony(ctx, ceremonyID, userID, "registration", data, session.Expires); err != nil {
		return nil, uuid.Nil, err
	}
	return options, ceremonyID, nil
}
func (p *Passkeys) FinishRegistration(ctx context.Context, expectedUserID, ceremonyID uuid.UUID, r *http.Request) error {
	userID, data, err := p.repo.TakeCeremony(ctx, ceremonyID, "registration")
	if err != nil {
		return err
	}
	if userID != expectedUserID {
		return ErrUnauthorized
	}
	u, err := p.repo.UserByID(ctx, userID)
	if err != nil {
		return err
	}
	var session webauthn.SessionData
	if err = json.Unmarshal(data, &session); err != nil {
		return err
	}
	credential, err := p.web.FinishRegistration(u, session, r)
	if err != nil {
		return fmt.Errorf("verify passkey registration: %w", err)
	}
	return p.repo.SaveCredential(ctx, userID, *credential)
}
func (p *Passkeys) BeginLogin(ctx context.Context, email string) (any, uuid.UUID, error) {
	u, err := p.repo.UserByEmail(ctx, email)
	if err != nil {
		return nil, uuid.Nil, ErrUnauthorized
	}
	policy, err := p.repo.AuthPolicyByUser(ctx, u.ID, 30*24*time.Hour)
	if err != nil || !policy.PasskeyEnabled {
		return nil, uuid.Nil, ErrUnauthorized
	}
	options, session, err := p.web.BeginLogin(u)
	if err != nil {
		return nil, uuid.Nil, err
	}
	data, err := json.Marshal(session)
	if err != nil {
		return nil, uuid.Nil, err
	}
	ceremonyID := id.New()
	if err = p.repo.SaveCeremony(ctx, ceremonyID, u.ID, "login", data, time.Now().Add(5*time.Minute)); err != nil {
		return nil, uuid.Nil, err
	}
	return options, ceremonyID, nil
}
func (p *Passkeys) FinishLogin(ctx context.Context, ceremonyID uuid.UUID, r *http.Request, sessionTTL time.Duration) (domain.Principal, string, error) {
	userID, data, err := p.repo.TakeCeremony(ctx, ceremonyID, "login")
	if err != nil {
		return domain.Principal{}, "", ErrUnauthorized
	}
	u, err := p.repo.UserByID(ctx, userID)
	if err != nil {
		return domain.Principal{}, "", ErrUnauthorized
	}
	policy, err := p.repo.AuthPolicyByUser(ctx, userID, sessionTTL)
	if err != nil || !policy.PasskeyEnabled {
		return domain.Principal{}, "", ErrUnauthorized
	}
	sessionTTL = policy.SessionTTL
	var session webauthn.SessionData
	if err = json.Unmarshal(data, &session); err != nil {
		return domain.Principal{}, "", err
	}
	credential, err := p.web.FinishLogin(u, session, r)
	if err != nil {
		return domain.Principal{}, "", ErrUnauthorized
	}
	if err = p.repo.SaveCredential(ctx, userID, *credential); err != nil {
		return domain.Principal{}, "", err
	}
	token, hash, err := newToken()
	if err != nil {
		return domain.Principal{}, "", err
	}
	if err = p.repo.CreateSession(ctx, userID, hash, time.Now().Add(sessionTTL), r.UserAgent()); err != nil {
		return domain.Principal{}, "", err
	}
	principal, err := p.repo.PrincipalBySessionHash(ctx, hash)
	return principal, token, err
}
