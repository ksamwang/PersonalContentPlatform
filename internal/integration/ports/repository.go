package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/ksamwang/PersonalContentPlatform/internal/integration/domain"
)

type Repository interface {
	BuildManifest(context.Context, uuid.UUID) (domain.Manifest, error)
	StartExport(context.Context, uuid.UUID, uuid.UUID) (uuid.UUID, error)
	CompleteExport(context.Context, uuid.UUID, domain.Manifest, error) error

	CreateWebhook(context.Context, uuid.UUID, domain.WebhookEndpoint) (domain.WebhookEndpoint, error)
	ListWebhooks(context.Context, uuid.UUID) ([]domain.WebhookEndpoint, error)
	SetWebhookEnabled(context.Context, uuid.UUID, uuid.UUID, bool) error
	DeleteWebhook(context.Context, uuid.UUID, uuid.UUID) error
	ClaimWebhookDelivery(context.Context, int) (*domain.WebhookDelivery, error)
	CompleteWebhookDelivery(context.Context, uuid.UUID, domain.DeliveryResult) error
}
