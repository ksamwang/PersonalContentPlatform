package ports

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/ksamwang/PersonalContentPlatform/internal/settings/domain"
)

type Repository interface {
	Get(context.Context, uuid.UUID) (domain.GeneralSettings, error)
	GetBySlug(context.Context, string) (domain.GeneralSettings, error)
	UpdateSection(context.Context, uuid.UUID, uuid.UUID, string, json.RawMessage) (domain.GeneralSettings, error)
	ActiveStorage(context.Context, uuid.UUID) (*domain.StorageProfile, error)
	StorageByID(context.Context, uuid.UUID, uuid.UUID) (*domain.StorageProfile, error)
	SaveStorage(context.Context, uuid.UUID, uuid.UUID, domain.StorageProfile) (domain.StorageProfile, error)
	AI(context.Context, uuid.UUID) (domain.AIConfig, error)
	SaveAI(context.Context, uuid.UUID, uuid.UUID, domain.AIConfig) (domain.AIConfig, error)
	Embedding(context.Context, uuid.UUID) (domain.EmbeddingConfig, error)
	SaveEmbedding(context.Context, uuid.UUID, uuid.UUID, domain.EmbeddingConfig) (domain.EmbeddingConfig, error)
	Media(context.Context, uuid.UUID, string) (domain.MediaConfig, error)
	SaveMedia(context.Context, uuid.UUID, uuid.UUID, domain.MediaConfig) (domain.MediaConfig, error)
}
