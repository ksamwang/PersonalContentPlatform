package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/ksamwang/PersonalContentPlatform/internal/content/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/content/ports"
	"regexp"
	"strings"
)

var ErrConflict = errors.New("version conflict")
var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type Service struct {
	repo      ports.Repository
	supported map[string]bool
}

func New(repo ports.Repository, locales []string) *Service {
	supported := map[string]bool{}
	for _, v := range locales {
		supported[v] = true
	}
	return &Service{repo: repo, supported: supported}
}
func (s *Service) Create(ctx context.Context, workspaceID, userID uuid.UUID, kind domain.Type, locale, slug, title string) (domain.Content, error) {
	if !kind.Valid() {
		return domain.Content{}, fmt.Errorf("unsupported content type")
	}
	if !s.supported[locale] {
		return domain.Content{}, fmt.Errorf("unsupported locale")
	}
	slug = strings.ToLower(strings.TrimSpace(slug))
	if !slugPattern.MatchString(slug) {
		return domain.Content{}, fmt.Errorf("slug must contain lowercase letters, numbers and hyphens")
	}
	return s.repo.Create(ctx, workspaceID, userID, kind, locale, slug, strings.TrimSpace(title))
}
func (s *Service) List(ctx context.Context, workspaceID uuid.UUID, locale, state string, limit int) ([]domain.Content, error) {
	if limit < 1 || limit > 100 {
		limit = 30
	}
	return s.repo.List(ctx, workspaceID, locale, state, limit)
}
func (s *Service) Get(ctx context.Context, workspaceID, contentID uuid.UUID) (domain.Content, error) {
	return s.repo.Get(ctx, workspaceID, contentID)
}
func (s *Service) GetDraft(ctx context.Context, workspaceID, localizationID uuid.UUID) (domain.Draft, error) {
	return s.repo.GetDraft(ctx, workspaceID, localizationID)
}
func (s *Service) SaveDraft(ctx context.Context, workspaceID, userID, localizationID uuid.UUID, expected int, title, summary string, body, metadata json.RawMessage) (domain.Draft, error) {
	if !json.Valid(body) || !json.Valid(metadata) {
		return domain.Draft{}, fmt.Errorf("body and metadata must be valid JSON")
	}
	return s.repo.SaveDraft(ctx, workspaceID, userID, localizationID, expected, title, summary, body, metadata)
}
func (s *Service) Seal(ctx context.Context, workspaceID, userID, localizationID uuid.UUID, expected int) (domain.Revision, error) {
	return s.repo.SealRevision(ctx, workspaceID, userID, localizationID, expected)
}
func (s *Service) MarkReady(ctx context.Context, workspaceID, localizationID uuid.UUID) error {
	return s.repo.MarkReady(ctx, workspaceID, localizationID)
}
func (s *Service) Publish(ctx context.Context, workspaceID, userID, contentID uuid.UUID, locale string, targetID uuid.UUID) (uuid.UUID, error) {
	return s.repo.Publish(ctx, workspaceID, userID, contentID, locale, targetID)
}
func (s *Service) Search(ctx context.Context, workspaceID uuid.UUID, locale, query string, limit int) ([]domain.Content, error) {
	if limit < 1 || limit > 100 {
		limit = 30
	}
	return s.repo.Search(ctx, workspaceID, locale, strings.TrimSpace(query), limit)
}
