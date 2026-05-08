CREATE TABLE IF NOT EXISTS endpoints (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    method TEXT NOT NULL,
    path TEXT NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    handler TEXT NOT NULL DEFAULT '',
    middleware TEXT NOT NULL DEFAULT '[]',
    parameters TEXT NOT NULL DEFAULT '[]',
    tags TEXT NOT NULL DEFAULT '[]',
    source_file TEXT NOT NULL DEFAULT '',
    source_line INTEGER NOT NULL DEFAULT 0,
    framework TEXT NOT NULL DEFAULT '',
    confidence REAL NOT NULL DEFAULT 0,
    scanned_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_endpoints_project ON endpoints(project_id);
CREATE INDEX IF NOT EXISTS idx_endpoints_method ON endpoints(method);
CREATE INDEX IF NOT EXISTS idx_endpoints_path ON endpoints(path);
