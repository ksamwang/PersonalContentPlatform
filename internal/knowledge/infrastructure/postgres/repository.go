package postgres

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ksamwang/PersonalContentPlatform/internal/knowledge/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/id"
)

type Repository struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Repository { return &Repository{db: db} }
func (r *Repository) CreateEntity(ctx context.Context, ws uuid.UUID, kind, name, description string) (domain.Entity, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Entity{}, err
	}
	defer tx.Rollback(ctx)
	entityID := id.New()
	if _, err = tx.Exec(ctx, `INSERT INTO objects(id,workspace_id,kind) VALUES($1,$2,'entity')`, entityID, ws); err != nil {
		return domain.Entity{}, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO entities(object_id,workspace_id,type,canonical_name,description) VALUES($1,$2,$3,$4,$5)`, entityID, ws, kind, name, description); err != nil {
		return domain.Entity{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.Entity{}, err
	}
	return domain.Entity{ID: entityID, Type: kind, CanonicalName: name, Description: description}, nil
}
func (r *Repository) ListEntities(ctx context.Context, ws uuid.UUID, q string, limit int) ([]domain.Entity, error) {
	rows, err := r.db.Query(ctx, `SELECT object_id,type,canonical_name,description,created_at FROM entities WHERE workspace_id=$1 AND ($2='' OR canonical_name ILIKE '%'||$2||'%' OR description ILIKE '%'||$2||'%') ORDER BY canonical_name LIMIT $3`, ws, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.Entity{}
	for rows.Next() {
		var e domain.Entity
		if err = rows.Scan(&e.ID, &e.Type, &e.CanonicalName, &e.Description, &e.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, e)
	}
	return items, rows.Err()
}
func (r *Repository) AddAlias(ctx context.Context, ws, entityID uuid.UUID, value, locale string) error {
	_, err := r.db.Exec(ctx, `INSERT INTO entity_aliases(id,workspace_id,entity_id,alias,locale) SELECT $1,$2,e.object_id,$3,$4 FROM entities e WHERE e.object_id=$5 AND e.workspace_id=$2`, id.New(), ws, value, locale, entityID)
	return err
}
func (r *Repository) CreateRelation(ctx context.Context, ws, source uuid.UUID, key string, target uuid.UUID, confirmed bool) (domain.Relation, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Relation{}, err
	}
	defer tx.Rollback(ctx)
	predicateID := id.New()
	if err = tx.QueryRow(ctx, `INSERT INTO relation_predicates(id,workspace_id,key,label_zh,label_en) VALUES($1,$2,$3,$3,$3) ON CONFLICT(workspace_id,key) DO UPDATE SET key=EXCLUDED.key RETURNING id`, predicateID, ws, key).Scan(&predicateID); err != nil {
		return domain.Relation{}, err
	}
	relationID := id.New()
	if _, err = tx.Exec(ctx, `INSERT INTO relations(id,workspace_id,source_object_id,predicate_id,target_object_id,confirmed) VALUES($1,$2,$3,$4,$5,$6)`, relationID, ws, source, predicateID, target, confirmed); err != nil {
		return domain.Relation{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.Relation{}, err
	}
	return domain.Relation{ID: relationID, SourceID: source, TargetID: target, PredicateKey: key, Confirmed: confirmed}, nil
}
func (r *Repository) ConfirmRelation(ctx context.Context, ws, idValue uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE relations SET confirmed=true WHERE id=$1 AND workspace_id=$2`, idValue, ws)
	return err
}
func (r *Repository) Relations(ctx context.Context, ws, objectID uuid.UUID) ([]domain.Relation, error) {
	rows, err := r.db.Query(ctx, `SELECT r.id,r.source_object_id,r.target_object_id,p.key,r.confirmed,r.created_at FROM relations r JOIN relation_predicates p ON p.id=r.predicate_id WHERE r.workspace_id=$1 AND (r.source_object_id=$2 OR r.target_object_id=$2) ORDER BY r.created_at DESC`, ws, objectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.Relation{}
	for rows.Next() {
		var v domain.Relation
		if err = rows.Scan(&v.ID, &v.SourceID, &v.TargetID, &v.PredicateKey, &v.Confirmed, &v.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, v)
	}
	return items, rows.Err()
}
