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
	"github.com/ksamwang/PersonalContentPlatform/internal/content/application"
	"github.com/ksamwang/PersonalContentPlatform/internal/content/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/id"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/outbox"
	"strings"
	"time"
)

type Repository struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

func (r *Repository) Create(ctx context.Context, workspaceID, userID uuid.UUID, kind domain.Type, locale, slug, title string) (domain.Content, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Content{}, err
	}
	defer tx.Rollback(ctx)
	contentID, localizationID, revisionID := id.New(), id.New(), id.New()
	body := json.RawMessage(`{"schemaVersion":1,"type":"doc","content":[]}`)
	metadata := json.RawMessage(`{}`)
	hash := revisionHash(title, "", body, metadata)
	if _, err = tx.Exec(ctx, `INSERT INTO objects(id,workspace_id,kind) VALUES($1,$2,'content')`, contentID, workspaceID); err != nil {
		return domain.Content{}, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO contents(object_id,workspace_id,type,default_locale) VALUES($1,$2,$3,$4)`, contentID, workspaceID, kind, locale); err != nil {
		return domain.Content{}, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO content_localizations(id,workspace_id,content_id,locale,slug) VALUES($1,$2,$3,$4,$5)`, localizationID, workspaceID, contentID, locale, slug); err != nil {
		return domain.Content{}, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO content_revisions(id,workspace_id,localization_id,seq,title,body_json,metadata_json,content_hash,created_by) VALUES($1,$2,$3,1,$4,$5,$6,$7,$8)`, revisionID, workspaceID, localizationID, title, body, metadata, hash, userID); err != nil {
		return domain.Content{}, err
	}
	if _, err = tx.Exec(ctx, `UPDATE content_localizations SET current_revision_id=$1 WHERE id=$2`, revisionID, localizationID); err != nil {
		return domain.Content{}, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO draft_buffers(localization_id,workspace_id,title,body_json,metadata_json,updated_by) VALUES($1,$2,$3,$4,$5,$6)`, localizationID, workspaceID, title, body, metadata, userID); err != nil {
		return domain.Content{}, err
	}
	if err = outbox.Add(ctx, tx, workspaceID, contentID, "content", "ContentCreated", map[string]any{"content_id": contentID, "revision_id": revisionID, "locale": locale}); err != nil {
		return domain.Content{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.Content{}, err
	}
	return r.Get(ctx, workspaceID, contentID)
}

