package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ksamwang/PersonalContentPlatform/internal/ai/ports"
	"net/http"
	"time"
)

type Provider struct {
	baseURL, key, model string
	client              *http.Client
}

func New(baseURL, key, model string) *Provider {
	return &Provider{baseURL: baseURL, key: key, model: model, client: &http.Client{Timeout: 90 * time.Second}}
}
func (p *Provider) Name() string { return "openai-compatible" }
func (p *Provider) Generate(ctx context.Context, in ports.Request) (ports.Response, error) {
	if p.baseURL == "" || p.key == "" || p.model == "" {
		return ports.Response{}, fmt.Errorf("AI provider is not configured")
	}
	body, _ := json.Marshal(map[string]any{"model": p.model, "messages": []map[string]string{{"role": "system", "content": in.System}, {"role": "user", "content": in.User}}, "temperature": 0.2})
	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return ports.Response{}, err
	}
	req.Header.Set("Authorization", "Bearer "+p.key)
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
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			Prompt     int `json:"prompt_tokens"`
			Completion int `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err = json.NewDecoder(res.Body).Decode(&out); err != nil {
		return ports.Response{}, err
	}
	if len(out.Choices) == 0 {
		return ports.Response{}, fmt.Errorf("provider returned no choices")
	}
	return ports.Response{Text: out.Choices[0].Message.Content, InputTokens: out.Usage.Prompt, OutputTokens: out.Usage.Completion}, nil
}
