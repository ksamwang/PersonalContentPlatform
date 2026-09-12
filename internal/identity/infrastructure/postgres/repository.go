package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ksamwang/PersonalContentPlatform/internal/identity/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/id"
	"time"
)

type Repository struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Repository { return &Repository{db: db} }
func (r *Repository) HasUsers(ctx context.Context) (bool, error) {
	var value bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users)`).Scan(&value)
	return value, err
}
func (r *Repository) CreateOwner(ctx context.Context, u domain.User, slug, name string) (domain.Principal, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Principal{}, err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(81745201)`); err != nil {
		return domain.Principal{}, err
	}
	var exists bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users)`).Scan(&exists); err != nil {
		return domain.Principal{}, err
	}
	if exists {
		return domain.Principal{}, fmt.Errorf("initial setup already completed")
	}
	workspaceID := id.New()
	if _, err = tx.Exec(ctx, `INSERT INTO workspaces(id,slug,name) VALUES($1,$2,$3)`, workspaceID, slug, name); err != nil {
		return domain.Principal{}, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO users(id,email,display_name,password_hash,status) VALUES($1,$2,$3,$4,$5)`, u.ID, u.Email, u.DisplayName, u.PasswordHash, u.Status); err != nil {
		return domain.Principal{}, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO memberships(workspace_id,user_id,role) VALUES($1,$2,'owner')`, workspaceID, u.ID); err != nil {
		return domain.Principal{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.Principal{}, err
	}
	return domain.Principal{UserID: u.ID, WorkspaceID: workspaceID, Email: u.Email, DisplayName: u.DisplayName, Role: "owner"}, nil
}
func (r *Repository) UserByEmail(ctx context.Context, email string) (domain.User, error) {
	return r.user(ctx, `u.email=$1`, email)
}
func (r *Repository) UserByID(ctx context.Context, userID uuid.UUID) (domain.User, error) {
	return r.user(ctx, `u.id=$1`, userID)
}
func (r *Repository) AuthPolicyByUser(ctx context.Context, userID uuid.UUID, fallback time.Duration) (domain.AuthPolicy, error) {
	v := domain.AuthPolicy{PasswordLoginEnabled: true, PasskeyEnabled: true, TOTPLoginEnabled: true, SessionTTL: fallback}
	var hours *int
	err := r.db.QueryRow(ctx, `SELECT COALESCE((w.settings_json#>>'{auth,password_login_enabled}')::boolean,true),COALESCE((w.settings_json#>>'{auth,passkey_enabled}')::boolean,true),COALESCE((w.settings_json#>>'{auth,totp_login_enabled}')::boolean,true),(w.settings_json#>>'{auth,session_ttl_hours}')::integer FROM memberships m JOIN workspaces w ON w.id=m.workspace_id WHERE m.user_id=$1 ORDER BY CASE m.role WHEN 'owner' THEN 1 WHEN 'editor' THEN 2 ELSE 3 END LIMIT 1`, userID).Scan(&v.PasswordLoginEnabled, &v.PasskeyEnabled, &v.TOTPLoginEnabled, &hours)
	if err != nil {
		return v, err
	}
	if hours != nil && *hours > 0 {
		v.SessionTTL = time.Duration(*hours) * time.Hour
	}
	return v, nil
}
func (r *Repository) user(ctx context.Context, where string, arg any) (domain.User, error) {
	var u domain.User
	err := r.db.QueryRow(ctx, `SELECT u.id,u.email,u.display_name,COALESCE(u.password_hash,''),u.status FROM users u WHERE `+where, arg).Scan(&u.ID, &u.Email, &u.DisplayName, &u.PasswordHash, &u.Status)
	if err != nil {
		return u, err
	}
	rows, err := r.db.Query(ctx, `SELECT credential_json FROM webauthn_credentials WHERE user_id=$1 ORDER BY created_at`, u.ID)
	if err != nil {
		return u, err
	}
	defer rows.Close()
	for rows.Next() {
		var data []byte
		if err = rows.Scan(&data); err != nil {
			return u, err
		}
		var c webauthn.Credential
		if err = json.Unmarshal(data, &c); err != nil {
			return u, err
		}
		u.Credentials = append(u.Credentials, c)
	}
	return u, rows.Err()
}
func (r *Repository) CreateSession(ctx context.Context, userID uuid.UUID, hash []byte, expires time.Time, userAgent string) error {
	_, err := r.db.Exec(ctx, `INSERT INTO sessions(id,user_id,token_hash,expires_at,user_agent) VALUES($1,$2,$3,$4,$5)`, id.New(), userID, hash, expires, userAgent)
	return err
}
func (r *Repository) PrincipalBySessionHash(ctx context.Context, hash []byte) (domain.Principal, error) {
	var p domain.Principal
	err := r.db.QueryRow(ctx, `SELECT u.id,m.workspace_id,u.email,u.display_name,m.role FROM sessions s JOIN users u ON u.id=s.user_id JOIN memberships m ON m.user_id=u.id WHERE s.token_hash=$1 AND s.expires_at>now() AND u.status='active' ORDER BY CASE m.role WHEN 'owner' THEN 1 WHEN 'editor' THEN 2 ELSE 3 END LIMIT 1`, hash).Scan(&p.UserID, &p.WorkspaceID, &p.Email, &p.DisplayName, &p.Role)
	if err == nil {
		_, _ = r.db.Exec(ctx, `UPDATE sessions SET last_used_at=now() WHERE token_hash=$1 AND last_used_at<now()-interval '5 minutes'`, hash)
	}
	return p, err
}
func (r *Repository) DeleteSession(ctx context.Context, hash []byte) error {
	_, err := r.db.Exec(ctx, `DELETE FROM sessions WHERE token_hash=$1`, hash)
	return err
}
func (r *Repository) SaveCeremony(ctx context.Context, ceremonyID, userID uuid.UUID, kind string, data json.RawMessage, expires time.Time) error {
	_, err := r.db.Exec(ctx, `INSERT INTO webauthn_ceremonies(id,user_id,kind,session_json,expires_at) VALUES($1,$2,$3,$4,$5)`, ceremonyID, userID, kind, data, expires)
	return err
}
func (r *Repository) TakeCeremony(ctx context.Context, ceremonyID uuid.UUID, kind string) (uuid.UUID, json.RawMessage, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return uuid.Nil, nil, err
	}
	defer tx.Rollback(ctx)
	var userID uuid.UUID
	var data json.RawMessage
	if err = tx.QueryRow(ctx, `DELETE FROM webauthn_ceremonies WHERE id=$1 AND kind=$2 AND expires_at>now() RETURNING user_id,session_json`, ceremonyID, kind).Scan(&userID, &data); err != nil {
		return uuid.Nil, nil, err
	}
	return userID, data, tx.Commit(ctx)
}
func (r *Repository) SaveCredential(ctx context.Context, userID uuid.UUID, credential webauthn.Credential) error {
	data, err := json.Marshal(credential)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `INSERT INTO webauthn_credentials(id,user_id,credential_json) VALUES($1,$2,$3) ON CONFLICT(id) DO UPDATE SET credential_json=EXCLUDED.credential_json,last_used_at=now()`, credential.ID, userID, data)
	return err
}

