package anthropic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ksamwang/PersonalContentPlatform/internal/ai/ports"
)

func TestGenerateUsesAnthropicCompatibleProtocol(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" || r.Header.Get("x-api-key") != "test-key" || r.Header.Get("anthropic-version") == "" {
			t.Fatalf("unexpected request: path=%s headers=%v", r.URL.Path, r.Header)
		}
		var input map[string]any
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Fatal(err)
		}
		if input["model"] != "claude-test" || input["system"] != "system" {
			t.Fatalf("unexpected payload: %#v", input)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"content":[{"type":"text","text":"OK"}],"usage":{"input_tokens":3,"output_tokens":1}}`))
	}))
	defer server.Close()

	result, err := New(server.URL+"/v1", "test-key", "claude-test").Generate(context.Background(), ports.Request{System: "system", User: "user"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Text != "OK" || result.InputTokens != 3 || result.OutputTokens != 1 {
		t.Fatalf("unexpected result: %#v", result)
	}
}
