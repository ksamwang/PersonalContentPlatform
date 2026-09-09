package domain

import (
	"time"

	"github.com/google/uuid"
)

type WorkspaceSettings struct {
	Name             string   `json:"name"`
	Slug             string   `json:"slug"`
	DefaultLocale    string   `json:"default_locale"`
	SupportedLocales []string `json:"supported_locales"`
	Timezone         string   `json:"timezone"`
}

type SiteSettings struct {
	Name        string `json:"name"`
	PublicURL   string `json:"public_url"`
	Description string `json:"description"`
	About       string `json:"about"`
	Footer      string `json:"footer"`
	RSSEnabled  bool   `json:"rss_enabled"`
}

type AuthSettings struct {
	PasswordLoginEnabled bool `json:"password_login_enabled"`
	PasskeyEnabled       bool `json:"passkey_enabled"`
	SessionTTLHours      int  `json:"session_ttl_hours"`
}

type PublicationSettings struct {
	DefaultChannel string `json:"default_channel"`
	AutoPublish    bool   `json:"auto_publish"`
}

type GeneralSettings struct {
	Workspace   WorkspaceSettings   `json:"workspace"`
	Site        SiteSettings        `json:"site"`
	Auth        AuthSettings        `json:"auth"`
	Publication PublicationSettings `json:"publication"`
}

type StorageProfile struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Provider      string    `json:"provider"`
	Endpoint      string    `json:"endpoint"`
	Region        string    `json:"region"`
	Bucket        string    `json:"bucket"`
	AccessKey     string    `json:"-"`
	SecretKey     string    `json:"-"`
	BasePath      string    `json:"base_path"`
	Active        bool      `json:"active"`
	AccessKeyMask string    `json:"access_key_mask"`
	SecretKeySet  bool      `json:"secret_key_set"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type AIConfig struct {
	Provider      string            `json:"provider"`
	BaseURL       string            `json:"base_url"`
	APIKey        string            `json:"-"`
	APIKeyMask    string            `json:"api_key_mask"`
	APIKeySet     bool              `json:"api_key_set"`
	Model         string            `json:"model"`
	PurposeModels map[string]string `json:"purpose_models"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

type AIModel struct {
	ID string `json:"id"`
}
type EmbeddingConfig struct {
	Provider   string    `json:"provider"`
	BaseURL    string    `json:"base_url"`
	APIKey     string    `json:"-"`
	APIKeyMask string    `json:"api_key_mask"`
	APIKeySet  bool      `json:"api_key_set"`
	Model      string    `json:"model"`
	Dimensions int       `json:"dimensions"`
	UpdatedAt  time.Time `json:"updated_at"`
}
type MediaConfig struct {
	Purpose    string    `json:"purpose"`
	Provider   string    `json:"provider"`
	BaseURL    string    `json:"base_url"`
	APIKey     string    `json:"-"`
	APIKeyMask string    `json:"api_key_mask"`
	APIKeySet  bool      `json:"api_key_set"`
	Model      string    `json:"model"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Settings struct {
	General   GeneralSettings `json:"general"`
	Storage   *StorageProfile `json:"storage"`
	AI        AIConfig        `json:"ai"`
	Embedding EmbeddingConfig `json:"embedding"`
	Media     []MediaConfig   `json:"media"`
	CanEdit   bool            `json:"can_edit"`
}
