CREATE TABLE project_auth (
    project_id TEXT PRIMARY KEY,
    scheme TEXT NOT NULL DEFAULT '',
    token TEXT NOT NULL DEFAULT '',
    token_path TEXT NOT NULL DEFAULT '',
    user_json TEXT NOT NULL DEFAULT '',
    cookies_json TEXT NOT NULL DEFAULT '',
    headers_json TEXT NOT NULL DEFAULT '',
    expires_at TIMESTAMP,
    captured_from_endpoint TEXT NOT NULL DEFAULT '',
    captured_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);
