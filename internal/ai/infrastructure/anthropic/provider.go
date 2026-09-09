package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ksamwang/PersonalContentPlatform/internal/ai/ports"
)

type Provider struct {
	baseURL, key, model string
	client              *http.Client
}

func New(baseURL, key, model string) *Provider {
	return &Provider{baseURL: strings.TrimRight(baseURL, "/"), key: key, model: model, client: &http.Client{Timeout: 90 * time.Second}}
}
func (p *Provider) Name() string { return "anthropic-compatible" }
func (p *Provider) Generate(ctx context.Context, in ports.Request) (ports.Response, error) {
	if p.baseURL == "" || p.key == "" || p.model == "" {
		return ports.Response{}, fmt.Errorf("AI provider is not configured")
	}
	model := p.model
	if in.Model != "" {
		model = in.Model
	}
	body, _ := json.Marshal(map[string]any{"model": model, "system": in.System, "messages": []map[string]string{{"role": "user", "content": in.User}}, "max_tokens": 1024})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return ports.Response{}, err
	}
	req.Header.Set("x-api-key", p.key)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")
	res, err := p.client.Do(req)
	if err != nil {
		return ports.Response{}, err
	}
	defer res.Body.Close()
	if res.StatusCode/100 != 2 {
		return ports.Response{}, fmt.Errorf("provider returned status %d", res.StatusCode)
	}
	var out struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			Input  int `json:"input_tokens"`
			Output int `json:"output_tokens"`
		} `json:"usage"`
	}
	if err = json.NewDecoder(res.Body).Decode(&out); err != nil {
		return ports.Response{}, err
	}
	for _, block := range out.Content {
		if block.Type == "text" && block.Text != "" {
			return ports.Response{Text: block.Text, InputTokens: out.Usage.Input, OutputTokens: out.Usage.Output}, nil
		}
	}
	return ports.Response{}, fmt.Errorf("provider returned no text content")
}
