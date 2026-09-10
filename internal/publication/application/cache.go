package application

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type CacheInvalidator interface {
	Invalidate(context.Context, uuid.UUID) error
}

type HTTPCacheInvalidator struct {
	origin, token string
	client        *http.Client
}

func NewCacheInvalidator(origin, token string) *HTTPCacheInvalidator {
	return &HTTPCacheInvalidator{origin: strings.TrimRight(origin, "/"), token: token, client: &http.Client{Timeout: 5 * time.Second}}
}

func (c *HTTPCacheInvalidator) Invalidate(ctx context.Context, workspaceID uuid.UUID) error {
	if c.origin == "" {
		return nil
	}
	body, _ := json.Marshal(map[string]string{"workspace_id": workspaceID.String()})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.origin+"/api/revalidate", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("public cache invalidation returned HTTP %d", resp.StatusCode)
	}
	return nil
}
