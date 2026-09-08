package ports

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/ksamwang/PersonalContentPlatform/internal/settings/domain"
)

type Repository interface {
	Get(context.Context, uuid.UUID) (domain.GeneralSettings, error)
	UpdateSection(context.Context, uuid.UUID, uuid.UUID, string, json.RawMessage) (domain.GeneralSettings, error)
	ActiveStorage(context.Context, uuid.UUID) (*domain.StorageProfile, error)
	StorageByID(context.Context, uuid.UUID, uuid.UUID) (*domain.StorageProfile, error)
	SaveStorage(context.Context, uuid.UUID, uuid.UUID, domain.StorageProfile) (domain.StorageProfile, error)
	AI(context.Context, uuid.UUID) (domain.AIConfig, error)
	SaveAI(context.Context, uuid.UUID, uuid.UUID, domain.AIConfig) (domain.AIConfig, error)
}
