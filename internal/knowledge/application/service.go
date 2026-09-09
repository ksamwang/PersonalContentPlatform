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
func (s *Service) UpdateEntity(ctx context.Context, ws, entityID uuid.UUID, kind, name, description string) error {
	if strings.TrimSpace(kind) == "" || strings.TrimSpace(name) == "" {
		return fmt.Errorf("type and canonical name are required")
	}
	return s.repo.UpdateEntity(ctx, ws, entityID, strings.TrimSpace(kind), strings.TrimSpace(name), strings.TrimSpace(description))
}
func (s *Service) DeleteEntity(ctx context.Context, ws, entityID uuid.UUID) error {
	return s.repo.DeleteEntity(ctx, ws, entityID)
}
func (s *Service) DeleteAlias(ctx context.Context, ws, aliasID uuid.UUID) error {
	return s.repo.DeleteAlias(ctx, ws, aliasID)
}
func (s *Service) DeleteRelation(ctx context.Context, ws, relationID uuid.UUID) error {
	return s.repo.DeleteRelation(ctx, ws, relationID)
}
func (s *Service) ExtractMentions(ctx context.Context, ws uuid.UUID) ([]domain.Mention, error) {
	if err := s.repo.ExtractMentions(ctx, ws); err != nil {
		return nil, err
	}
	return s.repo.ListMentions(ctx, ws, false)
}
func (s *Service) Mentions(ctx context.Context, ws uuid.UUID, confirmed bool) ([]domain.Mention, error) {
	return s.repo.ListMentions(ctx, ws, confirmed)
}
func (s *Service) ConfirmMention(ctx context.Context, ws, mentionID uuid.UUID) error {
	return s.repo.ConfirmMention(ctx, ws, mentionID)
}
