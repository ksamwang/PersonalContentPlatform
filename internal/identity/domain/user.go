package domain

import (
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

type User struct {
	ID                                       uuid.UUID
	Email, DisplayName, PasswordHash, Status string
	Credentials                              []webauthn.Credential
}

func (u User) WebAuthnID() []byte                         { return u.ID[:] }
func (u User) WebAuthnName() string                       { return u.Email }
func (u User) WebAuthnDisplayName() string                { return u.DisplayName }
func (u User) WebAuthnCredentials() []webauthn.Credential { return u.Credentials }

type Principal struct {
	UserID, WorkspaceID      uuid.UUID
	Email, DisplayName, Role string
}
