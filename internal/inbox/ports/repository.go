package ports

import (
	"context"

	"github.com/google/uuid"
	contentdomain "github.com/ksamwang/PersonalContentPlatform/internal/content/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/inbox/domain"
)

type Repository interface {
	Create(context.Context, uuid.UUID, uuid.UUID, domain.Kind, string, *string) (domain.Item, error)
	List(context.Context, uuid.UUID, domain.State, int) ([]domain.Item, error)
	Archive(context.Context, uuid.UUID, uuid.UUID) error
	Convert(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, contentdomain.Type, string, string, string) (domain.Conversion, error)
}
