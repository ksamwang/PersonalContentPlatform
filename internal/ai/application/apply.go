package application

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/id"
)

type ApplyResult struct {
	Applied      []string `json:"applied"`
	DraftVersion *int     `json:"draft_version,omitempty"`
}

type suggestionPayload struct {
	Text string `json:"text"`
}

func suggestionValue(raw json.RawMessage, target any) error {
	var payload suggestionPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return err
	}
	value := strings.TrimSpace(payload.Text)
	if strings.HasPrefix(value, "```") {
		if newline := strings.IndexByte(value, '\n'); newline >= 0 {
			value = value[newline+1:]
		}
		value = strings.TrimSuffix(strings.TrimSpace(value), "```")
	}
	if err := json.Unmarshal([]byte(value), target); err != nil {
		return fmt.Errorf("AI suggestion is not valid JSON: %w", err)
	}
	return nil
}

func applyImageAlt(node any, values map[string]string) int {
	object, ok := node.(map[string]any)
	if !ok {
		return 0
	}
	changed := 0
	if object["type"] == "image" {
		if attrs, ok := object["attrs"].(map[string]any); ok {
			src, _ := attrs["src"].(string)
			if alt := strings.TrimSpace(values[src]); alt != "" {
				attrs["alt"] = alt
				changed++
			}
		}
	}
	if children, ok := object["content"].([]any); ok {
		for _, child := range children {
			changed += applyImageAlt(child, values)
		}
	}
	return changed
}

