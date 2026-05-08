CREATE TABLE request_history (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    endpoint_id TEXT NOT NULL DEFAULT '',
    method TEXT NOT NULL,
    url TEXT NOT NULL,
    request_headers TEXT NOT NULL DEFAULT '{}',
    request_body TEXT NOT NULL DEFAULT '',
    response_status INTEGER NOT NULL DEFAULT 0,
    response_headers TEXT NOT NULL DEFAULT '{}',
    response_body TEXT NOT NULL DEFAULT '',
    duration_ms INTEGER NOT NULL DEFAULT 0,
    size_bytes INTEGER NOT NULL DEFAULT 0,
    error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL
);
CREATE INDEX idx_request_history_project ON request_history(project_id, created_at DESC);
CREATE INDEX idx_request_history_endpoint ON request_history(project_id, endpoint_id, created_at DESC);
