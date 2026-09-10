package application

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ksamwang/PersonalContentPlatform/internal/ai/ports"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/id"
	"time"
)

type Service struct {
	db        *pgxpool.Pool
	providers ProviderResolver
}

type ProviderResolver interface {
	Resolve(context.Context, uuid.UUID) (ports.Provider, string, error)
}

func New(db *pgxpool.Pool, p ProviderResolver) *Service { return &Service{db: db, providers: p} }
func (s *Service) Suggest(ctx context.Context, ws, user, target, localization uuid.UUID, purpose, input string) (uuid.UUID, error) {
	prompts := map[string]string{
		"summary":  "Create a concise, faithful summary for the supplied content. Return plain text only.",
		"tags":     "Suggest up to five precise tags for the supplied content. Return a JSON array of strings only.",
		"seo":      "Act as an SEO editor. Return one JSON object with seo_title, seo_description, slug, keywords, heading_suggestions and internal_link_suggestions. Keep claims faithful to the supplied content.",
		"entities": "Extract people, organizations, brands, places and topics. Return a JSON array with type, canonical_name, aliases and confidence.",
		"alt_text": "Write concise accessible alt text for every image represented in the supplied content. Return a JSON array with src and alt.",
		"related":  "Suggest related content based only on supplied candidates. Return a JSON array with object_id and reason.",
	}
	system, ok := prompts[purpose]
	if !ok {
		return uuid.Nil, fmt.Errorf("unsupported AI purpose")
	}
	runID := id.New()
	provider, model, err := s.providers.Resolve(ctx, ws)
	if err != nil {
		return uuid.Nil, err
	}
	refs, _ := json.Marshal(map[string]any{"target_object_id": target, "localization_id": localization})
	if _, err := s.db.Exec(ctx, `INSERT INTO ai_runs(id,workspace_id,purpose,provider,model,input_refs,status,created_by) VALUES($1,$2,$3,$4,$5,$6,'running',$7)`, runID, ws, purpose, provider.Name(), model, refs, user); err != nil {
		return uuid.Nil, err
	}
	result, err := provider.Generate(ctx, ports.Request{System: system, User: input, Model: model})
	if err != nil {
		_, _ = s.db.Exec(ctx, `UPDATE ai_runs SET status='failed',error_code='provider_error',completed_at=now() WHERE id=$1`, runID)
		return uuid.Nil, err
	}
	usage, _ := json.Marshal(map[string]int{"input_tokens": result.InputTokens, "output_tokens": result.OutputTokens})
	suggestionID := id.New()
	payload, _ := json.Marshal(map[string]string{"text": result.Text})
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `UPDATE ai_runs SET status='succeeded',usage_json=$2,completed_at=now() WHERE id=$1`, runID, usage); err != nil {
		return uuid.Nil, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO ai_suggestions(id,workspace_id,run_id,target_object_id,kind,payload) VALUES($1,$2,$3,$4,$5,$6)`, suggestionID, ws, runID, target, purpose, payload); err != nil {
		return uuid.Nil, err
	}
	return suggestionID, tx.Commit(ctx)
}

type Suggestion struct {
	ID          uuid.UUID       `json:"id"`
	TargetID    uuid.UUID       `json:"target_id"`
	TargetTitle string          `json:"target_title"`
	Kind        string          `json:"kind"`
	Payload     json.RawMessage `json:"payload"`
	State       string          `json:"state"`
	CreatedAt   time.Time       `json:"created_at"`
}
type Run struct {
	ID          uuid.UUID       `json:"id"`
	Purpose     string          `json:"purpose"`
	Provider    string          `json:"provider"`
	Model       string          `json:"model"`
	Status      string          `json:"status"`
	Usage       json.RawMessage `json:"usage"`
	ErrorCode   string          `json:"error_code"`
	StartedAt   time.Time       `json:"started_at"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
}

func (s *Service) Suggestions(ctx context.Context, ws uuid.UUID, state string) ([]Suggestion, error) {
	rows, err := s.db.Query(ctx, `SELECT s.id,s.target_object_id,COALESCE((SELECT rv.title FROM content_localizations l JOIN content_revisions rv ON rv.id=l.current_revision_id WHERE l.content_id=s.target_object_id ORDER BY (l.locale='zh-CN') DESC LIMIT 1),''),s.kind,s.payload,s.state,s.created_at FROM ai_suggestions s WHERE s.workspace_id=$1 AND ($2='' OR s.state=$2) ORDER BY s.created_at DESC LIMIT 100`, ws, state)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Suggestion{}
	for rows.Next() {
		var v Suggestion
		if err = rows.Scan(&v.ID, &v.TargetID, &v.TargetTitle, &v.Kind, &v.Payload, &v.State, &v.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, v)
	}
	return items, rows.Err()
}
func (s *Service) Runs(ctx context.Context, ws uuid.UUID) ([]Run, error) {
	rows, err := s.db.Query(ctx, `SELECT id,purpose,provider,model,status,usage_json,COALESCE(error_code,''),started_at,completed_at FROM ai_runs WHERE workspace_id=$1 ORDER BY started_at DESC LIMIT 100`, ws)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Run{}
	for rows.Next() {
		var v Run
		if err = rows.Scan(&v.ID, &v.Purpose, &v.Provider, &v.Model, &v.Status, &v.Usage, &v.ErrorCode, &v.StartedAt, &v.CompletedAt); err != nil {
			return nil, err
		}
		items = append(items, v)
	}
	return items, rows.Err()
}
func (s *Service) Review(ctx context.Context, ws, user, idValue uuid.UUID, state string) (ApplyResult, error) {
	if state != "accepted" && state != "rejected" {
		return ApplyResult{}, fmt.Errorf("invalid review state")
	}
	if state == "rejected" {
		tag, err := s.db.Exec(ctx, `UPDATE ai_suggestions SET state='rejected',reviewed_by=$1,reviewed_at=now() WHERE id=$2 AND workspace_id=$3 AND state='pending'`, user, idValue, ws)
		if err == nil && tag.RowsAffected() == 0 {
			return ApplyResult{}, fmt.Errorf("suggestion has already been reviewed")
		}
		return ApplyResult{Applied: []string{}}, err
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return ApplyResult{}, err
	}
	defer tx.Rollback(ctx)
	result, err := s.applySuggestion(ctx, tx, ws, user, idValue)
	if err != nil {
		return ApplyResult{}, err
	}
	return result, tx.Commit(ctx)
}
