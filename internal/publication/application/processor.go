package application

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ksamwang/PersonalContentPlatform/internal/content/document"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/id"
	"log/slog"
	"time"
)

type Processor struct {
	db       *pgxpool.Pool
	workerID string
	indexer  interface {
		IndexContent(context.Context, uuid.UUID, uuid.UUID) (int, error)
	}
}

func NewProcessor(db *pgxpool.Pool, workerID string) *Processor {
	return &Processor{db: db, workerID: workerID}
}

func (p *Processor) WithIndexer(indexer interface {
	IndexContent(context.Context, uuid.UUID, uuid.UUID) (int, error)
}) *Processor {
	p.indexer = indexer
	return p
}

type event struct {
	ID, WorkspaceID, AggregateID uuid.UUID
	Type                         string
	Payload                      json.RawMessage
	Attempts                     int
}

func (p *Processor) Run(ctx context.Context) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		processed, err := p.processOne(ctx)
		if err != nil {
			slog.Warn("process outbox event", "error", err)
		}
		if processed {
			continue
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}
func (p *Processor) processOne(ctx context.Context) (bool, error) {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)
	var e event
	err = tx.QueryRow(ctx, `SELECT id,workspace_id,aggregate_id,type,payload,attempts FROM outbox_events WHERE dispatched_at IS NULL AND available_at<=now() ORDER BY created_at FOR UPDATE SKIP LOCKED LIMIT 1`).Scan(&e.ID, &e.WorkspaceID, &e.AggregateID, &e.Type, &e.Payload, &e.Attempts)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	switch e.Type {
	case "ContentCreated", "ContentRevisionCreated":
		err = p.indexContent(ctx, tx, e)
		if err == nil && p.indexer != nil {
			if _, indexErr := p.indexer.IndexContent(ctx, e.WorkspaceID, e.AggregateID); indexErr != nil {
				slog.Warn("update semantic index", "content_id", e.AggregateID, "error", indexErr)
			}
		}
	case "PublicationRequested":
		err = p.publishWebsite(ctx, tx, e)
	default:
		err = nil
	}
	if err != nil {
		if e.Type == "PublicationRequested" {
			var attempt int
			_ = tx.QueryRow(ctx, `SELECT COALESCE(MAX(attempt_no),0)+1 FROM publication_attempts WHERE publication_id=$1`, e.AggregateID).Scan(&attempt)
			_, _ = tx.Exec(ctx, `INSERT INTO publication_attempts(id,publication_id,attempt_no,status,error_message) VALUES($1,$2,$3,'failed',$4)`, id.New(), e.AggregateID, attempt, err.Error())
			state := "retry_wait"
			if e.Attempts >= 4 {
				state = "dead"
			}
			_, _ = tx.Exec(ctx, `UPDATE publications SET state=$2,updated_at=now() WHERE id=$1`, e.AggregateID, state)
		}
		_, _ = tx.Exec(ctx, `UPDATE outbox_events SET attempts=attempts+1,last_error=$2,available_at=now()+interval '15 seconds' WHERE id=$1`, e.ID, err.Error())
		return true, tx.Commit(ctx)
	}
	if _, err = tx.Exec(ctx, `UPDATE outbox_events SET dispatched_at=now(),attempts=attempts+1,last_error=NULL WHERE id=$1`, e.ID); err != nil {
		return true, err
	}
	return true, tx.Commit(ctx)
}
func (p *Processor) indexContent(ctx context.Context, tx pgx.Tx, e event) error {
	rows, err := tx.Query(ctx, `SELECT c.object_id,r.id,l.locale,c.visibility,r.title,r.summary,r.body_json FROM contents c JOIN content_localizations l ON l.content_id=c.object_id JOIN content_revisions r ON r.id=l.current_revision_id WHERE c.workspace_id=$1 AND c.object_id=$2 AND c.deleted_at IS NULL AND l.deleted_at IS NULL AND l.state<>'archived'`, e.WorkspaceID, e.AggregateID)
	if err != nil {
		return err
	}
	type indexedContent struct {
		objectID, revisionID               uuid.UUID
		locale, visibility, title, summary string
		body                               []byte
	}
	items := []indexedContent{}
	for rows.Next() {
		var item indexedContent
		if err = rows.Scan(&item.objectID, &item.revisionID, &item.locale, &item.visibility, &item.title, &item.summary, &item.body); err != nil {
			rows.Close()
			return err
		}
		items = append(items, item)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM search_documents WHERE workspace_id=$1 AND object_id=$2`, e.WorkspaceID, e.AggregateID); err != nil {
		return err
	}
	for _, item := range items {
		if _, err = tx.Exec(ctx, `INSERT INTO search_documents(workspace_id,object_id,revision_id,locale,visibility,title,summary,body_text) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, e.WorkspaceID, item.objectID, item.revisionID, item.locale, item.visibility, item.title, item.summary, document.PlainText(item.body)); err != nil {
			return err
		}
	}
	return nil
}
func (p *Processor) publishWebsite(ctx context.Context, tx pgx.Tx, e event) error {
	publicationID := e.AggregateID
	var workspaceID, contentID, revisionID uuid.UUID
	var locale, slug, kind, title, summary, visibility string
	var body, metadata []byte
	tag, err := tx.Exec(ctx, `UPDATE publications SET state='publishing',updated_at=now() WHERE id=$1 AND state IN ('queued','retry_wait')`, publicationID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return nil
	}
	err = tx.QueryRow(ctx, `SELECT p.workspace_id,p.content_id,p.revision_id,p.locale,l.slug,c.type,r.title,r.summary,c.visibility,r.body_json,r.metadata_json FROM publications p JOIN contents c ON c.object_id=p.content_id JOIN content_localizations l ON l.content_id=p.content_id AND l.locale=p.locale JOIN content_revisions r ON r.id=p.revision_id JOIN publication_targets t ON t.id=p.target_id WHERE p.id=$1 AND t.channel='website'`, publicationID).Scan(&workspaceID, &contentID, &revisionID, &locale, &slug, &kind, &title, &summary, &visibility, &body, &metadata)
	if err != nil {
		return err
	}
	if visibility == "private" {
		if _, err = tx.Exec(ctx, `UPDATE contents SET visibility='public' WHERE object_id=$1`, contentID); err != nil {
			return err
		}
		visibility = "public"
	}
	rendered := document.HTML(body)
	_, err = tx.Exec(ctx, `INSERT INTO publication_views(workspace_id,content_id,locale,slug,type,revision_id,title,summary,rendered_html,metadata_json,published_at,visibility) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,now(),$11) ON CONFLICT(content_id,locale) DO UPDATE SET slug=EXCLUDED.slug,type=EXCLUDED.type,revision_id=EXCLUDED.revision_id,title=EXCLUDED.title,summary=EXCLUDED.summary,rendered_html=EXCLUDED.rendered_html,metadata_json=EXCLUDED.metadata_json,published_at=now(),visibility=EXCLUDED.visibility`, workspaceID, contentID, locale, slug, kind, revisionID, title, summary, rendered, metadata, visibility)
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE publications SET state='published',published_at=now(),updated_at=now() WHERE id=$1`, publicationID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE content_localizations SET state='published',translation_status='published',updated_at=now() WHERE content_id=$1 AND locale=$2`, contentID, locale); err != nil {
		return err
	}
	var attempt int
	if err = tx.QueryRow(ctx, `SELECT COALESCE(MAX(attempt_no),0)+1 FROM publication_attempts WHERE publication_id=$1`, publicationID).Scan(&attempt); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO publication_attempts(id,publication_id,attempt_no,status,external_id) VALUES($1,$2,$3,'succeeded',$4)`, id.New(), publicationID, attempt, fmt.Sprintf("/%s/%s/%s", locale, kind, slug))
	return err
}
