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
func (s *Service) CreateLocalization(ctx context.Context, workspaceID, userID, contentID uuid.UUID, locale, slug, sourceLocale string) (domain.Content, error) {
	supported := s.supported
	if s.settings != nil {
		if configured, err := s.settings.Get(ctx, workspaceID); err == nil {
			supported = map[string]bool{}
			for _, item := range configured.Workspace.SupportedLocales {
				supported[item] = true
			}
		}
	}
	if !supported[locale] || locale == sourceLocale {
		return domain.Content{}, fmt.Errorf("unsupported target locale")
	}
	slug = strings.ToLower(strings.TrimSpace(slug))
	if !slugPattern.MatchString(slug) {
		return domain.Content{}, fmt.Errorf("slug must contain lowercase letters, numbers and hyphens")
	}
	return s.repo.CreateLocalization(ctx, workspaceID, userID, contentID, locale, slug, sourceLocale)
}
func (s *Service) List(ctx context.Context, workspaceID uuid.UUID, filter domain.ListFilter) ([]domain.Content, error) {
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 30
	}
	return s.repo.List(ctx, workspaceID, filter)
}
func (s *Service) Get(ctx context.Context, workspaceID, contentID uuid.UUID) (domain.Content, error) {
	return s.repo.Get(ctx, workspaceID, contentID)
}
func (s *Service) Archive(ctx context.Context, workspaceID, contentID uuid.UUID) error {
	return s.repo.Archive(ctx, workspaceID, contentID)
}
func (s *Service) Restore(ctx context.Context, workspaceID, contentID uuid.UUID) error {
	return s.repo.Restore(ctx, workspaceID, contentID)
}
func (s *Service) SoftDelete(ctx context.Context, workspaceID, contentID uuid.UUID) error {
	return s.repo.SoftDelete(ctx, workspaceID, contentID)
}
func (s *Service) UpdateProperties(ctx context.Context, workspaceID, contentID, localizationID uuid.UUID, slug string, visibility domain.Visibility) error {
	slug = strings.ToLower(strings.TrimSpace(slug))
	if !slugPattern.MatchString(slug) {
		return fmt.Errorf("slug must contain lowercase letters, numbers and hyphens")
	}
	if visibility != domain.Private && visibility != domain.Unlisted && visibility != domain.Public {
		return fmt.Errorf("unsupported visibility")
	}
	return s.repo.UpdateProperties(ctx, workspaceID, contentID, localizationID, slug, visibility)
}
func (s *Service) Readiness(ctx context.Context, workspaceID, localizationID uuid.UUID) (domain.Readiness, error) {
	source, err := s.repo.ReadinessSource(ctx, workspaceID, localizationID)
	if err != nil {
		return domain.Readiness{}, err
	}
	issues := make([]domain.ReadinessIssue, 0)
	if strings.TrimSpace(source.Title) == "" {
		issues = append(issues, domain.ReadinessIssue{Code: "title_required", Message: "请填写标题", Severity: "error"})
	}
	if !slugPattern.MatchString(source.Slug) {
		issues = append(issues, domain.ReadinessIssue{Code: "slug_invalid", Message: "固定链接格式无效", Severity: "error"})
	}
	if source.Type == domain.TypeArticle && strings.TrimSpace(source.Summary) == "" {
		issues = append(issues, domain.ReadinessIssue{Code: "summary_recommended", Message: "文章建议填写摘要", Severity: "warning"})
	}
	var metadata map[string]any
	_ = json.Unmarshal(source.Metadata, &metadata)
	cover, _ := metadata["cover_asset_id"].(string)
	if source.Type == domain.TypeArticle && strings.TrimSpace(cover) == "" {
		issues = append(issues, domain.ReadinessIssue{Code: "cover_recommended", Message: "文章建议设置封面", Severity: "warning"})
	}
	if hasImageWithoutAlt(source.Body) {
		issues = append(issues, domain.ReadinessIssue{Code: "image_alt_recommended", Message: "正文中有图片缺少替代文字", Severity: "warning"})
	}
	ready := true
	for _, issue := range issues {
		if issue.Severity == "error" {
			ready = false
		}
	}
	return domain.Readiness{Ready: ready, Issues: issues}, nil
}
func hasImageWithoutAlt(raw json.RawMessage) bool {
	var value any
	if json.Unmarshal(raw, &value) != nil {
		return false
	}
	var visit func(any) bool
	visit = func(current any) bool {
		switch node := current.(type) {
		case map[string]any:
			if node["type"] == "image" {
				attrs, _ := node["attrs"].(map[string]any)
				alt, _ := attrs["alt"].(string)
				return strings.TrimSpace(alt) == ""
			}
			for _, child := range node {
				if visit(child) {
					return true
				}
			}
		case []any:
			for _, child := range node {
				if visit(child) {
					return true
				}
			}
		}
		return false
	}
	return visit(value)
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
	check, err := s.Readiness(ctx, workspaceID, localizationID)
	if err != nil {
		return err
	}
	if !check.Ready {
		return fmt.Errorf("localization has unresolved readiness errors")
	}
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
