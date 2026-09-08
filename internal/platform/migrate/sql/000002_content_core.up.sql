CREATE TABLE contents (
    object_id uuid PRIMARY KEY,
    workspace_id uuid NOT NULL,
    type text NOT NULL CHECK(type IN ('article','note','page')),
    default_locale text NOT NULL CHECK(default_locale IN ('zh-CN','en')),
    visibility text NOT NULL DEFAULT 'private' CHECK(visibility IN ('private','unlisted','public')),
    created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE(workspace_id,object_id), FOREIGN KEY(workspace_id,object_id) REFERENCES objects(workspace_id,id) ON DELETE CASCADE
);
CREATE INDEX contents_workspace_updated_idx ON contents(workspace_id,updated_at DESC);

CREATE TABLE content_localizations (
    id uuid PRIMARY KEY, workspace_id uuid NOT NULL, content_id uuid NOT NULL,
    locale text NOT NULL CHECK(locale IN ('zh-CN','en')),
    state text NOT NULL DEFAULT 'draft' CHECK(state IN ('draft','ready','scheduled','published','archived')),
    slug text NOT NULL, current_revision_id uuid,
    translation_status text NOT NULL DEFAULT 'draft' CHECK(translation_status IN ('missing','draft','needs_review','ready','published','outdated')),
    source_locale text, source_revision_id uuid, translated_from_hash text,
    created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(), deleted_at timestamptz,
    UNIQUE(content_id,locale), UNIQUE(workspace_id,locale,slug),
    FOREIGN KEY(workspace_id,content_id) REFERENCES contents(workspace_id,object_id) ON DELETE CASCADE
);
CREATE INDEX content_localizations_library_idx ON content_localizations(workspace_id,locale,state,updated_at DESC) WHERE deleted_at IS NULL;

CREATE TABLE content_revisions (
    id uuid PRIMARY KEY, workspace_id uuid NOT NULL, localization_id uuid NOT NULL,
    seq integer NOT NULL CHECK(seq>0), schema_version integer NOT NULL DEFAULT 1,
    title text NOT NULL, summary text NOT NULL DEFAULT '', body_json jsonb NOT NULL,
    metadata_json jsonb NOT NULL DEFAULT '{}'::jsonb, content_hash text NOT NULL,
    created_by uuid REFERENCES users(id) ON DELETE SET NULL, created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE(localization_id,seq), UNIQUE(localization_id,id),
    FOREIGN KEY(localization_id) REFERENCES content_localizations(id) ON DELETE CASCADE
);
ALTER TABLE content_localizations ADD CONSTRAINT content_localizations_current_revision_fk FOREIGN KEY(current_revision_id) REFERENCES content_revisions(id) DEFERRABLE INITIALLY DEFERRED;

