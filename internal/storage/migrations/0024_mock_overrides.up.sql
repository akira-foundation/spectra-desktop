CREATE TABLE mock_overrides (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    endpoint_id TEXT NOT NULL,
    enabled INTEGER NOT NULL DEFAULT 1,
    status INTEGER NOT NULL DEFAULT 200,
    latency_ms INTEGER NOT NULL DEFAULT 0,
    body TEXT NOT NULL DEFAULT '',
    headers_json TEXT NOT NULL DEFAULT '',
    source TEXT NOT NULL DEFAULT 'auto',
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    UNIQUE (project_id, endpoint_id)
);

CREATE INDEX idx_mock_overrides_project ON mock_overrides(project_id);
