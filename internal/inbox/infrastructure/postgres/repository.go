package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	contentdomain "github.com/ksamwang/PersonalContentPlatform/internal/content/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/inbox/application"
	"github.com/ksamwang/PersonalContentPlatform/internal/inbox/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/id"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/outbox"
)

type Repository struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

func (r *Repository) Create(ctx context.Context, workspaceID, userID uuid.UUID, kind domain.Kind, rawText string, sourceURL *string) (domain.Item, error) {
	item := domain.Item{ID: id.New(), WorkspaceID: workspaceID, Kind: kind, RawText: rawText, SourceURL: sourceURL, State: domain.StatePending}
	err := r.db.QueryRow(ctx, `INSERT INTO inbox_items(id,workspace_id,kind,raw_text,source_url,created_by) VALUES($1,$2,$3,$4,$5,$6) RETURNING created_at,updated_at`, item.ID, workspaceID, kind, rawText, sourceURL, userID).Scan(&item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (r *Repository) List(ctx context.Context, workspaceID uuid.UUID, state domain.State, limit int) ([]domain.Item, error) {
	query := `SELECT id,workspace_id,kind,raw_text,source_url,state,converted_content_id,created_at,updated_at FROM inbox_items WHERE workspace_id=$1`
	args := []any{workspaceID}
	if state != "" {
		args = append(args, state)
		query += fmt.Sprintf(" AND state=$%d", len(args))
	}
	args = append(args, limit)
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", len(args))
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.Item{}
	for rows.Next() {
		var item domain.Item
		if err := rows.Scan(&item.ID, &item.WorkspaceID, &item.Kind, &item.RawText, &item.SourceURL, &item.State, &item.ConvertedContentID, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) Archive(ctx context.Context, workspaceID, itemID uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `UPDATE inbox_items SET state='archived',updated_at=now() WHERE id=$1 AND workspace_id=$2 AND state='pending'`, itemID, workspaceID)
	if err == nil && tag.RowsAffected() == 0 {
		return application.ErrNotPending
	}
	return err
}

func (r *Repository) Convert(ctx context.Context, workspaceID, userID, itemID uuid.UUID, kind contentdomain.Type, locale, slug, title string) (domain.Conversion, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Conversion{}, err
	}
	defer tx.Rollback(ctx)
	var inboxState domain.State
	var rawText string
	var sourceURL *string
	var existingContentID *uuid.UUID
	if err = tx.QueryRow(ctx, `SELECT state,raw_text,source_url,converted_content_id FROM inbox_items WHERE id=$1 AND workspace_id=$2 FOR UPDATE`, itemID, workspaceID).Scan(&inboxState, &rawText, &sourceURL, &existingContentID); err != nil {
		return domain.Conversion{}, err
	}
	if inboxState == domain.StateConverted && existingContentID != nil {
		var localizationID uuid.UUID
		if err = tx.QueryRow(ctx, `SELECT id FROM content_localizations WHERE workspace_id=$1 AND content_id=$2 AND locale=$3`, workspaceID, *existingContentID, locale).Scan(&localizationID); err != nil {
			return domain.Conversion{}, err
		}
		return domain.Conversion{InboxID: itemID, ContentID: *existingContentID, LocalizationID: localizationID}, nil
	}
	if inboxState != domain.StatePending {
		return domain.Conversion{}, application.ErrNotPending
	}
	contentID, localizationID, revisionID := id.New(), id.New(), id.New()
	body := capturedBody(rawText, sourceURL)
	metadata := json.RawMessage(`{}`)
	hash := revisionHash(title, body, metadata)
	statements := []struct {
		query string
		args  []any
	}{
		{`INSERT INTO objects(id,workspace_id,kind) VALUES($1,$2,'content')`, []any{contentID, workspaceID}},
		{`INSERT INTO contents(object_id,workspace_id,type,default_locale) VALUES($1,$2,$3,$4)`, []any{contentID, workspaceID, kind, locale}},
		{`INSERT INTO content_localizations(id,workspace_id,content_id,locale,slug) VALUES($1,$2,$3,$4,$5)`, []any{localizationID, workspaceID, contentID, locale, slug}},
		{`INSERT INTO content_revisions(id,workspace_id,localization_id,seq,title,body_json,metadata_json,content_hash,created_by) VALUES($1,$2,$3,1,$4,$5,$6,$7,$8)`, []any{revisionID, workspaceID, localizationID, title, body, metadata, hash, userID}},
		{`UPDATE content_localizations SET current_revision_id=$1 WHERE id=$2`, []any{revisionID, localizationID}},
		{`INSERT INTO draft_buffers(localization_id,workspace_id,title,body_json,metadata_json,updated_by) VALUES($1,$2,$3,$4,$5,$6)`, []any{localizationID, workspaceID, title, body, metadata, userID}},
		{`UPDATE inbox_items SET state='converted',converted_content_id=$1,updated_at=now() WHERE id=$2`, []any{contentID, itemID}},
	}
	for _, statement := range statements {
		if _, err = tx.Exec(ctx, statement.query, statement.args...); err != nil {
			return domain.Conversion{}, err
		}
	}
	if err = outbox.Add(ctx, tx, workspaceID, contentID, "content", "ContentCreated", map[string]any{"content_id": contentID, "revision_id": revisionID, "locale": locale, "inbox_id": itemID}); err != nil {
		return domain.Conversion{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.Conversion{}, err
	}
	return domain.Conversion{InboxID: itemID, ContentID: contentID, LocalizationID: localizationID}, nil
}

func capturedBody(rawText string, sourceURL *string) json.RawMessage {
	content := []map[string]any{}
	if rawText != "" {
		content = append(content, map[string]any{"type": "paragraph", "content": []map[string]string{{"type": "text", "text": rawText}}})
	}
	if sourceURL != nil {
		content = append(content, map[string]any{"type": "paragraph", "content": []map[string]any{{"type": "text", "text": *sourceURL, "marks": []map[string]any{{"type": "link", "attrs": map[string]any{"href": *sourceURL}}}}}})
	}
	payload, _ := json.Marshal(map[string]any{"schemaVersion": 1, "type": "doc", "content": content})
	return payload
}

func revisionHash(title string, body, metadata []byte) string {
	sum := sha256.Sum256(append(append(append([]byte(title+"\n\n"), body...), '\n'), metadata...))
	return hex.EncodeToString(sum[:])
}
