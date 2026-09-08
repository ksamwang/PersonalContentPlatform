package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ksamwang/PersonalContentPlatform/internal/integration/domain"
)

func TestSenderSignsEnvelope(t *testing.T) {
	fixedTime := time.Unix(1_700_000_000, 0).UTC()
	var receivedBody []byte
	var receivedSignature string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		receivedBody, err = io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		receivedSignature = r.Header.Get("X-PCP-Signature")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	sender := NewSender()
	sender.now = func() time.Time { return fixedTime }
	sender.resolveSecret = func(name string) (string, bool) { return "test-secret", name == "PCP_WEBHOOK_SECRET_TEST" }
	result := sender.Send(context.Background(), domain.WebhookDelivery{
		EventID:   uuid.MustParse("32245403-756d-4c19-ac1c-497583904bf9"),
		EventType: "PublicationRequested",
		URL:       server.URL,
		SecretRef: "PCP_WEBHOOK_SECRET_TEST",
		Payload:   []byte(`{"publication_id":"123"}`),
		CreatedAt: fixedTime,
	})
	if result.Status != "succeeded" {
		t.Fatalf("expected succeeded, got %s: %s", result.Status, result.ResponseExcerpt)
	}
	mac := hmac.New(sha256.New, []byte("test-secret"))
	_, _ = mac.Write([]byte("1700000000."))
	_, _ = mac.Write(receivedBody)
	want := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if receivedSignature != want {
		t.Fatalf("signature mismatch: got %s want %s", receivedSignature, want)
	}
}

func TestSenderFailsWhenSecretIsMissing(t *testing.T) {
	sender := NewSender()
	sender.resolveSecret = func(string) (string, bool) { return "", false }
	result := sender.Send(context.Background(), domain.WebhookDelivery{SecretRef: "MISSING"})
	if result.Status != "failed" {
		t.Fatalf("expected failed, got %s", result.Status)
	}
}
