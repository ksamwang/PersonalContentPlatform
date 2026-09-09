package processing

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path"
	"strings"

	"github.com/ksamwang/PersonalContentPlatform/internal/settings/domain"
)

type MediaClient struct{ client *http.Client }

func NewMediaClient() *MediaClient { return &MediaClient{client: http.DefaultClient} }
func (c *MediaClient) OCR(ctx context.Context, cfg domain.MediaConfig, reader io.Reader, mime string) (string, error) {
	data, err := io.ReadAll(io.LimitReader(reader, 20<<20))
	if err != nil {
		return "", err
	}
	payload := map[string]any{"model": cfg.Model, "messages": []any{map[string]any{"role": "user", "content": []any{map[string]any{"type": "text", "text": "Extract all readable text. Preserve paragraphs and return text only."}, map[string]any{"type": "image_url", "image_url": map[string]string{"url": "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data)}}}}}, "temperature": 0}
	return c.chat(ctx, cfg, payload)
}
func (c *MediaClient) chat(ctx context.Context, cfg domain.MediaConfig, payload any) (string, error) {
	raw, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(cfg.BaseURL, "/")+"/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("provider returned HTTP %d: %s", resp.StatusCode, string(body))
	}
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err = json.Unmarshal(body, &out); err != nil || len(out.Choices) == 0 {
		return "", fmt.Errorf("invalid provider response")
	}
	return strings.TrimSpace(out.Choices[0].Message.Content), nil
}
func (c *MediaClient) Transcribe(ctx context.Context, cfg domain.MediaConfig, reader io.Reader, filename string) (string, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("model", cfg.Model)
	part, err := writer.CreateFormFile("file", path.Base(filename))
	if err != nil {
		return "", err
	}
	if _, err = io.Copy(part, io.LimitReader(reader, 100<<20)); err != nil {
		return "", err
	}
	_ = writer.Close()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(cfg.BaseURL, "/")+"/audio/transcriptions", &body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("provider returned HTTP %d: %s", resp.StatusCode, string(raw))
	}
	var out struct {
		Text string `json:"text"`
	}
	if err = json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.Text), nil
}
