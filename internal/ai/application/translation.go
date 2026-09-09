package application

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/ksamwang/PersonalContentPlatform/internal/ai/ports"
	contentdomain "github.com/ksamwang/PersonalContentPlatform/internal/content/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/id"
)

type translationPayload struct {
	Title    string          `json:"title"`
	Summary  string          `json:"summary"`
	Body     json.RawMessage `json:"body"`
	Metadata map[string]any  `json:"metadata"`
}

func (s *Service) Translate(ctx context.Context, workspaceID, userID, sourceLocalizationID, targetLocalizationID uuid.UUID) (contentdomain.Draft, error) {
	var sourceLocale, targetLocale, sourceHash, title, summary string
	var sourceRevisionID uuid.UUID
	var body, sourceMetadata, targetMetadata []byte
	var targetVersion int
	err := s.db.QueryRow(ctx, `SELECT sl.locale,tl.locale,sr.id,sr.content_hash,sr.title,sr.summary,sr.body_json,sr.metadata_json,td.version,td.metadata_json FROM content_localizations sl JOIN content_revisions sr ON sr.id=sl.current_revision_id JOIN content_localizations tl ON tl.content_id=sl.content_id AND tl.id=$3 JOIN draft_buffers td ON td.localization_id=tl.id WHERE sl.workspace_id=$1 AND sl.id=$2 AND tl.workspace_id=$1`, workspaceID, sourceLocalizationID, targetLocalizationID).Scan(&sourceLocale, &targetLocale, &sourceRevisionID, &sourceHash, &title, &summary, &body, &sourceMetadata, &targetVersion, &targetMetadata)
	if err != nil {
		return contentdomain.Draft{}, err
	}
	if sourceLocale == targetLocale {
		return contentdomain.Draft{}, fmt.Errorf("source and target locales must differ")
	}
	provider, model, err := s.providers.Resolve(ctx, workspaceID)
	if err != nil {
		return contentdomain.Draft{}, err
	}
	runID := id.New()
	refs, _ := json.Marshal(map[string]any{"source_revision_id": sourceRevisionID, "target_localization_id": targetLocalizationID})
	if _, err = s.db.Exec(ctx, `INSERT INTO ai_runs(id,workspace_id,purpose,provider,model,input_refs,status,created_by) VALUES($1,$2,'translation',$3,$4,$5,'running',$6)`, runID, workspaceID, provider.Name(), model, refs, userID); err != nil {
		return contentdomain.Draft{}, err
	}
	input, _ := json.Marshal(map[string]any{"source_locale": sourceLocale, "target_locale": targetLocale, "title": title, "summary": summary, "body": json.RawMessage(body), "metadata": json.RawMessage(sourceMetadata)})
	system := "Translate the supplied content faithfully. Return one JSON object only with title, summary, body and metadata. Preserve the TipTap body tree, node types, marks, URLs, asset identifiers and code exactly; translate only human-readable text and image alt/title text. Preserve metadata keys and non-text values. Do not add facts or markdown fences."
	result, err := provider.Generate(ctx, ports.Request{System: system, User: string(input), Model: model})
	if err != nil {
		s.failRun(ctx, runID)
		return contentdomain.Draft{}, err
	}
	var translated translationPayload
	if err = json.Unmarshal(extractJSONObject(result.Text), &translated); err != nil || !validDocument(translated.Body) || strings.TrimSpace(translated.Title) == "" {
		s.failRun(ctx, runID)
		return contentdomain.Draft{}, fmt.Errorf("AI provider returned an invalid translation document")
	}
	var metadata map[string]any
	_ = json.Unmarshal(sourceMetadata, &metadata)
	if metadata == nil {
		metadata = map[string]any{}
	}
	for _, key := range []string{"seo_title", "seo_description", "tags"} {
		if value, ok := translated.Metadata[key]; ok {
			metadata[key] = value
		}
	}
	if len(targetMetadata) > 0 {
		var current map[string]any
		if json.Unmarshal(targetMetadata, &current) == nil {
			for key, value := range current {
				if key != "seo_title" && key != "seo_description" && key != "tags" {
					metadata[key] = value
				}
			}
		}
	}
	metadataJSON, _ := json.Marshal(metadata)
	usage, _ := json.Marshal(map[string]int{"input_tokens": result.InputTokens, "output_tokens": result.OutputTokens})
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return contentdomain.Draft{}, err
	}
	defer tx.Rollback(ctx)
	var draft contentdomain.Draft
	err = tx.QueryRow(ctx, `UPDATE draft_buffers SET version=version+1,title=$1,summary=$2,body_json=$3,metadata_json=$4,updated_by=$5,updated_at=now() WHERE localization_id=$6 AND workspace_id=$7 AND version=$8 RETURNING localization_id,version,title,summary,body_json,metadata_json,updated_at`, strings.TrimSpace(translated.Title), strings.TrimSpace(translated.Summary), translated.Body, metadataJSON, userID, targetLocalizationID, workspaceID, targetVersion).Scan(&draft.LocalizationID, &draft.Version, &draft.Title, &draft.Summary, &draft.Body, &draft.Metadata, &draft.UpdatedAt)
	if err == pgx.ErrNoRows {
		return contentdomain.Draft{}, fmt.Errorf("target draft changed; reload before translating")
	}
	if err != nil {
		return contentdomain.Draft{}, err
	}
	if _, err = tx.Exec(ctx, `UPDATE content_localizations SET source_locale=$1,source_revision_id=$2,translated_from_hash=$3,translation_status='needs_review',state='draft',updated_at=now() WHERE id=$4 AND workspace_id=$5`, sourceLocale, sourceRevisionID, sourceHash, targetLocalizationID, workspaceID); err != nil {
		return contentdomain.Draft{}, err
	}
	if _, err = tx.Exec(ctx, `UPDATE ai_runs SET status='succeeded',usage_json=$2,completed_at=now() WHERE id=$1`, runID, usage); err != nil {
		return contentdomain.Draft{}, err
	}
	return draft, tx.Commit(ctx)
}

func (s *Service) failRun(ctx context.Context, runID uuid.UUID) {
	_, _ = s.db.Exec(ctx, `UPDATE ai_runs SET status='failed',error_code='translation_invalid',completed_at=now() WHERE id=$1`, runID)
}
func extractJSONObject(value string) []byte {
	value = strings.TrimSpace(value)
	start := strings.Index(value, "{")
	end := strings.LastIndex(value, "}")
	if start >= 0 && end > start {
		return []byte(value[start : end+1])
	}
	return []byte(value)
}
func validDocument(raw json.RawMessage) bool {
	var body struct {
		Type    string `json:"type"`
		Content []any  `json:"content"`
	}
	return json.Unmarshal(raw, &body) == nil && body.Type == "doc" && body.Content != nil
}
