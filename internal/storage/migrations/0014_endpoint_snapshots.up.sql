CREATE TABLE endpoint_snapshots (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    hash TEXT NOT NULL,
    payload_json TEXT NOT NULL,
    endpoint_count INTEGER NOT NULL DEFAULT 0,
    scanned_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL
);
CREATE INDEX idx_endpoint_snapshots_project ON endpoint_snapshots(project_id, scanned_at DESC);
