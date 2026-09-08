CREATE TABLE webhook_endpoints (
    id uuid PRIMARY KEY,
    workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name text NOT NULL CHECK(length(name)>0),
    url text NOT NULL CHECK(length(url)>0),
    secret_ref text NOT NULL CHECK(length(secret_ref)>0),
    enabled boolean NOT NULL DEFAULT true,
    event_types text[] NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX webhook_endpoints_workspace_idx ON webhook_endpoints(workspace_id,enabled);

CREATE TABLE webhook_attempts (
    id uuid PRIMARY KEY,
    endpoint_id uuid NOT NULL REFERENCES webhook_endpoints(id) ON DELETE CASCADE,
    event_id uuid NOT NULL REFERENCES outbox_events(id) ON DELETE CASCADE,
    attempt_no integer NOT NULL CHECK(attempt_no>0),
    status text NOT NULL CHECK(status IN ('pending','succeeded','failed')),
    response_code integer,
    response_excerpt text,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE(endpoint_id,event_id,attempt_no)
);
CREATE INDEX webhook_attempts_delivery_idx ON webhook_attempts(endpoint_id,event_id,created_at DESC);

CREATE TABLE export_runs (
    id uuid PRIMARY KEY,
    workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    format text NOT NULL,
    state text NOT NULL CHECK(state IN ('running','succeeded','failed')),
    manifest_json jsonb,
    created_by uuid REFERENCES users(id) ON DELETE SET NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    completed_at timestamptz
);
