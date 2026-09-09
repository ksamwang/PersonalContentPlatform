package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ksamwang/PersonalContentPlatform/internal/collection/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/id"
)

type Repository struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

func (r *Repository) Create(ctx context.Context, workspaceID uuid.UUID, title, slug, visibility string) (domain.Collection, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Collection{}, err
	}
	defer tx.Rollback(ctx)
	collectionID := id.New()
	if _, err = tx.Exec(ctx, `INSERT INTO objects(id,workspace_id,kind) VALUES($1,$2,'collection')`, collectionID, workspaceID); err != nil {
		return domain.Collection{}, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO collections(object_id,workspace_id,title,slug,visibility) VALUES($1,$2,$3,$4,$5)`, collectionID, workspaceID, title, slug, visibility); err != nil {
		return domain.Collection{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.Collection{}, err
	}
	return domain.Collection{ID: collectionID, WorkspaceID: workspaceID, Title: title, Slug: slug, Visibility: visibility, Sections: []domain.Section{}}, nil
}
func (r *Repository) List(ctx context.Context, workspaceID uuid.UUID) ([]domain.Collection, error) {
	rows, err := r.db.Query(ctx, `SELECT object_id,workspace_id,title,slug,visibility FROM collections WHERE workspace_id=$1 ORDER BY title,object_id`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.Collection{}
	for rows.Next() {
		var item domain.Collection
		if err := rows.Scan(&item.ID, &item.WorkspaceID, &item.Title, &item.Slug, &item.Visibility); err != nil {
			return nil, err
		}
		item.Sections = []domain.Section{}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (r *Repository) Get(ctx context.Context, workspaceID, collectionID uuid.UUID) (domain.Collection, error) {
	var result domain.Collection
	if err := r.db.QueryRow(ctx, `SELECT object_id,workspace_id,title,slug,visibility FROM collections WHERE workspace_id=$1 AND object_id=$2`, workspaceID, collectionID).Scan(&result.ID, &result.WorkspaceID, &result.Title, &result.Slug, &result.Visibility); err != nil {
		return result, err
	}
	rows, err := r.db.Query(ctx, `SELECT s.id,s.title,s.sort_key,i.id,i.object_id,COALESCE(rv.title,''),i.sort_key,i.annotation FROM collection_sections s LEFT JOIN collection_items i ON i.section_id=s.id LEFT JOIN content_localizations l ON l.content_id=i.object_id AND l.locale=(SELECT default_locale FROM contents WHERE object_id=i.object_id) LEFT JOIN content_revisions rv ON rv.id=l.current_revision_id WHERE s.workspace_id=$1 AND s.collection_id=$2 ORDER BY s.sort_key,i.sort_key`, workspaceID, collectionID)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	result.Sections = []domain.Section{}
	sectionIndex := map[uuid.UUID]int{}
	for rows.Next() {
		var section domain.Section
		var itemID, objectID *uuid.UUID
		var itemTitle, annotation *string
		var itemSort *float64
		if err := rows.Scan(&section.ID, &section.Title, &section.SortKey, &itemID, &objectID, &itemTitle, &itemSort, &annotation); err != nil {
			return result, err
		}
		idx, ok := sectionIndex[section.ID]
		if !ok {
			section.Items = []domain.Item{}
			result.Sections = append(result.Sections, section)
			idx = len(result.Sections) - 1
			sectionIndex[section.ID] = idx
		}
		if itemID != nil && objectID != nil {
			result.Sections[idx].Items = append(result.Sections[idx].Items, domain.Item{ID: *itemID, ObjectID: *objectID, Title: value(itemTitle), SortKey: valueFloat(itemSort), Annotation: value(annotation)})
		}
	}
	return result, rows.Err()
}
func value(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
func valueFloat(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}
func (r *Repository) Update(ctx context.Context, workspaceID, collectionID uuid.UUID, title, slug, visibility string) (domain.Collection, error) {
	tag, err := r.db.Exec(ctx, `UPDATE collections SET title=$1,slug=$2,visibility=$3 WHERE workspace_id=$4 AND object_id=$5`, title, slug, visibility, workspaceID, collectionID)
	if err == nil && tag.RowsAffected() == 0 {
		return domain.Collection{}, pgx.ErrNoRows
	}
	if err != nil {
		return domain.Collection{}, err
	}
	return r.Get(ctx, workspaceID, collectionID)
}

func (r *Repository) AddSection(ctx context.Context, workspaceID, collectionID uuid.UUID, title string) (domain.Section, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Section{}, err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `SELECT object_id FROM collections WHERE workspace_id=$1 AND object_id=$2 FOR UPDATE`, workspaceID, collectionID); err != nil {
		return domain.Section{}, err
	}
	var sort float64
	if err = tx.QueryRow(ctx, `SELECT COALESCE(MAX(sort_key),0)+1024 FROM collection_sections WHERE collection_id=$1`, collectionID).Scan(&sort); err != nil {
		return domain.Section{}, err
	}
	section := domain.Section{ID: id.New(), Title: title, SortKey: sort, Items: []domain.Item{}}
	if _, err = tx.Exec(ctx, `INSERT INTO collection_sections(id,workspace_id,collection_id,title,sort_key) VALUES($1,$2,$3,$4,$5)`, section.ID, workspaceID, collectionID, title, sort); err != nil {
		return domain.Section{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.Section{}, err
	}
	return section, nil
}
func (r *Repository) MoveSection(ctx context.Context, workspaceID, collectionID, sectionID uuid.UUID, direction int) error {
	return r.swap(ctx, workspaceID, `SELECT id,sort_key FROM collection_sections WHERE workspace_id=$1 AND collection_id=$2 ORDER BY sort_key FOR UPDATE`, []any{workspaceID, collectionID}, sectionID, direction, "collection_sections")
}
func (r *Repository) AddItem(ctx context.Context, workspaceID, sectionID, objectID uuid.UUID, annotation string) (domain.Item, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Item{}, err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `SELECT id FROM collection_sections WHERE workspace_id=$1 AND id=$2 FOR UPDATE`, workspaceID, sectionID); err != nil {
		return domain.Item{}, err
	}
	var title string
	if err = tx.QueryRow(ctx, `SELECT COALESCE(rv.title,'') FROM objects o JOIN contents c ON c.object_id=o.id JOIN content_localizations l ON l.content_id=c.object_id AND l.locale=c.default_locale LEFT JOIN content_revisions rv ON rv.id=l.current_revision_id WHERE o.workspace_id=$1 AND o.id=$2 AND o.kind='content'`, workspaceID, objectID).Scan(&title); err != nil {
		return domain.Item{}, err
	}
	var exists bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM collection_items WHERE section_id=$1 AND object_id=$2)`, sectionID, objectID).Scan(&exists); err != nil {
		return domain.Item{}, err
	}
	if exists {
		return domain.Item{}, errors.New("content is already in this section")
	}
	var sort float64
	if err = tx.QueryRow(ctx, `SELECT COALESCE(MAX(sort_key),0)+1024 FROM collection_items WHERE section_id=$1`, sectionID).Scan(&sort); err != nil {
		return domain.Item{}, err
	}
	item := domain.Item{ID: id.New(), ObjectID: objectID, Title: title, SortKey: sort, Annotation: annotation}
	if _, err = tx.Exec(ctx, `INSERT INTO collection_items(id,workspace_id,section_id,object_id,sort_key,annotation) VALUES($1,$2,$3,$4,$5,$6)`, item.ID, workspaceID, sectionID, objectID, sort, annotation); err != nil {
		return domain.Item{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.Item{}, err
	}
	return item, nil
}
func (r *Repository) MoveItem(ctx context.Context, workspaceID, itemID uuid.UUID, direction int) error {
	var sectionID uuid.UUID
	if err := r.db.QueryRow(ctx, `SELECT section_id FROM collection_items WHERE workspace_id=$1 AND id=$2`, workspaceID, itemID).Scan(&sectionID); err != nil {
		return err
	}
	return r.swap(ctx, workspaceID, `SELECT id,sort_key FROM collection_items WHERE workspace_id=$1 AND section_id=$2 ORDER BY sort_key FOR UPDATE`, []any{workspaceID, sectionID}, itemID, direction, "collection_items")
}
func (r *Repository) RemoveItem(ctx context.Context, workspaceID, itemID uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM collection_items WHERE workspace_id=$1 AND id=$2`, workspaceID, itemID)
	if err == nil && tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return err
}
func (r *Repository) swap(ctx context.Context, workspaceID uuid.UUID, query string, args []any, target uuid.UUID, direction int, table string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return err
	}
	var ids []uuid.UUID
	var keys []float64
	for rows.Next() {
		var rowID uuid.UUID
		var key float64
		if err = rows.Scan(&rowID, &key); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, rowID)
		keys = append(keys, key)
	}
	rows.Close()
	index := -1
	for i, rowID := range ids {
		if rowID == target {
			index = i
			break
		}
	}
	other := index + direction
	if index < 0 {
		return pgx.ErrNoRows
	}
	if other < 0 || other >= len(ids) {
		return tx.Commit(ctx)
	}
	temporary := keys[index] + 0.5
	if _, err = tx.Exec(ctx, `UPDATE `+table+` SET sort_key=$1 WHERE id=$2`, temporary, ids[index]); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE `+table+` SET sort_key=$1 WHERE id=$2`, keys[index], ids[other]); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE `+table+` SET sort_key=$1 WHERE id=$2`, keys[other], ids[index]); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
