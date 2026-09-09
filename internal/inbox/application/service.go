package application

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"

	"github.com/google/uuid"
	assetdomain "github.com/ksamwang/PersonalContentPlatform/internal/asset/domain"
	contentdomain "github.com/ksamwang/PersonalContentPlatform/internal/content/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/inbox/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/inbox/ports"
	"github.com/ksamwang/PersonalContentPlatform/internal/inbox/processing"
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
	assets    interface {
		Open(context.Context, uuid.UUID, uuid.UUID) (io.ReadCloser, assetdomain.StoredObject, error)
	}
	links *processing.LinkFetcher
	media *processing.MediaClient
}

func New(repo ports.Repository, locales []string, settings settingsports.Repository, assets ...interface {
	Open(context.Context, uuid.UUID, uuid.UUID) (io.ReadCloser, assetdomain.StoredObject, error)
}) *Service {
	supported := make(map[string]bool, len(locales))
	for _, locale := range locales {
		supported[locale] = true
	}
	var opener interface {
		Open(context.Context, uuid.UUID, uuid.UUID) (io.ReadCloser, assetdomain.StoredObject, error)
	}
	if len(assets) > 0 {
		opener = assets[0]
	}
	return &Service{repo: repo, supported: supported, settings: settings, assets: opener, links: processing.NewLinkFetcher(), media: processing.NewMediaClient()}
}

func (s *Service) Create(ctx context.Context, workspaceID, userID uuid.UUID, kind domain.Kind, rawText, sourceURL, assetID string) (domain.Item, error) {
	rawText = strings.TrimSpace(rawText)
	sourceURL = strings.TrimSpace(sourceURL)
	if !kind.Valid() {
		return domain.Item{}, fmt.Errorf("unsupported inbox kind")
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
	var asset *uuid.UUID
	if kind == domain.KindImage || kind == domain.KindAudio || kind == domain.KindFile {
		parsed, err := uuid.Parse(assetID)
		if err != nil {
			return domain.Item{}, fmt.Errorf("asset_id is required for file capture")
		}
		asset = &parsed
	}
	return s.repo.Create(ctx, workspaceID, userID, kind, rawText, source, asset)
}
func (s *Service) Process(ctx context.Context, workspaceID, itemID uuid.UUID) (item domain.Item, err error) {
	item, err = s.repo.Get(ctx, workspaceID, itemID)
	if err != nil {
		return item, err
	}
	if item.State != domain.StatePending {
		return item, ErrNotPending
	}
	if err = s.repo.BeginProcessing(ctx, workspaceID, itemID); err != nil {
		return item, err
	}
	defer func() {
		if err != nil {
			_ = s.repo.FailProcessing(ctx, workspaceID, itemID, err.Error())
		}
	}()
	var title, text string
	var cover *string
	switch item.Kind {
	case domain.KindLink:
		if item.SourceURL == nil {
			return item, fmt.Errorf("link has no source URL")
		}
		result, e := s.links.Fetch(ctx, *item.SourceURL)
		if e != nil {
			return item, e
		}
		title, text, cover = result.Title, result.Text, result.Cover
	case domain.KindImage, domain.KindAudio:
		if item.AssetID == nil || s.assets == nil {
			return item, fmt.Errorf("captured asset is unavailable")
		}
		reader, object, e := s.assets.Open(ctx, workspaceID, *item.AssetID)
		if e != nil {
			return item, e
		}
		defer reader.Close()
		purpose := "ocr"
		if item.Kind == domain.KindAudio {
			purpose = "transcription"
		}
		cfg, e := s.settings.Media(ctx, workspaceID, purpose)
		if e != nil {
			return item, e
		}
		if cfg.BaseURL == "" || cfg.APIKey == "" || cfg.Model == "" {
			return item, fmt.Errorf("%s provider configuration is incomplete", purpose)
		}
		if purpose == "ocr" {
			text, e = s.media.OCR(ctx, cfg, reader, object.MIME)
		} else {
			text, e = s.media.Transcribe(ctx, cfg, reader, item.RawText)
		}
		if e != nil {
			return item, e
		}
	case domain.KindFile:
		return item, fmt.Errorf("this file type has no automatic processor")
	default:
		text = item.RawText
	}
	duplicate, e := s.repo.FindDuplicate(ctx, workspaceID, itemID, item.SourceURL, text)
	if e != nil {
		return item, e
	}
	return s.repo.CompleteProcessing(ctx, workspaceID, itemID, title, text, cover, duplicate)
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
