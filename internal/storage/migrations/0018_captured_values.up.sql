CREATE TABLE captured_values (
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    value TEXT NOT NULL,
    endpoint_key TEXT NOT NULL DEFAULT '',
    captured_at TIMESTAMP NOT NULL,
    PRIMARY KEY (project_id, name)
);
