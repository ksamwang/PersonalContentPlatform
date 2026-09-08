CREATE TABLE storage_profiles (
    id uuid PRIMARY KEY,
    workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name text NOT NULL DEFAULT 'Default',
    provider text NOT NULL CHECK(provider IN ('filesystem','s3','r2','oss')),
    endpoint text NOT NULL DEFAULT '',
    region text NOT NULL DEFAULT '',
    bucket text NOT NULL DEFAULT '',
    access_key text NOT NULL DEFAULT '',
    secret_key text NOT NULL DEFAULT '',
    base_path text NOT NULL DEFAULT '',
    active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE(workspace_id,name)
);
CREATE UNIQUE INDEX storage_profiles_one_active_idx ON storage_profiles(workspace_id) WHERE active;

CREATE TABLE ai_provider_configs (
    workspace_id uuid PRIMARY KEY REFERENCES workspaces(id) ON DELETE CASCADE,
    provider text NOT NULL DEFAULT 'openai-compatible',
    base_url text NOT NULL DEFAULT '',
    api_key text NOT NULL DEFAULT '',
    model text NOT NULL DEFAULT '',
    purpose_models jsonb NOT NULL DEFAULT '{}'::jsonb,
    updated_at timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE upload_intents ADD COLUMN storage_profile_id uuid REFERENCES storage_profiles(id);
ALTER TABLE blobs ADD COLUMN storage_profile_id uuid REFERENCES storage_profiles(id);
ALTER TABLE webhook_endpoints ADD COLUMN secret_value text NOT NULL DEFAULT '';
