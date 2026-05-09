CREATE TABLE collections (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
CREATE INDEX idx_collections_project ON collections(project_id, sort_order);

CREATE TABLE collection_items (
    id TEXT PRIMARY KEY,
    collection_id TEXT NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    endpoint_id TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    body_override TEXT NOT NULL DEFAULT '',
    headers_override TEXT NOT NULL DEFAULT '',
    skip_on_failure INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
CREATE INDEX idx_collection_items_lookup ON collection_items(collection_id, sort_order);
