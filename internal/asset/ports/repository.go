package ports

import (
	"context"
	"github.com/google/uuid"
	"github.com/ksamwang/PersonalContentPlatform/internal/asset/domain"
)

type Repository interface {
	CreateIntent(context.Context, domain.UploadIntent, uuid.UUID) error
	Intent(context.Context, uuid.UUID, uuid.UUID) (domain.UploadIntent, error)
	Complete(context.Context, domain.UploadIntent, uuid.UUID, string, int64) (domain.Asset, error)
	List(context.Context, uuid.UUID, int) ([]domain.Asset, error)
	Get(context.Context, uuid.UUID, uuid.UUID) (domain.Asset, error)
	Usages(context.Context, uuid.UUID, uuid.UUID) ([]domain.Usage, error)
	Archive(context.Context, uuid.UUID, uuid.UUID) error
	Replace(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (domain.Asset, error)
	Object(context.Context, uuid.UUID) (domain.StoredObject, error)
}
