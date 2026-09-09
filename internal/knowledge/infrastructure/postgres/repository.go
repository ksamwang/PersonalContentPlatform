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
	if err = rows.Err(); err != nil {
		return nil, err
	}
	for i := range items {
		aliasRows, queryErr := r.db.Query(ctx, `SELECT id,alias,locale FROM entity_aliases WHERE workspace_id=$1 AND entity_id=$2 ORDER BY locale,alias`, ws, items[i].ID)
		if queryErr != nil {
			return nil, queryErr
		}
		for aliasRows.Next() {
			var alias domain.Alias
			if queryErr = aliasRows.Scan(&alias.ID, &alias.Value, &alias.Locale); queryErr != nil {
				aliasRows.Close()
				return nil, queryErr
			}
			items[i].Aliases = append(items[i].Aliases, alias)
		}
		aliasRows.Close()
	}
	return items, nil
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
func (r *Repository) UpdateEntity(ctx context.Context, ws, entityID uuid.UUID, kind, name, description string) error {
	_, err := r.db.Exec(ctx, `UPDATE entities SET type=$1,canonical_name=$2,description=$3,updated_at=now() WHERE workspace_id=$4 AND object_id=$5`, kind, name, description, ws, entityID)
	return err
}
func (r *Repository) DeleteEntity(ctx context.Context, ws, entityID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM objects WHERE workspace_id=$1 AND id=$2 AND kind='entity'`, ws, entityID)
	return err
}
func (r *Repository) DeleteAlias(ctx context.Context, ws, aliasID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM entity_aliases WHERE workspace_id=$1 AND id=$2`, ws, aliasID)
	return err
}
func (r *Repository) DeleteRelation(ctx context.Context, ws, relationID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM relations WHERE workspace_id=$1 AND id=$2`, ws, relationID)
	return err
}
func (r *Repository) ExtractMentions(ctx context.Context, ws uuid.UUID) error {
	_, err := r.db.Exec(ctx, `WITH terms AS (
		SELECT workspace_id,object_id AS entity_id,canonical_name AS term FROM entities
		UNION ALL SELECT workspace_id,entity_id,alias FROM entity_aliases
	), candidates AS (
		SELECT DISTINCT l.current_revision_id AS revision_id,t.entity_id,t.term
		FROM content_localizations l
		JOIN content_revisions rv ON rv.id=l.current_revision_id
		JOIN terms t ON t.workspace_id=l.workspace_id
		WHERE l.workspace_id=$1 AND length(t.term)>=2
		AND lower(rv.title||' '||rv.summary||' '||rv.body_json::text) LIKE '%'||lower(t.term)||'%'
	)
	INSERT INTO mentions(id,workspace_id,revision_id,entity_id,locator_json,confidence,confirmed,provenance)
	SELECT gen_random_uuid(),$1,revision_id,entity_id,jsonb_build_object('matched_text',term),0.8,false,'{"source":"term_match"}'::jsonb FROM candidates
	ON CONFLICT(revision_id,entity_id) DO UPDATE SET locator_json=EXCLUDED.locator_json,confidence=EXCLUDED.confidence`, ws)
	return err
}
func (r *Repository) ListMentions(ctx context.Context, ws uuid.UUID, confirmed bool) ([]domain.Mention, error) {
	rows, err := r.db.Query(ctx, `SELECT m.id,m.revision_id,m.entity_id,e.canonical_name,l.content_id,rv.title,l.locale,m.confidence,m.confirmed,m.created_at FROM mentions m JOIN entities e ON e.object_id=m.entity_id JOIN content_revisions rv ON rv.id=m.revision_id JOIN content_localizations l ON l.id=rv.localization_id WHERE m.workspace_id=$1 AND ($2 OR NOT m.confirmed) ORDER BY m.confirmed,m.created_at DESC`, ws, confirmed)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.Mention{}
	for rows.Next() {
		var item domain.Mention
		if err = rows.Scan(&item.ID, &item.RevisionID, &item.EntityID, &item.EntityName, &item.ContentID, &item.ContentTitle, &item.Locale, &item.Confidence, &item.Confirmed, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (r *Repository) ConfirmMention(ctx context.Context, ws, mentionID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE mentions SET confirmed=true WHERE workspace_id=$1 AND id=$2`, ws, mentionID)
	return err
}
