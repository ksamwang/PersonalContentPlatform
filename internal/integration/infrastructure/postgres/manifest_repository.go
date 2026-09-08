package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ksamwang/PersonalContentPlatform/internal/integration/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/id"
)

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) BuildManifest(ctx context.Context, workspaceID uuid.UUID) (domain.Manifest, error) {
	manifest := emptyManifest()
	manifest.ExportedAt = time.Now().UTC()
	if err := r.db.QueryRow(ctx, `SELECT id,slug,name,settings_json FROM workspaces WHERE id=$1`, workspaceID).Scan(
		&manifest.Workspace.ID,
		&manifest.Workspace.Slug,
		&manifest.Workspace.Name,
		&manifest.Workspace.Settings,
	); err != nil {
		return manifest, err
	}

	queries := []func() error{
		func() error {
			return collect(ctx, r.db, &manifest.Contents, `SELECT object_id AS id,type,default_locale,visibility,created_at,updated_at FROM contents WHERE workspace_id=$1 ORDER BY created_at`, workspaceID)
		},
		func() error {
			return collect(ctx, r.db, &manifest.Localizations, `SELECT id,content_id,locale,state,slug,current_revision_id,translation_status,source_locale,source_revision_id,translated_from_hash FROM content_localizations WHERE workspace_id=$1 ORDER BY created_at`, workspaceID)
		},
		func() error {
			return collect(ctx, r.db, &manifest.Revisions, `SELECT id,localization_id,seq AS sequence,schema_version,title,summary,body_json AS body,metadata_json AS metadata,content_hash,created_at FROM content_revisions WHERE workspace_id=$1 ORDER BY localization_id,seq`, workspaceID)
		},
		func() error {
			return collect(ctx, r.db, &manifest.Drafts, `SELECT localization_id,version,title,summary,body_json AS body,metadata_json AS metadata,updated_at FROM draft_buffers WHERE workspace_id=$1 ORDER BY localization_id`, workspaceID)
		},
		func() error {
			return collect(ctx, r.db, &manifest.Assets, `SELECT a.object_id AS id,a.blob_id,a.filename,a.media_type,a.state,b.mime,b.size,b.sha256,b.storage_key,a.metadata_json AS metadata FROM assets a JOIN blobs b ON b.id=a.blob_id WHERE a.workspace_id=$1 ORDER BY a.created_at`, workspaceID)
		},
		func() error {
			return collect(ctx, r.db, &manifest.AssetVariants, `SELECT v.id,v.asset_id,v.recipe,v.blob_id,b.mime,b.size,b.sha256,b.storage_key,v.width,v.height FROM asset_variants v JOIN blobs b ON b.id=v.blob_id WHERE v.workspace_id=$1 ORDER BY v.created_at`, workspaceID)
		},
		func() error {
			return collect(ctx, r.db, &manifest.AssetUsages, `SELECT id,asset_id,owner_object_id,role,locator_json AS locator FROM asset_usages WHERE workspace_id=$1 ORDER BY created_at`, workspaceID)
		},
		func() error {
			return collect(ctx, r.db, &manifest.Entities, `SELECT object_id AS id,type,canonical_name,description FROM entities WHERE workspace_id=$1 ORDER BY canonical_name`, workspaceID)
		},
		func() error {
			return collect(ctx, r.db, &manifest.EntityAliases, `SELECT id,entity_id,alias,locale FROM entity_aliases WHERE workspace_id=$1 ORDER BY entity_id,locale,alias`, workspaceID)
		},
		func() error {
			return collect(ctx, r.db, &manifest.Predicates, `SELECT id,key,label_zh,label_en FROM relation_predicates WHERE workspace_id=$1 ORDER BY key`, workspaceID)
		},
		func() error {
			return collect(ctx, r.db, &manifest.Relations, `SELECT id,source_object_id AS source_id,target_object_id AS target_id,predicate_id,confirmed,provenance FROM relations WHERE workspace_id=$1 ORDER BY created_at`, workspaceID)
		},
		func() error {
			return collect(ctx, r.db, &manifest.Publications, `SELECT p.id,p.content_id,p.locale,p.revision_id,t.channel,t.name AS target_name,p.state,p.scheduled_at,p.published_at FROM publications p JOIN publication_targets t ON t.id=p.target_id WHERE p.workspace_id=$1 ORDER BY p.created_at`, workspaceID)
		},
	}
	for _, query := range queries {
		if err := query(); err != nil {
			return manifest, err
		}
	}
	return manifest, nil
}

func (r *Repository) StartExport(ctx context.Context, workspaceID, actorID uuid.UUID) (uuid.UUID, error) {
	runID := id.New()
	_, err := r.db.Exec(ctx, `INSERT INTO export_runs(id,workspace_id,format,state,created_by) VALUES($1,$2,'manifest-json','running',$3)`, runID, workspaceID, actorID)
	return runID, err
}

func (r *Repository) CompleteExport(ctx context.Context, runID uuid.UUID, manifest domain.Manifest, buildErr error) error {
	if buildErr != nil {
		_, err := r.db.Exec(ctx, `UPDATE export_runs SET state='failed',completed_at=now() WHERE id=$1`, runID)
		return err
	}
	payload, err := json.Marshal(manifest)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `UPDATE export_runs SET state='succeeded',manifest_json=$2,completed_at=now() WHERE id=$1`, runID, payload)
	return err
}

type queryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func collect[T any](ctx context.Context, db queryer, destination *[]T, query string, args ...any) error {
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return err
	}
	values, err := pgx.CollectRows(rows, pgx.RowToStructByName[T])
	if err != nil {
		return err
	}
	*destination = values
	return nil
}

func emptyManifest() domain.Manifest {
	return domain.Manifest{
		SchemaVersion: 1,
		Contents:      []domain.Content{},
		Localizations: []domain.Localization{},
		Revisions:     []domain.Revision{},
		Drafts:        []domain.Draft{},
		Assets:        []domain.Asset{},
		AssetVariants: []domain.AssetVariant{},
		AssetUsages:   []domain.AssetUsage{},
		Entities:      []domain.Entity{},
		EntityAliases: []domain.EntityAlias{},
		Predicates:    []domain.Predicate{},
		Relations:     []domain.Relation{},
		Publications:  []domain.Publication{},
	}
}
