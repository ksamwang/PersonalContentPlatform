package domain

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

type Type string

const (
	TypeArticle Type = "article"
	TypeNote    Type = "note"
	TypePage    Type = "page"
)

func (t Type) Valid() bool { return t == TypeArticle || t == TypeNote || t == TypePage }

type Visibility string

const (
	Private  Visibility = "private"
	Unlisted Visibility = "unlisted"
	Public   Visibility = "public"
)

type Content struct {
	ID            uuid.UUID      `json:"id"`
	WorkspaceID   uuid.UUID      `json:"workspace_id"`
	Type          Type           `json:"type"`
	DefaultLocale string         `json:"default_locale"`
	Visibility    Visibility     `json:"visibility"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	Localizations []Localization `json:"localizations,omitempty"`
}

type ListFilter struct {
	Locale     string
	State      string
	Type       Type
	Visibility Visibility
	Tag        string
	Limit      int
}
type Localization struct {
	ID                uuid.UUID `json:"id"`
	ContentID         uuid.UUID `json:"content_id"`
	Locale            string    `json:"locale"`
	State             string    `json:"state"`
	Slug              string    `json:"slug"`
	TranslationStatus string    `json:"translation_status"`
	CurrentRevision   *Revision `json:"current_revision,omitempty"`
	UpdatedAt         time.Time `json:"updated_at"`
}
type Revision struct {
	ID             uuid.UUID       `json:"id"`
	LocalizationID uuid.UUID       `json:"localization_id"`
	Seq            int             `json:"seq"`
	SchemaVersion  int             `json:"schema_version"`
	Title          string          `json:"title"`
	Summary        string          `json:"summary"`
	Body           json.RawMessage `json:"body"`
	Metadata       json.RawMessage `json:"metadata"`
	ContentHash    string          `json:"content_hash"`
	CreatedAt      time.Time       `json:"created_at"`
}
type Draft struct {
	LocalizationID uuid.UUID       `json:"localization_id"`
	Version        int             `json:"version"`
	Title          string          `json:"title"`
	Summary        string          `json:"summary"`
	Body           json.RawMessage `json:"body"`
	Metadata       json.RawMessage `json:"metadata"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

type Preview struct {
	Type      Type            `json:"type"`
	Locale    string          `json:"locale"`
	Title     string          `json:"title"`
	Summary   string          `json:"summary"`
	Body      json.RawMessage `json:"-"`
	Metadata  json.RawMessage `json:"metadata"`
	HTML      string          `json:"html"`
	ExpiresAt time.Time       `json:"expires_at"`
}

type PreviewToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type ReadinessIssue struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

type Readiness struct {
	Ready  bool             `json:"ready"`
	Issues []ReadinessIssue `json:"issues"`
}

type ReadinessSource struct {
	Type     Type
	Slug     string
	Title    string
	Summary  string
	Body     json.RawMessage
	Metadata json.RawMessage
}
