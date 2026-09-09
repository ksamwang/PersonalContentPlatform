package ports

import (
	"context"
	"github.com/google/uuid"
	"github.com/ksamwang/PersonalContentPlatform/internal/knowledge/domain"
)

type Repository interface {
	CreateEntity(context.Context, uuid.UUID, string, string, string) (domain.Entity, error)
	ListEntities(context.Context, uuid.UUID, string, int) ([]domain.Entity, error)
	AddAlias(context.Context, uuid.UUID, uuid.UUID, string, string) error
	CreateRelation(context.Context, uuid.UUID, uuid.UUID, string, uuid.UUID, bool) (domain.Relation, error)
	ConfirmRelation(context.Context, uuid.UUID, uuid.UUID) error
	Relations(context.Context, uuid.UUID, uuid.UUID) ([]domain.Relation, error)
	UpdateEntity(context.Context, uuid.UUID, uuid.UUID, string, string, string) error
	DeleteEntity(context.Context, uuid.UUID, uuid.UUID) error
	DeleteAlias(context.Context, uuid.UUID, uuid.UUID) error
	DeleteRelation(context.Context, uuid.UUID, uuid.UUID) error
	ExtractMentions(context.Context, uuid.UUID) error
	ListMentions(context.Context, uuid.UUID, bool) ([]domain.Mention, error)
	ConfirmMention(context.Context, uuid.UUID, uuid.UUID) error
}
