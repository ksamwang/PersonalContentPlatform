ALTER TABLE contents ADD COLUMN deleted_at timestamptz;
CREATE INDEX contents_workspace_active_idx ON contents(workspace_id,updated_at DESC) WHERE deleted_at IS NULL;
