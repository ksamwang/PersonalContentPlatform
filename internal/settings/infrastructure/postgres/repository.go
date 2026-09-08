package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/id"
	"github.com/ksamwang/PersonalContentPlatform/internal/settings/domain"
)

type Repository struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

func defaults(name, slug string) domain.GeneralSettings {
	return domain.GeneralSettings{
		Workspace:   domain.WorkspaceSettings{Name: name, Slug: slug, DefaultLocale: "zh-CN", SupportedLocales: []string{"zh-CN", "en"}, Timezone: "Asia/Shanghai"},
		Site:        domain.SiteSettings{Name: name, RSSEnabled: true, Footer: "Capture · Connect · Publish"},
		Auth:        domain.AuthSettings{PasswordLoginEnabled: true, PasskeyEnabled: true, SessionTTLHours: 720},
		Publication: domain.PublicationSettings{DefaultChannel: "website"},
	}
}

func (r *Repository) Get(ctx context.Context, ws uuid.UUID) (domain.GeneralSettings, error) {
	var name, slug string
	var raw []byte
	if err := r.db.QueryRow(ctx, `SELECT name,slug,settings_json FROM workspaces WHERE id=$1`, ws).Scan(&name, &slug, &raw); err != nil {
		return domain.GeneralSettings{}, err
	}
	value := defaults(name, slug)
	if len(raw) > 2 {
		_ = json.Unmarshal(raw, &value)
	}
	value.Workspace.Name, value.Workspace.Slug = name, slug
	return value, nil
}

func (r *Repository) GetBySlug(ctx context.Context, slug string) (domain.GeneralSettings, error) {
	var ws uuid.UUID
	if err := r.db.QueryRow(ctx, `SELECT id FROM workspaces WHERE slug=$1`, slug).Scan(&ws); err != nil {
		return domain.GeneralSettings{}, err
	}
	return r.Get(ctx, ws)
}

func (r *Repository) UpdateSection(ctx context.Context, ws, actor uuid.UUID, section string, raw json.RawMessage) (domain.GeneralSettings, error) {
	current, err := r.Get(ctx, ws)
	if err != nil {
		return current, err
	}
	switch section {
	case "workspace":
		err = json.Unmarshal(raw, &current.Workspace)
	case "site":
		err = json.Unmarshal(raw, &current.Site)
	case "auth":
		err = json.Unmarshal(raw, &current.Auth)
	case "publication":
		err = json.Unmarshal(raw, &current.Publication)
	default:
		return current, errors.New("unsupported settings section")
	}
	if err != nil {
		return current, err
	}
	payload, err := json.Marshal(current)
	if err != nil {
		return current, err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return current, err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `UPDATE workspaces SET name=$2,slug=$3,settings_json=$4,updated_at=now() WHERE id=$1`, ws, current.Workspace.Name, current.Workspace.Slug, payload); err != nil {
		return current, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_entries(id,workspace_id,actor_id,action,metadata) VALUES($1,$2,$3,$4,$5)`, id.New(), ws, actor, "settings."+section+".updated", json.RawMessage(`{"source":"studio"}`)); err != nil {
		return current, err
	}
	return current, tx.Commit(ctx)
}

func scanStorage(row pgx.Row) (*domain.StorageProfile, error) {
	var v domain.StorageProfile
	err := row.Scan(&v.ID, &v.Name, &v.Provider, &v.Endpoint, &v.Region, &v.Bucket, &v.AccessKey, &v.SecretKey, &v.BasePath, &v.Active, &v.UpdatedAt)
	if err != nil {
		return nil, err
	}
	maskStorage(&v)
	return &v, nil
}
func (r *Repository) ActiveStorage(ctx context.Context, ws uuid.UUID) (*domain.StorageProfile, error) {
	return scanStorage(r.db.QueryRow(ctx, `SELECT id,name,provider,endpoint,region,bucket,access_key,secret_key,base_path,active,updated_at FROM storage_profiles WHERE workspace_id=$1 AND active`, ws))
}
func (r *Repository) StorageByID(ctx context.Context, ws, idv uuid.UUID) (*domain.StorageProfile, error) {
	return scanStorage(r.db.QueryRow(ctx, `SELECT id,name,provider,endpoint,region,bucket,access_key,secret_key,base_path,active,updated_at FROM storage_profiles WHERE workspace_id=$1 AND id=$2`, ws, idv))
}
func mask(value string) string {
	if len(value) <= 4 {
		if value == "" {
			return ""
		}
		return "****"
	}
	return value[:2] + "****" + value[len(value)-2:]
}
func maskStorage(v *domain.StorageProfile) {
	v.AccessKeyMask = mask(v.AccessKey)
	v.SecretKeySet = v.SecretKey != ""
}

func (r *Repository) SaveStorage(ctx context.Context, ws, actor uuid.UUID, v domain.StorageProfile) (domain.StorageProfile, error) {
	old, _ := r.ActiveStorage(ctx, ws)
	if v.ID == uuid.Nil {
		v.ID = id.New()
	}
	if old != nil {
		if v.AccessKey == "" {
			v.AccessKey = old.AccessKey
		}
		if v.SecretKey == "" {
			v.SecretKey = old.SecretKey
		}
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return v, err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `UPDATE storage_profiles SET active=false,updated_at=now() WHERE workspace_id=$1 AND id<>$2`, ws, v.ID); err != nil {
		return v, err
	}
	err = tx.QueryRow(ctx, `INSERT INTO storage_profiles(id,workspace_id,name,provider,endpoint,region,bucket,access_key,secret_key,base_path,active) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,true) ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name,provider=EXCLUDED.provider,endpoint=EXCLUDED.endpoint,region=EXCLUDED.region,bucket=EXCLUDED.bucket,access_key=EXCLUDED.access_key,secret_key=EXCLUDED.secret_key,base_path=EXCLUDED.base_path,active=true,updated_at=now() RETURNING updated_at`, v.ID, ws, v.Name, v.Provider, v.Endpoint, v.Region, v.Bucket, v.AccessKey, v.SecretKey, v.BasePath).Scan(&v.UpdatedAt)
	if err != nil {
		return v, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_entries(id,workspace_id,actor_id,action,metadata) VALUES($1,$2,$3,'settings.storage.updated','{}')`, id.New(), ws, actor); err != nil {
		return v, err
	}
	if err = tx.Commit(ctx); err != nil {
		return v, err
	}
	v.Active = true
	maskStorage(&v)
	return v, nil
}

