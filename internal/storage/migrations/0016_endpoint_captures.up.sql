CREATE TABLE endpoint_captures (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    endpoint_key TEXT NOT NULL,
    name TEXT NOT NULL,
    source TEXT NOT NULL,
    path TEXT NOT NULL,
    sort_order INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
CREATE INDEX idx_endpoint_captures_lookup ON endpoint_captures(project_id, endpoint_key, sort_order);
