CREATE TABLE endpoint_datasets (
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    endpoint_key TEXT NOT NULL,
    rows_json TEXT NOT NULL DEFAULT '[]',
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    PRIMARY KEY (project_id, endpoint_key)
);
