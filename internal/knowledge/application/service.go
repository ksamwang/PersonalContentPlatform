package application

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/ksamwang/PersonalContentPlatform/internal/knowledge/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/knowledge/ports"
	"strings"
)

type Service struct{ repo ports.Repository }

func New(repo ports.Repository) *Service { return &Service{repo: repo} }
func (s *Service) CreateEntity(ctx context.Context, ws uuid.UUID, kind, name, description string) (domain.Entity, error) {
	name = strings.TrimSpace(name)
	if name == "" || kind == "" {
		return domain.Entity{}, fmt.Errorf("type and canonical name are required")
	}
	return s.repo.CreateEntity(ctx, ws, kind, name, strings.TrimSpace(description))
}
func (s *Service) ListEntities(ctx context.Context, ws uuid.UUID, q string, limit int) ([]domain.Entity, error) {
	if limit < 1 || limit > 100 {
		limit = 50
	}
	return s.repo.ListEntities(ctx, ws, strings.TrimSpace(q), limit)
}
func (s *Service) AddAlias(ctx context.Context, ws, id uuid.UUID, value, locale string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("alias is required")
	}
	return s.repo.AddAlias(ctx, ws, id, strings.TrimSpace(value), locale)
}
func (s *Service) CreateRelation(ctx context.Context, ws, source uuid.UUID, predicate string, target uuid.UUID, confirmed bool) (domain.Relation, error) {
	if source == target {
		return domain.Relation{}, fmt.Errorf("self relations are not allowed")
	}
	return s.repo.CreateRelation(ctx, ws, source, predicate, target, confirmed)
}
func (s *Service) Confirm(ctx context.Context, ws, id uuid.UUID) error {
	return s.repo.ConfirmRelation(ctx, ws, id)
}
func (s *Service) Relations(ctx context.Context, ws, objectID uuid.UUID) ([]domain.Relation, error) {
	return s.repo.Relations(ctx, ws, objectID)
}
