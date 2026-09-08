package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type WebhookEndpoint struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	URL        string    `json:"url"`
	SecretRef  string    `json:"secret_ref"`
	Enabled    bool      `json:"enabled"`
	EventTypes []string  `json:"event_types"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type WebhookDelivery struct {
	AttemptID  uuid.UUID
	EndpointID uuid.UUID
	EventID    uuid.UUID
	EventType  string
	URL        string
	SecretRef  string
	AttemptNo  int
	Payload    json.RawMessage
	CreatedAt  time.Time
}

type DeliveryResult struct {
	Status          string
	ResponseCode    *int
	ResponseExcerpt string
}
