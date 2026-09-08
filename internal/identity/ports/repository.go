package ports

import (
	"context"
	"encoding/json"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/ksamwang/PersonalContentPlatform/internal/identity/domain"
	"time"
)

type Repository interface {
	HasUsers(context.Context) (bool, error)
	CreateOwner(context.Context, domain.User, string, string) (domain.Principal, error)
	UserByEmail(context.Context, string) (domain.User, error)
	UserByID(context.Context, uuid.UUID) (domain.User, error)
	CreateSession(context.Context, uuid.UUID, []byte, time.Time, string) error
	PrincipalBySessionHash(context.Context, []byte) (domain.Principal, error)
	DeleteSession(context.Context, []byte) error
	SaveCeremony(context.Context, uuid.UUID, uuid.UUID, string, json.RawMessage, time.Time) error
	TakeCeremony(context.Context, uuid.UUID, string) (uuid.UUID, json.RawMessage, error)
	SaveCredential(context.Context, uuid.UUID, webauthn.Credential) error
}
