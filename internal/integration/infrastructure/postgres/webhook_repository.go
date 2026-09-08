package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/ksamwang/PersonalContentPlatform/internal/integration/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/id"
)

func (r *Repository) CreateWebhook(ctx context.Context, workspaceID uuid.UUID, endpoint domain.WebhookEndpoint) (domain.WebhookEndpoint, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO webhook_endpoints(id,workspace_id,name,url,secret_ref,enabled,event_types)
		VALUES($1,$2,$3,$4,$5,$6,$7)
		RETURNING created_at,updated_at`,
		endpoint.ID, workspaceID, endpoint.Name, endpoint.URL, endpoint.SecretRef, endpoint.Enabled, endpoint.EventTypes,
	).Scan(&endpoint.CreatedAt, &endpoint.UpdatedAt)
	return endpoint, err
}

func (r *Repository) ListWebhooks(ctx context.Context, workspaceID uuid.UUID) ([]domain.WebhookEndpoint, error) {
	rows, err := r.db.Query(ctx, `SELECT id,name,url,secret_ref,enabled,event_types,created_at,updated_at FROM webhook_endpoints WHERE workspace_id=$1 ORDER BY created_at`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.WebhookEndpoint{}
	for rows.Next() {
		var endpoint domain.WebhookEndpoint
		if err = rows.Scan(&endpoint.ID, &endpoint.Name, &endpoint.URL, &endpoint.SecretRef, &endpoint.Enabled, &endpoint.EventTypes, &endpoint.CreatedAt, &endpoint.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, endpoint)
	}
	return items, rows.Err()
}

func (r *Repository) SetWebhookEnabled(ctx context.Context, workspaceID, endpointID uuid.UUID, enabled bool) error {
	command, err := r.db.Exec(ctx, `UPDATE webhook_endpoints SET enabled=$3,updated_at=now() WHERE id=$1 AND workspace_id=$2`, endpointID, workspaceID, enabled)
	if err == nil && command.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return err
}

func (r *Repository) DeleteWebhook(ctx context.Context, workspaceID, endpointID uuid.UUID) error {
	command, err := r.db.Exec(ctx, `DELETE FROM webhook_endpoints WHERE id=$1 AND workspace_id=$2`, endpointID, workspaceID)
	if err == nil && command.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return err
}

func (r *Repository) ClaimWebhookDelivery(ctx context.Context, maxAttempts int) (*domain.WebhookDelivery, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var delivery domain.WebhookDelivery
	err = tx.QueryRow(ctx, `
		SELECT e.id,o.id,o.type,e.url,e.secret_ref,o.payload,o.created_at,COALESCE(a.attempt_count,0)+1
		FROM outbox_events o
		JOIN webhook_endpoints e ON e.workspace_id=o.workspace_id
		LEFT JOIN LATERAL (
			SELECT count(*)::integer AS attempt_count,max(created_at) AS last_attempt_at,
			       bool_or(status='succeeded') AS succeeded,
			       bool_or(status='pending' AND created_at>now()-interval '5 minutes') AS active
			FROM webhook_attempts wa WHERE wa.endpoint_id=e.id AND wa.event_id=o.id
		) a ON true
		WHERE o.dispatched_at IS NOT NULL AND e.enabled
		  AND (cardinality(e.event_types)=0 OR o.type=ANY(e.event_types))
		  AND NOT COALESCE(a.succeeded,false)
		  AND NOT COALESCE(a.active,false)
		  AND COALESCE(a.attempt_count,0)<$1
		  AND (a.last_attempt_at IS NULL OR a.last_attempt_at<=now()-interval '15 seconds')
		ORDER BY o.created_at,e.created_at
		FOR UPDATE OF e SKIP LOCKED
		LIMIT 1`, maxAttempts,
	).Scan(
		&delivery.EndpointID,
		&delivery.EventID,
		&delivery.EventType,
		&delivery.URL,
		&delivery.SecretRef,
		&delivery.Payload,
		&delivery.CreatedAt,
		&delivery.AttemptNo,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	delivery.AttemptID = id.New()
	if _, err = tx.Exec(ctx, `INSERT INTO webhook_attempts(id,endpoint_id,event_id,attempt_no,status) VALUES($1,$2,$3,$4,'pending')`, delivery.AttemptID, delivery.EndpointID, delivery.EventID, delivery.AttemptNo); err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &delivery, nil
}

func (r *Repository) CompleteWebhookDelivery(ctx context.Context, attemptID uuid.UUID, result domain.DeliveryResult) error {
	_, err := r.db.Exec(ctx, `UPDATE webhook_attempts SET status=$2,response_code=$3,response_excerpt=$4 WHERE id=$1`, attemptID, result.Status, result.ResponseCode, result.ResponseExcerpt)
	return err
}
