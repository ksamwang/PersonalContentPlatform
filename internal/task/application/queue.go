package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/id"
)

type Queue struct{ db *pgxpool.Pool }

func NewQueue(db *pgxpool.Pool) *Queue { return &Queue{db: db} }

func (q *Queue) Enqueue(ctx context.Context, ws uuid.UUID, kind string, payload any) (uuid.UUID, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return uuid.Nil, err
	}
	jobID := id.New()
	_, err = q.db.Exec(ctx, `INSERT INTO jobs(id,workspace_id,queue,type,idempotency_key,payload,available_at) VALUES($1,$2,'default',$3,$4,$5,now())`, jobID, ws, kind, jobID.String(), raw)
	return jobID, err
}

type JobStatus struct {
	ID        uuid.UUID `json:"id"`
	Type      string    `json:"type"`
	State     string    `json:"state"`
	Attempts  int       `json:"attempts"`
	LastError string    `json:"last_error"`
	Result    string    `json:"result,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (q *Queue) Status(ctx context.Context, ws, jobID uuid.UUID) (JobStatus, error) {
	var value JobStatus
	err := q.db.QueryRow(ctx, `SELECT id,type,state,attempts,COALESCE(last_error_code,''),COALESCE(detail_ref,''),updated_at FROM jobs WHERE workspace_id=$1 AND id=$2`, ws, jobID).
		Scan(&value.ID, &value.Type, &value.State, &value.Attempts, &value.LastError, &value.Result, &value.UpdatedAt)
	return value, err
}

func RetryDelay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	if attempt > 4 {
		attempt = 4
	}
	return time.Duration(attempt*15) * time.Second
}

func ResultID(value uuid.UUID) string { return fmt.Sprintf(`{"id":"%s"}`, value) }
