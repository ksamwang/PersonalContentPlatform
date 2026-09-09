ALTER TABLE assets ADD COLUMN archived_at timestamptz;
CREATE INDEX assets_workspace_active_idx ON assets(workspace_id,created_at DESC) WHERE archived_at IS NULL;
