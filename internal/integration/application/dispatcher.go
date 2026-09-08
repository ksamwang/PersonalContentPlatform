package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/ksamwang/PersonalContentPlatform/internal/integration/ports"
)

type WebhookDispatcher struct {
	repository  ports.Repository
	sender      ports.WebhookSender
	maxAttempts int
}

func NewWebhookDispatcher(repository ports.Repository, sender ports.WebhookSender, maxAttempts int) *WebhookDispatcher {
	if maxAttempts < 1 {
		maxAttempts = 8
	}
	return &WebhookDispatcher{repository: repository, sender: sender, maxAttempts: maxAttempts}
}

func (d *WebhookDispatcher) Run(ctx context.Context) error {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		processed, err := d.processOne(ctx)
		if err != nil {
			slog.Warn("dispatch webhook", "error", err)
		}
		if processed {
			continue
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (d *WebhookDispatcher) processOne(ctx context.Context) (bool, error) {
	delivery, err := d.repository.ClaimWebhookDelivery(ctx, d.maxAttempts)
	if err != nil || delivery == nil {
		return false, err
	}
	result := d.sender.Send(ctx, *delivery)
	return true, d.repository.CompleteWebhookDelivery(ctx, delivery.AttemptID, result)
}
