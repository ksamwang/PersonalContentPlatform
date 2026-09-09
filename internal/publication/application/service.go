package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/id"
)

type Service struct{ db *pgxpool.Pool }

func NewService(db *pgxpool.Pool) *Service { return &Service{db: db} }

type Record struct {
	ID          uuid.UUID  `json:"id"`
	ContentID   uuid.UUID  `json:"content_id"`
	Title       string     `json:"title"`
	Locale      string     `json:"locale"`
	Channel     string     `json:"channel"`
	State       string     `json:"state"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	Attempts    int        `json:"attempts"`
	LastError   string     `json:"last_error"`
	CreatedAt   time.Time  `json:"created_at"`
}

func (s *Service) List(ctx context.Context, ws uuid.UUID) ([]Record, error) {
	rows, err := s.db.Query(ctx, `SELECT p.id,p.content_id,r.title,p.locale,t.channel,p.state,p.scheduled_at,p.published_at,(SELECT count(*) FROM publication_attempts a WHERE a.publication_id=p.id),COALESCE((SELECT NULLIF(a.error_message,'') FROM publication_attempts a WHERE a.publication_id=p.id ORDER BY a.attempt_no DESC LIMIT 1),(SELECT o.last_error FROM outbox_events o WHERE o.aggregate_id=p.id ORDER BY o.created_at DESC LIMIT 1),''),p.created_at FROM publications p JOIN publication_targets t ON t.id=p.target_id JOIN content_revisions r ON r.id=p.revision_id WHERE p.workspace_id=$1 ORDER BY p.created_at DESC LIMIT 100`, ws)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Record{}
	for rows.Next() {
		var v Record
		if err = rows.Scan(&v.ID, &v.ContentID, &v.Title, &v.Locale, &v.Channel, &v.State, &v.ScheduledAt, &v.PublishedAt, &v.Attempts, &v.LastError, &v.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, v)
	}
	return items, rows.Err()
}
func (s *Service) Schedule(ctx context.Context, ws, user, content uuid.UUID, locale string, at time.Time) (uuid.UUID, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)
	var revision, target uuid.UUID
	var state string
	if err = tx.QueryRow(ctx, `SELECT current_revision_id,state FROM content_localizations WHERE workspace_id=$1 AND content_id=$2 AND locale=$3`, ws, content, locale).Scan(&revision, &state); err != nil {
		return uuid.Nil, err
	}
	if state != "ready" && state != "published" {
		return uuid.Nil, fmt.Errorf("localization must be ready before scheduling")
	}
	if err = tx.QueryRow(ctx, `INSERT INTO publication_targets(id,workspace_id,channel,name) VALUES($1,$2,'website','Primary website') ON CONFLICT(workspace_id,channel,name) DO UPDATE SET enabled=true RETURNING id`, id.New(), ws).Scan(&target); err != nil {
		return uuid.Nil, err
	}
	publication := id.New()
	if err = tx.QueryRow(ctx, `INSERT INTO publications(id,workspace_id,content_id,locale,revision_id,target_id,state,scheduled_at,created_by) VALUES($1,$2,$3,$4,$5,$6,'queued',$7,$8) ON CONFLICT(target_id,revision_id) DO UPDATE SET state='queued',scheduled_at=EXCLUDED.scheduled_at,updated_at=now() RETURNING id`, publication, ws, content, locale, revision, target, at, user).Scan(&publication); err != nil {
		return uuid.Nil, err
	}
	payload := fmt.Sprintf(`{"publication_id":"%s"}`, publication)
	_, err = tx.Exec(ctx, `INSERT INTO outbox_events(id,workspace_id,aggregate_id,aggregate_type,type,payload,available_at) VALUES($1,$2,$3,'publication','PublicationRequested',$4,$5)`, id.New(), ws, publication, payload, at)
	if err != nil {
		return uuid.Nil, err
	}
	return publication, tx.Commit(ctx)
}
func (s *Service) Retry(ctx context.Context, ws, idv uuid.UUID) error {
	tag, err := s.db.Exec(ctx, `WITH changed AS (UPDATE publications SET state='queued',updated_at=now() WHERE workspace_id=$1 AND id=$2 AND state IN ('failed','dead','retry_wait','cancelled') RETURNING id) INSERT INTO outbox_events(id,workspace_id,aggregate_id,aggregate_type,type,payload,available_at) SELECT gen_random_uuid(),$1,id,'publication','PublicationRequested',jsonb_build_object('publication_id',id),now() FROM changed`, ws, idv)
	if err == nil && tag.RowsAffected() == 0 {
		return fmt.Errorf("publication cannot be retried")
	}
	return err
}
func (s *Service) Withdraw(ctx context.Context, ws, idv uuid.UUID) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var content uuid.UUID
	var locale string
	if err = tx.QueryRow(ctx, `UPDATE publications SET state='withdrawn',updated_at=now() WHERE workspace_id=$1 AND id=$2 AND state IN ('published','queued','retry_wait') RETURNING content_id,locale`, ws, idv).Scan(&content, &locale); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM publication_views WHERE workspace_id=$1 AND content_id=$2 AND locale=$3`, ws, content, locale); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `UPDATE content_localizations SET state='ready',translation_status='ready',updated_at=now() WHERE workspace_id=$1 AND content_id=$2 AND locale=$3`, ws, content, locale)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

type Channel struct {
	Channel string `json:"channel"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	Detail  string `json:"detail"`
}

func (s *Service) Channels(ctx context.Context, ws uuid.UUID) ([]Channel, error) {
	var rss bool
	var hooks int
	err := s.db.QueryRow(ctx, `SELECT COALESCE((settings_json#>>'{site,rss_enabled}')::boolean,true),(SELECT count(*) FROM webhook_endpoints WHERE workspace_id=$1 AND enabled) FROM workspaces WHERE id=$1`, ws).Scan(&rss, &hooks)
	if err != nil {
		return nil, err
	}
	rssState := "disabled"
	if rss {
		rssState = "ready"
	}
	webhookState := "not_configured"
	if hooks > 0 {
		webhookState = "ready"
	}
	return []Channel{{"website", "公开网站", "ready", "内置发布目标"}, {"rss", "RSS", rssState, "由站点设置控制"}, {"webhook", "Webhook", webhookState, fmt.Sprintf("%d 个启用端点", hooks)}, {"newsletter", "Newsletter", "planned", "等待具体服务商适配"}, {"social", "社交平台", "planned", "等待具体平台授权"}}, nil
}
