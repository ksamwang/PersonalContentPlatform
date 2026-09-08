package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ksamwang/PersonalContentPlatform/internal/integration/domain"
)

const responseExcerptLimit = 2048

type SecretResolver func(string) (string, bool)

type Sender struct {
	client        *http.Client
	resolveSecret SecretResolver
	now           func() time.Time
}

func NewSender() *Sender {
	return &Sender{
		client:        &http.Client{Timeout: 10 * time.Second},
		resolveSecret: os.LookupEnv,
		now:           time.Now,
	}
}

func (s *Sender) Send(ctx context.Context, delivery domain.WebhookDelivery) domain.DeliveryResult {
	secret, ok := s.resolveSecret(delivery.SecretRef)
	if !ok || secret == "" {
		return failed(fmt.Sprintf("secret environment variable %q is not configured", delivery.SecretRef), nil)
	}
	body, err := json.Marshal(map[string]any{
		"id":         delivery.EventID,
		"type":       delivery.EventType,
		"created_at": delivery.CreatedAt,
		"data":       json.RawMessage(delivery.Payload),
	})
	if err != nil {
		return failed(err.Error(), nil)
	}
	timestamp := strconv.FormatInt(s.now().UTC().Unix(), 10)
	signature := sign(secret, timestamp, body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, delivery.URL, bytes.NewReader(body))
	if err != nil {
		return failed(err.Error(), nil)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "PersonalContentPlatform-Webhook/1.0")
	req.Header.Set("X-PCP-Event-ID", delivery.EventID.String())
	req.Header.Set("X-PCP-Event-Type", delivery.EventType)
	req.Header.Set("X-PCP-Timestamp", timestamp)
	req.Header.Set("X-PCP-Signature", "sha256="+signature)

	response, err := s.client.Do(req)
	if err != nil {
		return failed(err.Error(), nil)
	}
	defer response.Body.Close()
	excerpt, readErr := io.ReadAll(io.LimitReader(response.Body, responseExcerptLimit))
	code := response.StatusCode
	if readErr != nil {
		return failed(readErr.Error(), &code)
	}
	result := domain.DeliveryResult{Status: "succeeded", ResponseCode: &code, ResponseExcerpt: strings.TrimSpace(string(excerpt))}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		result.Status = "failed"
	}
	return result
}

func sign(secret, timestamp string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(timestamp))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func failed(message string, code *int) domain.DeliveryResult {
	if len(message) > responseExcerptLimit {
		message = message[:responseExcerptLimit]
	}
	return domain.DeliveryResult{Status: "failed", ResponseCode: code, ResponseExcerpt: message}
}
