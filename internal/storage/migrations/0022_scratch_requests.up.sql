CREATE TABLE scratch_requests (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name TEXT NOT NULL DEFAULT '',
    method TEXT NOT NULL DEFAULT 'GET',
    url TEXT NOT NULL DEFAULT '',
    headers_json TEXT NOT NULL DEFAULT '[]',
    body TEXT NOT NULL DEFAULT '',
    response_json TEXT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_scratch_requests_project ON scratch_requests(project_id, sort_order);
