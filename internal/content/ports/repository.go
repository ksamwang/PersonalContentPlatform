package ports

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/ksamwang/PersonalContentPlatform/internal/content/domain"
)

type Repository interface {
	Create(context.Context, uuid.UUID, uuid.UUID, domain.Type, string, string, string) (domain.Content, error)
	List(context.Context, uuid.UUID, string, string, int) ([]domain.Content, error)
	Get(context.Context, uuid.UUID, uuid.UUID) (domain.Content, error)
	GetDraft(context.Context, uuid.UUID, uuid.UUID) (domain.Draft, error)
	SaveDraft(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, int, string, string, json.RawMessage, json.RawMessage) (domain.Draft, error)
	SealRevision(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, int) (domain.Revision, error)
	MarkReady(context.Context, uuid.UUID, uuid.UUID) error
	Publish(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, string, uuid.UUID) (uuid.UUID, error)
	Search(context.Context, uuid.UUID, string, string, int) ([]domain.Content, error)
}
