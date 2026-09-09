package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	contentdomain "github.com/ksamwang/PersonalContentPlatform/internal/content/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/inbox/application"
	"github.com/ksamwang/PersonalContentPlatform/internal/inbox/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/id"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/outbox"
)

type Repository struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

func (r *Repository) Create(ctx context.Context, workspaceID, userID uuid.UUID, kind domain.Kind, rawText string, sourceURL *string, assetID *uuid.UUID) (domain.Item, error) {
	item := domain.Item{ID: id.New(), WorkspaceID: workspaceID, Kind: kind, RawText: rawText, SourceURL: sourceURL, AssetID: assetID, State: domain.StatePending, ProcessingState: "idle"}
	err := r.db.QueryRow(ctx, `INSERT INTO inbox_items(id,workspace_id,kind,raw_text,source_url,asset_id,created_by) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING created_at,updated_at`, item.ID, workspaceID, kind, rawText, sourceURL, assetID, userID).Scan(&item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (r *Repository) List(ctx context.Context, workspaceID uuid.UUID, state domain.State, limit int) ([]domain.Item, error) {
	query := `SELECT id,workspace_id,kind,raw_text,source_url,asset_id,title,extracted_text,cover_url,processing_state,processing_error,duplicate_of,state,converted_content_id,created_at,updated_at FROM inbox_items WHERE workspace_id=$1`
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
		if err := rows.Scan(&item.ID, &item.WorkspaceID, &item.Kind, &item.RawText, &item.SourceURL, &item.AssetID, &item.Title, &item.ExtractedText, &item.CoverURL, &item.ProcessingState, &item.ProcessingError, &item.DuplicateOf, &item.State, &item.ConvertedContentID, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func scanItem(row pgx.Row) (domain.Item, error) {
	var v domain.Item
	err := row.Scan(&v.ID, &v.WorkspaceID, &v.Kind, &v.RawText, &v.SourceURL, &v.AssetID, &v.Title, &v.ExtractedText, &v.CoverURL, &v.ProcessingState, &v.ProcessingError, &v.DuplicateOf, &v.State, &v.ConvertedContentID, &v.CreatedAt, &v.UpdatedAt)
	return v, err
}
func (r *Repository) Get(ctx context.Context, ws, idv uuid.UUID) (domain.Item, error) {
	return scanItem(r.db.QueryRow(ctx, `SELECT id,workspace_id,kind,raw_text,source_url,asset_id,title,extracted_text,cover_url,processing_state,processing_error,duplicate_of,state,converted_content_id,created_at,updated_at FROM inbox_items WHERE workspace_id=$1 AND id=$2`, ws, idv))
}
func (r *Repository) BeginProcessing(ctx context.Context, ws, idv uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE inbox_items SET processing_state='processing',processing_error='',updated_at=now() WHERE workspace_id=$1 AND id=$2`, ws, idv)
	return err
}
func (r *Repository) CompleteProcessing(ctx context.Context, ws, idv uuid.UUID, title, text string, cover *string, duplicate *uuid.UUID) (domain.Item, error) {
	_, err := r.db.Exec(ctx, `UPDATE inbox_items SET title=$3,extracted_text=$4,cover_url=$5,duplicate_of=$6,processing_state='completed',processing_error='',updated_at=now() WHERE workspace_id=$1 AND id=$2`, ws, idv, title, text, cover, duplicate)
	if err != nil {
		return domain.Item{}, err
	}
	return r.Get(ctx, ws, idv)
}
func (r *Repository) FailProcessing(ctx context.Context, ws, idv uuid.UUID, message string) error {
	_, err := r.db.Exec(ctx, `UPDATE inbox_items SET processing_state='failed',processing_error=$3,updated_at=now() WHERE workspace_id=$1 AND id=$2`, ws, idv, message)
	return err
}
func (r *Repository) FindDuplicate(ctx context.Context, ws, exclude uuid.UUID, source *string, text string) (*uuid.UUID, error) {
	var result uuid.UUID
	err := r.db.QueryRow(ctx, `SELECT id FROM inbox_items WHERE workspace_id=$1 AND id<>$2 AND (($3::text IS NOT NULL AND source_url=$3) OR ($4<>'' AND md5(extracted_text)=md5($4))) ORDER BY created_at LIMIT 1`, ws, exclude, source, text).Scan(&result)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
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
	var rawText, extractedText string
	var sourceURL *string
	var existingContentID *uuid.UUID
	if err = tx.QueryRow(ctx, `SELECT state,raw_text,source_url,converted_content_id,extracted_text FROM inbox_items WHERE id=$1 AND workspace_id=$2 FOR UPDATE`, itemID, workspaceID).Scan(&inboxState, &rawText, &sourceURL, &existingContentID, &extractedText); err != nil {
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
	if extractedText != "" {
		if rawText != "" {
			rawText += "\n\n"
		}
		rawText += extractedText
	}
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