func (r *Repository) SaveTOTP(ctx context.Context, userID uuid.UUID, secret []byte) error {
	_, err := r.db.Exec(ctx, `INSERT INTO user_totp_credentials(user_id,secret_ciphertext) VALUES($1,$2) ON CONFLICT(user_id) DO UPDATE SET secret_ciphertext=EXCLUDED.secret_ciphertext,enabled_at=now(),updated_at=now()`, userID, secret)
	return err
}

func (r *Repository) TOTPSecretByUser(ctx context.Context, userID uuid.UUID) ([]byte, error) {
	var secret []byte
	err := r.db.QueryRow(ctx, `SELECT secret_ciphertext FROM user_totp_credentials WHERE user_id=$1`, userID).Scan(&secret)
	return secret, err
}

func (r *Repository) HasTOTP(ctx context.Context, userID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM user_totp_credentials WHERE user_id=$1)`, userID).Scan(&exists)
	return exists, err
}

func (r *Repository) DeleteTOTP(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM user_totp_credentials WHERE user_id=$1`, userID)
	return err
}

func (r *Repository) ReplaceRecoveryCodes(ctx context.Context, userID uuid.UUID, hashes [][]byte) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `DELETE FROM totp_recovery_codes WHERE user_id=$1`, userID); err != nil {
		return err
	}
	for _, hash := range hashes {
		if _, err = tx.Exec(ctx, `INSERT INTO totp_recovery_codes(user_id,code_hash) VALUES($1,$2)`, userID, hash); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *Repository) ConsumeRecoveryCode(ctx context.Context, userID uuid.UUID, hash []byte) (bool, error) {
	result, err := r.db.Exec(ctx, `UPDATE totp_recovery_codes SET used_at=now() WHERE user_id=$1 AND code_hash=$2 AND used_at IS NULL`, userID, hash)
	return result.RowsAffected() == 1, err
}

var _ pgx.Tx
