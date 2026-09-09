package ports

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/ksamwang/PersonalContentPlatform/internal/content/domain"
	"time"
)

type Repository interface {
	Create(context.Context, uuid.UUID, uuid.UUID, domain.Type, string, string, string) (domain.Content, error)
	CreateLocalization(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, string, string, string) (domain.Content, error)
	List(context.Context, uuid.UUID, domain.ListFilter) ([]domain.Content, error)
	Get(context.Context, uuid.UUID, uuid.UUID) (domain.Content, error)
	Archive(context.Context, uuid.UUID, uuid.UUID) error
	Restore(context.Context, uuid.UUID, uuid.UUID) error
	SoftDelete(context.Context, uuid.UUID, uuid.UUID) error
	UpdateProperties(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, string, domain.Visibility) error
	ReadinessSource(context.Context, uuid.UUID, uuid.UUID) (domain.ReadinessSource, error)
	GetDraft(context.Context, uuid.UUID, uuid.UUID) (domain.Draft, error)
	SaveDraft(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, int, string, string, json.RawMessage, json.RawMessage) (domain.Draft, error)
	SealRevision(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, int) (domain.Revision, error)
	MarkReady(context.Context, uuid.UUID, uuid.UUID) error
	Publish(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, string, uuid.UUID, string) (uuid.UUID, error)
	Search(context.Context, uuid.UUID, string, string, int) ([]domain.Content, error)
	ListRevisions(context.Context, uuid.UUID, uuid.UUID, int) ([]domain.Revision, error)
	GetRevision(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (domain.Revision, error)
	RestoreRevision(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, int) (domain.Draft, error)
	CreatePreview(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, []byte, time.Time) error
	GetPreview(context.Context, []byte) (domain.Preview, error)
}
