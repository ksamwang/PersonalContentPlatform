package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Manifest struct {
	SchemaVersion int            `json:"schema_version"`
	ExportedAt    time.Time      `json:"exported_at"`
	Workspace     Workspace      `json:"workspace"`
	Contents      []Content      `json:"contents"`
	Localizations []Localization `json:"localizations"`
	Revisions     []Revision     `json:"revisions"`
	Drafts        []Draft        `json:"drafts"`
	Assets        []Asset        `json:"assets"`
	AssetVariants []AssetVariant `json:"asset_variants"`
	AssetUsages   []AssetUsage   `json:"asset_usages"`
	Entities      []Entity       `json:"entities"`
	EntityAliases []EntityAlias  `json:"entity_aliases"`
	Predicates    []Predicate    `json:"predicates"`
	Relations     []Relation     `json:"relations"`
	Publications  []Publication  `json:"publications"`
}

type Workspace struct {
	ID       uuid.UUID       `json:"id" db:"id"`
	Slug     string          `json:"slug" db:"slug"`
	Name     string          `json:"name" db:"name"`
	Settings json.RawMessage `json:"settings" db:"settings"`
}

type Content struct {
	ID            uuid.UUID `json:"id" db:"id"`
	Type          string    `json:"type" db:"type"`
	DefaultLocale string    `json:"default_locale" db:"default_locale"`
	Visibility    string    `json:"visibility" db:"visibility"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type Localization struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	ContentID          uuid.UUID  `json:"content_id" db:"content_id"`
	Locale             string     `json:"locale" db:"locale"`
	State              string     `json:"state" db:"state"`
	Slug               string     `json:"slug" db:"slug"`
	CurrentRevisionID  *uuid.UUID `json:"current_revision_id,omitempty" db:"current_revision_id"`
	TranslationStatus  string     `json:"translation_status" db:"translation_status"`
	SourceLocale       *string    `json:"source_locale,omitempty" db:"source_locale"`
	SourceRevisionID   *uuid.UUID `json:"source_revision_id,omitempty" db:"source_revision_id"`
	TranslatedFromHash *string    `json:"translated_from_hash,omitempty" db:"translated_from_hash"`
}

type Revision struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	LocalizationID uuid.UUID       `json:"localization_id" db:"localization_id"`
	Sequence       int             `json:"sequence" db:"sequence"`
	SchemaVersion  int             `json:"schema_version" db:"schema_version"`
	Title          string          `json:"title" db:"title"`
	Summary        string          `json:"summary" db:"summary"`
	Body           json.RawMessage `json:"body" db:"body"`
	Metadata       json.RawMessage `json:"metadata" db:"metadata"`
	ContentHash    string          `json:"content_hash" db:"content_hash"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
}

type Draft struct {
	LocalizationID uuid.UUID       `json:"localization_id" db:"localization_id"`
	Version        int             `json:"version" db:"version"`
	Title          string          `json:"title" db:"title"`
	Summary        string          `json:"summary" db:"summary"`
	Body           json.RawMessage `json:"body" db:"body"`
	Metadata       json.RawMessage `json:"metadata" db:"metadata"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
}

type Asset struct {
	ID         uuid.UUID       `json:"id" db:"id"`
	BlobID     uuid.UUID       `json:"blob_id" db:"blob_id"`
	Filename   string          `json:"filename" db:"filename"`
	MediaType  string          `json:"media_type" db:"media_type"`
	State      string          `json:"state" db:"state"`
	MIME       string          `json:"mime" db:"mime"`
	Size       int64           `json:"size" db:"size"`
	SHA256     string          `json:"sha256" db:"sha256"`
	StorageKey string          `json:"storage_key" db:"storage_key"`
	Metadata   json.RawMessage `json:"metadata" db:"metadata"`
}

type AssetVariant struct {
	ID         uuid.UUID `json:"id" db:"id"`
	AssetID    uuid.UUID `json:"asset_id" db:"asset_id"`
	Recipe     string    `json:"recipe" db:"recipe"`
	BlobID     uuid.UUID `json:"blob_id" db:"blob_id"`
	MIME       string    `json:"mime" db:"mime"`
	Size       int64     `json:"size" db:"size"`
	SHA256     string    `json:"sha256" db:"sha256"`
	StorageKey string    `json:"storage_key" db:"storage_key"`
	Width      *int      `json:"width,omitempty" db:"width"`
	Height     *int      `json:"height,omitempty" db:"height"`
}

type AssetUsage struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	AssetID       uuid.UUID       `json:"asset_id" db:"asset_id"`
	OwnerObjectID uuid.UUID       `json:"owner_object_id" db:"owner_object_id"`
	Role          string          `json:"role" db:"role"`
	Locator       json.RawMessage `json:"locator" db:"locator"`
}

type Entity struct {
	ID            uuid.UUID `json:"id" db:"id"`
	Type          string    `json:"type" db:"type"`
	CanonicalName string    `json:"canonical_name" db:"canonical_name"`
	Description   string    `json:"description" db:"description"`
}

type EntityAlias struct {
	ID       uuid.UUID `json:"id" db:"id"`
	EntityID uuid.UUID `json:"entity_id" db:"entity_id"`
	Alias    string    `json:"alias" db:"alias"`
	Locale   string    `json:"locale" db:"locale"`
}

type Predicate struct {
	ID      uuid.UUID `json:"id" db:"id"`
	Key     string    `json:"key" db:"key"`
	LabelZH string    `json:"label_zh" db:"label_zh"`
	LabelEN string    `json:"label_en" db:"label_en"`
}

type Relation struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	SourceID    uuid.UUID       `json:"source_id" db:"source_id"`
	TargetID    uuid.UUID       `json:"target_id" db:"target_id"`
	PredicateID uuid.UUID       `json:"predicate_id" db:"predicate_id"`
	Confirmed   bool            `json:"confirmed" db:"confirmed"`
	Provenance  json.RawMessage `json:"provenance" db:"provenance"`
}

type Publication struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	ContentID   uuid.UUID  `json:"content_id" db:"content_id"`
	Locale      string     `json:"locale" db:"locale"`
	RevisionID  uuid.UUID  `json:"revision_id" db:"revision_id"`
	Channel     string     `json:"channel" db:"channel"`
	TargetName  string     `json:"target_name" db:"target_name"`
	State       string     `json:"state" db:"state"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty" db:"scheduled_at"`
	PublishedAt *time.Time `json:"published_at,omitempty" db:"published_at"`
}
