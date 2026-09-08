package application

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ksamwang/PersonalContentPlatform/internal/settings/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/settings/ports"
)

type Service struct{ repo ports.Repository }

func New(repo ports.Repository) *Service { return &Service{repo: repo} }

func (s *Service) Get(ctx context.Context, ws uuid.UUID, canEdit bool) (domain.Settings, error) {
	general, err := s.repo.Get(ctx, ws)
	if err != nil {
		return domain.Settings{}, err
	}
	storage, err := s.repo.ActiveStorage(ctx, ws)
	if err != nil {
		storage = nil
	}
	ai, err := s.repo.AI(ctx, ws)
	if err != nil {
		return domain.Settings{}, err
	}
	return domain.Settings{General: general, Storage: storage, AI: ai, CanEdit: canEdit}, nil
}

func (s *Service) UpdateGeneral(ctx context.Context, ws, actor uuid.UUID, section string, raw json.RawMessage) (domain.GeneralSettings, error) {
	if err := validateSection(section, raw); err != nil {
		return domain.GeneralSettings{}, err
	}
	return s.repo.UpdateSection(ctx, ws, actor, section, raw)
}

func validateSection(section string, raw json.RawMessage) error {
	switch section {
	case "workspace":
		var v domain.WorkspaceSettings
		if err := json.Unmarshal(raw, &v); err != nil {
			return err
		}
		if strings.TrimSpace(v.Name) == "" || strings.TrimSpace(v.Slug) == "" || len(v.SupportedLocales) == 0 {
			return fmt.Errorf("name, slug and supported locales are required")
		}
		found := false
		seen := map[string]bool{}
		for _, locale := range v.SupportedLocales {
			if strings.TrimSpace(locale) == "" || seen[locale] {
				return fmt.Errorf("supported locales must be unique")
			}
			seen[locale] = true
			if locale == v.DefaultLocale {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("default locale must be supported")
		}
		if _, err := time.LoadLocation(v.Timezone); err != nil {
			return fmt.Errorf("invalid timezone")
		}
	case "site":
		var v domain.SiteSettings
		if err := json.Unmarshal(raw, &v); err != nil {
			return err
		}
		if v.PublicURL != "" {
			u, err := url.ParseRequestURI(v.PublicURL)
			if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
				return fmt.Errorf("public URL must be an absolute HTTP or HTTPS URL")
			}
		}
	case "auth":
		var v domain.AuthSettings
		if err := json.Unmarshal(raw, &v); err != nil {
			return err
		}
		if v.SessionTTLHours < 1 || v.SessionTTLHours > 8760 {
			return fmt.Errorf("session TTL must be between 1 and 8760 hours")
		}
		if !v.PasswordLoginEnabled && !v.PasskeyEnabled {
			return fmt.Errorf("at least one login method must remain enabled")
		}
	case "publication":
		var v domain.PublicationSettings
		if err := json.Unmarshal(raw, &v); err != nil {
			return err
		}
		if v.DefaultChannel != "website" && v.DefaultChannel != "rss" {
			return fmt.Errorf("unsupported publication channel")
		}
	default:
		return fmt.Errorf("unsupported settings section")
	}
	return nil
}

func (s *Service) SaveStorage(ctx context.Context, ws, actor uuid.UUID, v domain.StorageProfile) (domain.StorageProfile, error) {
	v.Name = strings.TrimSpace(v.Name)
	if v.Name == "" {
		v.Name = "Default"
	}
	if v.Provider != "filesystem" && v.Provider != "s3" && v.Provider != "r2" && v.Provider != "oss" {
		return v, fmt.Errorf("unsupported storage provider")
	}
	if v.Provider == "filesystem" && strings.TrimSpace(v.BasePath) == "" {
		return v, fmt.Errorf("filesystem base path is required")
	}
	if v.Provider != "filesystem" && (v.Endpoint == "" || v.Bucket == "") {
		return v, fmt.Errorf("endpoint and bucket are required")
	}
	return s.repo.SaveStorage(ctx, ws, actor, v)
}
func (s *Service) SaveAI(ctx context.Context, ws, actor uuid.UUID, v domain.AIConfig) (domain.AIConfig, error) {
	v.Provider = strings.TrimSpace(v.Provider)
	v.BaseURL = strings.TrimRight(strings.TrimSpace(v.BaseURL), "/")
	v.Model = strings.TrimSpace(v.Model)
	if v.Provider == "" {
		v.Provider = "openai-compatible"
	}
	if v.Provider != "openai-compatible" {
		return v, fmt.Errorf("unsupported AI provider")
	}
	if v.BaseURL != "" {
		u, err := url.ParseRequestURI(v.BaseURL)
		if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
			return v, fmt.Errorf("AI base URL must be an absolute HTTP or HTTPS URL")
		}
	}
	keys := make([]string, 0, len(v.PurposeModels))
	for k := range v.PurposeModels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return s.repo.SaveAI(ctx, ws, actor, v)
}
