package application

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/ksamwang/PersonalContentPlatform/internal/integration/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/integration/ports"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/id"
)

type WebhookService struct {
	repository ports.Repository
}

func NewWebhookService(repository ports.Repository) *WebhookService {
	return &WebhookService{repository: repository}
}

func (s *WebhookService) Create(ctx context.Context, workspaceID uuid.UUID, name, endpointURL, secretRef string, eventTypes []string) (domain.WebhookEndpoint, error) {
	name = strings.TrimSpace(name)
	secretRef = strings.TrimSpace(secretRef)
	parsedURL, err := url.ParseRequestURI(strings.TrimSpace(endpointURL))
	if err != nil || parsedURL.Host == "" || (parsedURL.Scheme != "https" && parsedURL.Scheme != "http") {
		return domain.WebhookEndpoint{}, fmt.Errorf("webhook URL must be an absolute HTTP or HTTPS URL")
	}
	if name == "" || secretRef == "" {
		return domain.WebhookEndpoint{}, fmt.Errorf("name and secret_ref are required")
	}
	eventTypes = normalizeEventTypes(eventTypes)
	return s.repository.CreateWebhook(ctx, workspaceID, domain.WebhookEndpoint{
		ID:         id.New(),
		Name:       name,
		URL:        parsedURL.String(),
		SecretRef:  secretRef,
		Enabled:    true,
		EventTypes: eventTypes,
	})
}

func (s *WebhookService) List(ctx context.Context, workspaceID uuid.UUID) ([]domain.WebhookEndpoint, error) {
	return s.repository.ListWebhooks(ctx, workspaceID)
}

func (s *WebhookService) SetEnabled(ctx context.Context, workspaceID, endpointID uuid.UUID, enabled bool) error {
	return s.repository.SetWebhookEnabled(ctx, workspaceID, endpointID, enabled)
}

func (s *WebhookService) Delete(ctx context.Context, workspaceID, endpointID uuid.UUID) error {
	return s.repository.DeleteWebhook(ctx, workspaceID, endpointID)
}

func normalizeEventTypes(values []string) []string {
	unique := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			unique[value] = struct{}{}
		}
	}
	result := make([]string, 0, len(unique))
	for value := range unique {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
