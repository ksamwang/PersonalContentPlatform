package domain

import (
	"time"

	"github.com/google/uuid"
)

type Kind string

const (
	KindText  Kind = "text"
	KindLink  Kind = "link"
	KindImage Kind = "image"
	KindAudio Kind = "audio"
	KindFile  Kind = "file"
)

func (k Kind) Valid() bool {
	return k == KindText || k == KindLink || k == KindImage || k == KindAudio || k == KindFile
}

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
	AssetID            *uuid.UUID `json:"asset_id,omitempty"`
	Title              string     `json:"title"`
	ExtractedText      string     `json:"extracted_text"`
	CoverURL           *string    `json:"cover_url,omitempty"`
	ProcessingState    string     `json:"processing_state"`
	ProcessingError    string     `json:"processing_error"`
	DuplicateOf        *uuid.UUID `json:"duplicate_of,omitempty"`
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
