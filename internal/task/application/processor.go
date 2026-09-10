package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler func(context.Context, uuid.UUID, json.RawMessage) (string, error)

type Processor struct {
	db       *pgxpool.Pool
	workerID string
	handlers map[string]Handler
}

func NewProcessor(db *pgxpool.Pool, workerID string, handlers map[string]Handler) *Processor {
	return &Processor{db: db, workerID: workerID, handlers: handlers}
}

func (p *Processor) Run(ctx context.Context) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		processed, err := p.processOne(ctx)
		if err != nil {
			slog.Warn("process background job", "error", err)
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
	var jobID, ws uuid.UUID
	var kind string
	var payload json.RawMessage
	var attempts, maxAttempts int
	err := p.db.QueryRow(ctx, `UPDATE jobs SET state='running',lease_owner=$1,lease_until=now()+interval '2 minutes',attempts=attempts+1,updated_at=now() WHERE id=(SELECT id FROM jobs WHERE queue='default' AND ((state IN ('pending','retry_wait') AND available_at<=now()) OR (state='running' AND lease_until<now())) ORDER BY priority DESC,created_at FOR UPDATE SKIP LOCKED LIMIT 1) RETURNING id,workspace_id,type,payload,attempts,max_attempts`, p.workerID).
		Scan(&jobID, &ws, &kind, &payload, &attempts, &maxAttempts)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	handler := p.handlers[kind]
	if handler == nil {
		return true, p.fail(ctx, jobID, attempts, maxAttempts, fmt.Errorf("unsupported job type %s", kind))
	}
	result, runErr := handler(ctx, ws, payload)
	if runErr != nil {
		return true, p.fail(ctx, jobID, attempts, maxAttempts, runErr)
	}
	_, err = p.db.Exec(ctx, `UPDATE jobs SET state='succeeded',detail_ref=$2,last_error_code=NULL,lease_owner=NULL,lease_until=NULL,updated_at=now() WHERE id=$1`, jobID, result)
	return true, err
}

func (p *Processor) fail(ctx context.Context, jobID uuid.UUID, attempts, maxAttempts int, cause error) error {
	state := "retry_wait"
	if attempts >= maxAttempts {
		state = "dead"
	}
	_, err := p.db.Exec(ctx, `UPDATE jobs SET state=$2,last_error_code=$3,available_at=$4,lease_owner=NULL,lease_until=NULL,updated_at=now() WHERE id=$1`, jobID, state, cause.Error(), time.Now().Add(RetryDelay(attempts)))
	return err
}
