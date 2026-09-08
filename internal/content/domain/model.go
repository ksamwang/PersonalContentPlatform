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
