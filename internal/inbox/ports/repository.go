package ports

import (
	"context"

	"github.com/google/uuid"
	contentdomain "github.com/ksamwang/PersonalContentPlatform/internal/content/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/inbox/domain"
)

type Repository interface {
	Create(context.Context, uuid.UUID, uuid.UUID, domain.Kind, string, *string, *uuid.UUID) (domain.Item, error)
	List(context.Context, uuid.UUID, domain.State, int) ([]domain.Item, error)
	Archive(context.Context, uuid.UUID, uuid.UUID) error
	Convert(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, contentdomain.Type, string, string, string) (domain.Conversion, error)
	Get(context.Context, uuid.UUID, uuid.UUID) (domain.Item, error)
	BeginProcessing(context.Context, uuid.UUID, uuid.UUID) error
	CompleteProcessing(context.Context, uuid.UUID, uuid.UUID, string, string, *string, *uuid.UUID) (domain.Item, error)
	FailProcessing(context.Context, uuid.UUID, uuid.UUID, string) error
	FindDuplicate(context.Context, uuid.UUID, uuid.UUID, *string, string) (*uuid.UUID, error)
}