func (s *Service) applySuggestion(ctx context.Context, tx pgx.Tx, ws, user, suggestionID uuid.UUID) (ApplyResult, error) {
	var kind, state string
	var payload json.RawMessage
	var target, localization uuid.UUID
	err := tx.QueryRow(ctx, `
		SELECT s.kind,s.payload,s.target_object_id,s.state,
		       COALESCE(NULLIF(r.input_refs->>'localization_id','')::uuid,
		         (SELECT l.id FROM content_localizations l JOIN contents c ON c.object_id=l.content_id
		          WHERE l.workspace_id=s.workspace_id AND l.content_id=s.target_object_id AND l.deleted_at IS NULL
		          ORDER BY (l.locale=c.default_locale) DESC,l.created_at LIMIT 1))
		FROM ai_suggestions s JOIN ai_runs r ON r.id=s.run_id
		WHERE s.workspace_id=$1 AND s.id=$2 FOR UPDATE`, ws, suggestionID).
		Scan(&kind, &payload, &target, &state, &localization)
	if err != nil {
		return ApplyResult{}, err
	}
	if state != "pending" {
		return ApplyResult{}, fmt.Errorf("suggestion has already been reviewed")
	}

	result := ApplyResult{Applied: []string{kind}}
	switch kind {
	case "summary", "tags", "seo", "alt_text":
		var title, summary string
		var bodyRaw, metadataRaw []byte
		var version int
		if err = tx.QueryRow(ctx, `SELECT version,title,summary,body_json,metadata_json FROM draft_buffers WHERE workspace_id=$1 AND localization_id=$2 FOR UPDATE`, ws, localization).
			Scan(&version, &title, &summary, &bodyRaw, &metadataRaw); err != nil {
			return ApplyResult{}, err
		}
		var metadata map[string]any
		var body map[string]any
		if json.Unmarshal(metadataRaw, &metadata) != nil || metadata == nil {
			metadata = map[string]any{}
		}
		if err = json.Unmarshal(bodyRaw, &body); err != nil {
			return ApplyResult{}, fmt.Errorf("draft body is invalid: %w", err)
		}
		var wrapped suggestionPayload
		_ = json.Unmarshal(payload, &wrapped)
		switch kind {
		case "summary":
			summary = strings.TrimSpace(wrapped.Text)
			if summary == "" {
				return ApplyResult{}, fmt.Errorf("summary suggestion is empty")
			}
		case "tags":
			var tags []string
			if err = suggestionValue(payload, &tags); err != nil {
				return ApplyResult{}, err
			}
			clean := make([]string, 0, len(tags))
			for _, tag := range tags {
				if value := strings.TrimSpace(tag); value != "" {
					clean = append(clean, value)
				}
			}
			metadata["tags"] = clean
		case "seo":
			var seo map[string]any
			if err = suggestionValue(payload, &seo); err != nil {
				return ApplyResult{}, err
			}
			for _, key := range []string{"seo_title", "seo_description", "keywords", "heading_suggestions", "internal_link_suggestions"} {
				if value, exists := seo[key]; exists {
					metadata[key] = value
				}
			}
		case "alt_text":
			var items []struct {
				Src string `json:"src"`
				Alt string `json:"alt"`
			}
			if err = suggestionValue(payload, &items); err != nil {
				return ApplyResult{}, err
			}
			values := make(map[string]string, len(items))
			for _, item := range items {
				values[item.Src] = item.Alt
			}
			if applyImageAlt(body, values) == 0 {
				return ApplyResult{}, fmt.Errorf("no matching draft images were found")
			}
		}
		bodyRaw, _ = json.Marshal(body)
		metadataRaw, _ = json.Marshal(metadata)
		if err = tx.QueryRow(ctx, `UPDATE draft_buffers SET version=version+1,title=$1,summary=$2,body_json=$3,metadata_json=$4,updated_by=$5,updated_at=now() WHERE workspace_id=$6 AND localization_id=$7 RETURNING version`, title, summary, bodyRaw, metadataRaw, user, ws, localization).Scan(&version); err != nil {
			return ApplyResult{}, err
		}
		result.DraftVersion = &version
	case "entities":
		var entities []struct {
			Type          string   `json:"type"`
			CanonicalName string   `json:"canonical_name"`
			Aliases       []string `json:"aliases"`
		}
		if err = suggestionValue(payload, &entities); err != nil {
			return ApplyResult{}, err
		}
		var predicate uuid.UUID
		if err = tx.QueryRow(ctx, `INSERT INTO relation_predicates(id,workspace_id,key,label_zh,label_en) VALUES($1,$2,'mentions','提及','mentions') ON CONFLICT(workspace_id,key) DO UPDATE SET key=EXCLUDED.key RETURNING id`, id.New(), ws).Scan(&predicate); err != nil {
			return ApplyResult{}, err
		}
		for _, item := range entities {
			item.Type, item.CanonicalName = strings.TrimSpace(item.Type), strings.TrimSpace(item.CanonicalName)
			if item.Type == "" || item.CanonicalName == "" {
				continue
			}
			var entityID uuid.UUID
			err = tx.QueryRow(ctx, `SELECT object_id FROM entities WHERE workspace_id=$1 AND type=$2 AND canonical_name=$3`, ws, item.Type, item.CanonicalName).Scan(&entityID)
			if err == pgx.ErrNoRows {
				entityID = id.New()
				if _, err = tx.Exec(ctx, `INSERT INTO objects(id,workspace_id,kind) VALUES($1,$2,'entity')`, entityID, ws); err != nil {
					return ApplyResult{}, err
				}
				if _, err = tx.Exec(ctx, `INSERT INTO entities(object_id,workspace_id,type,canonical_name) VALUES($1,$2,$3,$4)`, entityID, ws, item.Type, item.CanonicalName); err != nil {
					return ApplyResult{}, err
				}
			} else if err != nil {
				return ApplyResult{}, err
			}
			for _, alias := range item.Aliases {
				if alias = strings.TrimSpace(alias); alias != "" {
					_, _ = tx.Exec(ctx, `INSERT INTO entity_aliases(id,workspace_id,entity_id,alias,locale) VALUES($1,$2,$3,$4,'zh-CN') ON CONFLICT DO NOTHING`, id.New(), ws, entityID, alias)
				}
			}
			_, _ = tx.Exec(ctx, `INSERT INTO relations(id,workspace_id,source_object_id,predicate_id,target_object_id,confirmed,provenance) VALUES($1,$2,$3,$4,$5,true,jsonb_build_object('source','ai_suggestion','suggestion_id',$6::text)) ON CONFLICT DO NOTHING`, id.New(), ws, target, predicate, entityID, suggestionID)
		}
	case "related":
		var related []struct {
			ObjectID uuid.UUID `json:"object_id"`
		}
		if err = suggestionValue(payload, &related); err != nil {
			return ApplyResult{}, err
		}
		var predicate uuid.UUID
		if err = tx.QueryRow(ctx, `INSERT INTO relation_predicates(id,workspace_id,key,label_zh,label_en) VALUES($1,$2,'related_to','相关','related to') ON CONFLICT(workspace_id,key) DO UPDATE SET key=EXCLUDED.key RETURNING id`, id.New(), ws).Scan(&predicate); err != nil {
			return ApplyResult{}, err
		}
		for _, item := range related {
			if item.ObjectID == uuid.Nil || item.ObjectID == target {
				continue
			}
			_, _ = tx.Exec(ctx, `INSERT INTO relations(id,workspace_id,source_object_id,predicate_id,target_object_id,confirmed,provenance) SELECT $1,$2,$3,$4,o.id,true,jsonb_build_object('source','ai_suggestion','suggestion_id',$6::text) FROM objects o WHERE o.workspace_id=$2 AND o.id=$5 ON CONFLICT DO NOTHING`, id.New(), ws, target, predicate, item.ObjectID, suggestionID)
		}
	default:
		return ApplyResult{}, fmt.Errorf("unsupported suggestion kind")
	}
	if _, err = tx.Exec(ctx, `UPDATE ai_suggestions SET state='accepted',reviewed_by=$1,reviewed_at=now() WHERE id=$2`, user, suggestionID); err != nil {
		return ApplyResult{}, err
	}
	_, _ = tx.Exec(ctx, `INSERT INTO audit_entries(id,workspace_id,actor_id,action,target_object_id,metadata) VALUES($1,$2,$3,'ai.suggestion.applied',$4,jsonb_build_object('suggestion_id',$5::text,'kind',$6))`, id.New(), ws, user, target, suggestionID, kind)
	return result, nil
}
