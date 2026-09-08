package application

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ksamwang/PersonalContentPlatform/internal/ai/ports"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/id"
)

type Service struct {
	db       *pgxpool.Pool
	provider ports.Provider
}

func New(db *pgxpool.Pool, p ports.Provider) *Service { return &Service{db: db, provider: p} }
func (s *Service) Suggest(ctx context.Context, ws, user, target uuid.UUID, purpose, input string) (uuid.UUID, error) {
	prompts := map[string]string{"summary": "Summarize the content concisely. Return plain text only.", "tags": "Suggest up to five precise tags as a JSON array.", "translation": "Translate faithfully while preserving structure. Return only the translation."}
	system, ok := prompts[purpose]
	if !ok {
		return uuid.Nil, fmt.Errorf("unsupported AI purpose")
	}
	runID := id.New()
	refs, _ := json.Marshal(map[string]any{"target_object_id": target})
	if _, err := s.db.Exec(ctx, `INSERT INTO ai_runs(id,workspace_id,purpose,provider,model,input_refs,status,created_by) VALUES($1,$2,$3,$4,'configured',$5,'running',$6)`, runID, ws, purpose, s.provider.Name(), refs, user); err != nil {
		return uuid.Nil, err
	}
	result, err := s.provider.Generate(ctx, ports.Request{System: system, User: input})
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
func (s *Service) Review(ctx context.Context, ws, user, idValue uuid.UUID, state string) error {
	if state != "accepted" && state != "rejected" {
		return fmt.Errorf("invalid review state")
	}
	_, err := s.db.Exec(ctx, `UPDATE ai_suggestions SET state=$1,reviewed_by=$2,reviewed_at=now() WHERE id=$3 AND workspace_id=$4 AND state='pending'`, state, user, idValue, ws)
	return err
}
