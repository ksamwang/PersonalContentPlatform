package application

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/google/uuid"
	contentdomain "github.com/ksamwang/PersonalContentPlatform/internal/content/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/inbox/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/inbox/ports"
	settingsports "github.com/ksamwang/PersonalContentPlatform/internal/settings/ports"
)

var (
	ErrNotPending = errors.New("inbox item is not pending")
	slugPattern   = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
)

type Service struct {
	repo      ports.Repository
	supported map[string]bool
	settings  settingsports.Repository
}

func New(repo ports.Repository, locales []string, settings settingsports.Repository) *Service {
	supported := make(map[string]bool, len(locales))
	for _, locale := range locales {
		supported[locale] = true
	}
	return &Service{repo: repo, supported: supported, settings: settings}
}

func (s *Service) Create(ctx context.Context, workspaceID, userID uuid.UUID, kind domain.Kind, rawText, sourceURL string) (domain.Item, error) {
	rawText = strings.TrimSpace(rawText)
	sourceURL = strings.TrimSpace(sourceURL)
	if !kind.Valid() {
		return domain.Item{}, fmt.Errorf("kind must be text or link")
	}
	if kind == domain.KindText && rawText == "" {
		return domain.Item{}, fmt.Errorf("text capture requires raw_text")
	}
	if kind == domain.KindLink {
		parsed, err := url.ParseRequestURI(sourceURL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return domain.Item{}, fmt.Errorf("link capture requires a valid http or https source_url")
		}
	}
	var source *string
	if sourceURL != "" {
		source = &sourceURL
	}
	return s.repo.Create(ctx, workspaceID, userID, kind, rawText, source)
}

func (s *Service) List(ctx context.Context, workspaceID uuid.UUID, state domain.State, limit int) ([]domain.Item, error) {
	if state != "" && state != domain.StatePending && state != domain.StateConverted && state != domain.StateArchived {
		return nil, fmt.Errorf("invalid inbox state")
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	return s.repo.List(ctx, workspaceID, state, limit)
}

func (s *Service) Archive(ctx context.Context, workspaceID, itemID uuid.UUID) error {
	return s.repo.Archive(ctx, workspaceID, itemID)
}

func (s *Service) Convert(ctx context.Context, workspaceID, userID, itemID uuid.UUID, kind contentdomain.Type, locale, slug, title string) (domain.Conversion, error) {
	if !kind.Valid() {
		return domain.Conversion{}, fmt.Errorf("unsupported content type")
	}
	if !s.localeSupported(ctx, workspaceID, locale) {
		return domain.Conversion{}, fmt.Errorf("unsupported locale")
	}
	slug = strings.ToLower(strings.TrimSpace(slug))
	if !slugPattern.MatchString(slug) {
		return domain.Conversion{}, fmt.Errorf("slug must contain lowercase letters, numbers and hyphens")
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return domain.Conversion{}, fmt.Errorf("title is required")
	}
	return s.repo.Convert(ctx, workspaceID, userID, itemID, kind, locale, slug, title)
}

func (s *Service) localeSupported(ctx context.Context, workspaceID uuid.UUID, locale string) bool {
	if s.settings != nil {
		if configured, err := s.settings.Get(ctx, workspaceID); err == nil {
			for _, candidate := range configured.Workspace.SupportedLocales {
				if candidate == locale {
					return true
				}
			}
			return false
		}
	}
	return s.supported[locale]
}