func (r *Repository) AI(ctx context.Context, ws uuid.UUID) (domain.AIConfig, error) {
	var v domain.AIConfig
	var raw []byte
	err := r.db.QueryRow(ctx, `SELECT provider,base_url,api_key,model,purpose_models,updated_at FROM ai_provider_configs WHERE workspace_id=$1`, ws).Scan(&v.Provider, &v.BaseURL, &v.APIKey, &v.Model, &raw, &v.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		v.Provider = "openai-compatible"
		v.PurposeModels = map[string]string{}
		return v, nil
	}
	if err != nil {
		return v, err
	}
	_ = json.Unmarshal(raw, &v.PurposeModels)
	v.APIKeyMask = mask(v.APIKey)
	v.APIKeySet = v.APIKey != ""
	return v, nil
}
func (r *Repository) SaveAI(ctx context.Context, ws, actor uuid.UUID, v domain.AIConfig) (domain.AIConfig, error) {
	old, _ := r.AI(ctx, ws)
	if v.APIKey == "" {
		v.APIKey = old.APIKey
	}
	raw, _ := json.Marshal(v.PurposeModels)
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return v, err
	}
	defer tx.Rollback(ctx)
	err = tx.QueryRow(ctx, `INSERT INTO ai_provider_configs(workspace_id,provider,base_url,api_key,model,purpose_models) VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT(workspace_id) DO UPDATE SET provider=EXCLUDED.provider,base_url=EXCLUDED.base_url,api_key=EXCLUDED.api_key,model=EXCLUDED.model,purpose_models=EXCLUDED.purpose_models,updated_at=now() RETURNING updated_at`, ws, v.Provider, v.BaseURL, v.APIKey, v.Model, raw).Scan(&v.UpdatedAt)
	if err != nil {
		return v, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_entries(id,workspace_id,actor_id,action,metadata) VALUES($1,$2,$3,'settings.ai.updated','{}')`, id.New(), ws, actor); err != nil {
		return v, err
	}
	if err = tx.Commit(ctx); err != nil {
		return v, err
	}
	v.APIKeyMask = mask(v.APIKey)
	v.APIKeySet = v.APIKey != ""
	return v, nil
}
