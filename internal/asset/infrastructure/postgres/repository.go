package postgres

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ksamwang/PersonalContentPlatform/internal/asset/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/id"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/outbox"
)

type Repository struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Repository { return &Repository{db: db} }
func (r *Repository) CreateIntent(ctx context.Context, v domain.UploadIntent, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `INSERT INTO upload_intents(id,workspace_id,storage_key,filename,mime,expected_size,expires_at,created_by) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, v.ID, v.WorkspaceID, v.StorageKey, v.Filename, v.MIME, v.ExpectedSize, v.ExpiresAt, userID)
	return err
}
func (r *Repository) Intent(ctx context.Context, workspaceID, intentID uuid.UUID) (domain.UploadIntent, error) {
	var v domain.UploadIntent
	err := r.db.QueryRow(ctx, `SELECT id,workspace_id,storage_key,filename,mime,expected_size,expires_at FROM upload_intents WHERE id=$1 AND workspace_id=$2 AND completed_at IS NULL AND expires_at>now()`, intentID, workspaceID).Scan(&v.ID, &v.WorkspaceID, &v.StorageKey, &v.Filename, &v.MIME, &v.ExpectedSize, &v.ExpiresAt)
	return v, err
}
func (r *Repository) Complete(ctx context.Context, v domain.UploadIntent, userID uuid.UUID, sha string, size int64) (domain.Asset, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Asset{}, err
	}
	defer tx.Rollback(ctx)
	blobID := id.New()
	if err = tx.QueryRow(ctx, `INSERT INTO blobs(id,sha256,size,storage_key,mime) VALUES($1,$2,$3,$4,$5) ON CONFLICT(sha256,size) DO UPDATE SET sha256=EXCLUDED.sha256 RETURNING id`, blobID, sha, size, v.StorageKey, v.MIME).Scan(&blobID); err != nil {
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
	return domain.Asset{ID: assetID, WorkspaceID: v.WorkspaceID, BlobID: blobID, Filename: v.Filename, MediaType: mediaType, State: "ready", MIME: v.MIME, Size: size, SHA256: sha}, nil
}
func (r *Repository) List(ctx context.Context, workspaceID uuid.UUID, limit int) ([]domain.Asset, error) {
	rows, err := r.db.Query(ctx, `SELECT a.object_id,a.workspace_id,a.blob_id,a.filename,a.media_type,a.state,b.mime,b.size,b.sha256,a.created_at FROM assets a JOIN blobs b ON b.id=a.blob_id WHERE a.workspace_id=$1 ORDER BY a.created_at DESC LIMIT $2`, workspaceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.Asset{}
	for rows.Next() {
		var a domain.Asset
		if err = rows.Scan(&a.ID, &a.WorkspaceID, &a.BlobID, &a.Filename, &a.MediaType, &a.State, &a.MIME, &a.Size, &a.SHA256, &a.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, a)
	}
	return items, rows.Err()
}