func (r *Repository) CreateLocalization(ctx context.Context, workspaceID, userID, contentID uuid.UUID, locale, slug, sourceLocale string) (domain.Content, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Content{}, err
	}
	defer tx.Rollback(ctx)
	var sourceRevisionID *uuid.UUID
	if err = tx.QueryRow(ctx, `SELECT l.current_revision_id FROM content_localizations l WHERE l.workspace_id=$1 AND l.content_id=$2 AND l.locale=$3 AND l.deleted_at IS NULL`, workspaceID, contentID, sourceLocale).Scan(&sourceRevisionID); err != nil {
		return domain.Content{}, err
	}
	localizationID := id.New()
	body := json.RawMessage(`{"schemaVersion":1,"type":"doc","content":[]}`)
	if _, err = tx.Exec(ctx, `INSERT INTO content_localizations(id,workspace_id,content_id,locale,slug,translation_status,source_locale,source_revision_id) VALUES($1,$2,$3,$4,$5,'missing',$6,$7)`, localizationID, workspaceID, contentID, locale, slug, sourceLocale, sourceRevisionID); err != nil {
		return domain.Content{}, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO draft_buffers(localization_id,workspace_id,body_json,updated_by) VALUES($1,$2,$3,$4)`, localizationID, workspaceID, body, userID); err != nil {
		return domain.Content{}, err
	}
	if _, err = tx.Exec(ctx, `UPDATE contents SET updated_at=now() WHERE object_id=$1 AND workspace_id=$2`, contentID, workspaceID); err != nil {
		return domain.Content{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.Content{}, err
	}
	return r.Get(ctx, workspaceID, contentID)
}

func (r *Repository) List(ctx context.Context, workspaceID uuid.UUID, filter domain.ListFilter) ([]domain.Content, error) {
	query := `SELECT c.object_id,c.workspace_id,c.type,c.default_locale,c.visibility,c.created_at,c.updated_at,l.id,l.locale,l.state,l.slug,l.translation_status,l.updated_at,rv.id,rv.seq,rv.schema_version,rv.title,rv.summary,rv.body_json,rv.metadata_json,rv.content_hash,rv.created_at FROM contents c JOIN content_localizations l ON l.content_id=c.object_id LEFT JOIN content_revisions rv ON rv.id=l.current_revision_id WHERE c.workspace_id=$1 AND c.deleted_at IS NULL AND l.deleted_at IS NULL`
	args := []any{workspaceID}
	if filter.Locale != "" {
		args = append(args, filter.Locale)
		query += fmt.Sprintf(" AND l.locale=$%d", len(args))
	}
	if filter.State != "" {
		args = append(args, filter.State)
		query += fmt.Sprintf(" AND l.state=$%d", len(args))
	}
	if filter.Type != "" {
		args = append(args, filter.Type)
		query += fmt.Sprintf(" AND c.type=$%d", len(args))
	}
	if filter.Visibility != "" {
		args = append(args, filter.Visibility)
		query += fmt.Sprintf(" AND c.visibility=$%d", len(args))
	}
	if filter.Tag != "" {
		args = append(args, filter.Tag)
		query += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM jsonb_array_elements_text(CASE WHEN jsonb_typeof(rv.metadata_json->'tags')='array' THEN rv.metadata_json->'tags' ELSE '[]'::jsonb END) tag(value) WHERE lower(tag.value)=lower($%d))", len(args))
	}
	args = append(args, filter.Limit)
	query += fmt.Sprintf(" ORDER BY c.updated_at DESC LIMIT $%d", len(args))
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := scanContentRows(rows)
	return aggregateContents(items), err
}
func (r *Repository) Get(ctx context.Context, workspaceID, contentID uuid.UUID) (domain.Content, error) {
	rows, err := r.db.Query(ctx, `SELECT c.object_id,c.workspace_id,c.type,c.default_locale,c.visibility,c.created_at,c.updated_at,l.id,l.locale,l.state,l.slug,l.translation_status,l.updated_at,rv.id,rv.seq,rv.schema_version,rv.title,rv.summary,rv.body_json,rv.metadata_json,rv.content_hash,rv.created_at FROM contents c JOIN content_localizations l ON l.content_id=c.object_id LEFT JOIN content_revisions rv ON rv.id=l.current_revision_id WHERE c.workspace_id=$1 AND c.object_id=$2 AND c.deleted_at IS NULL AND l.deleted_at IS NULL ORDER BY l.locale`, workspaceID, contentID)
	if err != nil {
		return domain.Content{}, err
	}
	defer rows.Close()
	items, err := scanContentRows(rows)
	if err != nil {
		return domain.Content{}, err
	}
	if len(items) == 0 {
		return domain.Content{}, pgx.ErrNoRows
	}
	result := items[0]
	for _, item := range items[1:] {
		result.Localizations = append(result.Localizations, item.Localizations...)
	}
	return result, nil
}

func (r *Repository) Archive(ctx context.Context, workspaceID, contentID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `UPDATE content_localizations SET state='archived',updated_at=now() WHERE workspace_id=$1 AND content_id=$2 AND deleted_at IS NULL AND state<>'archived'`, workspaceID, contentID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	if _, err = tx.Exec(ctx, `UPDATE contents SET updated_at=now() WHERE workspace_id=$1 AND object_id=$2 AND deleted_at IS NULL`, workspaceID, contentID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM search_documents WHERE workspace_id=$1 AND object_id=$2`, workspaceID, contentID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM content_chunks WHERE workspace_id=$1 AND content_id=$2`, workspaceID, contentID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM publication_views WHERE workspace_id=$1 AND content_id=$2`, workspaceID, contentID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) Restore(ctx context.Context, workspaceID, contentID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `UPDATE content_localizations SET state='draft',translation_status=CASE WHEN source_locale IS NULL THEN 'draft' WHEN translation_status='published' THEN 'needs_review' ELSE translation_status END,updated_at=now() WHERE workspace_id=$1 AND content_id=$2 AND deleted_at IS NULL AND state='archived'`, workspaceID, contentID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	if _, err = tx.Exec(ctx, `UPDATE contents SET updated_at=now() WHERE workspace_id=$1 AND object_id=$2 AND deleted_at IS NULL`, workspaceID, contentID); err != nil {
		return err
	}
	if err = outbox.Add(ctx, tx, workspaceID, contentID, "content", "ContentRevisionCreated", map[string]any{"content_id": contentID, "reason": "restored"}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) SoftDelete(ctx context.Context, workspaceID, contentID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `UPDATE contents SET deleted_at=now(),updated_at=now() WHERE workspace_id=$1 AND object_id=$2 AND deleted_at IS NULL`, workspaceID, contentID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	if _, err = tx.Exec(ctx, `UPDATE content_localizations SET deleted_at=now(),updated_at=now() WHERE workspace_id=$1 AND content_id=$2 AND deleted_at IS NULL`, workspaceID, contentID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM search_documents WHERE workspace_id=$1 AND object_id=$2`, workspaceID, contentID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM content_chunks WHERE workspace_id=$1 AND content_id=$2`, workspaceID, contentID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM publication_views WHERE workspace_id=$1 AND content_id=$2`, workspaceID, contentID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (r *Repository) UpdateProperties(ctx context.Context, workspaceID, contentID, localizationID uuid.UUID, slug string, visibility domain.Visibility) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `UPDATE content_localizations SET slug=$1,updated_at=now() WHERE id=$2 AND content_id=$3 AND workspace_id=$4 AND deleted_at IS NULL`, slug, localizationID, contentID, workspaceID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	if _, err = tx.Exec(ctx, `UPDATE contents SET visibility=$1,updated_at=now() WHERE object_id=$2 AND workspace_id=$3`, visibility, contentID, workspaceID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (r *Repository) ReadinessSource(ctx context.Context, workspaceID, localizationID uuid.UUID) (domain.ReadinessSource, error) {
	var source domain.ReadinessSource
	err := r.db.QueryRow(ctx, `SELECT c.type,l.slug,r.title,r.summary,r.body_json,r.metadata_json FROM content_localizations l JOIN contents c ON c.object_id=l.content_id JOIN content_revisions r ON r.id=l.current_revision_id WHERE l.id=$1 AND l.workspace_id=$2 AND l.deleted_at IS NULL`, localizationID, workspaceID).Scan(&source.Type, &source.Slug, &source.Title, &source.Summary, &source.Body, &source.Metadata)
	return source, err
}
func (r *Repository) GetDraft(ctx context.Context, workspaceID, localizationID uuid.UUID) (domain.Draft, error) {
	var d domain.Draft
	err := r.db.QueryRow(ctx, `SELECT localization_id,version,title,summary,body_json,metadata_json,updated_at FROM draft_buffers WHERE workspace_id=$1 AND localization_id=$2`, workspaceID, localizationID).Scan(&d.LocalizationID, &d.Version, &d.Title, &d.Summary, &d.Body, &d.Metadata, &d.UpdatedAt)
	return d, err
}
func scanContentRows(rows pgx.Rows) ([]domain.Content, error) {
	result := []domain.Content{}
	for rows.Next() {
		var c domain.Content
		var l domain.Localization
		var rv domain.Revision
		var revisionID *uuid.UUID
		var seq, schema *int
		var title, summary, hash *string
		var body, metadata []byte
		var revisionCreated *json.RawMessage
		_ = revisionCreated
		var revisionTime *time.Time
		if err := rows.Scan(&c.ID, &c.WorkspaceID, &c.Type, &c.DefaultLocale, &c.Visibility, &c.CreatedAt, &c.UpdatedAt, &l.ID, &l.Locale, &l.State, &l.Slug, &l.TranslationStatus, &l.UpdatedAt, &revisionID, &seq, &schema, &title, &summary, &body, &metadata, &hash, &revisionTime); err != nil {
			return nil, err
		}
		l.ContentID = c.ID
		if revisionID != nil {
			rv.ID = *revisionID
			rv.LocalizationID = l.ID
			rv.Seq = *seq
			rv.SchemaVersion = *schema
			rv.Title = *title
			rv.Summary = *summary
			rv.Body = body
			rv.Metadata = metadata
			rv.ContentHash = *hash
			if revisionTime != nil {
				rv.CreatedAt = *revisionTime
			}
			l.CurrentRevision = &rv
		}
		c.Localizations = []domain.Localization{l}
		result = append(result, c)
	}
	return result, rows.Err()
}
func aggregateContents(items []domain.Content) []domain.Content {
	result := make([]domain.Content, 0, len(items))
	positions := map[uuid.UUID]int{}
	for _, item := range items {
		if position, ok := positions[item.ID]; ok {
			result[position].Localizations = append(result[position].Localizations, item.Localizations...)
			continue
		}
		positions[item.ID] = len(result)
		result = append(result, item)
	}
	return result
}

func (r *Repository) SaveDraft(ctx context.Context, workspaceID, userID, localizationID uuid.UUID, expected int, title, summary string, body, metadata json.RawMessage) (domain.Draft, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Draft{}, err
	}
	defer tx.Rollback(ctx)
	var d domain.Draft
	err = tx.QueryRow(ctx, `UPDATE draft_buffers SET version=version+1,title=$1,summary=$2,body_json=$3,metadata_json=$4,updated_by=$5,updated_at=now() WHERE localization_id=$6 AND workspace_id=$7 AND version=$8 RETURNING localization_id,version,title,summary,body_json,metadata_json,updated_at`, title, summary, body, metadata, userID, localizationID, workspaceID, expected).Scan(&d.LocalizationID, &d.Version, &d.Title, &d.Summary, &d.Body, &d.Metadata, &d.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return d, application.ErrConflict
	}
	if err != nil {
		return d, err
	}
	var contentID uuid.UUID
	if err = tx.QueryRow(ctx, `SELECT content_id FROM content_localizations WHERE workspace_id=$1 AND id=$2`, workspaceID, localizationID).Scan(&contentID); err != nil {
		return d, err
	}
	if err = syncAssetUsages(ctx, tx, workspaceID, contentID); err != nil {
		return d, err
	}
	return d, tx.Commit(ctx)
}

func syncAssetUsages(ctx context.Context, tx pgx.Tx, workspaceID, contentID uuid.UUID) error {
	rows, err := tx.Query(ctx, `SELECT d.body_json,d.metadata_json FROM draft_buffers d JOIN content_localizations l ON l.id=d.localization_id WHERE l.workspace_id=$1 AND l.content_id=$2 AND l.deleted_at IS NULL`, workspaceID, contentID)
	if err != nil {
		return err
	}
	assets := map[uuid.UUID]string{}
	for rows.Next() {
		var body, metadata []byte
		if err = rows.Scan(&body, &metadata); err != nil {
			rows.Close()
			return err
		}
		collectAssetIDs(body, "body", assets)
		collectAssetIDs(metadata, "cover", assets)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM asset_usages WHERE workspace_id=$1 AND owner_object_id=$2`, workspaceID, contentID); err != nil {
		return err
	}
	for assetID, role := range assets {
		_, err = tx.Exec(ctx, `INSERT INTO asset_usages(id,workspace_id,asset_id,owner_object_id,role,locator_json) SELECT $1,$2,a.object_id,$3,$4,'{}'::jsonb FROM assets a WHERE a.workspace_id=$2 AND a.object_id=$5 AND a.archived_at IS NULL`, id.New(), workspaceID, contentID, role, assetID)
		if err != nil {
			return err
		}
	}
	return nil
}
func collectAssetIDs(raw []byte, role string, result map[uuid.UUID]string) {
	var value any
	if json.Unmarshal(raw, &value) != nil {
		return
	}
	var visit func(any)
	visit = func(current any) {
		switch node := current.(type) {
		case map[string]any:
			if role == "cover" {
				if rawID, ok := node["cover_asset_id"].(string); ok {
					if assetID, err := uuid.Parse(rawID); err == nil {
						result[assetID] = "cover"
					}
				}
			}
			if node["type"] == "image" || node["type"] == "asset" {
				if attrs, ok := node["attrs"].(map[string]any); ok {
					rawID, _ := attrs["assetId"].(string)
					if rawID == "" {
						if src, ok := attrs["src"].(string); ok {
							rawID = strings.TrimPrefix(src, "/media/")
							rawID = strings.SplitN(rawID, "/", 2)[0]
						}
					}
					if assetID, err := uuid.Parse(rawID); err == nil {
						if result[assetID] != "cover" {
							result[assetID] = "body"
						}
					}
				}
			}
			for _, child := range node {
				visit(child)
			}
		case []any:
			for _, child := range node {
				visit(child)
			}
		}
	}
	visit(value)
}
func (r *Repository) SealRevision(ctx context.Context, workspaceID, userID, localizationID uuid.UUID, expected int) (domain.Revision, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Revision{}, err
	}
	defer tx.Rollback(ctx)
	var d domain.Draft
	if err = tx.QueryRow(ctx, `SELECT localization_id,version,title,summary,body_json,metadata_json,updated_at FROM draft_buffers WHERE localization_id=$1 AND workspace_id=$2 FOR UPDATE`, localizationID, workspaceID).Scan(&d.LocalizationID, &d.Version, &d.Title, &d.Summary, &d.Body, &d.Metadata, &d.UpdatedAt); err != nil {
		return domain.Revision{}, err
	}
	if d.Version != expected {
		return domain.Revision{}, application.ErrConflict
	}
	var contentID uuid.UUID
	var locale string
	var nextSeq int
	if err = tx.QueryRow(ctx, `SELECT l.content_id,l.locale,COALESCE(MAX(r.seq),0)+1 FROM content_localizations l LEFT JOIN content_revisions r ON r.localization_id=l.id WHERE l.id=$1 AND l.workspace_id=$2 GROUP BY l.content_id,l.locale`, localizationID, workspaceID).Scan(&contentID, &locale, &nextSeq); err != nil {
		return domain.Revision{}, err
	}
	revisionID := id.New()
	hash := revisionHash(d.Title, d.Summary, d.Body, d.Metadata)
	if _, err = tx.Exec(ctx, `INSERT INTO content_revisions(id,workspace_id,localization_id,seq,title,summary,body_json,metadata_json,content_hash,created_by) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, revisionID, workspaceID, localizationID, nextSeq, d.Title, d.Summary, d.Body, d.Metadata, hash, userID); err != nil {
		return domain.Revision{}, err
	}
	if _, err = tx.Exec(ctx, `UPDATE content_localizations SET current_revision_id=$1,translation_status=CASE WHEN source_locale IS NOT NULL AND translation_status='missing' THEN 'draft' ELSE translation_status END,updated_at=now() WHERE id=$2`, revisionID, localizationID); err != nil {
		return domain.Revision{}, err
	}
	if _, err = tx.Exec(ctx, `UPDATE contents SET updated_at=now() WHERE object_id=$1`, contentID); err != nil {
		return domain.Revision{}, err
	}
	if _, err = tx.Exec(ctx, `UPDATE content_localizations SET translation_status='outdated',updated_at=now() WHERE content_id=$1 AND source_locale=$2 AND id<>$3 AND translation_status IN ('draft','needs_review','ready','published')`, contentID, locale, localizationID); err != nil {
		return domain.Revision{}, err
	}
	if err = outbox.Add(ctx, tx, workspaceID, contentID, "content", "ContentRevisionCreated", map[string]any{"content_id": contentID, "revision_id": revisionID}); err != nil {
		return domain.Revision{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.Revision{}, err
	}
	return domain.Revision{ID: revisionID, LocalizationID: localizationID, Seq: nextSeq, SchemaVersion: 1, Title: d.Title, Summary: d.Summary, Body: d.Body, Metadata: d.Metadata, ContentHash: hash}, nil
}
func revisionHash(title, summary string, body, metadata []byte) string {
	sum := sha256.Sum256(append(append(append([]byte(title+"\n"+summary+"\n"), body...), byte('\n')), metadata...))
	return hex.EncodeToString(sum[:])
}
func (r *Repository) MarkReady(ctx context.Context, workspaceID, localizationID uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `UPDATE content_localizations SET state='ready',translation_status='ready',updated_at=now() WHERE id=$1 AND workspace_id=$2 AND current_revision_id IS NOT NULL`, localizationID, workspaceID)
	if err == nil && tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return err
}
func (r *Repository) Publish(ctx context.Context, workspaceID, userID, contentID uuid.UUID, locale string, targetID uuid.UUID, channel string) (uuid.UUID, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)
	var revisionID uuid.UUID
	var state string
	if err = tx.QueryRow(ctx, `SELECT current_revision_id,state FROM content_localizations WHERE workspace_id=$1 AND content_id=$2 AND locale=$3 FOR UPDATE`, workspaceID, contentID, locale).Scan(&revisionID, &state); err != nil {
		return uuid.Nil, err
	}
	if state != "ready" && state != "published" {
		return uuid.Nil, fmt.Errorf("content localization is not ready")
	}
	if targetID == uuid.Nil {
		if err = tx.QueryRow(ctx, `INSERT INTO publication_targets(id,workspace_id,channel,name) VALUES($1,$2,$3,$4) ON CONFLICT(workspace_id,channel,name) DO UPDATE SET enabled=true RETURNING id`, id.New(), workspaceID, channel, "Primary "+channel).Scan(&targetID); err != nil {
			return uuid.Nil, err
		}
	}
	publicationID := id.New()
	if err = tx.QueryRow(ctx, `INSERT INTO publications(id,workspace_id,content_id,locale,revision_id,target_id,state,created_by) VALUES($1,$2,$3,$4,$5,$6,'queued',$7) ON CONFLICT(target_id,revision_id) DO UPDATE SET state='queued',updated_at=now() RETURNING id`, publicationID, workspaceID, contentID, locale, revisionID, targetID, userID).Scan(&publicationID); err != nil {
		return uuid.Nil, err
	}
	if err = outbox.Add(ctx, tx, workspaceID, publicationID, "publication", "PublicationRequested", map[string]any{"publication_id": publicationID}); err != nil {
		return uuid.Nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	return publicationID, nil
}
func (r *Repository) Search(ctx context.Context, workspaceID uuid.UUID, locale, query string, limit int) ([]domain.Content, error) {
	rows, err := r.db.Query(ctx, `SELECT c.object_id,c.workspace_id,c.type,c.default_locale,c.visibility,c.created_at,c.updated_at,l.id,l.locale,l.state,l.slug,l.translation_status,l.updated_at,rv.id,rv.seq,rv.schema_version,rv.title,rv.summary,rv.body_json,rv.metadata_json,rv.content_hash,rv.created_at FROM search_documents s JOIN contents c ON c.object_id=s.object_id JOIN content_localizations l ON l.content_id=c.object_id AND l.locale=s.locale JOIN content_revisions rv ON rv.id=s.revision_id WHERE s.workspace_id=$1 AND ($2='' OR s.locale=$2) AND ($3='' OR s.tsv @@ websearch_to_tsquery('simple',$3)) ORDER BY CASE WHEN $3='' THEN 0 ELSE ts_rank(s.tsv,websearch_to_tsquery('simple',$3)) END DESC,c.updated_at DESC LIMIT $4`, workspaceID, locale, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := scanContentRows(rows)
	return aggregateContents(items), err
}

func (r *Repository) ListRevisions(ctx context.Context, workspaceID, localizationID uuid.UUID, limit int) ([]domain.Revision, error) {
	rows, err := r.db.Query(ctx, `SELECT id,localization_id,seq,schema_version,title,summary,body_json,metadata_json,content_hash,created_at FROM content_revisions WHERE workspace_id=$1 AND localization_id=$2 ORDER BY seq DESC LIMIT $3`, workspaceID, localizationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.Revision{}
	for rows.Next() {
		var revision domain.Revision
		if err := rows.Scan(&revision.ID, &revision.LocalizationID, &revision.Seq, &revision.SchemaVersion, &revision.Title, &revision.Summary, &revision.Body, &revision.Metadata, &revision.ContentHash, &revision.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, revision)
	}
	return items, rows.Err()
}
func (r *Repository) GetRevision(ctx context.Context, workspaceID, localizationID, revisionID uuid.UUID) (domain.Revision, error) {
	var revision domain.Revision
	err := r.db.QueryRow(ctx, `SELECT id,localization_id,seq,schema_version,title,summary,body_json,metadata_json,content_hash,created_at FROM content_revisions WHERE workspace_id=$1 AND localization_id=$2 AND id=$3`, workspaceID, localizationID, revisionID).Scan(&revision.ID, &revision.LocalizationID, &revision.Seq, &revision.SchemaVersion, &revision.Title, &revision.Summary, &revision.Body, &revision.Metadata, &revision.ContentHash, &revision.CreatedAt)
	return revision, err
}
func (r *Repository) RestoreRevision(ctx context.Context, workspaceID, userID, localizationID, revisionID uuid.UUID, expectedVersion int) (domain.Draft, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Draft{}, err
	}
	defer tx.Rollback(ctx)
	var revision domain.Revision
	if err = tx.QueryRow(ctx, `SELECT title,summary,body_json,metadata_json FROM content_revisions WHERE workspace_id=$1 AND localization_id=$2 AND id=$3`, workspaceID, localizationID, revisionID).Scan(&revision.Title, &revision.Summary, &revision.Body, &revision.Metadata); err != nil {
		return domain.Draft{}, err
	}
	var draft domain.Draft
	err = tx.QueryRow(ctx, `UPDATE draft_buffers SET version=version+1,title=$1,summary=$2,body_json=$3,metadata_json=$4,updated_by=$5,updated_at=now() WHERE workspace_id=$6 AND localization_id=$7 AND version=$8 RETURNING localization_id,version,title,summary,body_json,metadata_json,updated_at`, revision.Title, revision.Summary, revision.Body, revision.Metadata, userID, workspaceID, localizationID, expectedVersion).Scan(&draft.LocalizationID, &draft.Version, &draft.Title, &draft.Summary, &draft.Body, &draft.Metadata, &draft.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Draft{}, application.ErrConflict
	}
	if err != nil {
		return domain.Draft{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.Draft{}, err
	}
	return draft, nil
}
func (r *Repository) CreatePreview(ctx context.Context, workspaceID, userID, localizationID uuid.UUID, tokenHash []byte, expiresAt time.Time) error {
	tag, err := r.db.Exec(ctx, `INSERT INTO preview_tokens(token_hash,workspace_id,localization_id,content_type,locale,title,summary,body_json,metadata_json,created_by,expires_at) SELECT $1,d.workspace_id,d.localization_id,c.type,l.locale,d.title,d.summary,d.body_json,d.metadata_json,$2,$3 FROM draft_buffers d JOIN content_localizations l ON l.id=d.localization_id JOIN contents c ON c.object_id=l.content_id WHERE d.workspace_id=$4 AND d.localization_id=$5`, tokenHash, userID, expiresAt, workspaceID, localizationID)
	if err == nil && tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return err
}
func (r *Repository) GetPreview(ctx context.Context, tokenHash []byte) (domain.Preview, error) {
	var preview domain.Preview
	err := r.db.QueryRow(ctx, `SELECT content_type,locale,title,summary,body_json,metadata_json,expires_at FROM preview_tokens WHERE token_hash=$1 AND expires_at>now()`, tokenHash).Scan(&preview.Type, &preview.Locale, &preview.Title, &preview.Summary, &preview.Body, &preview.Metadata, &preview.ExpiresAt)
	return preview, err
}
