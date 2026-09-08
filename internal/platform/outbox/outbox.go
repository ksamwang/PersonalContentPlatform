package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/id"
)

func Add(ctx context.Context, tx pgx.Tx, workspaceID, aggregateID uuid.UUID, aggregateType, eventType string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal event payload: %w", err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO outbox_events (id,workspace_id,aggregate_id,aggregate_type,type,payload,available_at) VALUES ($1,$2,$3,$4,$5,$6,now())`, id.New(), workspaceID, aggregateID, aggregateType, eventType, data)
	return err
}
