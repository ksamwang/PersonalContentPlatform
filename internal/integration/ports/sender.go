package ports

import (
	"context"

	"github.com/ksamwang/PersonalContentPlatform/internal/integration/domain"
)

type WebhookSender interface {
	Send(context.Context, domain.WebhookDelivery) domain.DeliveryResult
}