CREATE TABLE draft_buffers (
    localization_id uuid PRIMARY KEY REFERENCES content_localizations(id) ON DELETE CASCADE,
    workspace_id uuid NOT NULL, version integer NOT NULL DEFAULT 1,
    title text NOT NULL DEFAULT '', summary text NOT NULL DEFAULT '', body_json jsonb NOT NULL DEFAULT '{"schemaVersion":1,"type":"doc","content":[]}'::jsonb,
    metadata_json jsonb NOT NULL DEFAULT '{}'::jsonb, updated_by uuid REFERENCES users(id) ON DELETE SET NULL, updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE tags (id uuid PRIMARY KEY, workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE, name text NOT NULL, slug text NOT NULL, UNIQUE(workspace_id,slug));
CREATE TABLE object_tags (workspace_id uuid NOT NULL, object_id uuid NOT NULL, tag_id uuid NOT NULL REFERENCES tags(id) ON DELETE CASCADE, PRIMARY KEY(object_id,tag_id), FOREIGN KEY(workspace_id,object_id) REFERENCES objects(workspace_id,id) ON DELETE CASCADE);

CREATE TABLE collections (object_id uuid PRIMARY KEY, workspace_id uuid NOT NULL, title text NOT NULL, slug text NOT NULL, visibility text NOT NULL DEFAULT 'private' CHECK(visibility IN ('private','unlisted','public')), FOREIGN KEY(workspace_id,object_id) REFERENCES objects(workspace_id,id) ON DELETE CASCADE, UNIQUE(workspace_id,slug));
CREATE TABLE collection_sections (id uuid PRIMARY KEY, workspace_id uuid NOT NULL, collection_id uuid NOT NULL REFERENCES collections(object_id) ON DELETE CASCADE, title text NOT NULL, sort_key numeric NOT NULL, UNIQUE(collection_id,sort_key));
CREATE TABLE collection_items (id uuid PRIMARY KEY, workspace_id uuid NOT NULL, section_id uuid NOT NULL REFERENCES collection_sections(id) ON DELETE CASCADE, object_id uuid NOT NULL, sort_key numeric NOT NULL, annotation text NOT NULL DEFAULT '', UNIQUE(section_id,sort_key), FOREIGN KEY(workspace_id,object_id) REFERENCES objects(workspace_id,id));

CREATE TABLE inbox_items (id uuid PRIMARY KEY, workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE, kind text NOT NULL CHECK(kind IN ('text','link')), raw_text text NOT NULL DEFAULT '', source_url text, state text NOT NULL DEFAULT 'pending' CHECK(state IN ('pending','converted','archived')), converted_content_id uuid REFERENCES contents(object_id), created_by uuid REFERENCES users(id) ON DELETE SET NULL, created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now());
CREATE INDEX inbox_pending_idx ON inbox_items(workspace_id,state,created_at DESC);

CREATE TABLE search_documents (workspace_id uuid NOT NULL, object_id uuid NOT NULL, revision_id uuid NOT NULL REFERENCES content_revisions(id) ON DELETE CASCADE, locale text NOT NULL, visibility text NOT NULL, title text NOT NULL, summary text NOT NULL, body_text text NOT NULL, tsv tsvector GENERATED ALWAYS AS (to_tsvector('simple',coalesce(title,'')||' '||coalesce(summary,'')||' '||coalesce(body_text,''))) STORED, updated_at timestamptz NOT NULL DEFAULT now(), PRIMARY KEY(workspace_id,object_id,locale));
CREATE INDEX search_documents_tsv_idx ON search_documents USING GIN(tsv);

CREATE TABLE publication_targets (id uuid PRIMARY KEY, workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE, channel text NOT NULL CHECK(channel IN ('website','rss')), name text NOT NULL, config_ref text, enabled boolean NOT NULL DEFAULT true, UNIQUE(workspace_id,channel,name));
CREATE TABLE publications (id uuid PRIMARY KEY, workspace_id uuid NOT NULL, content_id uuid NOT NULL, locale text NOT NULL, revision_id uuid NOT NULL REFERENCES content_revisions(id), target_id uuid NOT NULL REFERENCES publication_targets(id), state text NOT NULL CHECK(state IN ('draft','queued','publishing','published','retry_wait','failed','dead','cancelled','withdrawn')), scheduled_at timestamptz, published_at timestamptz, created_by uuid REFERENCES users(id) ON DELETE SET NULL, created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(), UNIQUE(target_id,revision_id), FOREIGN KEY(workspace_id,content_id) REFERENCES contents(workspace_id,object_id));
CREATE TABLE publication_attempts (id uuid PRIMARY KEY, publication_id uuid NOT NULL REFERENCES publications(id) ON DELETE CASCADE, attempt_no integer NOT NULL, status text NOT NULL, external_id text, response_meta jsonb NOT NULL DEFAULT '{}'::jsonb, created_at timestamptz NOT NULL DEFAULT now(), UNIQUE(publication_id,attempt_no));
CREATE TABLE publication_views (workspace_id uuid NOT NULL, content_id uuid NOT NULL, locale text NOT NULL, slug text NOT NULL, type text NOT NULL, revision_id uuid NOT NULL, title text NOT NULL, summary text NOT NULL, rendered_html text NOT NULL, metadata_json jsonb NOT NULL DEFAULT '{}'::jsonb, published_at timestamptz NOT NULL, visibility text NOT NULL, PRIMARY KEY(workspace_id,locale,slug), UNIQUE(content_id,locale));
