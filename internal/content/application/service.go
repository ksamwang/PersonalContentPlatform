package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/ksamwang/PersonalContentPlatform/internal/content/document"
	"github.com/ksamwang/PersonalContentPlatform/internal/content/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/content/ports"
	settingsports "github.com/ksamwang/PersonalContentPlatform/internal/settings/ports"
	"regexp"
	"strings"
	"time"
)

var ErrConflict = errors.New("version conflict")
var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type Service struct {
	repo      ports.Repository
	supported map[string]bool
	settings  settingsports.Repository
}

func New(repo ports.Repository, locales []string, settings settingsports.Repository) *Service {
	supported := map[string]bool{}
	for _, v := range locales {
		supported[v] = true
	}
	return &Service{repo: repo, supported: supported, settings: settings}
}
func (s *Service) Create(ctx context.Context, workspaceID, userID uuid.UUID, kind domain.Type, locale, slug, title string) (domain.Content, error) {
	if !kind.Valid() {
		return domain.Content{}, fmt.Errorf("unsupported content type")
	}
	supported := s.supported
	if s.settings != nil {
		if configured, err := s.settings.Get(ctx, workspaceID); err == nil {
			supported = map[string]bool{}
			for _, item := range configured.Workspace.SupportedLocales {
				supported[item] = true
			}
		}
	}
	if !supported[locale] {
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
	channel := "website"
	if s.settings != nil {
		if configured, err := s.settings.Get(ctx, workspaceID); err == nil && configured.Publication.DefaultChannel != "" {
			channel = configured.Publication.DefaultChannel
		}
	}
	return s.repo.Publish(ctx, workspaceID, userID, contentID, locale, targetID, channel)
}
func (s *Service) Search(ctx context.Context, workspaceID uuid.UUID, locale, query string, limit int) ([]domain.Content, error) {
	if limit < 1 || limit > 100 {
		limit = 30
	}
	return s.repo.Search(ctx, workspaceID, locale, strings.TrimSpace(query), limit)
}

func (s *Service) ListRevisions(ctx context.Context, workspaceID, localizationID uuid.UUID, limit int) ([]domain.Revision, error) {
	if limit < 1 || limit > 100 {
		limit = 30
	}
	return s.repo.ListRevisions(ctx, workspaceID, localizationID, limit)
}
func (s *Service) GetRevision(ctx context.Context, workspaceID, localizationID, revisionID uuid.UUID) (domain.Revision, error) {
	return s.repo.GetRevision(ctx, workspaceID, localizationID, revisionID)
}
func (s *Service) RestoreRevision(ctx context.Context, workspaceID, userID, localizationID, revisionID uuid.UUID, expectedVersion int) (domain.Draft, error) {
	if expectedVersion < 1 {
		return domain.Draft{}, fmt.Errorf("draft version is required")
	}
	return s.repo.RestoreRevision(ctx, workspaceID, userID, localizationID, revisionID, expectedVersion)
}
func (s *Service) CreatePreview(ctx context.Context, workspaceID, userID, localizationID uuid.UUID) (domain.PreviewToken, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return domain.PreviewToken{}, err
	}
	hash := sha256.Sum256(raw)
	expiresAt := time.Now().UTC().Add(30 * time.Minute)
	if err := s.repo.CreatePreview(ctx, workspaceID, userID, localizationID, hash[:], expiresAt); err != nil {
		return domain.PreviewToken{}, err
	}
	return domain.PreviewToken{Token: base64.RawURLEncoding.EncodeToString(raw), ExpiresAt: expiresAt}, nil
}
func (s *Service) GetPreview(ctx context.Context, token string) (domain.Preview, error) {
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(raw) != 32 {
		return domain.Preview{}, fmt.Errorf("invalid preview token")
	}
	hash := sha256.Sum256(raw)
	preview, err := s.repo.GetPreview(ctx, hash[:])
	if err != nil {
		return domain.Preview{}, err
	}
	preview.HTML = document.HTML(preview.Body)
	return preview, nil
}
