package application

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ksamwang/PersonalContentPlatform/internal/settings/domain"
)

func (s *Service) ListAIModels(ctx context.Context, workspaceID uuid.UUID, provider, baseURL, apiKey string) ([]domain.AIModel, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	apiKey = strings.TrimSpace(apiKey)
	stored, err := s.repo.AI(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	provider = strings.TrimSpace(provider)
	if provider == "" {
		provider = stored.Provider
	}
	if provider != "openai-compatible" && provider != "anthropic-compatible" {
		return nil, fmt.Errorf("unsupported AI provider")
	}
	if baseURL == "" {
		baseURL = strings.TrimRight(strings.TrimSpace(stored.BaseURL), "/")
	}
	if apiKey == "" {
		apiKey = stored.APIKey
	}
	parsed, err := url.ParseRequestURI(baseURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, fmt.Errorf("AI base URL must be an absolute HTTP or HTTPS URL")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required to retrieve models")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/models", nil)
	if err != nil {
		return nil, err
	}
	if provider == "anthropic-compatible" {
		req.Header.Set("x-api-key", apiKey)
		req.Header.Set("anthropic-version", "2023-06-01")
	} else {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	res, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("model endpoint is unavailable: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode/100 != 2 {
		return nil, fmt.Errorf("model endpoint returned status %d", res.StatusCode)
	}
	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("model endpoint returned invalid JSON")
	}
	seen := map[string]bool{}
	models := make([]domain.AIModel, 0, len(payload.Data))
	for _, item := range payload.Data {
		modelID := strings.TrimSpace(item.ID)
		if modelID != "" && !seen[modelID] {
			seen[modelID] = true
			models = append(models, domain.AIModel{ID: modelID})
		}
	}
	sort.Slice(models, func(i, j int) bool { return models[i].ID < models[j].ID })
	if len(models) == 0 {
		return nil, fmt.Errorf("model endpoint returned no models")
	}
	return models, nil
}
