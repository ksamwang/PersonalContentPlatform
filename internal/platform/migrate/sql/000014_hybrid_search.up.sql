CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE embedding_provider_configs (
    workspace_id uuid PRIMARY KEY REFERENCES workspaces(id) ON DELETE CASCADE,
    provider text NOT NULL DEFAULT 'openai-compatible',
    base_url text NOT NULL,
    api_key text NOT NULL,
    model text NOT NULL,
    dimensions integer NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE content_chunks (
    id uuid PRIMARY KEY,
    workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    content_id uuid NOT NULL,
    localization_id uuid NOT NULL REFERENCES content_localizations(id) ON DELETE CASCADE,
    revision_id uuid NOT NULL REFERENCES content_revisions(id) ON DELETE CASCADE,
    locale text NOT NULL,
    chunk_index integer NOT NULL,
    text text NOT NULL,
    tsv tsvector GENERATED ALWAYS AS (to_tsvector('simple', text)) STORED,
    embedding vector NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE(revision_id,chunk_index),
    FOREIGN KEY(workspace_id,content_id) REFERENCES contents(workspace_id,object_id)
);
CREATE INDEX content_chunks_fts_idx ON content_chunks USING gin(tsv);
CREATE INDEX content_chunks_workspace_idx ON content_chunks(workspace_id,locale);
