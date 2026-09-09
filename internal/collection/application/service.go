package application

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/ksamwang/PersonalContentPlatform/internal/collection/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/collection/ports"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type Service struct{ repo ports.Repository }

func New(repo ports.Repository) *Service { return &Service{repo: repo} }

func (s *Service) Create(ctx context.Context, workspaceID uuid.UUID, title, slug, visibility string) (domain.Collection, error) {
	title, slug, visibility, err := validate(title, slug, visibility)
	if err != nil {
		return domain.Collection{}, err
	}
	return s.repo.Create(ctx, workspaceID, title, slug, visibility)
}
func (s *Service) List(ctx context.Context, workspaceID uuid.UUID) ([]domain.Collection, error) {
	return s.repo.List(ctx, workspaceID)
}
func (s *Service) Get(ctx context.Context, workspaceID, collectionID uuid.UUID) (domain.Collection, error) {
	return s.repo.Get(ctx, workspaceID, collectionID)
}
func (s *Service) Update(ctx context.Context, workspaceID, collectionID uuid.UUID, title, slug, visibility string) (domain.Collection, error) {
	title, slug, visibility, err := validate(title, slug, visibility)
	if err != nil {
		return domain.Collection{}, err
	}
	return s.repo.Update(ctx, workspaceID, collectionID, title, slug, visibility)
}
func (s *Service) AddSection(ctx context.Context, workspaceID, collectionID uuid.UUID, title string) (domain.Section, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return domain.Section{}, fmt.Errorf("section title is required")
	}
	return s.repo.AddSection(ctx, workspaceID, collectionID, title)
}
func (s *Service) MoveSection(ctx context.Context, workspaceID, collectionID, sectionID uuid.UUID, direction int) error {
	if direction != -1 && direction != 1 {
		return fmt.Errorf("direction must be -1 or 1")
	}
	return s.repo.MoveSection(ctx, workspaceID, collectionID, sectionID, direction)
}
func (s *Service) AddItem(ctx context.Context, workspaceID, sectionID, objectID uuid.UUID, annotation string) (domain.Item, error) {
	return s.repo.AddItem(ctx, workspaceID, sectionID, objectID, strings.TrimSpace(annotation))
}
func (s *Service) MoveItem(ctx context.Context, workspaceID, itemID uuid.UUID, direction int) error {
	if direction != -1 && direction != 1 {
		return fmt.Errorf("direction must be -1 or 1")
	}
	return s.repo.MoveItem(ctx, workspaceID, itemID, direction)
}
func (s *Service) RemoveItem(ctx context.Context, workspaceID, itemID uuid.UUID) error {
	return s.repo.RemoveItem(ctx, workspaceID, itemID)
}

func validate(title, slug, visibility string) (string, string, string, error) {
	title = strings.TrimSpace(title)
	slug = strings.ToLower(strings.TrimSpace(slug))
	if title == "" {
		return "", "", "", fmt.Errorf("title is required")
	}
	if !slugPattern.MatchString(slug) {
		return "", "", "", fmt.Errorf("slug must contain lowercase letters, numbers and hyphens")
	}
	if visibility == "" {
		visibility = "private"
	}
	if visibility != "private" && visibility != "unlisted" && visibility != "public" {
		return "", "", "", fmt.Errorf("invalid visibility")
	}
	return title, slug, visibility, nil
}
