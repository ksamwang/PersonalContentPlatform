package smoke_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	assetdomain "github.com/ksamwang/PersonalContentPlatform/internal/asset/domain"
	contentdomain "github.com/ksamwang/PersonalContentPlatform/internal/content/domain"
	inboxdomain "github.com/ksamwang/PersonalContentPlatform/internal/inbox/domain"
	integrationdomain "github.com/ksamwang/PersonalContentPlatform/internal/integration/domain"
	knowledgedomain "github.com/ksamwang/PersonalContentPlatform/internal/knowledge/domain"
	settingsdomain "github.com/ksamwang/PersonalContentPlatform/internal/settings/domain"
)

func TestCorePlatformWorkflows(t *testing.T) {
	fixture := newSmokeFixture(t)
	client := fixture.client
	workspaceBase := "/v1/workspaces/" + fixture.workspaceID.String()

	var principal map[string]any
	client.json(http.MethodGet, "/v1/auth/me", nil, &principal, http.StatusOK)
	if principal["WorkspaceID"] != fixture.workspaceID.String() {
		t.Fatalf("unexpected principal: %#v", principal)
	}

	var settings settingsdomain.Settings
	client.json(http.MethodGet, workspaceBase+"/settings", nil, &settings, http.StatusOK)
	if !settings.CanEdit || settings.General.Workspace.DefaultLocale != "zh-CN" {
		t.Fatalf("unexpected settings: %#v", settings)
	}
	settings.General.Site.Name = "Smoke Notes"
	settings.General.Site.PublicURL = "https://example.invalid"
	settings.General.Site.Description = "Smoke description"
	settings.General.Site.About = "Smoke about"
	client.json(http.MethodPut, workspaceBase+"/settings/site", settings.General.Site, &settings.General, http.StatusOK)
	settings.General.Workspace.Name = "Smoke Workspace"
	settings.General.Workspace.SupportedLocales = []string{"zh-CN", "en"}
	settings.General.Workspace.Timezone = "Asia/Shanghai"
	client.json(http.MethodPut, workspaceBase+"/settings/workspace", settings.General.Workspace, &settings.General, http.StatusOK)
	var storage settingsdomain.StorageProfile
	client.json(http.MethodPut, workspaceBase+"/settings/storage", map[string]any{"name": "Smoke Files", "provider": "filesystem", "base_path": fixture.objectRoot}, &storage, http.StatusOK)
	client.json(http.MethodPost, workspaceBase+"/settings/storage:test", nil, &map[string]any{}, http.StatusOK)
	if storage.ID.String() == "" {
		t.Fatal("storage settings did not return an id")
	}
	var publicSettings map[string]any
	client.json(http.MethodGet, "/v1/public/"+fixture.workspace+"/settings", nil, &publicSettings, http.StatusOK)
	if publicSettings["site"].(map[string]any)["name"] != "Smoke Notes" {
		t.Fatalf("public settings not applied: %#v", publicSettings)
	}
	client.json(http.MethodPut, workspaceBase+"/settings/ai", map[string]any{"provider": "openai-compatible", "base_url": "https://example.invalid/v1", "api_key": "smoke-secret", "model": "smoke-model", "purpose_models": map[string]string{"summary": "smoke-summary"}}, &settings.AI, http.StatusOK)
	if !settings.AI.APIKeySet || settings.AI.APIKey != "" || settings.AI.APIKeyMask == "" {
		t.Fatalf("AI secret was not masked: %#v", settings.AI)
	}
	client.json(http.MethodPost, workspaceBase+"/settings/ai:test", nil, nil, http.StatusUnprocessableEntity)
	if _, err := fixture.db.Exec(t.Context(), `UPDATE memberships SET role='editor' WHERE workspace_id=$1 AND user_id=$2`, fixture.workspaceID, fixture.userID); err != nil {
		t.Fatal(err)
	}
	client.json(http.MethodPut, workspaceBase+"/settings/site", settings.General.Site, nil, http.StatusForbidden)
	client.json(http.MethodPost, workspaceBase+"/webhooks", map[string]any{"name": "Forbidden", "url": "https://example.invalid", "secret": "nope", "event_types": []string{}}, nil, http.StatusForbidden)
	if _, err := fixture.db.Exec(t.Context(), `UPDATE memberships SET role='owner' WHERE workspace_id=$1 AND user_id=$2`, fixture.workspaceID, fixture.userID); err != nil {
		t.Fatal(err)
	}

	var content contentdomain.Content
	client.json(http.MethodPost, workspaceBase+"/contents", map[string]any{
		"type": "article", "locale": "zh-CN", "slug": "smoke-test", "title": "Smoke Test",
	}, &content, http.StatusCreated)
	if len(content.Localizations) != 1 {
		t.Fatalf("expected one localization, got %d", len(content.Localizations))
	}
	localizationID := content.Localizations[0].ID

	var draft contentdomain.Draft
	client.json(http.MethodGet, workspaceBase+"/localizations/"+localizationID.String()+"/draft", nil, &draft, http.StatusOK)
	body := json.RawMessage(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Smoke body"}]}]}`)
	client.json(http.MethodPut, workspaceBase+"/localizations/"+localizationID.String()+"/draft", map[string]any{
		"version": draft.Version, "title": "Smoke Test", "summary": "Smoke summary", "body": body, "metadata": json.RawMessage(`{}`),
	}, &draft, http.StatusOK)

	var revision contentdomain.Revision
	client.json(http.MethodPost, workspaceBase+"/localizations/"+localizationID.String()+"/revisions", map[string]any{"version": draft.Version}, &revision, http.StatusCreated)
	client.json(http.MethodPost, workspaceBase+"/localizations/"+localizationID.String()+":mark-ready", nil, nil, http.StatusNoContent)
	client.json(http.MethodPost, workspaceBase+"/publications", map[string]any{"content_id": content.ID, "locale": "zh-CN"}, &map[string]any{}, http.StatusAccepted)
	waitForPublishedPage(t, client.base, fixture.workspace, "zh-CN", "article", "smoke-test")

	var contentItems struct {
		Items []contentdomain.Content `json:"items"`
	}
	client.json(http.MethodGet, workspaceBase+"/contents", nil, &contentItems, http.StatusOK)
	if len(contentItems.Items) != 1 {
		t.Fatalf("expected one content item, got %d", len(contentItems.Items))
	}
	waitForSearch(t, client, workspaceBase)

	var captured inboxdomain.Item
	client.json(http.MethodPost, workspaceBase+"/inbox/", map[string]any{
		"kind": "text", "raw_text": "Captured directly into the inbox",
	}, &captured, http.StatusCreated)
	var capturedLink inboxdomain.Item
	client.json(http.MethodPost, workspaceBase+"/inbox/", map[string]any{
		"kind": "link", "raw_text": "Reference", "source_url": "https://example.com/reference",
	}, &capturedLink, http.StatusCreated)
	var inboxItems struct {
		Items []inboxdomain.Item `json:"items"`
	}
	client.json(http.MethodGet, workspaceBase+"/inbox/?state=pending", nil, &inboxItems, http.StatusOK)
	if len(inboxItems.Items) != 2 {
		t.Fatalf("expected two pending inbox items, got %d", len(inboxItems.Items))
	}
	var conversion inboxdomain.Conversion
	client.json(http.MethodPost, workspaceBase+"/inbox/"+captured.ID.String()+":convert", map[string]any{
		"type": "note", "locale": "zh-CN", "slug": "captured-note", "title": "Captured Note",
	}, &conversion, http.StatusCreated)
	if conversion.ContentID == uuid.Nil || conversion.LocalizationID == uuid.Nil {
		t.Fatalf("unexpected inbox conversion: %#v", conversion)
	}
	var convertedDraft contentdomain.Draft
	client.json(http.MethodGet, workspaceBase+"/localizations/"+conversion.LocalizationID.String()+"/draft", nil, &convertedDraft, http.StatusOK)
	if !strings.Contains(string(convertedDraft.Body), "Captured directly into the inbox") {
		t.Fatalf("converted draft lost captured text: %s", convertedDraft.Body)
	}
	var repeated inboxdomain.Conversion
	client.json(http.MethodPost, workspaceBase+"/inbox/"+captured.ID.String()+":convert", map[string]any{
		"type": "note", "locale": "zh-CN", "slug": "captured-note", "title": "Captured Note",
	}, &repeated, http.StatusCreated)
	if repeated.ContentID != conversion.ContentID {
		t.Fatal("repeated conversion was not idempotent")
	}
	client.json(http.MethodPost, workspaceBase+"/inbox/"+capturedLink.ID.String()+":archive", nil, nil, http.StatusNoContent)
	if _, err := fixture.db.Exec(t.Context(), `UPDATE memberships SET role='viewer' WHERE workspace_id=$1 AND user_id=$2`, fixture.workspaceID, fixture.userID); err != nil {
		t.Fatal(err)
	}
	client.json(http.MethodPost, workspaceBase+"/inbox/", map[string]any{"kind": "text", "raw_text": "forbidden"}, nil, http.StatusForbidden)
	if _, err := fixture.db.Exec(t.Context(), `UPDATE memberships SET role='owner' WHERE workspace_id=$1 AND user_id=$2`, fixture.workspaceID, fixture.userID); err != nil {
		t.Fatal(err)
	}

	image := []byte("\x89PNG\r\n\x1a\nsmoke-image-" + fixture.workspaceID.String())
	var plan assetdomain.UploadPlan
	client.json(http.MethodPost, workspaceBase+"/assets:prepare-upload", map[string]any{
		"filename": "smoke.png", "mime": "image/png", "size": len(image),
	}, &plan, http.StatusCreated)
	client.raw(http.MethodPut, plan.URL, "image/png", image, http.StatusNoContent)
	var asset assetdomain.Asset
	client.json(http.MethodPost, workspaceBase+"/assets:finalize-upload", map[string]any{"upload_id": plan.UploadID}, &asset, http.StatusCreated)
	fixture.blobID = asset.BlobID
	if asset.State != "ready" || asset.SHA256 == "" {
		t.Fatalf("unexpected asset: %#v", asset)
	}
	if asset.StorageProfileID == nil || *asset.StorageProfileID != storage.ID {
		t.Fatalf("asset did not retain storage profile: %#v", asset)
	}
	var switchedStorage settingsdomain.StorageProfile
	client.json(http.MethodPut, workspaceBase+"/settings/storage", map[string]any{"id": storage.ID, "name": "Smoke Files Next", "provider": "filesystem", "base_path": "next"}, &switchedStorage, http.StatusOK)
	if switchedStorage.ID == storage.ID {
		t.Fatal("changing storage provider configuration must create a new profile")
	}
	var retainedProfileID *uuid.UUID
	if err := fixture.db.QueryRow(t.Context(), `SELECT storage_profile_id FROM blobs WHERE id=$1`, asset.BlobID).Scan(&retainedProfileID); err != nil {
		t.Fatal(err)
	}
	if retainedProfileID == nil || *retainedProfileID != storage.ID {
		t.Fatalf("old blob lost its storage profile: %v", retainedProfileID)
	}
	var assetItems struct {
		Items []assetdomain.Asset `json:"items"`
	}
	client.json(http.MethodGet, workspaceBase+"/assets/", nil, &assetItems, http.StatusOK)
	if len(assetItems.Items) != 1 {
		t.Fatalf("expected one asset, got %d", len(assetItems.Items))
	}

	firstEntity := createEntity(t, client, workspaceBase, "concept", "Smoke Source")
	secondEntity := createEntity(t, client, workspaceBase, "concept", "Smoke Target")
	var relation knowledgedomain.Relation
	client.json(http.MethodPost, workspaceBase+"/knowledge/relations", map[string]any{
		"source_id": firstEntity.ID, "target_id": secondEntity.ID, "predicate": "relates-to", "confirmed": false,
	}, &relation, http.StatusCreated)
	client.json(http.MethodPost, workspaceBase+"/knowledge/relations/"+relation.ID.String()+":confirm", nil, nil, http.StatusNoContent)
	var relationItems struct {
		Items []knowledgedomain.Relation `json:"items"`
	}
	client.json(http.MethodGet, workspaceBase+"/knowledge/objects/"+firstEntity.ID.String()+"/relations", nil, &relationItems, http.StatusOK)
	if len(relationItems.Items) != 1 {
		t.Fatalf("expected one relation, got %d", len(relationItems.Items))
	}

	var manifest integrationdomain.Manifest
	client.json(http.MethodGet, workspaceBase+"/exports/manifest.json", nil, &manifest, http.StatusOK)
	if len(manifest.Contents) != 2 || len(manifest.Assets) != 1 || len(manifest.Relations) != 1 {
		t.Fatalf("manifest does not contain core records: contents=%d assets=%d relations=%d", len(manifest.Contents), len(manifest.Assets), len(manifest.Relations))
	}

	var webhook integrationdomain.WebhookEndpoint
	client.json(http.MethodPost, workspaceBase+"/webhooks", map[string]any{
		"name": "Smoke Webhook", "url": "https://example.invalid/smoke", "secret": "smoke-webhook-secret", "event_types": []string{"NeverEmitted"},
	}, &webhook, http.StatusCreated)
	var webhookItems struct {
		Items []integrationdomain.WebhookEndpoint `json:"items"`
	}
	client.json(http.MethodGet, workspaceBase+"/webhooks", nil, &webhookItems, http.StatusOK)
	if len(webhookItems.Items) != 1 {
		t.Fatalf("expected one webhook, got %d", len(webhookItems.Items))
	}
	if !webhookItems.Items[0].SecretSet || webhookItems.Items[0].SecretValue != "" {
		t.Fatalf("webhook secret was exposed: %#v", webhookItems.Items[0])
	}
	client.json(http.MethodPatch, workspaceBase+"/webhooks/"+webhook.ID.String(), map[string]any{"enabled": false}, nil, http.StatusNoContent)
	client.json(http.MethodDelete, workspaceBase+"/webhooks/"+webhook.ID.String(), nil, nil, http.StatusNoContent)
	// Clear the intentionally unreachable provider so the normal unavailable path remains deterministic.
	if _, err := fixture.db.Exec(t.Context(), `DELETE FROM ai_provider_configs WHERE workspace_id=$1`, fixture.workspaceID); err != nil {
		t.Fatal(err)
	}
	client.json(http.MethodPost, workspaceBase+"/ai/suggestions", map[string]any{
		"target_id": content.ID, "purpose": "summary", "input": "Smoke body",
	}, nil, http.StatusServiceUnavailable)

	var passkeyOptions map[string]any
	client.json(http.MethodPost, "/v1/auth/passkeys/register/options", map[string]any{}, &passkeyOptions, http.StatusOK)
	if passkeyOptions["ceremony_id"] == nil {
		t.Fatal("passkey registration did not return ceremony_id")
	}

	assertRSS(t, client.base, fixture.workspace)
	client.json(http.MethodPost, "/v1/auth/logout", nil, nil, http.StatusNoContent)
	client.json(http.MethodGet, "/v1/auth/me", nil, nil, http.StatusUnauthorized)
}

func createEntity(t *testing.T, client apiClient, workspaceBase, kind, name string) knowledgedomain.Entity {
	t.Helper()
	var entity knowledgedomain.Entity
	client.json(http.MethodPost, workspaceBase+"/knowledge/entities", map[string]any{
		"type": kind, "canonical_name": name, "description": "smoke test",
	}, &entity, http.StatusCreated)
	return entity
}

func waitForPublishedPage(t *testing.T, base, workspace, locale, kind, slug string) {
	t.Helper()
	path := fmt.Sprintf("%s/v1/public/%s/%s/%s/%s", base, workspace, locale, kind, slug)
	for attempt := 0; attempt < 30; attempt++ {
		response, err := http.Get(path)
		if err == nil {
			_ = response.Body.Close()
			if response.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatal("publication did not become publicly readable")
}

func waitForSearch(t *testing.T, client apiClient, workspaceBase string) {
	t.Helper()
	for attempt := 0; attempt < 30; attempt++ {
		var result struct {
			Items []contentdomain.Content `json:"items"`
		}
		client.json(http.MethodGet, workspaceBase+"/search?q="+url.QueryEscape("Smoke"), nil, &result, http.StatusOK)
		if len(result.Items) > 0 {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatal("content did not become searchable")
}

func assertRSS(t *testing.T, base, workspace string) {
	t.Helper()
	response, err := http.Get(base + "/v1/public/" + workspace + "/zh-CN/rss.xml")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	payload, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || !strings.Contains(string(payload), "<title>Smoke Test</title>") {
		t.Fatalf("unexpected RSS response %d: %s", response.StatusCode, payload)
	}
}
