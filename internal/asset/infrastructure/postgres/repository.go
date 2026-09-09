package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ksamwang/PersonalContentPlatform/internal/asset/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/id"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/outbox"
)

type Repository struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Repository { return &Repository{db: db} }
func (r *Repository) CreateIntent(ctx context.Context, v domain.UploadIntent, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `INSERT INTO upload_intents(id,workspace_id,storage_key,filename,mime,expected_size,expires_at,created_by,storage_profile_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, v.ID, v.WorkspaceID, v.StorageKey, v.Filename, v.MIME, v.ExpectedSize, v.ExpiresAt, userID, v.StorageProfileID)
	return err
}
func (r *Repository) Intent(ctx context.Context, workspaceID, intentID uuid.UUID) (domain.UploadIntent, error) {
	var v domain.UploadIntent
	err := r.db.QueryRow(ctx, `SELECT id,workspace_id,storage_key,filename,mime,expected_size,expires_at,storage_profile_id FROM upload_intents WHERE id=$1 AND workspace_id=$2 AND completed_at IS NULL AND expires_at>now()`, intentID, workspaceID).Scan(&v.ID, &v.WorkspaceID, &v.StorageKey, &v.Filename, &v.MIME, &v.ExpectedSize, &v.ExpiresAt, &v.StorageProfileID)
	return v, err
}
func (r *Repository) Complete(ctx context.Context, v domain.UploadIntent, userID uuid.UUID, sha string, size int64) (domain.Asset, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Asset{}, err
	}
	defer tx.Rollback(ctx)
	blobID := id.New()
	if err = tx.QueryRow(ctx, `INSERT INTO blobs(id,sha256,size,storage_key,mime,storage_profile_id) VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT(sha256,size) DO UPDATE SET sha256=EXCLUDED.sha256 RETURNING id,storage_profile_id`, blobID, sha, size, v.StorageKey, v.MIME, v.StorageProfileID).Scan(&blobID, &v.StorageProfileID); err != nil {
		return domain.Asset{}, err
	}
	assetID := id.New()
	mediaType := "file"
	if len(v.MIME) > 6 && v.MIME[:6] == "image/" {
		mediaType = "image"
	}
	if _, err = tx.Exec(ctx, `INSERT INTO objects(id,workspace_id,kind) VALUES($1,$2,'asset')`, assetID, v.WorkspaceID); err != nil {
		return domain.Asset{}, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO assets(object_id,workspace_id,blob_id,media_type,state,filename) VALUES($1,$2,$3,$4,'ready',$5)`, assetID, v.WorkspaceID, blobID, mediaType, v.Filename); err != nil {
		return domain.Asset{}, err
	}
	if _, err = tx.Exec(ctx, `UPDATE upload_intents SET completed_at=now() WHERE id=$1 AND completed_at IS NULL`, v.ID); err != nil {
		return domain.Asset{}, err
	}
	if err = outbox.Add(ctx, tx, v.WorkspaceID, assetID, "asset", "AssetReady", map[string]any{"asset_id": assetID}); err != nil {
		return domain.Asset{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.Asset{}, err
	}
	return domain.Asset{ID: assetID, WorkspaceID: v.WorkspaceID, BlobID: blobID, StorageProfileID: v.StorageProfileID, Filename: v.Filename, MediaType: mediaType, State: "ready", MIME: v.MIME, Size: size, SHA256: sha}, nil
}
func (r *Repository) List(ctx context.Context, workspaceID uuid.UUID, limit int) ([]domain.Asset, error) {
	rows, err := r.db.Query(ctx, `SELECT a.object_id,a.workspace_id,a.blob_id,b.storage_profile_id,a.filename,a.media_type,a.state,b.mime,b.size,b.sha256,a.created_at,(SELECT count(*) FROM asset_usages u WHERE u.asset_id=a.object_id),a.archived_at,COALESCE((a.metadata_json->>'width')::int,0),COALESCE((a.metadata_json->>'height')::int,0),COALESCE((SELECT jsonb_agg(jsonb_build_object('recipe',v.recipe,'width',v.width,'height',v.height) ORDER BY v.recipe) FROM asset_variants v WHERE v.asset_id=a.object_id),'[]'::jsonb) FROM assets a JOIN blobs b ON b.id=a.blob_id WHERE a.workspace_id=$1 AND a.archived_at IS NULL ORDER BY a.created_at DESC LIMIT $2`, workspaceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.Asset{}
	for rows.Next() {
		var a domain.Asset
		var variants []byte
		if err = rows.Scan(&a.ID, &a.WorkspaceID, &a.BlobID, &a.StorageProfileID, &a.Filename, &a.MediaType, &a.State, &a.MIME, &a.Size, &a.SHA256, &a.CreatedAt, &a.UsageCount, &a.ArchivedAt, &a.Width, &a.Height, &variants); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(variants, &a.Variants)
		items = append(items, a)
	}
	return items, rows.Err()
}
func (r *Repository) Get(ctx context.Context, workspaceID, assetID uuid.UUID) (domain.Asset, error) {
	var a domain.Asset
	var variants []byte
	err := r.db.QueryRow(ctx, `SELECT a.object_id,a.workspace_id,a.blob_id,b.storage_profile_id,a.filename,a.media_type,a.state,b.mime,b.size,b.sha256,a.created_at,(SELECT count(*) FROM asset_usages u WHERE u.asset_id=a.object_id),a.archived_at,COALESCE((a.metadata_json->>'width')::int,0),COALESCE((a.metadata_json->>'height')::int,0),COALESCE((SELECT jsonb_agg(jsonb_build_object('recipe',v.recipe,'width',v.width,'height',v.height) ORDER BY v.recipe) FROM asset_variants v WHERE v.asset_id=a.object_id),'[]'::jsonb) FROM assets a JOIN blobs b ON b.id=a.blob_id WHERE a.workspace_id=$1 AND a.object_id=$2 AND a.archived_at IS NULL`, workspaceID, assetID).Scan(&a.ID, &a.WorkspaceID, &a.BlobID, &a.StorageProfileID, &a.Filename, &a.MediaType, &a.State, &a.MIME, &a.Size, &a.SHA256, &a.CreatedAt, &a.UsageCount, &a.ArchivedAt, &a.Width, &a.Height, &variants)
	_ = json.Unmarshal(variants, &a.Variants)
	return a, err
}
func (r *Repository) Usages(ctx context.Context, workspaceID, assetID uuid.UUID) ([]domain.Usage, error) {
	rows, err := r.db.Query(ctx, `SELECT u.id,u.owner_object_id,u.role,COALESCE((SELECT rv.title FROM contents c JOIN content_localizations l ON l.content_id=c.object_id AND l.locale=c.default_locale LEFT JOIN content_revisions rv ON rv.id=l.current_revision_id WHERE c.object_id=u.owner_object_id),'') FROM asset_usages u WHERE u.workspace_id=$1 AND u.asset_id=$2 ORDER BY u.created_at DESC`, workspaceID, assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.Usage{}
	for rows.Next() {
		var item domain.Usage
		if err = rows.Scan(&item.ID, &item.OwnerID, &item.Role, &item.OwnerTitle); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (r *Repository) Archive(ctx context.Context, workspaceID, assetID uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `UPDATE assets SET archived_at=now(),updated_at=now() WHERE workspace_id=$1 AND object_id=$2 AND archived_at IS NULL AND NOT EXISTS(SELECT 1 FROM asset_usages WHERE asset_id=$2)`, workspaceID, assetID)
	if err == nil && tag.RowsAffected() == 0 {
		return fmt.Errorf("asset is in use or was not found")
	}
	return err
}
func (r *Repository) Replace(ctx context.Context, workspaceID, assetID, replacementID uuid.UUID) (domain.Asset, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Asset{}, err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `UPDATE assets target SET blob_id=replacement.blob_id,media_type=replacement.media_type,state='ready',metadata_json=replacement.metadata_json,updated_at=now() FROM assets replacement WHERE target.workspace_id=$1 AND target.object_id=$2 AND target.archived_at IS NULL AND replacement.workspace_id=$1 AND replacement.object_id=$3 AND replacement.archived_at IS NULL`, workspaceID, assetID, replacementID)
	if err != nil {
		return domain.Asset{}, err
	}
	if tag.RowsAffected() == 0 {
		return domain.Asset{}, fmt.Errorf("asset was not found")
	}
	if _, err = tx.Exec(ctx, `UPDATE assets SET archived_at=now(),updated_at=now() WHERE workspace_id=$1 AND object_id=$2`, workspaceID, replacementID); err != nil {
		return domain.Asset{}, err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM asset_variants WHERE asset_id=$1`, assetID); err != nil {
		return domain.Asset{}, err
	}
	if _, err = tx.Exec(ctx, `UPDATE asset_variants SET asset_id=$1,workspace_id=$2 WHERE asset_id=$3`, assetID, workspaceID, replacementID); err != nil {
		return domain.Asset{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.Asset{}, err
	}
	return r.Get(ctx, workspaceID, assetID)
}
func (r *Repository) Object(ctx context.Context, assetID uuid.UUID) (domain.StoredObject, error) {
	var value domain.StoredObject
	err := r.db.QueryRow(ctx, `SELECT a.workspace_id,b.storage_profile_id,b.storage_key,b.mime,b.size FROM assets a JOIN blobs b ON b.id=a.blob_id WHERE a.object_id=$1 AND a.state='ready' AND a.archived_at IS NULL`, assetID).Scan(&value.WorkspaceID, &value.StorageProfileID, &value.StorageKey, &value.MIME, &value.Size)
	return value, err
}
func (r *Repository) VariantObject(ctx context.Context, assetID uuid.UUID, recipe string) (domain.StoredObject, error) {
	var value domain.StoredObject
	err := r.db.QueryRow(ctx, `SELECT a.workspace_id,b.storage_profile_id,b.storage_key,b.mime,b.size FROM asset_variants v JOIN assets a ON a.object_id=v.asset_id JOIN blobs b ON b.id=v.blob_id WHERE v.asset_id=$1 AND v.recipe=$2 AND a.archived_at IS NULL`, assetID, recipe).Scan(&value.WorkspaceID, &value.StorageProfileID, &value.StorageKey, &value.MIME, &value.Size)
	return value, err
}
func (r *Repository) SaveVariant(ctx context.Context, workspaceID, assetID uuid.UUID, profileID *uuid.UUID, recipe, key, sha string, size int64, width, height, originalWidth, originalHeight int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	blobID := id.New()
	if err = tx.QueryRow(ctx, `INSERT INTO blobs(id,sha256,size,storage_key,mime,storage_profile_id) VALUES($1,$2,$3,$4,'image/jpeg',$5) ON CONFLICT(sha256,size) DO UPDATE SET sha256=EXCLUDED.sha256 RETURNING id`, blobID, sha, size, key, profileID).Scan(&blobID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO asset_variants(id,workspace_id,asset_id,recipe,blob_id,width,height) VALUES($1,$2,$3,$4,$5,$6,$7) ON CONFLICT(asset_id,recipe) DO UPDATE SET blob_id=EXCLUDED.blob_id,width=EXCLUDED.width,height=EXCLUDED.height`, id.New(), workspaceID, assetID, recipe, blobID, width, height); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE assets SET metadata_json=metadata_json||jsonb_build_object('width',$1::int,'height',$2::int),updated_at=now() WHERE workspace_id=$3 AND object_id=$4`, originalWidth, originalHeight, workspaceID, assetID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
