CREATE TABLE preview_tokens (
    token_hash bytea PRIMARY KEY,
    workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    localization_id uuid NOT NULL REFERENCES content_localizations(id) ON DELETE CASCADE,
    content_type text NOT NULL CHECK(content_type IN ('article','note','page')),
    locale text NOT NULL,
    title text NOT NULL,
    summary text NOT NULL DEFAULT '',
    body_json jsonb NOT NULL,
    metadata_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_by uuid REFERENCES users(id) ON DELETE SET NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL
);
CREATE INDEX preview_tokens_expiry_idx ON preview_tokens(expires_at);
