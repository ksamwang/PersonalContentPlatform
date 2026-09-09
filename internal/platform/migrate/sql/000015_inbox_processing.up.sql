CREATE TABLE media_provider_configs (
    workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    purpose text NOT NULL CHECK(purpose IN ('ocr','transcription')),
    provider text NOT NULL DEFAULT 'openai-compatible',
    base_url text NOT NULL,
    api_key text NOT NULL,
    model text NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY(workspace_id,purpose)
);

ALTER TABLE inbox_items DROP CONSTRAINT inbox_items_kind_check;
ALTER TABLE inbox_items ADD CONSTRAINT inbox_items_kind_check CHECK(kind IN ('text','link','image','audio','file'));
ALTER TABLE inbox_items
    ADD COLUMN asset_id uuid REFERENCES assets(object_id) ON DELETE SET NULL,
    ADD COLUMN title text NOT NULL DEFAULT '',
    ADD COLUMN extracted_text text NOT NULL DEFAULT '',
    ADD COLUMN cover_url text,
    ADD COLUMN processing_state text NOT NULL DEFAULT 'idle' CHECK(processing_state IN ('idle','processing','completed','failed')),
    ADD COLUMN processing_error text NOT NULL DEFAULT '',
    ADD COLUMN duplicate_of uuid REFERENCES inbox_items(id) ON DELETE SET NULL;
CREATE INDEX inbox_asset_idx ON inbox_items(workspace_id,asset_id) WHERE asset_id IS NOT NULL;
CREATE INDEX inbox_source_url_idx ON inbox_items(workspace_id,source_url) WHERE source_url IS NOT NULL;
