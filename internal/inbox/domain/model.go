package domain

import (
	"time"

	"github.com/google/uuid"
)

type Kind string

const (
	KindText Kind = "text"
	KindLink Kind = "link"
)

func (k Kind) Valid() bool { return k == KindText || k == KindLink }

type State string

const (
	StatePending   State = "pending"
	StateConverted State = "converted"
	StateArchived  State = "archived"
)

type Item struct {
	ID                 uuid.UUID  `json:"id"`
	WorkspaceID        uuid.UUID  `json:"workspace_id"`
	Kind               Kind       `json:"kind"`
	RawText            string     `json:"raw_text"`
	SourceURL          *string    `json:"source_url,omitempty"`
	State              State      `json:"state"`
	ConvertedContentID *uuid.UUID `json:"converted_content_id,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type Conversion struct {
	InboxID        uuid.UUID `json:"inbox_id"`
	ContentID      uuid.UUID `json:"content_id"`
	LocalizationID uuid.UUID `json:"localization_id"`
}
