ALTER TABLE publication_attempts ADD COLUMN error_message text NOT NULL DEFAULT '';
ALTER TABLE publication_targets DROP CONSTRAINT publication_targets_channel_check;
ALTER TABLE publication_targets ADD CONSTRAINT publication_targets_channel_check CHECK(channel IN ('website','rss','webhook','newsletter','social'));
CREATE INDEX publications_workspace_state_idx ON publications(workspace_id,state,created_at DESC);
