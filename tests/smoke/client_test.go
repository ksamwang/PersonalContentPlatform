package smoke_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

type apiClient struct {
	t      *testing.T
	base   string
	cookie *http.Cookie
}

func (c apiClient) json(method, path string, input, output any, expected int) {
	c.t.Helper()
	var body io.Reader
	if input != nil {
		payload, err := json.Marshal(input)
		if err != nil {
			c.t.Fatal(err)
		}
		body = bytes.NewReader(payload)
	}
	request, err := http.NewRequest(method, c.base+path, body)
	if err != nil {
		c.t.Fatal(err)
	}
	request.AddCookie(c.cookie)
	if input != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		c.t.Fatal(err)
	}
	defer response.Body.Close()
	payload, err := io.ReadAll(response.Body)
	if err != nil {
		c.t.Fatal(err)
	}
	if response.StatusCode != expected {
		c.t.Fatalf("%s %s: got %d want %d: %s", method, path, response.StatusCode, expected, payload)
	}
	if output != nil && len(payload) > 0 {
		if err = json.Unmarshal(payload, output); err != nil {
			c.t.Fatalf("decode %s %s: %v: %s", method, path, err, payload)
		}
	}
}

func (c apiClient) raw(method, path, contentType string, payload []byte, expected int) {
	c.t.Helper()
	request, err := http.NewRequest(method, c.base+path, bytes.NewReader(payload))
	if err != nil {
		c.t.Fatal(err)
	}
	request.AddCookie(c.cookie)
	request.Header.Set("Content-Type", contentType)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		c.t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != expected {
		body, _ := io.ReadAll(response.Body)
		c.t.Fatalf("%s %s: got %d want %d: %s", method, path, response.StatusCode, expected, body)
	}
}
