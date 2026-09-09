package domain

import (
	"github.com/google/uuid"
	"time"
)

type Entity struct {
	ID            uuid.UUID `json:"id"`
	Type          string    `json:"type"`
	CanonicalName string    `json:"canonical_name"`
	Description   string    `json:"description"`
	Aliases       []Alias   `json:"aliases,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}
type Alias struct {
	ID     uuid.UUID `json:"id"`
	Value  string    `json:"value"`
	Locale string    `json:"locale"`
}
type Relation struct {
	ID           uuid.UUID `json:"id"`
	SourceID     uuid.UUID `json:"source_id"`
	TargetID     uuid.UUID `json:"target_id"`
	PredicateKey string    `json:"predicate"`
	Confirmed    bool      `json:"confirmed"`
	CreatedAt    time.Time `json:"created_at"`
}
type Mention struct {
	ID           uuid.UUID `json:"id"`
	RevisionID   uuid.UUID `json:"revision_id"`
	EntityID     uuid.UUID `json:"entity_id"`
	EntityName   string    `json:"entity_name"`
	ContentID    uuid.UUID `json:"content_id"`
	ContentTitle string    `json:"content_title"`
	Locale       string    `json:"locale"`
	Confidence   float32   `json:"confidence"`
	Confirmed    bool      `json:"confirmed"`
	CreatedAt    time.Time `json:"created_at"`
}
