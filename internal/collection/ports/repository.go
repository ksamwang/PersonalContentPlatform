package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/ksamwang/PersonalContentPlatform/internal/collection/domain"
)

type Repository interface {
	Create(context.Context, uuid.UUID, string, string, string) (domain.Collection, error)
	List(context.Context, uuid.UUID) ([]domain.Collection, error)
	Get(context.Context, uuid.UUID, uuid.UUID) (domain.Collection, error)
	Update(context.Context, uuid.UUID, uuid.UUID, string, string, string) (domain.Collection, error)
	AddSection(context.Context, uuid.UUID, uuid.UUID, string) (domain.Section, error)
	MoveSection(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, int) error
	AddItem(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, string) (domain.Item, error)
	MoveItem(context.Context, uuid.UUID, uuid.UUID, int) error
	RemoveItem(context.Context, uuid.UUID, uuid.UUID) error
}
