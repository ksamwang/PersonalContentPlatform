package domain

import (
	"github.com/google/uuid"
	"time"
)

type UploadIntent struct {
	ID, WorkspaceID            uuid.UUID
	StorageProfileID           *uuid.UUID
	StorageKey, Filename, MIME string
	ExpectedSize               int64
	ExpiresAt                  time.Time
}
type Asset struct {
	ID               uuid.UUID  `json:"id"`
	WorkspaceID      uuid.UUID  `json:"workspace_id"`
	BlobID           uuid.UUID  `json:"blob_id"`
	StorageProfileID *uuid.UUID `json:"storage_profile_id,omitempty"`
	Filename         string     `json:"filename"`
	MediaType        string     `json:"media_type"`
	State            string     `json:"state"`
	MIME             string     `json:"mime"`
	Size             int64      `json:"size"`
	SHA256           string     `json:"sha256"`
	CreatedAt        time.Time  `json:"created_at"`
}
type UploadPlan struct {
	UploadID  uuid.UUID         `json:"upload_id"`
	Method    string            `json:"method"`
	URL       string            `json:"url"`
	Headers   map[string]string `json:"headers"`
	ExpiresAt time.Time         `json:"expires_at"`
}
